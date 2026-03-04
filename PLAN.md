# go-stats-generator: WebAssembly (WASM) Deployment Plan

A step-by-step implementation plan for deploying `go-stats-generator` as a client-side WebAssembly application hosted on GitHub Pages. After deployment, all analysis runs entirely in the visitor's browser with zero server-side processing.

---

## Phase 1: WASM Build Target

### Goal

Compile the Go analysis engine to `GOOS=js GOARCH=wasm`, producing a `.wasm` binary that exposes analysis functions to JavaScript. Isolate and replace packages incompatible with the WASM target.

### Steps

1. **✅ Audit package compatibility** — Catalog every import used by the analysis path (`internal/analyzer/*`, `internal/metrics/*`, `internal/reporter/*`, `internal/scanner/*`, `internal/config/*`, `pkg/go-stats-generator/*`). Classify each as:
   - *WASM-safe* — Go standard library AST packages (`go/parser`, `go/ast`, `go/token`), `html/template`, `encoding/json`, `encoding/csv`, `fmt`, `strings`, `time`, `math`, etc.
   - *Needs adaptation* — `internal/scanner` (uses `os`, `filepath.Walk`), `internal/storage` (SQLite, PostgreSQL, MongoDB), `cmd/*` (Cobra CLI), `internal/api` (HTTP server), `fsnotify`.
   - *Exclude entirely* — `modernc.org/sqlite`, `github.com/lib/pq`, `go.mongodb.org/mongo-driver`, `github.com/spf13/cobra`, `github.com/spf13/viper`, `github.com/fsnotify/fsnotify`, `internal/multirepo`.
   - **Status:** Complete (2026-03-04). Audit report: 28 WASM-safe packages, 3 need adaptation, 7 external deps to exclude. All core analyzers (function, struct, interface, package, concurrency, duplication, naming, placement, burden, documentation) are pure AST and require zero changes. See session artifact for full analysis.

2. **✅ Create a WASM-specific entry point** — Add a new file `cmd/wasm/main.go` with build tag `//go:build js && wasm`. This file:
   - Imports `syscall/js` to register Go functions on the JavaScript `globalThis`.
   - Exposes an `analyzeCode(filesJSON)` function that accepts a JSON array of `{path, content}` objects (the fetched repository files).
   - Internally constructs an in-memory `token.FileSet`, parses each file with `go/parser.ParseFile`, and feeds results through the existing analyzer pipeline.
   - Returns the `metrics.Report` serialized as JSON (or the rendered HTML string from `internal/reporter`).
   - Blocks with `select {}` to keep the Go runtime alive.
   - **Status:** Complete (2026-03-04). Created `cmd/wasm/main.go` and `cmd/wasm/doc.go` with full JavaScript API exposure, configurable analysis parameters, HTML/JSON output formats, and proper error handling. All functions under 30 lines with complexity ≤10. WASM binary compiles successfully at 5.8MB. Placeholder for in-memory file analysis ready for integration with steps 3-4.

3. **✅ Create a WASM-compatible scanner shim** — Add `internal/scanner/discover_wasm.go` (with build tag `//go:build js && wasm`) that replaces `filepath.Walk` and `os` calls with a `DiscoverFilesFromMemory(files []MemoryFile)` function. The existing `discover.go` keeps the build tag `//go:build !js || !wasm` so native builds are unchanged.
   - **Status:** Complete (2026-03-04). Created `internal/scanner/discover_wasm.go` with `DiscoverFilesFromMemory`, `analyzeMemoryFile`, `ParseMemoryFile`, and `shouldSkipMemoryDirectory` functions. Added `internal/scanner/types.go` and `internal/scanner/helpers.go` to share common code between native and WASM builds. All new functions are under 30 lines with complexity ≤6. Native and WASM builds both compile successfully. All existing scanner tests pass (16/16).

4. **✅ Create a WASM-compatible worker shim** — Add `internal/scanner/worker_wasm.go` that provides a single-threaded `ProcessFiles` implementation. The WASM target does not support OS-level concurrency, so files are processed sequentially in a simple loop. The native `worker.go` retains the `//go:build !js || !wasm` tag.
   - **Status:** Complete (2026-03-04). Created `internal/scanner/worker_wasm.go` with sequential file processing (9 functions, all ≤20 lines, complexity ≤4). Added build tag `//go:build !js || !wasm` to `worker.go`. Created `pkg/go-stats-generator/api_wasm.go` with `AnalyzeMemoryFiles` method and `pkg/go-stats-generator/api_common.go` to share analysis logic between native and WASM builds. Updated `cmd/wasm/main.go` to use the new API. Both native and WASM builds compile successfully (WASM binary: 7.5MB). All scanner tests pass. Zero regressions in unchanged code. New code quality: functions over 30 lines = 0, complexity over 10 = 0.

