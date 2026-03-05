/**
 * Zipball Repository Fetcher for go-stats-generator WASM
 *
 * Downloads an entire GitHub repository as a ZIP archive in a single
 * API request, then extracts Go source files in-memory using the
 * browser's native DecompressionStream API. This replaces the per-blob
 * GitHub REST API approach which consumed one API request per file
 * (hitting the 60 req/hr unauthenticated limit on small repos).
 *
 * With this approach, fetching a repository always costs exactly
 * ONE API request regardless of repository size.
 */

// ---------------------------------------------------------------------------
// Minimal ZIP reader – extracts files from a ZIP archive using only native
// browser APIs (DataView for parsing, DecompressionStream for inflate).
// ---------------------------------------------------------------------------

class ZipReader {
  /**
   * @param {ArrayBuffer} buffer - Raw ZIP archive bytes.
   */
  constructor(buffer) {
    this.buffer = buffer;
    this.view = new DataView(buffer);
  }

  /**
   * Locate the End of Central Directory (EOCD) record.
   * The EOCD is at the end of the ZIP and starts with signature 0x06054b50.
   * We search backwards because a ZIP comment (variable length) may follow.
   * @returns {number} Byte offset of the EOCD record.
   */
  findEOCD() {
    // EOCD is at least 22 bytes; max comment is 65535 bytes.
    const minOffset = Math.max(0, this.buffer.byteLength - 22 - 65535);
    for (let i = this.buffer.byteLength - 22; i >= minOffset; i--) {
      if (this.view.getUint32(i, true) === 0x06054b50) {
        return i;
      }
    }
    throw new Error('Invalid ZIP archive: End of Central Directory not found');
  }

  /**
   * Parse all Central Directory file entries.
   * @returns {Array<{name: string, method: number, compressedSize: number, uncompressedSize: number, localHeaderOffset: number}>}
   */
  entries() {
    const eocdOff = this.findEOCD();
    const cdCount = this.view.getUint16(eocdOff + 10, true);
    const cdOff = this.view.getUint32(eocdOff + 16, true);

    const result = [];
    let off = cdOff;

    for (let i = 0; i < cdCount; i++) {
      if (this.view.getUint32(off, true) !== 0x02014b50) {
        throw new Error('Invalid ZIP: bad central directory entry signature');
      }

      const method = this.view.getUint16(off + 10, true);
      const compSize = this.view.getUint32(off + 20, true);
      const uncompSize = this.view.getUint32(off + 24, true);
      const nameLen = this.view.getUint16(off + 28, true);
      const extraLen = this.view.getUint16(off + 30, true);
      const commentLen = this.view.getUint16(off + 32, true);
      const localOff = this.view.getUint32(off + 42, true);

      const nameBytes = new Uint8Array(this.buffer, off + 46, nameLen);
      const name = new TextDecoder().decode(nameBytes);

      result.push({
        name,
        method,
        compressedSize: compSize,
        uncompressedSize: uncompSize,
        localHeaderOffset: localOff,
      });

      off += 46 + nameLen + extraLen + commentLen;
    }

    return result;
  }

  /**
   * Read the raw (possibly compressed) data for a file entry.
   * @param {{compressedSize: number, localHeaderOffset: number}} entry
   * @returns {Uint8Array} Raw file data bytes.
   */
  getFileData(entry) {
    const off = entry.localHeaderOffset;
    if (this.view.getUint32(off, true) !== 0x04034b50) {
      throw new Error('Invalid ZIP: bad local file header signature');
    }
    const nameLen = this.view.getUint16(off + 26, true);
    const extraLen = this.view.getUint16(off + 28, true);
    const dataOff = off + 30 + nameLen + extraLen;
    return new Uint8Array(this.buffer, dataOff, entry.compressedSize);
  }

