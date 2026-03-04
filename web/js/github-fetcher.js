/**
 * GitHub Repository Fetcher for go-stats-generator WASM
 *
 * Fetches Go source code from GitHub repositories entirely client-side
 * using the GitHub REST API. Supports branches, tags, commit SHAs, and
 * handles rate limiting with optional authentication.
 */

/** Maximum concurrent blob download requests. */
const BLOB_BATCH_SIZE = 6;

/** Duration (ms) to keep cached blobs in localStorage. */
const BLOB_CACHE_TTL_MS = 7 * 24 * 60 * 60 * 1000; // 7 days

class GitHubFetcher {
  /**
   * @param {string|null} token - GitHub personal access token (optional).
   */
  constructor(token = null) {
    this.token = token;
    this.baseURL = 'https://api.github.com';
    this.rateLimitRemaining = null;
    this.rateLimitReset = null;
    this.cacheEnabled = true;

    /** @type {AbortController|null} Active request controller for cancellation. */
    this.abortController = null;
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
   * Make an authenticated request to the GitHub API.
   * @param {string} endpoint - API endpoint path.
   * @param {string|null} etag - Optional ETag for conditional requests.
   * @returns {Promise<Response>} Fetch response.
   */
  async request(endpoint, etag = null) {
    const headers = {
      'Accept': 'application/vnd.github.v3+json',
    };

    if (this.token) {
      headers['Authorization'] = `Bearer ${this.token}`;
    }

    if (etag && this.cacheEnabled) {
      headers['If-None-Match'] = etag;
    }

    const fetchOptions = { headers };
    if (this.abortController) {
      fetchOptions.signal = this.abortController.signal;
    }

    const response = await fetch(`${this.baseURL}${endpoint}`, fetchOptions);

    // Update rate limit tracking
    this.rateLimitRemaining = parseInt(response.headers.get('X-RateLimit-Remaining') || '0', 10);
    this.rateLimitReset = parseInt(response.headers.get('X-RateLimit-Reset') || '0', 10);

    if (response.status === 304) {
      return response;
    }

    if (!response.ok) {
      if (response.status === 403 && this.rateLimitRemaining === 0) {
        const resetDate = new Date(this.rateLimitReset * 1000);
        throw new Error(
          `GitHub API rate limit exceeded. Resets at ${resetDate.toLocaleTimeString()}. ` +
          'Consider providing a personal access token.'
        );
      }
      throw new Error(`GitHub API error: ${response.status} ${response.statusText}`);
    }

    return response;
  }

  /**
   * Resolve a ref (branch, tag, or commit SHA) to a tree SHA.
   * @param {string} owner - Repository owner.
   * @param {string} repo  - Repository name.
   * @param {string|null} ref - Branch, tag, commit SHA, or null for default branch.
   * @returns {Promise<string>} Tree SHA.
   */
  async resolveTreeSHA(owner, repo, ref = null) {
    if (!ref) {
      ref = await this.fetchDefaultBranch(owner, repo);
    }

    // Try branch → tag → raw commit SHA, in order.
    const treeSHA =
      (await this.tryResolveBranch(owner, repo, ref)) ||
      (await this.tryResolveTag(owner, repo, ref)) ||
      (await this.tryResolveCommit(owner, repo, ref));

    if (!treeSHA) {
      throw new Error(`Could not resolve "${ref}" as a branch, tag, or commit SHA`);
    }
    return treeSHA;
  }

  /**
   * Fetch the default branch name for a repository.
   * @param {string} owner
   * @param {string} repo
   * @returns {Promise<string>}
   */
  async fetchDefaultBranch(owner, repo) {
    const response = await this.request(`/repos/${owner}/${repo}`);
    const data = await response.json();
    return data.default_branch;
  }

  /**
   * Try to resolve a ref as a branch name.
   * @returns {Promise<string|null>} Tree SHA or null.
   */
  async tryResolveBranch(owner, repo, ref) {
    try {
      const response = await this.request(`/repos/${owner}/${repo}/git/ref/heads/${ref}`);
      const data = await response.json();
      return this.commitToTreeSHA(owner, repo, data.object.sha);
    } catch {
      return null;
    }
  }

  /**
   * Try to resolve a ref as a tag name, handling both lightweight and
   * annotated tags.
   * @returns {Promise<string|null>} Tree SHA or null.
   */
  async tryResolveTag(owner, repo, ref) {
    try {
      const response = await this.request(`/repos/${owner}/${repo}/git/ref/tags/${ref}`);
      const data = await response.json();

      let commitSHA = data.object.sha;

      // Annotated tags point to a tag object, not a commit. Dereference.
      if (data.object.type === 'tag') {
        const tagResponse = await this.request(`/repos/${owner}/${repo}/git/tags/${commitSHA}`);
        const tagData = await tagResponse.json();
        commitSHA = tagData.object.sha;
      }

      return this.commitToTreeSHA(owner, repo, commitSHA);
    } catch {
      return null;
    }
  }

  /**
   * Try to resolve a ref as a raw commit SHA.
   * @returns {Promise<string|null>} Tree SHA or null.
   */
  async tryResolveCommit(owner, repo, ref) {
    try {
      return this.commitToTreeSHA(owner, repo, ref);
    } catch {
      return null;
    }
  }

  /**
   * Given a commit SHA, return its tree SHA.
   * @param {string} owner
   * @param {string} repo
   * @param {string} commitSHA
   * @returns {Promise<string>} Tree SHA.
   */
  async commitToTreeSHA(owner, repo, commitSHA) {
    const response = await this.request(`/repos/${owner}/${repo}/git/commits/${commitSHA}`);
    const data = await response.json();
    return data.tree.sha;
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
   * Set cached blob data in localStorage.
   * @param {string} cacheKey - localStorage key.
   * @param {string} content  - Decoded blob content.
   * @param {string} etag     - ETag from the response.
   */
  setCachedBlob(cacheKey, content, etag) {
    if (!this.cacheEnabled) return;
    try {
      localStorage.setItem(cacheKey, JSON.stringify({
        content,
        etag,
        expiresAt: Date.now() + BLOB_CACHE_TTL_MS,
      }));
    } catch (e) {
      console.warn('Failed to cache blob:', e);
    }
  }

  /**
   * Decode a base64-encoded string to a proper UTF-8 string.
   * Unlike bare `atob()`, this handles multi-byte UTF-8 characters
   * correctly (e.g. comments or strings in CJK, emoji, etc.).
   * @param {string} base64 - Base64-encoded data (may contain newlines).
   * @returns {string} Decoded UTF-8 string.
   */
  decodeBase64UTF8(base64) {
    const binaryString = atob(base64.replace(/\n/g, ''));
    const bytes = Uint8Array.from(binaryString, (ch) => ch.charCodeAt(0));
    return new TextDecoder().decode(bytes);
  }

  /**
   * Fetch blob content for a single file.
   * @param {string} owner - Repository owner.
   * @param {string} repo  - Repository name.
   * @param {string} sha   - Blob SHA.
   * @returns {Promise<string>} File content as a UTF-8 string.
   */
  async fetchBlob(owner, repo, sha) {
    const cacheKey = this.getBlobCacheKey(owner, repo, sha);
    const cached = this.getCachedBlob(cacheKey);

    // If we have a cached copy, make a conditional request using its ETag.
    if (cached) {
      const response = await this.request(
        `/repos/${owner}/${repo}/git/blobs/${sha}`,
        cached.etag,
      );
      if (response.status === 304) {
        return cached.content;
      }
      // Cache is stale – use the response we already received instead of
      // making a second, redundant request.
      return this.parseBlobResponse(response, cacheKey);
    }

    // No cache – fetch fresh.
    const response = await this.request(`/repos/${owner}/${repo}/git/blobs/${sha}`);
    return this.parseBlobResponse(response, cacheKey);
  }

  /**
   * Extract and cache the UTF-8 content from a blob API response.
   * @param {Response} response - Fetch response from the blobs endpoint.
   * @param {string} cacheKey  - localStorage cache key.
   * @returns {Promise<string>} Decoded file content.
   */
  async parseBlobResponse(response, cacheKey) {
    const data = await response.json();
    const content = this.decodeBase64UTF8(data.content);
    const etag = response.headers.get('ETag');

    if (etag) {
      this.setCachedBlob(cacheKey, content, etag);
    }
    return content;
  }

  /**
   * Fetch all Go file contents with progress tracking.
   * Downloads blobs in parallel batches of {@link BLOB_BATCH_SIZE}.
   * @param {string} owner - Repository owner.
   * @param {string} repo  - Repository name.
   * @param {Array} files  - File entries from the tree.
   * @param {Function|null} onProgress - Callback receiving (current, total).
   * @returns {Promise<Array<{path: string, content: string}>>}
   */
  async fetchFiles(owner, repo, files, onProgress = null) {
    const results = [];

    for (let i = 0; i < files.length; i += BLOB_BATCH_SIZE) {
      const batch = files.slice(i, i + BLOB_BATCH_SIZE);
      const batchResults = await Promise.all(
        batch.map(async (file) => {
          try {
            const content = await this.fetchBlob(owner, repo, file.sha);
            return { path: file.path, content };
          } catch (error) {
            console.error(`Failed to fetch ${file.path}:`, error);
            return null;
          }
        }),
      );

      for (const r of batchResults) {
        if (r) results.push(r);
      }

      if (onProgress) {
        onProgress(results.length, files.length);
      }
    }

    return results;
  }

  /**
   * Fetch a GitHub repository's Go source files.
   * @param {string} repoURL - GitHub repository URL.
   * @param {string|null} ref - Branch, tag, or commit SHA (null for default).
   * @param {boolean} includeTests - Whether to include test files.
   * @param {Function|null} onProgress - Progress callback.
   * @returns {Promise<{files: Array, stats: Object}>}
   */
  async fetchRepository(repoURL, ref = null, includeTests = false, onProgress = null) {
    // Create a fresh AbortController for this request cycle.
    this.abortController = new AbortController();

    try {
      const { owner, repo } = this.parseRepoURL(repoURL);

      if (onProgress) onProgress({ stage: 'resolving', message: 'Resolving ref…' });
      const treeSHA = await this.resolveTreeSHA(owner, repo, ref);

      if (onProgress) onProgress({ stage: 'fetching_tree', message: 'Fetching repository tree…' });
      const tree = await this.fetchTree(owner, repo, treeSHA);
      const goFiles = this.filterGoFiles(tree, includeTests);

      if (goFiles.length === 0) {
        throw new Error('No Go source files found in repository');
      }

      const files = await this.fetchFiles(owner, repo, goFiles, (current, total) => {
        if (onProgress) {
          onProgress({
            stage: 'downloading',
            message: `Downloading files (${current}/${total})…`,
            current,
            total,
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
          treeSHA,
        },
      };
    } finally {
      this.abortController = null;
    }
  }

  /**
   * Abort any in-flight fetch requests started by {@link fetchRepository}.
   */
  abort() {
    if (this.abortController) {
      this.abortController.abort();
      this.abortController = null;
    }
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