5. **✅ Exclude storage from the WASM build** — Add build tags to `internal/storage/sqlite.go`, `internal/storage/json.go`, and `internal/storage/memory.go` so they are excluded from `js/wasm`. Provide a stub `internal/storage/storage_wasm.go` returning `ErrNotSupported` for any storage call. Since the browser UI performs one-shot analysis (no baseline/diff/trend), storage is not needed.
   - **Status:** Complete (2026-03-04). Added `//go:build !js || !wasm` tags to `sqlite.go`, `json.go`, and `memory.go`. Created `internal/storage/storage_wasm.go` with stub types (SQLiteStorage, JSONStorage, MemoryStorage) and constructor functions that return `ErrNotSupported`. All functions under 10 lines. Both native and WASM builds compile successfully. All storage tests pass (42/42). Zero regressions in unchanged code. Documentation added explaining WASM limitations.

6. **✅ Add Makefile target** — Add a `build-wasm` target:
   - **Status:** Complete (2026-03-04). Added `build-wasm` target to Makefile with WASM binary compilation and wasm_exec.js copying. Updated help section to document the new target. Target successfully builds 7.5MB WASM binary to `build/wasm/go-stats-generator.wasm`.

7. **✅ Verify compilation** — Run `GOOS=js GOARCH=wasm go build ./cmd/wasm/` locally. Fix any remaining import errors by adding build tags or shims until the binary compiles cleanly.
   - **Status:** Complete (2026-03-04). WASM compilation verified successful with `GOOS=js GOARCH=wasm go build ./cmd/wasm/`. Native build confirmed working with `go build .`. Binary size: 7.5MB uncompressed. All build targets functional.

### Dependencies

- Go 1.23.2+ (already in use).
- No new Go module dependencies.

### Open Questions / Risks

- **Binary size**: The WASM blob may exceed 15 MB before compression. Mitigation: apply `wasm-opt` from Binaryen for size optimization; serve with Brotli/gzip compression (GitHub Pages supports gzip automatically). Aim for < 5 MB compressed.
- **`go/parser` in WASM**: The standard library parser works in WASM, but parsing thousands of files may be slow in a single thread. Large repositories (> 5,000 files) could take 30+ seconds. The UI should show a progress indicator.
- **Goroutine support**: `GOOS=js` supports goroutines cooperatively (single-threaded), so `sync.WaitGroup` and channels still compile. However, true parallelism is unavailable. The sequential worker shim avoids relying on parallelism.

---

## Phase 2: Client-Side Repository Fetching

### Goal

Fetch the contents of a remote Go repository entirely in the browser, without any server-side proxy, supporting HEAD of the default branch, specific branches, tags, and commit SHAs.

### Steps

1. **Choose the fetching strategy** — Use the **GitHub REST API (Trees endpoint)** as the primary mechanism:
   - `GET /repos/{owner}/{repo}/git/trees/{tree_sha}?recursive=1` returns the full file tree in a single request.
   - Individual file contents are fetched via `GET /repos/{owner}/{repo}/git/blobs/{sha}` (base64-encoded).
   - This avoids pulling the full `.git` history (which `isomorphic-git` would do) and keeps network transfer minimal.
   - **Fallback consideration**: For repositories exceeding the GitHub tree API's truncation limit (~100,000 entries), implement paginated directory traversal using `GET /repos/{owner}/{repo}/contents/{path}?ref={ref}`.
   - **Bulk download alternative**: For very large repositories, use the zipball endpoint (`GET /repos/{owner}/{repo}/zipball/{ref}`) and extract `.go` files client-side with `fflate` (preferred over JSZip for its smaller bundle size and better performance in a browser/WASM context). This reduces API requests to a single call at the cost of a larger download.

