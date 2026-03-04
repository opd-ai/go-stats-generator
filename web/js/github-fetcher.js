/**
 * GitHub Repository Fetcher for go-stats-generator WASM
 * 
 * Fetches Go source code from GitHub repositories entirely client-side using the GitHub REST API.
 * Supports branches, tags, commit SHAs, and handles rate limiting with optional authentication.
 */

class GitHubFetcher {
  constructor(token = null) {
    this.token = token;
    this.baseURL = 'https://api.github.com';
    this.rateLimitRemaining = null;
    this.rateLimitReset = null;
    this.cacheEnabled = true;
  }

  /**
   * Parse a GitHub repository URL into owner and repo components
   * @param {string} url - GitHub repository URL (e.g., https://github.com/owner/repo)
   * @returns {{owner: string, repo: string}} - Parsed components
   */
  parseRepoURL(url) {
    const match = url.match(/github\.com\/([^/]+)\/([^/]+?)(?:\.git)?(?:\/|$)/);
    if (!match) {
      throw new Error('Invalid GitHub URL format. Expected: https://github.com/owner/repo');
    }
    return { owner: match[1], repo: match[2] };
  }

  /**
   * Make an authenticated request to the GitHub API
   * @param {string} endpoint - API endpoint path
   * @param {string} etag - Optional ETag for conditional requests
   * @returns {Promise<Response>} - Fetch response
   */
  async request(endpoint, etag = null) {
    const headers = {
      'Accept': 'application/vnd.github.v3+json'
    };
    
    if (this.token) {
      headers['Authorization'] = `Bearer ${this.token}`;
    }
    
    if (etag && this.cacheEnabled) {
      headers['If-None-Match'] = etag;
    }

    const response = await fetch(`${this.baseURL}${endpoint}`, { headers });
    
    // Update rate limit tracking
    this.rateLimitRemaining = parseInt(response.headers.get('X-RateLimit-Remaining') || '0');
    this.rateLimitReset = parseInt(response.headers.get('X-RateLimit-Reset') || '0');
    
    if (response.status === 304) {
      // Not modified - return cached response marker
      return response;
    }
    
    if (!response.ok) {
      if (response.status === 403 && this.rateLimitRemaining === 0) {
        const resetDate = new Date(this.rateLimitReset * 1000);
        throw new Error(`GitHub API rate limit exceeded. Resets at ${resetDate.toLocaleTimeString()}. Consider providing a personal access token.`);
      }
      throw new Error(`GitHub API error: ${response.status} ${response.statusText}`);
    }
    
    return response;
  }

  /**
   * Resolve a ref (branch, tag, or commit SHA) to a tree SHA
   * @param {string} owner - Repository owner
   * @param {string} repo - Repository name
   * @param {string} ref - Branch name, tag name, commit SHA, or null for default branch
   * @returns {Promise<string>} - Tree SHA
   */
  async resolveTreeSHA(owner, repo, ref = null) {
    if (!ref) {
      // Get default branch
      const response = await this.request(`/repos/${owner}/${repo}`);
      const repoData = await response.json();
      ref = repoData.default_branch;
    }

    // Try as branch first
    try {
      const response = await this.request(`/repos/${owner}/${repo}/git/ref/heads/${ref}`);
      const data = await response.json();
      const commitSHA = data.object.sha;
      
      // Get commit to extract tree SHA
      const commitResponse = await this.request(`/repos/${owner}/${repo}/git/commits/${commitSHA}`);
      const commitData = await commitResponse.json();
      return commitData.tree.sha;
    } catch (e) {
      // Not a branch, try as tag
      try {
        const response = await this.request(`/repos/${owner}/${repo}/git/ref/tags/${ref}`);
        const data = await response.json();
        const commitSHA = data.object.sha;
        
        const commitResponse = await this.request(`/repos/${owner}/${repo}/git/commits/${commitSHA}`);
        const commitData = await commitResponse.json();
        return commitData.tree.sha;
      } catch (e2) {
        // Assume it's a commit SHA directly
        try {
          const commitResponse = await this.request(`/repos/${owner}/${repo}/git/commits/${ref}`);
          const commitData = await commitResponse.json();
          return commitData.tree.sha;
        } catch (e3) {
          throw new Error(`Could not resolve ref "${ref}" as branch, tag, or commit SHA`);
        }
      }
    }
  }

