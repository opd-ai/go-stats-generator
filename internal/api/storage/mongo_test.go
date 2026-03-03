// Package storage provides storage abstractions for API analysis results.
package storage

import (
	"os"
	"testing"

	"github.com/opd-ai/go-stats-generator/internal/metrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// getMongoTestConnectionString returns MongoDB connection string from environment
// or skips the test if not available.
func getMongoTestConnectionString(t *testing.T) string {
	connStr := os.Getenv("MONGO_TEST_CONNECTION_STRING")
	if connStr == "" {
		t.Skip("Skipping MongoDB test: MONGO_TEST_CONNECTION_STRING not set")
	}
	return connStr
}

func TestNewMongo(t *testing.T) {
	connStr := getMongoTestConnectionString(t)

	store, err := NewMongo(connStr)
	require.NoError(t, err)
	require.NotNil(t, store)
	defer store.Close()

	assert.NotNil(t, store.client)
	assert.NotNil(t, store.collection)
}

func TestNewMongo_InvalidConnection(t *testing.T) {
	store, err := NewMongo("mongodb://invalid:27017")
	assert.Error(t, err)
	assert.Nil(t, store)
}

func TestMongoStore_StoreAndGet(t *testing.T) {
	connStr := getMongoTestConnectionString(t)

	store, err := NewMongo(connStr)
	require.NoError(t, err)
	defer store.Close()
	defer store.Clear()

	result := &AnalysisResult{
		ID:     "test-1",
		Status: "completed",
		Report: &metrics.Report{
			Metadata: metrics.ReportMetadata{
				ToolVersion: "1.0.0",
			},
		},
	}

	store.Store(result)

	retrieved, found := store.Get("test-1")
	assert.True(t, found)
	assert.Equal(t, "test-1", retrieved.ID)
	assert.Equal(t, "completed", retrieved.Status)
	assert.NotNil(t, retrieved.Report)
}

func TestMongoStore_GetNotFound(t *testing.T) {
	connStr := getMongoTestConnectionString(t)

	store, err := NewMongo(connStr)
	require.NoError(t, err)
	defer store.Close()
	defer store.Clear()

	retrieved, found := store.Get("nonexistent")
	assert.False(t, found)
	assert.Nil(t, retrieved)
}

func TestMongoStore_UpdateExisting(t *testing.T) {
	connStr := getMongoTestConnectionString(t)

	store, err := NewMongo(connStr)
	require.NoError(t, err)
	defer store.Close()
	defer store.Clear()

	result := &AnalysisResult{
		ID:     "test-2",
		Status: "pending",
	}

	store.Store(result)

	result.Status = "completed"
	store.Store(result)

	retrieved, found := store.Get("test-2")
	assert.True(t, found)
	assert.Equal(t, "completed", retrieved.Status)
}

func TestMongoStore_List(t *testing.T) {
	connStr := getMongoTestConnectionString(t)

	store, err := NewMongo(connStr)
	require.NoError(t, err)
	defer store.Close()
	defer store.Clear()

	result1 := &AnalysisResult{
		ID:     "test-3",
		Status: "completed",
	}
	result2 := &AnalysisResult{
		ID:     "test-4",
		Status: "pending",
	}

	store.Store(result1)
	store.Store(result2)

	results := store.List()
	assert.Len(t, results, 2)
}

func TestMongoStore_Delete(t *testing.T) {
	connStr := getMongoTestConnectionString(t)

	store, err := NewMongo(connStr)
	require.NoError(t, err)
	defer store.Close()
	defer store.Clear()

	result := &AnalysisResult{
		ID:     "test-5",
		Status: "completed",
	}

	store.Store(result)

	deleted := store.Delete("test-5")
	assert.True(t, deleted)

	_, found := store.Get("test-5")
	assert.False(t, found)
}

func TestMongoStore_DeleteNonexistent(t *testing.T) {
	connStr := getMongoTestConnectionString(t)

	store, err := NewMongo(connStr)
	require.NoError(t, err)
	defer store.Close()
	defer store.Clear()

	deleted := store.Delete("nonexistent")
	assert.False(t, deleted)
}

func TestMongoStore_Clear(t *testing.T) {
	connStr := getMongoTestConnectionString(t)

	store, err := NewMongo(connStr)
	require.NoError(t, err)
	defer store.Close()

	result1 := &AnalysisResult{
		ID:     "test-6",
		Status: "completed",
	}
	result2 := &AnalysisResult{
		ID:     "test-7",
		Status: "pending",
	}

	store.Store(result1)
	store.Store(result2)

	store.Clear()

	results := store.List()
	assert.Len(t, results, 0)
}

func TestMongoStore_WithError(t *testing.T) {
	connStr := getMongoTestConnectionString(t)

	store, err := NewMongo(connStr)
	require.NoError(t, err)
	defer store.Close()
	defer store.Clear()

	result := &AnalysisResult{
		ID:     "test-8",
		Status: "failed",
		Error:  assert.AnError,
	}

	store.Store(result)

	retrieved, found := store.Get("test-8")
	assert.True(t, found)
	assert.NotNil(t, retrieved.Error)
}

func TestMongoStore_ThreadSafety(t *testing.T) {
	connStr := getMongoTestConnectionString(t)

	store, err := NewMongo(connStr)
	require.NoError(t, err)
	defer store.Close()
	defer store.Clear()

	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(id int) {
			result := &AnalysisResult{
				ID:     string(rune('a' + id)),
				Status: "completed",
			}
			store.Store(result)
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	results := store.List()
	assert.GreaterOrEqual(t, len(results), 1)
}
