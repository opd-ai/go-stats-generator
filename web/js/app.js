/**
 * Main Application Controller for go-stats-generator Web Interface
 *
 * Orchestrates: UI events → GitHub fetcher → WASM analyzer → results rendering
 */

/** Progress percentages for non-download stages. */
const STAGE_PERCENT = { resolving: 5, fetching_tree: 10 };

class App {
  constructor() {
    this.wasmLoader = new WASMLoader();
    this.fetcher = null;
    this.isAnalyzing = false;
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
    const tokenInput = document.getElementById('github-token');
    const cancelBtn = document.getElementById('cancel-btn');

    if (analyzeBtn) {
      analyzeBtn.addEventListener('click', () => this.handleAnalyze());
    }
    if (tokenInput) {
      tokenInput.addEventListener('input', (e) => this.ensureFetcher(e.target.value));
    }
    if (cancelBtn) {
      cancelBtn.addEventListener('click', () => this.handleCancel());
    }
  }

  /**
   * Ensure a {@link GitHubFetcher} instance exists and has the latest token.
   * @param {string} token
   */
  ensureFetcher(token) {
    if (!this.fetcher) {
      this.fetcher = new GitHubFetcher(token);
    } else {
      this.fetcher.setToken(token);
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
    UI.setAnalyzeButtonState(false);

    try {
      const inputs = this.gatherFormInputs();
      if (!inputs.repoURL) throw new Error('Please enter a repository URL');

      this.ensureFetcher(inputs.token);

      UI.show('progress-area');
      UI.hide('results-area');

      const { files, stats } = await this.fetcher.fetchRepository(
        inputs.repoURL,
        inputs.ref,
        inputs.includeTests,
        (progress) => this.handleFetchProgress(progress),
      );

      UI.updateRateLimit(this.fetcher.getRateLimitStatus());
      UI.updateProgress(80, 'Running analysis…');

      const result = await this.wasmLoader.analyze(files, {
        format: inputs.format,
        skipTests: !inputs.includeTests,
      });

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
      UI.setAnalyzeButtonState(true);
      UI.hide('progress-area');
    }
  }

  /**
   * Map fetcher progress events to the progress bar.
   * @param {Object} progress
   */
  handleFetchProgress(progress) {
    if (progress.stage === 'downloading') {
      const pct = 10 + (progress.current / progress.total) * 70;
      UI.updateProgress(pct, progress.message);
    } else {
      UI.updateProgress(STAGE_PERCENT[progress.stage] || 0, progress.message);
    }
  }

  /**
   * Cancel in-flight fetch requests.  The abort signal causes
   * {@link handleAnalyze} to reject with an AbortError, and its
   * finally block resets `isAnalyzing` and re-enables the button.
   */
  handleCancel() {
    if (this.fetcher) {
      this.fetcher.abort();
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
    h3.textContent = `Repository: ${stats.owner}/${stats.repo}`;
    el.appendChild(h3);

    for (const line of [
      `Ref: ${stats.ref}`,
      `Files analyzed: ${stats.totalFiles}`,
      `Total size: ${(stats.totalSize / 1024).toFixed(2)} KB`,
    ]) {
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