  /**
   * Fetch the repository tree recursively
   * @param {string} owner - Repository owner
   * @param {string} repo - Repository name
   * @param {string} treeSHA - Tree SHA to fetch
   * @returns {Promise<Array>} - Array of tree entries
   */
  async fetchTree(owner, repo, treeSHA) {
    const response = await this.request(`/repos/${owner}/${repo}/git/trees/${treeSHA}?recursive=1`);
    const data = await response.json();
    
    if (data.truncated) {
      console.warn('Repository tree was truncated by GitHub API. Some files may be missing. Consider using zipball endpoint for very large repositories.');
    }
    
    return data.tree;
  }

  /**
   * Filter tree entries to only include Go source files
   * @param {Array} tree - Tree entries from GitHub API
   * @param {boolean} includeTests - Whether to include _test.go files (default: false)
   * @returns {Array} - Filtered tree entries
   */
  filterGoFiles(tree, includeTests = false) {
    return tree.filter(entry => {
      if (entry.type !== 'blob') return false;
      if (!entry.path.endsWith('.go')) return false;
      
      // Exclude vendor directories
      if (entry.path.includes('/vendor/')) return false;
      
      // Exclude test files if requested
      if (!includeTests && entry.path.endsWith('_test.go')) return false;
      
      // Exclude generated files
      if (entry.path.includes('generated') || entry.path.includes('.pb.go')) return false;
      
      return true;
    });
  }

  /**
   * Get cache key for a blob
   * @param {string} owner - Repository owner
   * @param {string} repo - Repository name
   * @param {string} sha - Blob SHA
   * @returns {string} - Cache key
   */
  getBlobCacheKey(owner, repo, sha) {
    return `gh-blob:${owner}/${repo}:${sha}`;
  }

  /**
   * Get cached blob data from localStorage
   * @param {string} cacheKey - Cache key
   * @returns {Object|null} - Cached data or null
   */
  getCachedBlob(cacheKey) {
    if (!this.cacheEnabled) return null;
    try {
      const cached = localStorage.getItem(cacheKey);
      if (!cached) return null;
      const data = JSON.parse(cached);
      if (data.expiresAt && Date.now() > data.expiresAt) {
        localStorage.removeItem(cacheKey);
        return null;
      }
      return data;
    } catch (e) {
      return null;
    }
  }

  /**
   * Set cached blob data in localStorage
   * @param {string} cacheKey - Cache key
   * @param {string} content - Blob content
   * @param {string} etag - ETag from response
   */
  setCachedBlob(cacheKey, content, etag) {
    if (!this.cacheEnabled) return;
    try {
      const data = {
        content,
        etag,
        expiresAt: Date.now() + (7 * 24 * 60 * 60 * 1000) // 7 days
      };
      localStorage.setItem(cacheKey, JSON.stringify(data));
    } catch (e) {
      console.warn('Failed to cache blob:', e);
    }
  }

  /**
   * Fetch blob content for a single file
   * @param {string} owner - Repository owner
   * @param {string} repo - Repository name
   * @param {string} sha - Blob SHA
   * @returns {Promise<string>} - File content as UTF-8 string
   */
  async fetchBlob(owner, repo, sha) {
    const cacheKey = this.getBlobCacheKey(owner, repo, sha);
    const cached = this.getCachedBlob(cacheKey);
    
    if (cached) {
      const response = await this.request(`/repos/${owner}/${repo}/git/blobs/${sha}`, cached.etag);
      if (response.status === 304) {
        return cached.content;
      }
    }
    
    const response = await this.request(`/repos/${owner}/${repo}/git/blobs/${sha}`);
    const data = await response.json();
    const content = atob(data.content);
    const etag = response.headers.get('ETag');
    
    if (etag) {
      this.setCachedBlob(cacheKey, content, etag);
    }
    
    return content;
  }

