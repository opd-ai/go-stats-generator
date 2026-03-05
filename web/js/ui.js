/**
 * UI Helper Module for go-stats-generator Web Interface
 *
 * Provides DOM manipulation utilities for showing/hiding sections,
 * updating progress, and displaying results.
 */

const UI = {
  // ---------------------------------------------------------------------------
  // Visibility helpers
  // ---------------------------------------------------------------------------

  /**
   * Show an element by removing the 'hidden' class.
   * @param {string} id - Element ID.
   */
  show(id) {
    const el = document.getElementById(id);
    if (el) el.classList.remove('hidden');
  },

  /**
   * Hide an element by adding the 'hidden' class.
   * @param {string} id - Element ID.
   */
  hide(id) {
    const el = document.getElementById(id);
    if (el) el.classList.add('hidden');
  },

  // ---------------------------------------------------------------------------
  // Progress
  // ---------------------------------------------------------------------------

  /**
   * Update the progress bar width and status text.
   * @param {number} percent - 0–100.
   * @param {string} message - Human-readable status.
   */
  updateProgress(percent, message) {
    const bar = document.getElementById('progress-bar');
    const text = document.getElementById('progress-text');

    if (bar) bar.style.width = `${percent}%`;
    if (text) text.textContent = message;
  },

  // ---------------------------------------------------------------------------
  // Error display
  // ---------------------------------------------------------------------------

  /** Show an error message in the designated alert area. */
  showError(message) {
    const el = document.getElementById('error-message');
    if (el) {
      el.textContent = message;
      el.classList.remove('hidden');
    }
  },

  /** Clear any visible error message. */
  clearError() {
    const el = document.getElementById('error-message');
    if (el) {
      el.textContent = '';
      el.classList.add('hidden');
    }
  },

  // ---------------------------------------------------------------------------
  // Results rendering
  // ---------------------------------------------------------------------------

  /**
   * Render an HTML report inside an iframe for complete CSS isolation.
   * The generated report contains its own <style> tags which would otherwise
   * leak into – and conflict with – the host page styles.
   * @param {string} html - HTML string produced by the WASM analyzer.
   */
  renderHTMLReport(html) {
    const container = document.getElementById('results');
    if (!container) return;

    container.textContent = '';

    const iframe = document.createElement('iframe');
    iframe.className = 'report-frame';
    iframe.setAttribute('sandbox', 'allow-scripts');
    iframe.srcdoc = html;
    iframe.title = 'Analysis Report';

    iframe.addEventListener('load', () => {
      try {
        const doc = iframe.contentDocument || iframe.contentWindow.document;
        iframe.style.height = doc.documentElement.scrollHeight + 'px';
      } catch (e) {
        // Sandboxed iframe without allow-same-origin cannot be measured;
        // fall back to a generous viewport-relative height.
        iframe.style.height = '80vh';
      }
    });

    container.appendChild(iframe);
  },

  /**
   * Render a JSON report with syntax formatting and a download button.
   * @param {string} json - JSON string produced by the WASM analyzer.
   */
  renderJSONReport(json) {
    const container = document.getElementById('results');
    if (!container) return;

    container.textContent = '';

    // Download button
    const downloadBtn = document.createElement('button');
    downloadBtn.textContent = 'Download JSON';
    downloadBtn.className = 'download-btn';
    downloadBtn.addEventListener('click', () => this.downloadJSON(json));
    container.appendChild(downloadBtn);

    // Formatted code block
    const pre = document.createElement('pre');
    const code = document.createElement('code');
    code.className = 'json';
    try {
      code.textContent = JSON.stringify(JSON.parse(json), null, 2);
    } catch {
      code.textContent = json;
    }
    pre.appendChild(code);
    container.appendChild(pre);
  },

  /**
   * Trigger a browser download of the JSON report.
   * @param {string} json
   */
  downloadJSON(json) {
    const blob = new Blob([json], { type: 'application/json' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = 'go-stats-report.json';
    a.click();
    URL.revokeObjectURL(url);
  },

  // ---------------------------------------------------------------------------
  // Status indicators
  // ---------------------------------------------------------------------------

  /**
   * Update the GitHub API rate-limit indicator in the footer.
   * Only shown when the API fallback is used.
   * @param {{remaining: number|null, reset: Date|null, authenticated: boolean}} status
   */
  updateRateLimit(status) {
    const el = document.getElementById('rate-limit');
    if (!el || status.remaining === null) return;

    const resetTime = status.reset ? status.reset.toLocaleTimeString() : 'unknown';
    const authLabel = status.authenticated ? 'authenticated' : 'unauthenticated';
    el.textContent =
      `GitHub API fallback: ${status.remaining} requests remaining (${authLabel}) – resets at ${resetTime}`;
    el.classList.remove('hidden');
  },

  /**
   * Enable or disable the analyze button.
   * @param {boolean} enabled
   */
  setAnalyzeButtonState(enabled) {
    const btn = document.getElementById('analyze-btn');
    if (btn) btn.disabled = !enabled;
  },

  /**
   * Show or hide the cancel button. Hidden during git-clone analysis
   * because the WASM goroutine cannot be cancelled from JavaScript.
   * @param {boolean} visible
   */
  setCancelVisible(visible) {
    const btn = document.getElementById('cancel-btn');
    if (btn) {
      if (visible) {
        btn.classList.remove('hidden');
      } else {
        btn.classList.add('hidden');
      }
    }
  },
};

// Export for use in other modules.
if (typeof module !== 'undefined' && module.exports) {
  module.exports = UI;
}