2. **Resolve refs to a tree SHA** — Before fetching the tree:
   - If the user provides a branch or tag name: `GET /repos/{owner}/{repo}/git/ref/heads/{branch}` or `.../tags/{tag}` to get the commit SHA, then `GET /repos/{owner}/{repo}/git/commits/{sha}` to get the tree SHA.
   - If the user provides a commit SHA directly: fetch the commit to get the tree SHA.
   - If no ref is specified: `GET /repos/{owner}/{repo}` to discover the default branch, then resolve as above.

3. **Filter the tree for Go files** — From the recursive tree response, select only entries where `path` ends in `.go` and `type === "blob"`. Exclude files matching the same filters the CLI uses (vendor directories, `_test.go` if configured, generated files).

4. **Fetch file contents in parallel batches** — Use `Promise.all` with concurrency limiting (e.g., 6 concurrent fetches) to download blob contents. Decode from base64 to UTF-8 strings. Assemble an array of `{path, content}` objects to pass to the WASM `analyzeCode` function.

5. **Handle rate limiting and authentication**:
   - Unauthenticated GitHub API: 60 requests/hour. A medium repository (200 Go files) requires ~201 requests (1 tree + 200 blobs). This will hit the limit quickly.
   - Provide an optional **GitHub Personal Access Token** input field in the UI. When provided, include it as `Authorization: Bearer <token>` to raise the limit to 5,000 requests/hour.
   - Display the current rate limit status (`X-RateLimit-Remaining` header) in the UI.
   - On `403` rate-limit responses, show a clear message prompting the user to supply a token.

6. **Optimize with conditional requests** — Use `localStorage` to cache the tree SHA and blob contents with ETags. On repeat analysis of the same repository, send `If-None-Match` headers to avoid re-downloading unchanged files.

7. **Support non-GitHub hosts (future)** — Initially, only GitHub-hosted repositories are supported. Document this limitation and note that GitLab/Bitbucket API adapters could be added later using the same `{path, content}` interface.

### Dependencies

- GitHub REST API v3 (no additional libraries required; use browser `fetch()`).
- No npm packages needed for fetching.

### Open Questions / Risks

- **Large repositories**: Repos with thousands of Go files will require thousands of blob requests. Even with a token (5,000/hour), analyzing a 3,000-file repo consumes most of the hourly budget. Mitigation: use the zipball endpoint (described in Step 1 above) with `fflate` for client-side extraction, reducing API requests to 1 at the cost of a larger download.
- **Private repositories**: Require a token with `repo` scope. The UI should never store tokens in `localStorage`; use `sessionStorage` only, and clear on page unload.
- **CORS**: GitHub API supports CORS for browser requests. No proxy needed.

---

## Phase 3: Browser UI

### Goal

Build a single-page application that lets users input a repository URL, select a ref, trigger analysis, and view rich results — all running client-side with static assets.

### Steps

1. **Technology choice** — Use **vanilla HTML/CSS/JavaScript** (no framework). Rationale:
   - The UI is a single page with a form, a progress area, and a results container. A framework adds bundle complexity without meaningful benefit.
   - No build step required (no webpack/vite/rollup); files are served as-is from GitHub Pages.
   - Keeps the deployment pipeline simple (copy static files + WASM binary).

2. **Page layout** — Create `web/index.html` with the following sections:
   - **Header**: Tool name (`go-stats-generator`), brief description, link to the CLI repository.
   - **Input area**: Repository URL text field (e.g., `https://github.com/owner/repo`), ref selector (text input with placeholder "branch, tag, or SHA — leave blank for default branch"), optional GitHub token field (password type), and an "Analyze" button.
   - **Progress area** (hidden by default): Progress bar, status text ("Fetching repository tree...", "Downloading files (42/200)...", "Running analysis..."), cancel button.
   - **Results area** (hidden by default): Rendered analysis report.
   - **Footer**: Rate limit status, link to source code, version info.

3. **JavaScript modules** — Organize code in `web/js/`:
   - `wasm-loader.js` — Loads `wasm_exec.js`, instantiates the WASM binary, and wraps the exposed Go functions in async JS wrappers.
   - `github-fetcher.js` — Implements the repository fetching logic from Phase 2 (tree resolution, blob downloading, progress callbacks).
   - `app.js` — Glues UI events to the fetcher and WASM analyzer. Orchestrates the flow: parse URL → resolve ref → fetch tree → download blobs → call WASM → render results.
   - `ui.js` — DOM manipulation helpers (show/hide sections, update progress bar, render error messages).