  /**
   * Fetch all Go file contents with progress tracking
   * @param {string} owner - Repository owner
   * @param {string} repo - Repository name
   * @param {Array} files - Array of file entries from tree
   * @param {Function} onProgress - Progress callback (current, total)
   * @returns {Promise<Array>} - Array of {path, content} objects
   */
  async fetchFiles(owner, repo, files, onProgress = null) {
    const results = [];
    const batchSize = 6; // Concurrent requests
    
    for (let i = 0; i < files.length; i += batchSize) {
      const batch = files.slice(i, i + batchSize);
      const promises = batch.map(async (file) => {
        try {
          const content = await this.fetchBlob(owner, repo, file.sha);
          return { path: file.path, content };
        } catch (error) {
          console.error(`Failed to fetch ${file.path}:`, error);
          return { path: file.path, content: '', error: error.message };
        }
      });
      
      const batchResults = await Promise.all(promises);
      results.push(...batchResults);
      
      if (onProgress) {
        onProgress(results.length, files.length);
      }
    }
    
    return results.filter(r => !r.error);
  }

  /**
   * Fetch a GitHub repository's Go source files
   * @param {string} repoURL - GitHub repository URL
   * @param {string} ref - Branch, tag, or commit SHA (null for default)
   * @param {boolean} includeTests - Whether to include test files
   * @param {Function} onProgress - Progress callback
   * @returns {Promise<{files: Array, stats: Object}>} - Repository files and stats
   */
  async fetchRepository(repoURL, ref = null, includeTests = false, onProgress = null) {
    const { owner, repo } = this.parseRepoURL(repoURL);
    
    // Resolve ref to tree SHA
    if (onProgress) onProgress({ stage: 'resolving', message: 'Resolving ref...' });
    const treeSHA = await this.resolveTreeSHA(owner, repo, ref);
    
    // Fetch tree
    if (onProgress) onProgress({ stage: 'fetching_tree', message: 'Fetching repository tree...' });
    const tree = await this.fetchTree(owner, repo, treeSHA);
    
    // Filter to Go files
    const goFiles = this.filterGoFiles(tree, includeTests);
    
    if (goFiles.length === 0) {
      throw new Error('No Go source files found in repository');
    }
    
    // Fetch file contents
    const files = await this.fetchFiles(owner, repo, goFiles, (current, total) => {
      if (onProgress) {
        onProgress({ 
          stage: 'downloading', 
          message: `Downloading files (${current}/${total})...`,
          current,
          total
        });
      }
    });
    
    return {
      files,
      stats: {
        totalFiles: files.length,
        totalSize: files.reduce((sum, f) => sum + f.content.length, 0),
        owner,
        repo,
        ref: ref || 'default branch',
        treeSHA
      }
    };
  }

  /**
   * Get current rate limit status
   * @returns {{remaining: number, reset: Date, authenticated: boolean}}
   */
  getRateLimitStatus() {
    return {
      remaining: this.rateLimitRemaining,
      reset: this.rateLimitReset ? new Date(this.rateLimitReset * 1000) : null,
      authenticated: !!this.token
    };
  }

  /**
   * Update authentication token
   * @param {string} token - GitHub personal access token
   */
  setToken(token) {
    this.token = token;
  }

  /**
   * Enable or disable localStorage caching
   * @param {boolean} enabled - Whether to enable caching
   */
  setCacheEnabled(enabled) {
    this.cacheEnabled = enabled;
  }

  /**
   * Clear all cached repository data
   */
  clearCache() {
    try {
      const keys = Object.keys(localStorage);
      for (const key of keys) {
        if (key.startsWith('gh-blob:')) {
          localStorage.removeItem(key);
        }
      }
    } catch (e) {
      console.warn('Failed to clear cache:', e);
    }
  }
}

// Export for use in other modules
if (typeof module !== 'undefined' && module.exports) {
  module.exports = GitHubFetcher;
}
