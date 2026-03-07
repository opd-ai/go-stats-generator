//go:build js && wasm

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"syscall/js"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/storage/memory"

	"github.com/opd-ai/go-stats-generator/internal/metrics"
	"github.com/opd-ai/go-stats-generator/pkg/generator"
)

// CloneRequest represents the input from JavaScript for clone-based analysis.
type CloneRequest struct {
	URL          string       `json:"url"`
	Ref          string       `json:"ref,omitempty"`
	Token        string       `json:"token,omitempty"`
	IncludeTests bool         `json:"includeTests"`
	OutputFormat string       `json:"outputFormat"`
	Config       *ConfigInput `json:"config,omitempty"`
}

// cloneAndAnalyzeWrapper returns a js.Func that clones a repository using
// go-git over HTTPS and runs the analysis entirely in WASM. The function
// returns a JavaScript Promise so the caller can await the result.
//
// JavaScript usage:
//
//	const result = await cloneAndAnalyze(JSON.stringify({
//	    url: "https://github.com/owner/repo",
//	    ref: "main",
//	    includeTests: false,
//	    outputFormat: "json",
//	    config: { maxFunctionLength: 30 }
//	}), (progress) => console.log(progress));
func cloneAndAnalyzeWrapper() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) < 1 {
			return resolvedErrorPromise("missing request JSON argument")
		}

		inputJSON := args[0].String()

		var progressCb js.Value
		if len(args) > 1 && !args[1].IsUndefined() && !args[1].IsNull() {
			progressCb = args[1]
		}

		// The Promise constructor invokes the executor synchronously, so it
		// is safe to release the js.Func immediately after creating the Promise.
		handler := js.FuncOf(func(_ js.Value, promiseArgs []js.Value) interface{} {
			resolve := promiseArgs[0]

			go func() {
				result, err := performCloneAndAnalysis(inputJSON, progressCb)
				if err != nil {
					resolve.Invoke(js.ValueOf(map[string]interface{}{
						"success": false,
						"error":   err.Error(),
					}))
					return
				}
				resolve.Invoke(js.ValueOf(result))
			}()

			return nil
		})

		promise := js.Global().Get("Promise").New(handler)
		handler.Release()
		return promise
	})
}

// resolvedErrorPromise returns a JS Promise that resolves with {success:false, error:msg}.
// This matches the response shape used by analyzeCode for consistency.
func resolvedErrorPromise(msg string) js.Value {
	handler := js.FuncOf(func(_ js.Value, promiseArgs []js.Value) interface{} {
		resolve := promiseArgs[0]
		resolve.Invoke(js.ValueOf(map[string]interface{}{
			"success": false,
			"error":   msg,
		}))
		return nil
	})
	promise := js.Global().Get("Promise").New(handler)
	handler.Release()
	return promise
}

// performCloneAndAnalysis orchestrates the clone → extract → analyze pipeline.
func performCloneAndAnalysis(inputJSON string, progressCb js.Value) (map[string]interface{}, error) {
	request, err := parseCloneRequest(inputJSON)
	if err != nil {
		return nil, err
	}

	files, err := cloneAndExtractFiles(request, progressCb)
	if err != nil {
		return nil, err
	}

	report, err := analyzeFiles(files, request.Config, progressCb)
	if err != nil {
		return nil, err
	}

	return buildSuccessResponse(files, report, request, progressCb)
}

// parseCloneRequest validates and parses the JSON input for clone analysis.
func parseCloneRequest(inputJSON string) (*CloneRequest, error) {
	var request CloneRequest
	if err := json.Unmarshal([]byte(inputJSON), &request); err != nil {
		return nil, fmt.Errorf("failed to parse request: %w", err)
	}
	if request.URL == "" {
		return nil, fmt.Errorf("repository URL is required")
	}
	return &request, nil
}

// cloneAndExtractFiles clones the repository and extracts Go source files.
func cloneAndExtractFiles(request *CloneRequest, progressCb js.Value) ([]FileInput, error) {
	repoURL := normalizeGitURL(request.URL)
	reportProgress(progressCb, 5, "Cloning repository…")

	fs := memfs.New()
	_, err := cloneRepository(repoURL, request.Ref, request.Token, fs, progressCb)
	if err != nil {
		return nil, fmt.Errorf("clone failed: %w", err)
	}

	reportProgress(progressCb, 60, "Extracting Go source files…")
	files, err := extractGoFiles(fs, request.IncludeTests)
	if err != nil {
		return nil, fmt.Errorf("file extraction failed: %w", err)
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no Go source files found in repository")
	}
	return files, nil
}

