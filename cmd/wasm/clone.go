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
	"github.com/go-git/go-git/v5/storage/memory"

	"github.com/opd-ai/go-stats-generator/pkg/generator"
)

// CloneRequest represents the input from JavaScript for clone-based analysis.
type CloneRequest struct {
	URL          string       `json:"url"`
	Ref          string       `json:"ref,omitempty"`
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
			return rejectedPromise("missing request JSON argument")
		}

		inputJSON := args[0].String()

		var progressCb js.Value
		if len(args) > 1 && !args[1].IsUndefined() && !args[1].IsNull() {
			progressCb = args[1]
		}

		// Return a Promise that resolves with the analysis result.
		handler := js.FuncOf(func(_ js.Value, promiseArgs []js.Value) interface{} {
			resolve := promiseArgs[0]
			reject := promiseArgs[1]

			go func() {
				result, err := performCloneAndAnalysis(inputJSON, progressCb)
				if err != nil {
					reject.Invoke(js.Global().Get("Error").New(err.Error()))
					return
				}
				resolve.Invoke(js.ValueOf(result))
			}()

			return nil
		})

		return js.Global().Get("Promise").New(handler)
	})
}

// rejectedPromise returns a JS Promise that immediately rejects with msg.
func rejectedPromise(msg string) js.Value {
	handler := js.FuncOf(func(_ js.Value, promiseArgs []js.Value) interface{} {
		reject := promiseArgs[1]
		reject.Invoke(js.Global().Get("Error").New(msg))
		return nil
	})
	promise := js.Global().Get("Promise").New(handler)
	handler.Release()
	return promise
}

// performCloneAndAnalysis orchestrates the clone → extract → analyze pipeline.
func performCloneAndAnalysis(inputJSON string, progressCb js.Value) (map[string]interface{}, error) {
	var request CloneRequest
	if err := json.Unmarshal([]byte(inputJSON), &request); err != nil {
		return nil, fmt.Errorf("failed to parse request: %w", err)
	}

	if request.URL == "" {
		return nil, fmt.Errorf("repository URL is required")
	}

	// Ensure the URL ends with .git for the smart HTTP protocol.
	repoURL := normalizeGitURL(request.URL)

	reportProgress(progressCb, 5, "Cloning repository…")

	fs := memfs.New()
	_, err := cloneRepository(repoURL, request.Ref, fs, progressCb)
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

	reportProgress(progressCb, 75, fmt.Sprintf("Analyzing %d files…", len(files)))

	cfg := buildConfig(request.Config)
	analyzer := generator.NewAnalyzerWithConfig(cfg)

	memFiles := make([]generator.MemoryFile, len(files))
	for i, f := range files {
		memFiles[i] = generator.MemoryFile{Path: f.Path, Content: f.Content}
	}

	report, err := analyzer.AnalyzeMemoryFiles(context.Background(), memFiles, "/")
	if err != nil {
		return nil, fmt.Errorf("analysis failed: %w", err)
	}

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

// cloneRepository performs a shallow HTTPS clone into in-memory storage.
func cloneRepository(url, ref string, fs billy.Filesystem, progressCb js.Value) (*git.Repository, error) {
	opts := &git.CloneOptions{
		URL:   url,
		Depth: 1,
		Tags:  git.NoTags,
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
	if err != nil && ref != "" {
		// Fall back to clone without specific ref name (e.g. commit SHA).
		opts.ReferenceName = ""
		opts.SingleBranch = false
		repo, err = git.Clone(memory.NewStorage(), fs, opts)
	}
	return repo, err
}

// jsProgressWriter adapts clone progress output to a JS callback.
type jsProgressWriter struct {
	cb js.Value
}

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
func normalizeGitURL(raw string) string {
	u := strings.TrimSpace(raw)
	// Convert SSH URLs to HTTPS.
	if strings.HasPrefix(u, "git@") {
		u = strings.Replace(u, ":", "/", 1)
		u = strings.Replace(u, "git@", "https://", 1)
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
