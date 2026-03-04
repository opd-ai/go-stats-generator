// Package api provides REST API types and structures for go-stats-generator.
package api

import (
	"github.com/opd-ai/go-stats-generator/internal/api/storage"
	"github.com/opd-ai/go-stats-generator/internal/metrics"
)

// AnalysisResult is an alias for storage.AnalysisResult for backward compatibility.
type AnalysisResult = storage.AnalysisResult

// AnalyzeRequest represents a request to analyze a repository.
type AnalyzeRequest struct {
	Path              string   `json:"path"`
	Include           []string `json:"include,omitempty"`
	Exclude           []string `json:"exclude,omitempty"`
	SkipTests         bool     `json:"skip_tests,omitempty"`
	MaxFunctionLength int      `json:"max_function_length,omitempty"`
	MaxComplexity     int      `json:"max_complexity,omitempty"`
}

// AnalyzeResponse represents the response from an analysis request.
type AnalyzeResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"` // "pending", "running", "completed", "failed"
}

// ReportResponse represents a completed analysis report.
type ReportResponse struct {
	ID     string          `json:"id"`
	Status string          `json:"status"`
	Report *metrics.Report `json:"report,omitempty"`
	Error  string          `json:"error,omitempty"`
}

// HealthResponse represents the health status of the API.
type HealthResponse struct {
	Status  string `json:"status"`
	Version string `json:"version"`
}
