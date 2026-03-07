// Package main provides WebAssembly entry point for go-stats-generator.
//
// This package enables go-stats-generator to run entirely in a web browser
// by compiling to WebAssembly. It exposes a JavaScript API that accepts
// in-memory file contents and returns analysis results without requiring
// any server-side processing.
//
// Two analysis modes are supported:
//
// 1. Git Clone (recommended): Clones the repository over HTTPS using go-git
// into browser memory. This avoids GitHub API rate limits entirely.
//
//	const result = await cloneAndAnalyze(JSON.stringify({
//	    url: "https://github.com/owner/repo",
//	    ref: "main",
//	    includeTests: false,
//	    outputFormat: "json",
//	    config: {
//	        maxFunctionLength: 30,
//	        maxCyclomaticComplexity: 10,
//	        minDocumentationCoverage: 0.7,
//	        skipTestFiles: true
//	    }
//	}), (progress) => console.log(progress.message));
//
//	console.log(result.data);   // JSON string or HTML string
//	console.log(result.stats);  // { totalFiles, totalSize, owner, repo, ref, method }
//
// 2. Pre-fetched Files: Accepts files already fetched by JavaScript (e.g. via
// the GitHub REST API). This serves as a fallback when git clone fails.
//
//	const result = analyzeCode(JSON.stringify({
//	    files: [
//	        {path: "main.go", content: "package main\n..."},
//	        {path: "util.go", content: "package main\n..."}
//	    ],
//	    outputFormat: "json",  // or "html"
//	    config: {
//	        maxFunctionLength: 30,
//	        maxCyclomaticComplexity: 10,
//	        minDocumentationCoverage: 0.7,
//	        skipTestFiles: true
//	    }
//	}));
//
//	if (result.success) {
//	    console.log(result.data);  // JSON string or HTML string
//	} else {
//	    console.error(result.error);
//	}
//
// Build instructions:
//
//	GOOS=js GOARCH=wasm go build -o go-stats-generator.wasm ./cmd/wasm/
//	cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" .
//
// The WASM build supports all core analyzers (functions, structs, interfaces,
// packages, patterns, concurrency, duplication, naming, documentation) and
// produces identical results to the CLI build.
package main