4. **Results rendering** — Two display modes:
   - **HTML report mode** (default): The WASM `analyzeCode` function returns the HTML report string generated by `internal/reporter/html.go` (which already includes Chart.js and tabbed navigation). Inject this into an `<iframe srcdoc="...">` or a shadow DOM container to isolate its styles.
   - **JSON mode**: The WASM function returns raw JSON. Render a collapsible JSON tree viewer and offer a "Download JSON" button.
   - Add interactive enhancements on top of the HTML report:
     - **Summary cards** at the top showing key metrics (total files, functions, structs, average complexity, duplication ratio).
     - **Sortable tables** for function/struct listings (use a lightweight vanilla JS sort on `<th>` click).
     - **Collapsible sections** for each analysis category (functions, structs, packages, patterns, etc.).

5. **Styling** — Create `web/css/style.css`:
   - Responsive design (flexbox/grid) that works on desktop and tablet.
   - Dark/light mode toggle (using `prefers-color-scheme` media query with a manual override).
   - Minimal, professional appearance. No heavy CSS framework; use CSS custom properties for theming.

6. **Error handling** — Display user-friendly messages for:
   - Invalid repository URL format.
   - Repository not found (404).
   - Rate limit exceeded (403 with `X-RateLimit-Remaining: 0`).
   - WASM load failure.
   - Analysis errors (parse failures in individual files should be reported but not block the overall analysis).

7. **Accessibility** — Ensure keyboard navigation, ARIA labels on interactive elements, and sufficient color contrast.

### Dependencies

- `wasm_exec.js` (shipped with Go, copied during build).
- Chart.js v4.4.0 (already used in the HTML report template; loaded via CDN in the report HTML).
- No npm dependencies.

### Open Questions / Risks

- **WASM load time**: The initial WASM instantiation can take 2–5 seconds on slower connections. Show a loading spinner during instantiation. Consider using `WebAssembly.instantiateStreaming` for faster startup.
- **Memory usage**: Parsing a large repository in-browser may consume significant memory. Monitor with `performance.memory` (Chrome) and warn users if the repository is very large (> 2,000 files).
- **HTML report isolation**: The existing HTML report template is a full page with its own `<html>`, `<head>`, `<body>`. Rendering inside the host page requires either an `<iframe>` (easiest, provides full style isolation) or stripping the outer document structure and injecting into a `<div>`.

---

## Phase 4: GitHub Pages Deployment

### Goal

Create a GitHub Actions workflow that compiles the WASM binary, assembles the static site, and deploys it to GitHub Pages automatically on every push to `main`.

### Steps

1. **Define the site output directory** — Use `web/` as the source directory for static assets during development. The CI workflow assembles the final deployable site into `dist/`:
   ```
   dist/
   ├── index.html
   ├── css/
   │   └── style.css
   ├── js/
   │   ├── wasm_exec.js
   │   ├── wasm-loader.js
   │   ├── github-fetcher.js
   │   ├── app.js
   │   └── ui.js
   └── wasm/
       └── go-stats-generator.<content-hash>.wasm
   ```

2. **Content-hash the WASM binary** — In the build script, compute the SHA-256 hash of the `.wasm` file and rename it to `go-stats-generator.<first-8-chars-of-hash>.wasm`. Update a generated `wasm-manifest.json` file that the JavaScript loader reads to discover the current filename. This ensures browsers cache the WASM blob aggressively and only re-download when it changes.

