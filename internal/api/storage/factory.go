// Package storage provides storage abstractions for API analysis results.
package storage

import "github.com/opd-ai/go-stats-generator/internal/config"

// Backend represents the type of storage backend.
type Backend string

const (
	// BackendMemory represents in-memory storage.
	BackendMemory Backend = "memory"
	// BackendPostgres represents PostgreSQL storage (future).
	BackendPostgres Backend = "postgres"
	// BackendMongo represents MongoDB storage (future).
	BackendMongo Backend = "mongo"
)

// New creates a new ResultStore based on configuration.
// Currently only supports in-memory storage.
// PostgreSQL and MongoDB backends will be added in future releases.
func New(cfg *config.Config) ResultStore {
	// For now, always return in-memory storage
	// Future: read cfg.Storage.Backend to select implementation
	return NewMemory()
}
