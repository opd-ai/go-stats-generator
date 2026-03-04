/**
 * WASM Loader for go-stats-generator
 *
 * Loads and initializes the Go WebAssembly binary, providing a
 * JavaScript interface to the analysis functions.
 */

/** Maximum time (ms) to wait for the Go runtime to expose the API. */
const WASM_API_TIMEOUT_MS = 5000;

/** Polling interval (ms) when waiting for the WASM API. */
const WASM_API_POLL_MS = 50;

class WASMLoader {
  constructor() {
    this.go = null;
    this.instance = null;
    this.ready = false;
    this.analysisAPI = null;
  }

  /**
   * Load the WASM binary and start the Go runtime.
   * @param {string} wasmPath - URL of the .wasm file.
   */
  async load(wasmPath) {
    if (!globalThis.Go) {
      throw new Error('wasm_exec.js must be loaded before WASMLoader');
    }

    this.go = new Go();

    // Prefer streaming instantiation; fall back to buffered if it fails.
    this.instance = await this.instantiate(wasmPath);

    // Start the Go runtime (blocks via `select {}` in main).
    this.go.run(this.instance);

    await this.waitForAPI();
    this.ready = true;
  }

  /**
   * Instantiate the WASM binary, trying streaming first for speed.
   * @param {string} wasmPath
   * @returns {Promise<WebAssembly.Instance>}
   */
  async instantiate(wasmPath) {
    if (typeof WebAssembly.instantiateStreaming === 'function') {
      try {
        const result = await WebAssembly.instantiateStreaming(
          fetch(wasmPath),
          this.go.importObject,
        );
        return result.instance;
      } catch (e) {
        console.warn('Streaming instantiation failed, falling back to buffer:', e);
      }
    }

    const response = await fetch(wasmPath);
    if (!response.ok) {
      throw new Error(`Failed to fetch WASM binary: HTTP ${response.status} for ${wasmPath}`);
    }
    const bytes = await response.arrayBuffer();
    const result = await WebAssembly.instantiate(bytes, this.go.importObject);
    return result.instance;
  }

  /**
   * Poll until the Go runtime exposes `globalThis.analyzeCode`.
   */
  async waitForAPI() {
    let elapsed = 0;
    while (!globalThis.analyzeCode && elapsed < WASM_API_TIMEOUT_MS) {
      await new Promise((resolve) => setTimeout(resolve, WASM_API_POLL_MS));
      elapsed += WASM_API_POLL_MS;
    }

    if (!globalThis.analyzeCode) {
      throw new Error('WASM API did not initialize within timeout');
    }

    this.analysisAPI = { analyzeCode: globalThis.analyzeCode };
  }

  /**
   * Run analysis on in-memory Go source files.
   * @param {Array<{path: string, content: string}>} files
   * @param {Object} options
   * @returns {Promise<string>} JSON or HTML result string.
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
    } = options;

    const request = {
      files,
      outputFormat: format,
      config: {
        maxFunctionLength,
        maxCyclomaticComplexity: maxComplexity,
        minDocumentationCoverage: minDocCoverage,
        skipTestFiles: skipTests,
      },
    };

    const result = this.analysisAPI.analyzeCode(JSON.stringify(request));

    if (!result.success) {
      throw new Error(result.error || 'Analysis failed');
    }
    return result.data;
  }

  /** @returns {boolean} Whether the WASM runtime is ready. */
  isReady() {
    return this.ready;
  }
}

// Export for use in other modules.
if (typeof module !== 'undefined' && module.exports) {
  module.exports = WASMLoader;
}
