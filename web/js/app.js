/**
 * Main Application Controller for go-stats-generator Web Interface
 * 
 * Orchestrates the flow: UI events → GitHub fetcher → WASM analyzer → results rendering
 */

class App {
  constructor() {
    this.wasmLoader = new WASMLoader();
    this.fetcher = null;
    this.isAnalyzing = false;
  }

  /**
   * Resolve the base path for the deployed site.
   * GitHub Pages may serve from a subpath (e.g. /go-stats-generator/).
   * @returns {string} base path with trailing slash
   */
  getBasePath() {
    // Use <base> href if set, otherwise derive from document location
    const baseEl = document.querySelector('base[href]');
    if (baseEl) {
      return baseEl.getAttribute('href');
    }
    // Fallback: current directory of the page
    const path = window.location.pathname || '/';
    return path.endsWith('/') ? path : path.substring(0, path.lastIndexOf('/') + 1) || '/';
  }

  /**
   * Initialize the application
   */
  async init() {
    try {
      UI.updateProgress(10, 'Loading WebAssembly...');
      UI.show('progress-area');

      const basePath = this.getBasePath();
      let wasmPath;

      // Load WASM manifest to get content-hashed filename
      try {
        const manifestResponse = await fetch(`${basePath}wasm/wasm-manifest.json`);
        if (!manifestResponse.ok) {
          throw new Error(`manifest returned HTTP ${manifestResponse.status}`);
        }
        const manifest = await manifestResponse.json();
        wasmPath = `${basePath}wasm/${manifest.wasmFile}`;
      } catch (manifestError) {
        // Fallback: try the default non-hashed filename
        console.warn('Could not load WASM manifest, using default filename:', manifestError);
        wasmPath = `${basePath}wasm/go-stats-generator.wasm`;
      }

      await this.wasmLoader.load(wasmPath);
      
      UI.updateProgress(100, 'Ready');
      UI.hide('progress-area');
      
      // Set up event listeners
      this.setupEventListeners();
    } catch (error) {
      UI.showError(`Failed to initialize: ${error.message}`);
      console.error('Initialization error:', error);
    }
  }

  /**
   * Set up UI event listeners
   */
  setupEventListeners() {
    const analyzeBtn = document.getElementById('analyze-btn');
    const tokenInput = document.getElementById('github-token');
    const cancelBtn = document.getElementById('cancel-btn');

    if (analyzeBtn) {
      analyzeBtn.addEventListener('click', () => this.handleAnalyze());
    }

    if (tokenInput) {
      tokenInput.addEventListener('input', (e) => {
        if (!this.fetcher) {
          this.fetcher = new GitHubFetcher(e.target.value);
        } else {
          this.fetcher.setToken(e.target.value);
        }
      });
    }

    if (cancelBtn) {
      cancelBtn.addEventListener('click', () => this.handleCancel());
    }
  }

  /**
   * Handle analyze button click
   */
  async handleAnalyze() {
    if (this.isAnalyzing) return;

    UI.clearError();
    this.isAnalyzing = true;
    UI.setAnalyzeButtonState(false);

    try {
      const repoURL = document.getElementById('repo-url').value.trim();
      const ref = document.getElementById('repo-ref').value.trim() || null;
      const token = document.getElementById('github-token').value.trim() || null;
      const includeTests = document.getElementById('include-tests').checked;
      const format = document.querySelector('input[name="format"]:checked').value;

      if (!repoURL) {
        throw new Error('Please enter a repository URL');
      }

      // Initialize fetcher with token
      if (!this.fetcher) {
        this.fetcher = new GitHubFetcher(token);
      } else {
        this.fetcher.setToken(token);
      }

      // Show progress area
      UI.show('progress-area');
      UI.hide('results-area');

      // Fetch repository
      const { files, stats } = await this.fetcher.fetchRepository(
        repoURL,
        ref,
        includeTests,
        (progress) => this.handleFetchProgress(progress)
      );

      // Update rate limit display
      UI.updateRateLimit(this.fetcher.getRateLimitStatus());

      // Run analysis
      UI.updateProgress(80, 'Running analysis...');
      
      const result = await this.wasmLoader.analyze(files, {
        format,
        skipTests: !includeTests
      });

      // Display results
      UI.updateProgress(100, 'Complete');
      UI.hide('progress-area');
      UI.show('results-area');

      if (format === 'html') {
        UI.renderHTMLReport(result);
      } else {
        UI.renderJSONReport(result);
      }

      // Show stats summary
      this.displayStatsSummary(stats);

    } catch (error) {
      UI.showError(`Analysis failed: ${error.message}`);
      console.error('Analysis error:', error);
    } finally {
      this.isAnalyzing = false;
      UI.setAnalyzeButtonState(true);
      UI.hide('progress-area');
    }
  }

  /**
   * Handle fetch progress updates
   * @param {Object} progress - Progress information
   */
  handleFetchProgress(progress) {
    if (progress.stage === 'resolving') {
      UI.updateProgress(5, progress.message);
    } else if (progress.stage === 'fetching_tree') {
      UI.updateProgress(10, progress.message);
    } else if (progress.stage === 'downloading') {
      const percent = 10 + ((progress.current / progress.total) * 70);
      UI.updateProgress(percent, progress.message);
    }
  }

  /**
   * Handle cancel button click
   */
  handleCancel() {
    // Note: Can't truly cancel ongoing fetch/WASM operations
    // This just hides the progress UI
    UI.hide('progress-area');
    this.isAnalyzing = false;
    UI.setAnalyzeButtonState(true);
  }

  /**
   * Display repository stats summary
   * @param {Object} stats - Repository statistics
   */
  displayStatsSummary(stats) {
    const summaryDiv = document.getElementById('stats-summary');
    if (summaryDiv) {
      summaryDiv.innerHTML = `
        <h3>Repository: ${stats.owner}/${stats.repo}</h3>
        <p>Ref: ${stats.ref}</p>
        <p>Files analyzed: ${stats.totalFiles}</p>
        <p>Total size: ${(stats.totalSize / 1024).toFixed(2)} KB</p>
      `;
      summaryDiv.classList.remove('hidden');
    }
  }
}

// Initialize app when DOM is ready
document.addEventListener('DOMContentLoaded', async () => {
  const app = new App();
  await app.init();
});