3. **Create the GitHub Actions workflow** — Add `.github/workflows/deploy-pages.yml`:
   ```yaml
   name: Deploy to GitHub Pages

   on:
     push:
       branches: [main]
     workflow_dispatch:

   permissions:
     contents: read
     pages: write
     id-token: write

   concurrency:
     group: pages
     cancel-in-progress: true

   jobs:
     build-and-deploy:
       runs-on: ubuntu-latest
       environment:
         name: github-pages
         url: ${{ steps.deployment.outputs.page_url }}
       steps:
         - uses: actions/checkout@v4

         - name: Set up Go
           uses: actions/setup-go@v5
           with:
             go-version: '1.23'

         - name: Build WASM binary
           run: |
             mkdir -p dist/wasm
             GOOS=js GOARCH=wasm go build -ldflags "-s -w" -o dist/wasm/go-stats-generator.wasm ./cmd/wasm/
             HASH=$(sha256sum dist/wasm/go-stats-generator.wasm | head -c 8)
             HASHED_NAME="go-stats-generator.${HASH}.wasm"
             mv dist/wasm/go-stats-generator.wasm "dist/wasm/${HASHED_NAME}"
             echo "{\"wasmFile\": \"${HASHED_NAME}\"}" > dist/wasm/wasm-manifest.json

         - name: Copy static assets
           run: |
             cp web/index.html dist/
             cp -r web/css dist/
             cp -r web/js dist/
             cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" dist/js/

         - name: Optimize WASM (optional)
           run: |
             WASM_FILE=$(cat dist/wasm/wasm-manifest.json | grep -o '"go-stats-generator\.[^"]*\.wasm"' | tr -d '"')
             if command -v wasm-opt &> /dev/null; then
               wasm-opt -Oz "dist/wasm/${WASM_FILE}" -o dist/wasm/optimized.wasm
               mv dist/wasm/optimized.wasm "dist/wasm/${WASM_FILE}"
             fi
           continue-on-error: true

         - name: Upload artifact
           uses: actions/upload-pages-artifact@v3
           with:
             path: dist

         - name: Deploy to GitHub Pages
           id: deployment
           uses: actions/deploy-pages@v4
   ```

4. **Enable GitHub Pages** — In repository Settings → Pages, select "GitHub Actions" as the source (not branch-based). The workflow handles deployment via the `actions/deploy-pages` action.

5. **Add cache headers via `_headers` file** — Create `web/_headers` (copied to `dist/` during build):
   ```
   /wasm/*
     Cache-Control: public, max-age=31536000, immutable

   /js/wasm_exec.js
     Cache-Control: public, max-age=86400

   /*.html
     Cache-Control: public, max-age=300
   ```
   Note: GitHub Pages has limited support for custom headers. If needed, the WASM manifest approach (content-hashed filenames) provides equivalent cache-busting without server-side header configuration.

6. **Add a `404.html`** — GitHub Pages serves this for unknown routes. Redirect to `index.html` for SPA-like behavior (though the app is single-page, this handles direct links).

### Dependencies

- GitHub Actions runners with Go 1.23+.
- `actions/checkout@v4`, `actions/setup-go@v5`, `actions/upload-pages-artifact@v3`, `actions/deploy-pages@v4`.

### Open Questions / Risks

- **Binaryen `wasm-opt`**: Not pre-installed on GitHub Actions runners. Either install it in the workflow (`apt-get install binaryen` or download a release binary) or skip optimization. The `continue-on-error: true` makes it optional.
- **GitHub Pages size limits**: GitHub Pages has a soft limit of 1 GB per site. A single WASM binary (even unoptimized) is well within this limit.
- **Custom domain**: If the project later uses a custom domain, update the `<base>` tag in `index.html` and CNAME file accordingly.

---

## Phase 5: Constraints and Limitations

### Goal

Document analysis features that cannot work in WASM and define alternatives or graceful degradations.

### Steps

1. **Concurrent worker pools** — The `internal/scanner/worker.go` worker pool uses goroutines and channels for parallel file processing. In the WASM build:
   - *Limitation*: `GOOS=js` runs all goroutines on a single OS thread. True parallelism is unavailable.
   - *Alternative*: The WASM worker shim (Phase 1, Step 4) processes files sequentially. Performance impact is mitigated because parsing is CPU-bound and single-threaded WASM is not significantly slower than single-threaded native Go for AST parsing.
   - *Future improvement*: Use Web Workers to run multiple WASM instances in parallel, each processing a subset of files, then merge results.

2. **Filesystem scanning** — The `internal/scanner/discover.go` discoverer uses `filepath.Walk` and `os.Stat`.
   - *Limitation*: No filesystem access in the browser.
   - *Alternative*: The WASM entry point receives an in-memory file list from JavaScript (Phase 1, Step 3). The `DiscoverFilesFromMemory` function constructs `FileInfo` structs from the provided data.

3. **SQLite / PostgreSQL / MongoDB storage** — The `internal/storage` package uses `modernc.org/sqlite`, `github.com/lib/pq`, and `go.mongodb.org/mongo-driver`.
   - *Limitation*: None of these compile to or are usable in `GOOS=js`.
   - *Alternative*: Exclude all storage backends from the WASM build via build tags. The browser application performs one-shot analysis with no historical storage. If baseline/diff features are desired in the future, use `IndexedDB` via `syscall/js` or store snapshots in `localStorage` as JSON.