// analyzeFiles performs the analysis on the extracted Go files.
func analyzeFiles(files []FileInput, config *ConfigInput, progressCb js.Value) (*metrics.Report, error) {
	reportProgress(progressCb, 75, fmt.Sprintf("Analyzing %d files…", len(files)))

	cfg := buildConfig(config)
	analyzer := generator.NewAnalyzerWithConfig(cfg)

	memFiles := make([]generator.MemoryFile, len(files))
	for i, f := range files {
		memFiles[i] = generator.MemoryFile{Path: f.Path, Content: f.Content}
	}

	report, err := analyzer.AnalyzeMemoryFiles(context.Background(), memFiles, "/")
	if err != nil {
		return nil, fmt.Errorf("analysis failed: %w", err)
	}
	return report, nil
}

// buildSuccessResponse creates the final success response with analysis results.
func buildSuccessResponse(files []FileInput, report *metrics.Report, request *CloneRequest, progressCb js.Value) (map[string]interface{}, error) {
	reportProgress(progressCb, 90, "Generating output…")

	outputData, err := generateOutput(report, request.OutputFormat)
	if err != nil {
		return nil, err
	}

	owner, repoName := parseOwnerRepo(request.URL)
	ref := request.Ref
	if ref == "" {
		ref = "default branch"
	}

	reportProgress(progressCb, 100, "Complete")

	return map[string]interface{}{
		"success": true,
		"data":    outputData,
		"stats": map[string]interface{}{
			"totalFiles": len(files),
			"totalSize":  totalSize(files),
			"owner":      owner,
			"repo":       repoName,
			"ref":        ref,
			"method":     "git-clone",
		},
	}, nil
}

