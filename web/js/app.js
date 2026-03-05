/**
 * Main Application Controller for go-stats-generator Web Interface
 *
 * Orchestrates: UI events → WASM git clone → WASM analyzer → results rendering
 */

class App {
  constructor() {
    this.wasmLoader = new WASMLoader();
    this.isAnalyzing = false;
    this.cloneErrorDetail = null;
    /** @type {string|null} Error detail from failed zipball download. */
    this.zipballErrorDetail = null;
    /** @type {ZipballFetcher|null} Active fetcher for cancellation support. */
    this.currentFetcher = null;
  }

  // ---------------------------------------------------------------------------
  // Initialization
  // ---------------------------------------------------------------------------

  /**
   * Resolve the base path for the deployed site.
   * GitHub Pages may serve from a subpath (e.g. /go-stats-generator/).
   * @returns {string} Base path with trailing slash.
   */
  getBasePath() {
    const baseEl = document.querySelector('base[href]');
    if (baseEl) {
      return baseEl.getAttribute('href');
    }
    const path = window.location.pathname || '/';
    return path.endsWith('/') ? path : path.substring(0, path.lastIndexOf('/') + 1) || '/';
  }

  /**
   * Initialize the application: load WASM binary and bind event listeners.
   */
  async init() {
    try {
      UI.updateProgress(10, 'Loading WebAssembly…');
      UI.show('progress-area');

      const wasmPath = await this.resolveWASMPath();
      await this.wasmLoader.load(wasmPath);

      UI.updateProgress(100, 'Ready');
      UI.hide('progress-area');

      this.setupEventListeners();
    } catch (error) {
      UI.showError(`Failed to initialize: ${error.message}`);
      console.error('Initialization error:', error);
    }
  }

  /**
   * Determine the URL for the WASM binary, using the content-hashed
   * manifest when available and falling back to a default filename.
   * @returns {Promise<string>}
   */
  async resolveWASMPath() {
    const basePath = this.getBasePath();
    try {
      const res = await fetch(`${basePath}wasm/wasm-manifest.json`);
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      const manifest = await res.json();
      return `${basePath}wasm/${manifest.wasmFile}`;
    } catch (err) {
      console.warn('Could not load WASM manifest, using default filename:', err);
      return `${basePath}wasm/go-stats-generator.wasm`;
    }
  }

  // ---------------------------------------------------------------------------
  // Event listeners
  // ---------------------------------------------------------------------------

  /** Bind DOM event listeners. */
  setupEventListeners() {
    const analyzeBtn = document.getElementById('analyze-btn');
    const cancelBtn = document.getElementById('cancel-btn');

    if (analyzeBtn) {
      analyzeBtn.addEventListener('click', () => this.handleAnalyze());
    }
    if (cancelBtn) {
      cancelBtn.addEventListener('click', () => this.handleCancel());
    }
  }

  // ---------------------------------------------------------------------------
  // Analysis flow
  // ---------------------------------------------------------------------------

  /** Gather form inputs and return a plain object. */
  gatherFormInputs() {
    return {
      repoURL: document.getElementById('repo-url').value.trim(),
      ref: document.getElementById('repo-ref').value.trim() || null,
      token: document.getElementById('github-token').value.trim() || null,
      includeTests: document.getElementById('include-tests').checked,
      format: document.querySelector('input[name="format"]:checked').value,
    };
  }

