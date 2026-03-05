/**
 * Main Application Controller for go-stats-generator Web Interface
 *
 * Orchestrates: UI events → WASM git clone → WASM analyzer → results rendering
 */

class App {
  constructor() {
    this.wasmLoader = new WASMLoader();
    this.isAnalyzing = false;
    this.lastCloneError = null;
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
    UI.setAnalyzeButtonState(false);

    try {
      const inputs = this.gatherFormInputs();
      if (!inputs.repoURL) throw new Error('Please enter a repository URL');

      UI.show('progress-area');
      UI.hide('results-area');

      // Use the WASM git-clone path (no API rate limits).
      // The GitHub REST API fallback has been removed to prevent HTTP 429
      // rate-limit errors. Repositories are cloned directly over HTTPS
      // using go-git in the WASM binary.
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
        const detail = this.lastCloneError || 'unknown error';
        throw new Error(
          `Git clone failed: ${detail}. ` +
          'For private repositories, provide a personal access token.',
        );
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
      this.lastCloneError = String(err);
      return null;
    }

    if (!response || !response.success) {
      const errMsg = (response && response.error) || 'unknown error';
      console.error('Git clone failed:', errMsg);
      this.lastCloneError = errMsg;
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
   * Cancel in-flight operations. The WASM git clone goroutine cannot be
   * interrupted from JS, so we only hide the progress area.
   */
  handleCancel() {
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