4. **Cobra CLI and Viper configuration** — The `cmd/` package uses `github.com/spf13/cobra` and `github.com/spf13/viper`.
   - *Limitation*: CLI argument parsing is irrelevant in the browser.
   - *Alternative*: The WASM entry point (`cmd/wasm/main.go`) bypasses Cobra entirely and directly uses `pkg/go-stats-generator` (the public API) and `internal/analyzer` packages.

5. **HTTP API server** — The `internal/api` package provides a REST API via `net/http`.
   - *Limitation*: Cannot bind to a network port in the browser.
   - *Alternative*: Excluded from the WASM build. Not needed for client-side analysis.

6. **File watching (`fsnotify`)** — The `cmd/watch.go` command uses `github.com/fsnotify/fsnotify`.
   - *Limitation*: No filesystem events in the browser.
   - *Alternative*: Excluded from the WASM build. Users re-trigger analysis manually.

7. **Multi-repository analysis** — The `internal/multirepo` package analyzes multiple repositories.
   - *Limitation*: Fetching multiple repositories sequentially is slow and rate-limit-intensive.
   - *Alternative*: Initial WASM deployment supports single-repository analysis only. Multi-repo can be added later with a queue UI.

8. **Report format limitations** — Console output (`internal/reporter/console.go`) uses terminal-specific formatting.
   - *Limitation*: No terminal in the browser.
   - *Alternative*: The WASM build supports HTML and JSON output only. Console, CSV, and Markdown reporters are available but less useful in the browser context. The UI defaults to HTML.

9. **Large repository performance** — The CLI tool is designed for 50,000+ files within 60 seconds using concurrent workers.
   - *Limitation*: Single-threaded WASM processing of 50,000 files could take 10+ minutes.
   - *Alternative*: Display a warning for repositories exceeding a configurable file count threshold (default: 5,000 files). Allow the user to proceed but set expectations. Show per-file progress.

10. **Naming convention** — The tool must always be referred to as `go-stats-generator` in all UI text, documentation, error messages, and code comments. Never use the abbreviation `gostats` or any other shortened form.

### Dependencies

None (this phase is documentation and planning).

### Open Questions / Risks

- **Go standard library WASM stability**: The `go/parser` and `go/ast` packages are stable, but edge cases in WASM execution (e.g., `go/types` if later needed) should be tested.
- **Browser compatibility**: The application targets modern evergreen browsers (Chrome 90+, Firefox 90+, Safari 15+, Edge 90+). Older browsers without `WebAssembly.instantiateStreaming` support fall back to `WebAssembly.instantiate`.

---

## Success Criteria for Minimum Viable Deployment

A developer can consider the deployment minimally viable when **all** of the following are true:

1. **WASM binary compiles** — `GOOS=js GOARCH=wasm go build ./cmd/wasm/` succeeds without errors, producing a functional `.wasm` file.

2. **End-to-end analysis works** — A user can visit the GitHub Pages URL, enter a public GitHub repository URL (e.g., `https://github.com/golang/example`), click "Analyze", and receive a rendered HTML report showing function metrics, struct analysis, package dependencies, complexity scores, and pattern detection — all computed client-side.

3. **No server-side processing** — After the GitHub Pages deployment, the site consists entirely of static files (HTML, CSS, JS, WASM). All repository fetching and analysis happens in the browser.

4. **GitHub Pages CI/CD works** — Pushing to `main` triggers the GitHub Actions workflow, which compiles the WASM binary, assembles the site, and deploys to GitHub Pages without manual intervention.

5. **Rate limiting is handled** — The UI clearly communicates GitHub API rate limit status, prompts for an optional token when limits are reached, and gracefully handles `403` responses.

6. **Content-hashed WASM caching** — The WASM binary filename includes a content hash, ensuring returning visitors get cached versions and only download new binaries when the analysis engine changes.

7. **Progress feedback** — The UI shows meaningful progress during repository fetching (file count) and analysis (current phase), preventing the user from thinking the page is frozen.

8. **Error resilience** — Parse errors in individual files are reported in the results but do not abort the entire analysis. Network errors during fetching are retried once and then reported clearly.

9. **Correct naming** — All user-visible text refers to the tool as `go-stats-generator`, never `gostats` or any other abbreviated form.

10. **Tested with representative repositories** — The deployment has been manually verified against at least three public Go repositories of varying sizes (small: < 20 files, medium: 50–200 files, large: 500+ files).
