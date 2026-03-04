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

    while (!globalThis.goStatsAnalyze && elapsed < maxWaitMs) {
      await new Promise(resolve => setTimeout(resolve, pollIntervalMs));
      elapsed += pollIntervalMs;
    }

    if (!globalThis.goStatsAnalyze) {
      throw new Error('WASM API did not initialize within timeout');
    }

    this.analysisAPI = {
      analyze: globalThis.goStatsAnalyze,
      analyzeHTML: globalThis.goStatsAnalyzeHTML,
      analyzeJSON: globalThis.goStatsAnalyzeJSON
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
      skipTests = true,
      sections = 'functions,structs,interfaces,packages,patterns,duplication,documentation'
    } = options;

    const config = {
      files: JSON.stringify(files),
      maxFunctionLength,
      maxComplexity,
      minDocCoverage,
      skipTests,
      sections
    };

    let result;
    
    if (format === 'html') {
      result = await this.analysisAPI.analyzeHTML(config);
    } else {
      result = await this.analysisAPI.analyzeJSON(config);
    }

    return result;
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