  /** Handle the "Analyze" button click. */
  async handleAnalyze() {
    if (this.isAnalyzing) return;

    UI.clearError();
    this.isAnalyzing = true;
    this.usingClone = false;
    this.cloneErrorDetail = null;
    this.zipballErrorDetail = null;
    UI.setAnalyzeButtonState(false);

    try {
      const inputs = this.gatherFormInputs();
      if (!inputs.repoURL) throw new Error('Please enter a repository URL');

      UI.show('progress-area');
      UI.hide('results-area');

      // Use the WASM git-clone path (no API rate limits).
      // If clone fails with a network/CORS error, fall back to
      // downloading the repository as a ZIP archive (single API request).
      let result, stats;
      const cloneAvailable = typeof globalThis.cloneAndAnalyze === 'function';

      if (!cloneAvailable) {
        throw new Error(
          'WASM git clone is not available. Please reload the page and try again.',
        );
      }

      this.usingClone = true;
      UI.setCancelVisible(false);
      const cloneResult = await this.analyzeViaClone(inputs);
      if (cloneResult) {
        ({ result, stats } = cloneResult);
      } else {
        const detail = this.cloneErrorDetail || 'unknown error';
        if (this.isNetworkError(detail)) {
          // Clone failed due to CORS / network – download ZIP archive instead.
          console.warn(
            'Git clone failed with network error, trying zipball fallback:',
            detail,
          );
          this.usingClone = false;
          UI.setCancelVisible(true);
          UI.updateProgress(10, 'Git clone unavailable, downloading ZIP archive…');
          const zipResult = await this.analyzeViaZipball(inputs);
          if (zipResult) {
            ({ result, stats } = zipResult);
          } else {
            const zipDetail = this.zipballErrorDetail || 'unknown error';
            throw new Error(
              'Git clone was blocked by the browser (CORS) and the ZIP archive ' +
              `download also failed (${zipDetail}). If this is a private ` +
              'repository, provide a personal access token.',
            );
          }
        } else {
          throw new Error(
            `Git clone failed: ${detail}. ` +
            'If this is a private repository, provide a personal access token.',
          );
        }
      }

      UI.updateProgress(100, 'Complete');
      UI.hide('progress-area');
      UI.show('results-area');

      if (inputs.format === 'html') {
        UI.renderHTMLReport(result);
      } else {
        UI.renderJSONReport(result);
      }

      this.displayStatsSummary(stats);
    } catch (error) {
      if (error.name !== 'AbortError') {
        UI.showError(`Analysis failed: ${error.message}`);
        console.error('Analysis error:', error);
      }
    } finally {
      this.isAnalyzing = false;
      this.usingClone = false;
      UI.setAnalyzeButtonState(true);
      UI.setCancelVisible(true);
      UI.hide('progress-area');
    }
  }

  /**
   * Clone the repository in WASM via go-git and run analysis.
   * This avoids GitHub API rate limits entirely by using the git
   * smart HTTP protocol directly.
   * @param {Object} inputs - Form inputs.
   * @returns {Promise<{result: string, stats: Object}|null>} Resolves with analysis results, or null if clone failed.
   */
  async analyzeViaClone(inputs) {
    const request = {
      url: inputs.repoURL,
      ref: inputs.ref || '',
      token: inputs.token || '',
      includeTests: inputs.includeTests,
      outputFormat: inputs.format,
      config: {
        maxFunctionLength: 30,
        maxCyclomaticComplexity: 10,
        minDocumentationCoverage: 0.7,
        skipTestFiles: !inputs.includeTests,
      },
    };

    let response;
    try {
      response = await globalThis.cloneAndAnalyze(
        JSON.stringify(request),
        (progress) => this.handleCloneProgress(progress),
      );
    } catch (err) {
      console.error('Git clone threw:', err);
      this.cloneErrorDetail = String(err);
      return null;
    }

    if (!response || !response.success) {
      const errMsg = (response && response.error) || 'unknown error';
      console.error('Git clone failed:', errMsg);
      this.cloneErrorDetail = errMsg;
      return null;
    }

    return {
      result: response.data,
      stats: response.stats || {},
    };
  }

  /**
   * Map clone progress events to the progress bar.
   * @param {Object} progress - {percent, message}
   */
  handleCloneProgress(progress) {
    if (typeof progress.percent === 'number' && progress.percent >= 0) {
      UI.updateProgress(progress.percent, progress.message);
    } else {
      // Clone output without explicit percent – show message only.
      const text = document.getElementById('progress-text');
      if (text) text.textContent = progress.message;
    }
  }

