// Package api provides REST API storage for analysis results.
package api

import (
	"sync"

	"github.com/opd-ai/go-stats-generator/internal/metrics"
)

// AnalysisResult stores the result of an analysis operation.
type AnalysisResult struct {
	ID     string
	Status string // "pending", "running", "completed", "failed"
	Report *metrics.Report
	Error  error
}

// Storage provides thread-safe in-memory storage for analysis results.
type Storage struct {
	mu      sync.RWMutex
	results map[string]*AnalysisResult
}

// NewStorage creates a new Storage instance.
func NewStorage() *Storage {
	return &Storage{
		results: make(map[string]*AnalysisResult),
	}
}

// Store saves an analysis result.
func (s *Storage) Store(result *AnalysisResult) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.results[result.ID] = result
}

// Get retrieves an analysis result by ID.
func (s *Storage) Get(id string) (*AnalysisResult, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result, ok := s.results[id]
	return result, ok
}
