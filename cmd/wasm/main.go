//go:build js && wasm

// Package main provides WebAssembly entry point for go-stats-generator
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"syscall/js"

	"github.com/opd-ai/go-stats-generator/internal/config"
	"github.com/opd-ai/go-stats-generator/internal/metrics"
	"github.com/opd-ai/go-stats-generator/internal/reporter"
	"github.com/opd-ai/go-stats-generator/pkg/generator"
)

// FileInput represents a single file from the browser
type FileInput struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

// AnalysisRequest represents the input from JavaScript
type AnalysisRequest struct {
	Files        []FileInput  `json:"files"`
	OutputFormat string       `json:"outputFormat"` // "json" or "html"
	Config       *ConfigInput `json:"config,omitempty"`
}

// ConfigInput allows browser to customize analysis configuration
type ConfigInput struct {
	MaxFunctionLength        int     `json:"maxFunctionLength,omitempty"`
	MaxCyclomaticComplexity  int     `json:"maxCyclomaticComplexity,omitempty"`
	MinDocumentationCoverage float64 `json:"minDocumentationCoverage,omitempty"`
	SkipTestFiles            bool    `json:"skipTestFiles,omitempty"`
}

// AnalysisResponse represents the output to JavaScript
type AnalysisResponse struct {
	Success bool   `json:"success"`
	Data    string `json:"data"` // JSON string or HTML string
	Error   string `json:"error,omitempty"`
}

// analyzeCodeWrapper handles JavaScript calls to analyzeCode
func analyzeCodeWrapper() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) < 1 {
			return errorResponse("missing filesJSON argument")
		}

		// Parse input JSON
		inputJSON := args[0].String()
		var request AnalysisRequest
		if err := json.Unmarshal([]byte(inputJSON), &request); err != nil {
			return errorResponse(fmt.Sprintf("failed to parse input: %v", err))
		}

		// Build configuration and create analyzer
		cfg := buildConfig(request.Config)
		analyzer := generator.NewAnalyzerWithConfig(cfg)

		// Analyze files from memory
		report, err := analyzeFilesFromMemory(analyzer, request.Files)
		if err != nil {
			return errorResponse(fmt.Sprintf("analysis failed: %v", err))
		}

		// Generate output in requested format
		outputData, err := generateOutput(report, request.OutputFormat)
		if err != nil {
			return errorResponse(err.Error())
		}

		return successResponse(outputData)
	})
}

// errorResponse creates a standardized error response
func errorResponse(msg string) map[string]interface{} {
	return map[string]interface{}{
		"success": false,
		"error":   msg,
	}
}

// successResponse creates a standardized success response
func successResponse(data string) map[string]interface{} {
	return map[string]interface{}{
		"success": true,
		"data":    data,
	}
}

// generateOutput produces formatted output based on the requested format
func generateOutput(report *metrics.Report, format string) (string, error) {
	switch format {
	case "html":
		return generateHTMLOutput(report)
	default: // JSON by default
		return generateJSONOutput(report)
	}
}

// generateHTMLOutput creates HTML report output
func generateHTMLOutput(report *metrics.Report) (string, error) {
	htmlReporter := reporter.NewHTMLReporterWithConfig(nil)
	var buf bytes.Buffer
	err := htmlReporter.Generate(report, &buf)
	if err != nil {
		return "", fmt.Errorf("HTML generation failed: %v", err)
	}
	return buf.String(), nil
}

// generateJSONOutput creates JSON report output
func generateJSONOutput(report *metrics.Report) (string, error) {
	jsonBytes, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return "", fmt.Errorf("JSON serialization failed: %v", err)
	}
	return string(jsonBytes), nil
}

// buildConfig creates a Config from browser input
func buildConfig(input *ConfigInput) *config.Config {
	cfg := config.DefaultConfig()

	if input != nil {
		if input.MaxFunctionLength > 0 {
			cfg.Analysis.MaxFunctionLength = input.MaxFunctionLength
		}
		if input.MaxCyclomaticComplexity > 0 {
			cfg.Analysis.MaxCyclomaticComplexity = input.MaxCyclomaticComplexity
		}
		if input.MinDocumentationCoverage > 0 {
			cfg.Analysis.MinDocumentationCoverage = input.MinDocumentationCoverage
		}
		if input.SkipTestFiles {
			cfg.Filters.SkipTestFiles = true
		}
	}

	return cfg
}

// analyzeFilesFromMemory performs analysis on in-memory files (WASM path)
func analyzeFilesFromMemory(analyzer *generator.Analyzer, files []FileInput) (*metrics.Report, error) {
	memFiles := convertToMemoryFiles(files)
	return analyzer.AnalyzeMemoryFiles(context.Background(), memFiles, "/")
}

// convertToMemoryFiles converts FileInput to MemoryFile
func convertToMemoryFiles(files []FileInput) []generator.MemoryFile {
	result := make([]generator.MemoryFile, len(files))
	for i, f := range files {
		result[i] = generator.MemoryFile{
			Path:    f.Path,
			Content: f.Content,
		}
	}
	return result
}

func main() {
	// Register the analyzeCode function on the global JavaScript object
	js.Global().Set("analyzeCode", analyzeCodeWrapper())

	// Log to browser console
	console := js.Global().Get("console")
	console.Call("log", "go-stats-generator WASM module loaded successfully")

	// Block forever to keep the Go runtime alive
	select {}
}
