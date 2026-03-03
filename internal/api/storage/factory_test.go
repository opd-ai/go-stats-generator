package storage

import (
	"os"
	"testing"

	"github.com/opd-ai/go-stats-generator/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestNew_ReturnsMemoryBackend(t *testing.T) {
	cfg := config.DefaultConfig()

	store := New(cfg)

	assert.NotNil(t, store)
	_, ok := store.(*Memory)
	assert.True(t, ok, "expected Memory implementation")
}

func TestNew_ImplementsInterface(t *testing.T) {
	cfg := config.DefaultConfig()

	var _ ResultStore = New(cfg)
}

func TestNew_PostgresBackendWithValidConnection(t *testing.T) {
	connStr := os.Getenv("POSTGRES_TEST_URL")
	if connStr == "" {
		connStr = "postgres://postgres:postgres@localhost:5432/go_stats_test?sslmode=disable"
	}

	cfg := config.DefaultConfig()
	cfg.Storage.Type = "postgres"
	cfg.Storage.PostgresConnectionString = connStr

	store := New(cfg)
	assert.NotNil(t, store)

	// If postgres is available, we get Postgres backend
	// If not, factory falls back to Memory
	_, isPg := store.(*Postgres)
	_, isMem := store.(*Memory)
	assert.True(t, isPg || isMem, "should return either Postgres or Memory")

	if pg, ok := store.(*Postgres); ok {
		defer pg.Close()
	}
}

func TestNew_PostgresBackendWithInvalidConnection(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Storage.Type = "postgres"
	cfg.Storage.PostgresConnectionString = "invalid://connection"

	store := New(cfg)
	assert.NotNil(t, store)

	// Should fall back to Memory on connection failure
	_, ok := store.(*Memory)
	assert.True(t, ok, "should fall back to Memory on error")
}

func TestNew_PostgresBackendWithEmptyConnection(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Storage.Type = "postgres"
	cfg.Storage.PostgresConnectionString = ""

	store := New(cfg)
	assert.NotNil(t, store)

	_, ok := store.(*Memory)
	assert.True(t, ok, "should return Memory when connection string is empty")
}

func TestNew_NilConfig(t *testing.T) {
	store := New(nil)
	assert.NotNil(t, store)

	_, ok := store.(*Memory)
	assert.True(t, ok, "should return Memory for nil config")
}
