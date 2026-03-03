// Package storage provides storage abstractions for API analysis results.
package storage

import (
	"fmt"
	"os"
	"testing"

	"github.com/opd-ai/go-stats-generator/internal/metrics"
	"github.com/stretchr/testify/assert"
)

// getTestConnectionString returns PostgreSQL connection string from environment.
// If not set, tests are skipped.
func getTestConnectionString() (string, bool) {
	connStr := os.Getenv("POSTGRES_TEST_URL")
	if connStr == "" {
		connStr = "postgres://postgres:postgres@localhost:5432/go_stats_test?sslmode=disable"
	}
	return connStr, true
}

func TestPostgres_NewPostgres(t *testing.T) {
	connStr, ok := getTestConnectionString()
	if !ok {
		t.Skip("PostgreSQL test connection not available")
	}

	pg, err := NewPostgres(connStr)
	if err != nil {
		t.Skipf("PostgreSQL not available for testing: %v", err)
	}
	defer pg.Close()

	assert.NotNil(t, pg)
	assert.NotNil(t, pg.db)
}

func TestPostgres_StoreAndGet(t *testing.T) {
	connStr, ok := getTestConnectionString()
	if !ok {
		t.Skip("PostgreSQL test connection not available")
	}

	pg, err := NewPostgres(connStr)
	if err != nil {
		t.Skipf("PostgreSQL not available: %v", err)
	}
	defer pg.Close()
	defer pg.Clear()

	result := &AnalysisResult{
		ID:     "test-123",
		Status: "completed",
		Report: &metrics.Report{},
		Error:  nil,
	}

	pg.Store(result)

	retrieved, found := pg.Get("test-123")
	assert.True(t, found)
	assert.NotNil(t, retrieved)
	assert.Equal(t, "test-123", retrieved.ID)
	assert.Equal(t, "completed", retrieved.Status)
}

func TestPostgres_GetNonExistent(t *testing.T) {
	connStr, ok := getTestConnectionString()
	if !ok {
		t.Skip("PostgreSQL test connection not available")
	}

	pg, err := NewPostgres(connStr)
	if err != nil {
		t.Skipf("PostgreSQL not available: %v", err)
	}
	defer pg.Close()
	defer pg.Clear()

	retrieved, found := pg.Get("non-existent-id")
	assert.False(t, found)
	assert.Nil(t, retrieved)
}

func TestPostgres_List(t *testing.T) {
	connStr, ok := getTestConnectionString()
	if !ok {
		t.Skip("PostgreSQL test connection not available")
	}

	pg, err := NewPostgres(connStr)
	if err != nil {
		t.Skipf("PostgreSQL not available: %v", err)
	}
	defer pg.Close()
	defer pg.Clear()

	pg.Store(&AnalysisResult{ID: "id1", Status: "pending"})
	pg.Store(&AnalysisResult{ID: "id2", Status: "completed"})
	pg.Store(&AnalysisResult{ID: "id3", Status: "failed"})

	results := pg.List()
	assert.Len(t, results, 3)
}

func TestPostgres_Delete(t *testing.T) {
	connStr, ok := getTestConnectionString()
	if !ok {
		t.Skip("PostgreSQL test connection not available")
	}

	pg, err := NewPostgres(connStr)
	if err != nil {
		t.Skipf("PostgreSQL not available: %v", err)
	}
	defer pg.Close()
	defer pg.Clear()

	pg.Store(&AnalysisResult{ID: "delete-test", Status: "pending"})

	deleted := pg.Delete("delete-test")
	assert.True(t, deleted)

	_, found := pg.Get("delete-test")
	assert.False(t, found)
}

func TestPostgres_DeleteNonExistent(t *testing.T) {
	connStr, ok := getTestConnectionString()
	if !ok {
		t.Skip("PostgreSQL test connection not available")
	}

	pg, err := NewPostgres(connStr)
	if err != nil {
		t.Skipf("PostgreSQL not available: %v", err)
	}
	defer pg.Close()

	deleted := pg.Delete("non-existent")
	assert.False(t, deleted)
}

func TestPostgres_Clear(t *testing.T) {
	connStr, ok := getTestConnectionString()
	if !ok {
		t.Skip("PostgreSQL test connection not available")
	}

	pg, err := NewPostgres(connStr)
	if err != nil {
		t.Skipf("PostgreSQL not available: %v", err)
	}
	defer pg.Close()

	pg.Store(&AnalysisResult{ID: "clear1", Status: "pending"})
	pg.Store(&AnalysisResult{ID: "clear2", Status: "pending"})

	pg.Clear()

	results := pg.List()
	assert.Empty(t, results)
}

func TestPostgres_UpdateExisting(t *testing.T) {
	connStr, ok := getTestConnectionString()
	if !ok {
		t.Skip("PostgreSQL test connection not available")
	}

	pg, err := NewPostgres(connStr)
	if err != nil {
		t.Skipf("PostgreSQL not available: %v", err)
	}
	defer pg.Close()
	defer pg.Clear()

	pg.Store(&AnalysisResult{ID: "update-test", Status: "pending"})
	pg.Store(&AnalysisResult{ID: "update-test", Status: "completed"})

	result, found := pg.Get("update-test")
	assert.True(t, found)
	assert.Equal(t, "completed", result.Status)
}

func TestPostgres_ErrorHandling(t *testing.T) {
	connStr, ok := getTestConnectionString()
	if !ok {
		t.Skip("PostgreSQL test connection not available")
	}

	pg, err := NewPostgres(connStr)
	if err != nil {
		t.Skipf("PostgreSQL not available: %v", err)
	}
	defer pg.Close()
	defer pg.Clear()

	testErr := fmt.Errorf("test error message")
	pg.Store(&AnalysisResult{
		ID:     "error-test",
		Status: "failed",
		Error:  testErr,
	})

	result, found := pg.Get("error-test")
	assert.True(t, found)
	assert.NotNil(t, result.Error)
	assert.Equal(t, "test error message", result.Error.Error())
}

func TestPostgres_InvalidConnection(t *testing.T) {
	_, err := NewPostgres("invalid://connection")
	assert.Error(t, err)
}
