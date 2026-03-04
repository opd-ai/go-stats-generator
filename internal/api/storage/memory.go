package storage

import "sync"

// Memory provides thread-safe in-memory storage for analysis results.
type Memory struct {
	mu      sync.RWMutex
	results map[string]*AnalysisResult
}

// NewMemory creates a new in-memory storage instance for thread-safe temporary caching of analysis results.
// Uses sync.RWMutex for concurrent access protection, making it safe for multi-threaded API server environments.
// All data is lost when the process terminates, making this unsuitable for long-term persistence.
func NewMemory() *Memory {
	return &Memory{
		results: make(map[string]*AnalysisResult),
	}
}

// Store saves an analysis result.
func (m *Memory) Store(result *AnalysisResult) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.results[result.ID] = result
}

// Get retrieves an analysis result by ID.
func (m *Memory) Get(id string) (*AnalysisResult, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result, ok := m.results[id]
	return result, ok
}

// List returns all stored analysis results.
func (m *Memory) List() []*AnalysisResult {
	m.mu.RLock()
	defer m.mu.RUnlock()

	results := make([]*AnalysisResult, 0, len(m.results))
	for _, result := range m.results {
		results = append(results, result)
	}
	return results
}

// Delete removes an analysis result by ID.
func (m *Memory) Delete(id string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.results[id]; exists {
		delete(m.results, id)
		return true
	}
	return false
}

// Clear removes all stored analysis results.
func (m *Memory) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.results = make(map[string]*AnalysisResult)
}
