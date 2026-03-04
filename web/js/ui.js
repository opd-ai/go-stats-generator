/**
 * UI Helper Module for go-stats-generator Web Interface
 * 
 * Provides DOM manipulation utilities for showing/hiding sections,
 * updating progress, and displaying results.
 */

const UI = {
  /**
   * Show an element by removing the 'hidden' class
   * @param {string} elementId - Element ID
   */
  show(elementId) {
    const el = document.getElementById(elementId);
    if (el) el.classList.remove('hidden');
  },

  /**
   * Hide an element by adding the 'hidden' class
   * @param {string} elementId - Element ID
   */
  hide(elementId) {
    const el = document.getElementById(elementId);
    if (el) el.classList.add('hidden');
  },

  /**
   * Update progress bar and status text
   * @param {number} percent - Progress percentage (0-100)
   * @param {string} message - Status message
   */
  updateProgress(percent, message) {
    const bar = document.getElementById('progress-bar');
    const text = document.getElementById('progress-text');
    
    if (bar) bar.style.width = `${percent}%`;
    if (text) text.textContent = message;
  },

  /**
   * Display an error message
   * @param {string} message - Error message
   */
  showError(message) {
    const errorDiv = document.getElementById('error-message');
    if (errorDiv) {
      errorDiv.textContent = message;
      errorDiv.classList.remove('hidden');
    }
  },

  /**
   * Clear error message
   */
  clearError() {
    const errorDiv = document.getElementById('error-message');
    if (errorDiv) {
      errorDiv.textContent = '';
      errorDiv.classList.add('hidden');
    }
  },

  /**
   * Render HTML report in results area
   * @param {string} html - HTML report content
   */
  renderHTMLReport(html) {
    const resultsDiv = document.getElementById('results');
    if (resultsDiv) {
      resultsDiv.innerHTML = html;
    }
  },

  /**
   * Render JSON report in results area
   * @param {string} json - JSON report content
   */
  renderJSONReport(json) {
    const resultsDiv = document.getElementById('results');
    if (resultsDiv) {
      const pre = document.createElement('pre');
      const code = document.createElement('code');
      code.className = 'json';
      
      try {
        const formatted = JSON.stringify(JSON.parse(json), null, 2);
        code.textContent = formatted;
      } catch (e) {
        code.textContent = json;
      }
      
      pre.appendChild(code);
      resultsDiv.innerHTML = '';
      resultsDiv.appendChild(pre);
      
      // Add download button
      const downloadBtn = document.createElement('button');
      downloadBtn.textContent = 'Download JSON';
      downloadBtn.className = 'download-btn';
      downloadBtn.onclick = () => this.downloadJSON(json);
      resultsDiv.insertBefore(downloadBtn, pre);
    }
  },

  /**
   * Download JSON as a file
   * @param {string} json - JSON content
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

  /**
   * Update rate limit status display
   * @param {Object} status - Rate limit status {remaining, reset, authenticated}
   */
  updateRateLimit(status) {
    const rateLimitDiv = document.getElementById('rate-limit');
    if (rateLimitDiv && status.remaining !== null) {
      const resetTime = status.reset ? status.reset.toLocaleTimeString() : 'unknown';
      const authStatus = status.authenticated ? 'authenticated' : 'unauthenticated';
      rateLimitDiv.textContent = `GitHub API: ${status.remaining} requests remaining (${authStatus}) - Resets at ${resetTime}`;
    }
  },

  /**
   * Enable or disable the analyze button
   * @param {boolean} enabled - Whether to enable the button
   */
  setAnalyzeButtonState(enabled) {
    const btn = document.getElementById('analyze-btn');
    if (btn) {
      btn.disabled = !enabled;
    }
  }
};

// Export for use in other modules
if (typeof module !== 'undefined' && module.exports) {
  module.exports = UI;
}