  /**
   * Extract and decode a file entry as a UTF-8 string.
   * Supports method 0 (stored) and method 8 (deflate) via
   * the native DecompressionStream API.
   * @param {{method: number, compressedSize: number, localHeaderOffset: number}} entry
   * @returns {Promise<string>} Decoded file content.
   */
  async readFile(entry) {
    const data = this.getFileData(entry);

    if (entry.method === 0) {
      // Stored – no compression.
      return new TextDecoder().decode(data);
    }

    if (entry.method === 8) {
      // Deflate – use the browser's native DecompressionStream.
      if (typeof DecompressionStream === 'undefined') {
        throw new Error(
          'ZIP inflate (compression method 8) requires browser support for DecompressionStream. ' +
          'Run go-stats-generator in a modern browser that implements the DecompressionStream API ' +
          'or use an environment that does not rely on ZIP-based repository fetching.'
        );
      }
      const ds = new DecompressionStream('deflate-raw');
      const writer = ds.writable.getWriter();
      await writer.write(data);
      await writer.close();

      const reader = ds.readable.getReader();
      const chunks = [];
      for (;;) {
        const { done, value } = await reader.read();
        if (done) break;
        chunks.push(value);
      }

      // Concatenate chunks into a single Uint8Array.
      const totalLen = chunks.reduce((s, c) => s + c.byteLength, 0);
      const merged = new Uint8Array(totalLen);
      let pos = 0;
      for (const chunk of chunks) {
        merged.set(chunk, pos);
        pos += chunk.byteLength;
      }
      return new TextDecoder().decode(merged);
    }

    throw new Error(`Unsupported ZIP compression method: ${entry.method}`);
  }
}

// ---------------------------------------------------------------------------
// ZipballFetcher – downloads a GitHub repo as a ZIP and extracts Go files.
// ---------------------------------------------------------------------------

class ZipballFetcher {
  /**
   * @param {string|null} token - GitHub personal access token (optional).
   */
  constructor(token = null) {
    this.token = token;

    /** @type {AbortController|null} Active controller for cancellation. */
    this.abortController = null;
  }

  /**
   * Parse a GitHub repository URL into owner and repo components.
   * @param {string} url - e.g. https://github.com/owner/repo
   * @returns {{owner: string, repo: string}}
   */
  parseRepoURL(url) {
    const match = url.match(/github\.com\/([^/]+)\/([^/]+?)(?:\.git)?(?:\/|$)/);
    if (!match) {
      throw new Error('Invalid GitHub URL format. Expected: https://github.com/owner/repo');
    }
    return { owner: match[1], repo: match[2] };
  }

