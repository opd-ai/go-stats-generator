/**
 * WASM Loader for go-stats-generator
 * 
 * Loads and initializes the go-stats-generator WebAssembly binary,
 * providing a JavaScript interface to the analysis functions.
 */

class WASMLoader {
  constructor() {
    this.go = null;
    this.instance = null;
    this.ready = false;
    this.analysisAPI = null;
  }

  /**
   * Load the WASM binary and initialize the Go runtime
   * @param {string} wasmPath - Path to the .wasm file
   * @returns {Promise<void>}
   */
  async load(wasmPath) {
    // Load the Go WASM exec JavaScript runtime
    if (!globalThis.Go) {
      throw new Error('wasm_exec.js must be loaded before WASMLoader');
    }

    this.go = new Go();
    
    let wasmBytes;
    
    // Try streaming instantiation first (faster)
    if (typeof WebAssembly.instantiateStreaming === 'function') {
      try {
        const result = await WebAssembly.instantiateStreaming(
          fetch(wasmPath),
          this.go.importObject
        );
        this.instance = result.instance;
      } catch (e) {
        console.warn('Streaming instantiation failed, falling back to buffer:', e);
        // Fall through to buffer-based loading
      }
    }
    
    // Fallback: buffer-based instantiation
    if (!this.instance) {
      const response = await fetch(wasmPath);
      if (!response.ok) {
        throw new Error(`Failed to fetch WASM binary: HTTP ${response.status} for ${wasmPath}`);
      }
      wasmBytes = await response.arrayBuffer();
      const result = await WebAssembly.instantiate(wasmBytes, this.go.importObject);
      this.instance = result.instance;
    }

    // Run the Go program (this starts the runtime and blocks via select{})
    this.go.run(this.instance);
    
    // Wait for the Go runtime to expose the analysis API
    await this.waitForAPI();
    
    this.ready = true;
  }

  /**
   * Wait for the Go WASM to expose the analysis API on globalThis
   * @returns {Promise<void>}
   */
  async waitForAPI() {
    const maxWaitMs = 5000;
    const pollIntervalMs = 50;
    let elapsed = 0;

    while (!globalThis.analyzeCode && elapsed < maxWaitMs) {
      await new Promise(resolve => setTimeout(resolve, pollIntervalMs));
      elapsed += pollIntervalMs;
    }

    if (!globalThis.analyzeCode) {
      throw new Error('WASM API did not initialize within timeout');
    }

    this.analysisAPI = {
      analyzeCode: globalThis.analyzeCode
    };
  }

  /**
   * Analyze Go source files and return results in the specified format
   * @param {Array} files - Array of {path: string, content: string} objects
   * @param {Object} options - Analysis options
   * @returns {Promise<string>} - Analysis result (JSON or HTML string)
   */
  async analyze(files, options = {}) {
    if (!this.ready) {
      throw new Error('WASM not loaded. Call load() first.');
    }

    const {
      format = 'json',
      maxFunctionLength = 30,
      maxComplexity = 10,
      minDocCoverage = 0.7,
      skipTests = true
    } = options;

    // Build the AnalysisRequest matching Go's expected JSON structure
    const request = {
      files: files,
      outputFormat: format,
      config: {
        maxFunctionLength,
        maxCyclomaticComplexity: maxComplexity,
        minDocumentationCoverage: minDocCoverage,
        skipTestFiles: skipTests
      }
    };

    const result = await this.analysisAPI.analyzeCode(JSON.stringify(request));

    if (!result.success) {
      throw new Error(result.error || 'Analysis failed');
    }

    return result.data;
  }

  /**
   * Check if WASM is loaded and ready
   * @returns {boolean}
   */
  isReady() {
    return this.ready;
  }
}

// Export for use in other modules
if (typeof module !== 'undefined' && module.exports) {
  module.exports = WASMLoader;
}
