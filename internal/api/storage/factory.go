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

// New creates a ResultStore backend instance dynamically based on application configuration, routing to the appropriate implementation.
// It supports in-memory storage (for testing/development), PostgreSQL (for production persistence with ACID guarantees), and MongoDB
// (for NoSQL document storage). The factory pattern enables runtime backend selection via config files or environment variables,
// allowing seamless switching between storage implementations without code changes. Defaults to in-memory if configuration is unspecified.
func New(cfg *config.Config) ResultStore {
	if shouldUseMemory(cfg) {
		return NewMemory()
	}

	creator := getBackendCreator(cfg.Storage.Type)
	store := creator(cfg)
	if store != nil {
		return store
	}

	return NewMemory()
}

// shouldUseMemory determines if memory backend should be used.
func shouldUseMemory(cfg *config.Config) bool {
	return cfg == nil || cfg.Storage.Type == "" || cfg.Storage.Type == string(BackendMemory)
}

// getBackendCreator returns a backend creator function for the given type.
func getBackendCreator(backendType string) func(*config.Config) ResultStore {
	creators := map[string]func(*config.Config) ResultStore{
		string(BackendPostgres): createPostgresBackend,
		string(BackendMongo):    createMongoBackend,
	}

	if creator, exists := creators[backendType]; exists {
		return creator
	}
	return func(*config.Config) ResultStore { return nil }
}

// createPostgresBackend creates a PostgreSQL backend or returns nil on error.
func createPostgresBackend(cfg *config.Config) ResultStore {
	connStr := cfg.Storage.PostgresConnectionString
	if connStr == "" {
		return nil
	}

	store, err := NewPostgres(connStr)
	if err != nil {
		return nil
	}
	return store
}

// createMongoBackend creates a MongoDB backend or returns nil on error.
func createMongoBackend(cfg *config.Config) ResultStore {
	connStr := cfg.Storage.MongoConnectionString
	if connStr == "" {
		return nil
	}

	store, err := NewMongo(connStr)
	if err != nil {
		return nil
	}
	return store
}