  /**
   * Filter a list of ZIP entries to Go source files, applying the same
   * exclusion rules as the WASM git-clone path:
   *  - vendor/ directories
   *  - test files (unless includeTests)
   *  - generated / protobuf files
   *
   * GitHub zipball entries are prefixed with "{owner}-{repo}-{sha}/"
   * which is stripped to produce clean relative paths.
   *
   * @param {Array} entries - ZIP central directory entries.
   * @param {boolean} includeTests
   * @returns {Array<{entry: Object, path: string}>} Matching entries with clean paths.
   */
  filterGoEntries(entries, includeTests) {
    const results = [];

    for (const entry of entries) {
      // Skip directories.
      if (entry.name.endsWith('/')) continue;

      // Strip the top-level directory added by GitHub.
      const path = entry.name.replace(/^[^/]+\//, '');
      if (!path) continue;

      if (!path.endsWith('.go')) continue;
      if (path.includes('/vendor/') || path.startsWith('vendor/')) continue;
      if (!includeTests && path.endsWith('_test.go')) continue;
      if (path.includes('generated') || path.endsWith('.pb.go')) continue;

      results.push({ entry, path });
    }

    return results;
  }

  /**
   * Fetch a GitHub repository's Go source files via the zipball endpoint.
   * This uses exactly ONE API request regardless of repository size.
   *
   * @param {string} repoURL - GitHub repository URL.
   * @param {string|null} ref - Branch, tag, or commit SHA (null for default branch).
   * @param {boolean} includeTests - Whether to include _test.go files.
   * @param {Function|null} onProgress - Progress callback.
   * @returns {Promise<{files: Array<{path: string, content: string}>, stats: Object}>}
   */
  async fetchRepository(repoURL, ref = null, includeTests = false, onProgress = null) {
    this.abortController = new AbortController();

    try {
      const { owner, repo } = this.parseRepoURL(repoURL);

      // --- Download ZIP archive (1 API request) ---
      if (onProgress) {
        onProgress({ stage: 'downloading', message: 'Downloading repository ZIP archive…' });
      }

      let endpoint = `https://api.github.com/repos/${owner}/${repo}/zipball`;
      if (ref) {
        endpoint += `/${encodeURIComponent(ref)}`;
      }

      const headers = {};
      if (this.token) {
        headers['Authorization'] = `Bearer ${this.token}`;
      }

      // Perform initial request with manual redirect handling so that
      // Authorization headers are preserved when GitHub redirects to
      // codeload.github.com for the actual ZIP payload.
      let response = await fetch(endpoint, {
        headers,
        signal: this.abortController.signal,
        redirect: 'manual',
      });

      // If GitHub responds with a redirect (e.g., to codeload.github.com),
      // manually follow it while reusing the same headers and abort signal.
      if (response.status >= 300 && response.status < 400) {
        const location =
          response.headers.get('Location') || response.headers.get('location');
        if (!location) {
          throw new Error('GitHub API error: redirect response missing Location header');
        }
        response = await fetch(location, {
          headers,
          signal: this.abortController.signal,
        });
      }

      if (!response.ok) {
        if (response.status === 404) {
          throw new Error(
            `Repository not found: ${owner}/${repo}` +
            (ref ? ` (ref: ${ref})` : '') +
            '. Check the URL and ensure the repository is public, or provide a token.',
          );
        }
        if (response.status === 403 || response.status === 429) {
          throw new Error(
            'GitHub API rate limit exceeded. Provide a personal access token for higher limits.',
          );
        }
        throw new Error(`GitHub API error: ${response.status} ${response.statusText}`);
      }

      if (onProgress) {
        onProgress({ stage: 'downloading', message: 'Reading ZIP archive…' });
      }

      const buffer = await response.arrayBuffer();

      // --- Extract Go files from ZIP ---
      if (onProgress) {
        onProgress({ stage: 'extracting', message: 'Extracting Go source files…' });
      }

      const zip = new ZipReader(buffer);
      const allEntries = zip.entries();
      const goEntries = this.filterGoEntries(allEntries, includeTests);

      if (goEntries.length === 0) {
        throw new Error('No Go source files found in repository');
      }

      const files = [];
      for (let i = 0; i < goEntries.length; i++) {
        const { entry, path } = goEntries[i];
        const content = await zip.readFile(entry);
        files.push({ path, content });

        if (onProgress && (i % 20 === 0 || i === goEntries.length - 1)) {
          onProgress({
            stage: 'extracting',
            message: `Extracting files (${i + 1}/${goEntries.length})…`,
            current: i + 1,
            total: goEntries.length,
          });
        }
      }

      return {
        files,
        stats: {
          totalFiles: files.length,
          totalSize: files.reduce((sum, f) => sum + new TextEncoder().encode(f.content).byteLength, 0),
          owner,
          repo,
          ref: ref || 'default branch',
        },
      };
    } finally {
      this.abortController = null;
    }
  }

  /**
   * Abort any in-flight download.
   */
  abort() {
    if (this.abortController) {
      this.abortController.abort();
    }
  }
}

// Export for use in other modules / tests.
if (typeof module !== 'undefined' && module.exports) {
  module.exports = { ZipballFetcher, ZipReader };
}
