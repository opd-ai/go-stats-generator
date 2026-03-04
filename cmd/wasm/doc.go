// Package main provides WebAssembly entry point for go-stats-generator.
//
// This package enables go-stats-generator to run entirely in a web browser
// by compiling to WebAssembly. It exposes a JavaScript API that accepts
// in-memory file contents and returns analysis results without requiring
// any server-side processing.
//
// JavaScript API:
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
// Note: This implementation requires the WASM-compatible scanner shim
// (steps 3-4 in PLAN.md) to be completed before it can analyze files.
// Currently, it returns an error indicating the dependency.
package main
