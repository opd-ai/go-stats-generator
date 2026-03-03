// Package storage provides storage abstractions for API analysis results.
package storage

import "github.com/opd-ai/go-stats-generator/internal/config"

// Backend represents the type of storage backend.
type Backend string

const (
	// BackendMemory represents in-memory storage.
	BackendMemory Backend = "memory"
	// BackendPostgres represents PostgreSQL storage.
	BackendPostgres Backend = "postgres"
	// BackendMongo represents MongoDB storage.
	BackendMongo Backend = "mongo"
)

// New creates a new ResultStore based on configuration.
// Supports in-memory, PostgreSQL, and MongoDB backends.
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

	if cfg.Storage.Type == string(BackendMongo) {
		connStr := cfg.Storage.MongoConnectionString
		if connStr == "" {
			return NewMemory()
		}

		store, err := NewMongo(connStr)
		if err != nil {
			return NewMemory()
		}
		return store
	}

	return NewMemory()
}
