package storage

import "github.com/opd-ai/go-stats-generator/internal/metrics"

// AnalysisResult stores the result of an analysis operation.
type AnalysisResult struct {
	ID     string
	Status string // "pending", "running", "completed", "failed"
	Report *metrics.Report
	Error  error
}

// ResultStore defines the interface for storing and retrieving analysis results.
// Implementations must be thread-safe.
type ResultStore interface {
	// Store saves an analysis result.
	Store(result *AnalysisResult)

	// Get retrieves an analysis result by ID.
	// Returns the result and true if found, nil and false otherwise.
	Get(id string) (*AnalysisResult, bool)

	// List returns all stored analysis results.
	// Returns an empty slice if no results are stored.
	List() []*AnalysisResult

	// Delete removes an analysis result by ID.
	// Returns true if the result was deleted, false if it didn't exist.
	Delete(id string) bool

	// Clear removes all stored analysis results.
	Clear()
}