  /**
   * Detect whether a clone error is a browser network / CORS failure.
   * In WASM, Go's net/http delegates to the browser fetch() API which
   * surfaces opaque "NetworkError" messages when CORS blocks the request.
   * @param {string} detail - Error detail string from the clone attempt.
   * @returns {boolean}
   */
  isNetworkError(detail) {
    if (!detail) return false;
    const lower = detail.toLowerCase();
    return (
      lower.includes('networkerror') ||
      lower.includes('fetch() failed') ||
      lower.includes('network error') ||
      lower.includes('failed to fetch') ||
      lower.includes('cors')
    );
  }

  /**
   * Fallback: download the repository as a ZIP archive from the GitHub
   * API zipball endpoint and analyze the Go files with the in-memory
   * WASM analyzer. This uses exactly ONE API request regardless of
   * repository size, avoiding the per-blob rate-limit problem entirely.
   * @param {Object} inputs - Form inputs.
   * @returns {Promise<{result: string, stats: Object}|null>}
   */
  async analyzeViaZipball(inputs) {
    const fetcher = new ZipballFetcher(inputs.token);
    this.currentFetcher = fetcher;

    try {
      const { files, stats } = await fetcher.fetchRepository(
        inputs.repoURL,
        inputs.ref,
        inputs.includeTests,
        (progress) => {
          if (progress.current && progress.total) {
            const pct = 10 + Math.round((progress.current / progress.total) * 55);
            UI.updateProgress(pct, progress.message);
          } else if (progress.stage === 'downloading') {
            UI.updateProgress(15, progress.message);
          } else {
            UI.updateProgress(40, progress.message);
          }
        },
      );

      if (!files || files.length === 0) {
        throw new Error('No Go source files found in repository');
      }

      UI.updateProgress(70, `Analyzing ${files.length} files…`);

      const request = {
        files,
        outputFormat: inputs.format,
        config: {
          maxFunctionLength: 30,
          maxCyclomaticComplexity: 10,
          minDocumentationCoverage: 0.7,
          skipTestFiles: !inputs.includeTests,
        },
      };

      const response = globalThis.analyzeCode(JSON.stringify(request));

      if (!response || !response.success) {
        throw new Error((response && response.error) || 'Analysis failed');
      }

      return { result: response.data, stats: { ...stats, method: 'zipball' } };
    } catch (err) {
      if (err && err.name === 'AbortError') throw err;
      console.error('Zipball fallback failed:', err);
      // Preserve the error message for the caller so the user sees
      // the specific reason the ZIP download failed, not just a
      // generic "also failed" message.
      this.zipballErrorDetail = (err && err.message) || String(err);
      return null;
    } finally {
      this.currentFetcher = null;
    }
  }

  /**
   * Cancel in-flight operations. Aborts the zipball fetcher if
   * active; the WASM git clone goroutine cannot be interrupted from JS.
   */
  handleCancel() {
    if (this.currentFetcher) {
      this.currentFetcher.abort();
    }
    UI.hide('progress-area');
  }

  // ---------------------------------------------------------------------------
  // Results display
  // ---------------------------------------------------------------------------

  /**
   * Show a short summary of the fetched repository.
   * Uses textContent (not innerHTML) to avoid XSS from repo metadata.
   * @param {Object} stats
   */
  displayStatsSummary(stats) {
    const el = document.getElementById('stats-summary');
    if (!el) return;

    el.textContent = '';

    const h3 = document.createElement('h3');
    h3.textContent = `Repository: ${stats.owner || ''}/${stats.repo || ''}`;
    el.appendChild(h3);

    const lines = [
      `Ref: ${stats.ref || 'default branch'}`,
      `Files analyzed: ${stats.totalFiles || 0}`,
      `Total size: ${((stats.totalSize || 0) / 1024).toFixed(2)} KB`,
    ];
    if (stats.method) {
      lines.push(`Fetch method: ${stats.method}`);
    }

    for (const line of lines) {
      const p = document.createElement('p');
      p.textContent = line;
      el.appendChild(p);
    }

    el.classList.remove('hidden');
  }
}

// Initialize app when DOM is ready.
document.addEventListener('DOMContentLoaded', async () => {
  const app = new App();
  await app.init();
});