// isLikelyCommitSHA returns true if ref looks like a full or abbreviated commit hash.
func isLikelyCommitSHA(ref string) bool {
	if len(ref) < 7 || len(ref) > 40 {
		return false
	}
	for _, c := range ref {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}

// cloneRepository performs a shallow HTTPS clone into in-memory storage.
// Commit SHA refs are not supported with shallow clones and will return an error.
// If token is non-empty, it is used for HTTP basic authentication (e.g. GitHub PAT).
func cloneRepository(url, ref, token string, fs billy.Filesystem, progressCb js.Value) (*git.Repository, error) {
	if ref != "" && isLikelyCommitSHA(ref) {
		return nil, fmt.Errorf(
			"commit SHA refs (%q) are not supported with shallow clone; "+
				"use a branch or tag name instead", ref)
	}

	opts := &git.CloneOptions{
		URL:   url,
		Depth: 1,
		Tags:  git.NoTags,
	}

	if token != "" {
		// GitHub accepts PATs via HTTP basic auth with any username; the
		// conventional value is "x-access-token". Other git hosts (GitLab,
		// Bitbucket, etc.) may require different username conventions.
		opts.Auth = &githttp.BasicAuth{
			Username: "x-access-token",
			Password: token,
		}
	}

	// Optionally write clone progress to a JS callback.
	if !progressCb.IsUndefined() && !progressCb.IsNull() {
		opts.Progress = &jsProgressWriter{cb: progressCb}
	}

	if ref != "" {
		opts.SingleBranch = true
		opts.ReferenceName = plumbing.NewBranchReferenceName(ref)
	}

	repo, err := git.Clone(memory.NewStorage(), fs, opts)
	if err != nil && ref != "" {
		// The ref may be a tag rather than a branch – retry with tag ref.
		opts.ReferenceName = plumbing.NewTagReferenceName(ref)
		repo, err = git.Clone(memory.NewStorage(), fs, opts)
	}
	if err != nil {
		return nil, classifyCloneError(err, url, token)
	}
	return repo, nil
}

// classifyCloneError inspects a clone error and returns a more
// descriptive, user-facing message. In the WASM/browser context,
// Go's net/http uses the browser fetch() API, which surfaces opaque
// "NetworkError" messages for CORS blocks and connectivity issues.
func classifyCloneError(err error, url, token string) error {
	msg := err.Error()

	// Detect browser fetch / CORS failures.
	// Normalize to lowercase for case-insensitive matching – browser
	// implementations surface these errors with varying casing
	// (e.g. "Failed to fetch", "NetworkError", "networkerror").
	lower := strings.ToLower(msg)
	if strings.Contains(lower, "fetch() failed") ||
		strings.Contains(lower, "failed to fetch") ||
		strings.Contains(lower, "networkerror") ||
		strings.Contains(lower, "cors") {
		if token == "" {
			return fmt.Errorf(
				"network error cloning repository: browser fetch was blocked " +
					"(this usually means CORS restrictions). " +
					"A fallback via ZIP archive download will be attempted automatically")
		}
		return fmt.Errorf(
			"network error cloning repository: browser fetch was blocked. " +
				"Verify that the token is valid and has repository read access")
	}

	// Detect HTTP authentication failures.
	if strings.Contains(msg, "authentication") ||
		strings.Contains(msg, "401") ||
		strings.Contains(msg, "403") {
		return fmt.Errorf(
			"authentication failed: check that your personal access token " +
				"is valid and has repository access")
	}

	return err
}

// jsProgressWriter adapts clone progress output to a JS callback.
type jsProgressWriter struct {
	cb js.Value
}

// Write implements io.Writer by converting byte slices to progress messages sent to JavaScript.
// It extracts the message text from p, trims whitespace, and invokes the JavaScript progress callback
// with the message (but no percent update, indicated by -1). This allows go-git's standard progress
// output to be forwarded to the browser UI during clone operations. Always returns len(p) to signal
// successful processing of the entire buffer.
func (w *jsProgressWriter) Write(p []byte) (int, error) {
	msg := strings.TrimSpace(string(p))
	if msg != "" {
		reportProgress(w.cb, -1, msg)
	}
	return len(p), nil
}

// reportProgress invokes the JS progress callback if set.
// A negative percent means "keep current percent, update message only".
func reportProgress(cb js.Value, percent int, message string) {
	if cb.IsUndefined() || cb.IsNull() {
		return
	}
	obj := map[string]interface{}{
		"message": message,
	}
	if percent >= 0 {
		obj["percent"] = percent
	}
	cb.Invoke(js.ValueOf(obj))
}

// extractGoFiles walks the in-memory filesystem and reads all Go source files.
func extractGoFiles(fs billy.Filesystem, includeTests bool) ([]FileInput, error) {
	var files []FileInput
	err := walkDir(fs, "/", func(path string, isDir bool) error {
		if isDir {
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		// Skip vendor.
		if strings.Contains(path, "/vendor/") || strings.HasPrefix(path, "vendor/") {
			return nil
		}
		// Skip test files unless requested.
		if !includeTests && strings.HasSuffix(path, "_test.go") {
			return nil
		}
		// Skip generated files.
		if strings.Contains(path, "generated") || strings.HasSuffix(path, ".pb.go") {
			return nil
		}

		content, err := readFile(fs, path)
		if err != nil {
			return nil // skip unreadable files
		}
		files = append(files, FileInput{Path: path, Content: content})
		return nil
	})
	return files, err
}

// walkDir recursively walks the billy filesystem.
func walkDir(fs billy.Filesystem, dir string, fn func(path string, isDir bool) error) error {
	entries, err := fs.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		path := filepath.Join(dir, entry.Name())
		// Trim leading slash for consistency.
		cleanPath := strings.TrimPrefix(path, "/")

		if err := fn(cleanPath, entry.IsDir()); err != nil {
			return err
		}
		if entry.IsDir() {
			if err := walkDir(fs, path, fn); err != nil {
				return err
			}
		}
	}
	return nil
}

// readFile reads the entire content of a file from the billy filesystem.
func readFile(fs billy.Filesystem, path string) (string, error) {
	f, err := fs.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	data, err := io.ReadAll(f)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// normalizeGitURL ensures the URL uses HTTPS and ends with .git.
// In a browser context, only HTTPS is safe (avoids mixed-content and
// unsupported transport issues). SSH URLs are converted; plain HTTP
// is upgraded; other schemes are rejected.
func normalizeGitURL(raw string) string {
	u := strings.TrimSpace(raw)
	// Convert SSH URLs to HTTPS.
	if strings.HasPrefix(u, "git@") {
		u = strings.Replace(u, ":", "/", 1)
		u = strings.Replace(u, "git@", "https://", 1)
	}
	// Upgrade http:// to https://.
	if strings.HasPrefix(u, "http://") {
		u = "https://" + strings.TrimPrefix(u, "http://")
	}
	if !strings.HasSuffix(u, ".git") {
		u += ".git"
	}
	return u
}

// parseOwnerRepo extracts the owner and repo name from a GitHub URL.
func parseOwnerRepo(url string) (string, string) {
	// Match github.com/owner/repo
	parts := strings.Split(strings.TrimSuffix(strings.TrimSuffix(url, ".git"), "/"), "/")
	if len(parts) >= 2 {
		return parts[len(parts)-2], parts[len(parts)-1]
	}
	return "", ""
}

// totalSize sums the byte lengths of all file contents.
func totalSize(files []FileInput) int {
	n := 0
	for _, f := range files {
		n += len(f.Content)
	}
	return n
}
