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
// Supports in-memory and PostgreSQL backends.
// MongoDB backend will be added in future releases.
func New(cfg *config.Config) ResultStore {
	if cfg == nil || cfg.Storage.Type == "" || cfg.Storage.Type == string(BackendMemory) {
		return NewMemory()
	}

	if cfg.Storage.Type == string(BackendPostgres) {
		connStr := cfg.Storage.PostgresConnectionString
		if connStr == "" {
			return NewMemory()
		}

		store, err := NewPostgres(connStr)
		if err != nil {
			return NewMemory()
		}
		return store
	}

	return NewMemory()
}
