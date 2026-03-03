package storage

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemory_Store(t *testing.T) {
	store := NewMemory()
	result := &AnalysisResult{
		ID:     "test-123",
		Status: "pending",
	}

	store.Store(result)

	got, ok := store.Get("test-123")
	require.True(t, ok)
	assert.Equal(t, "test-123", got.ID)
	assert.Equal(t, "pending", got.Status)
}

func TestMemory_Get_NotFound(t *testing.T) {
	store := NewMemory()

	_, ok := store.Get("nonexistent")
	assert.False(t, ok)
}

func TestMemory_List(t *testing.T) {
	store := NewMemory()

	result1 := &AnalysisResult{ID: "id1", Status: "pending"}
	result2 := &AnalysisResult{ID: "id2", Status: "running"}
	result3 := &AnalysisResult{ID: "id3", Status: "completed"}

	store.Store(result1)
	store.Store(result2)
	store.Store(result3)

	results := store.List()
	assert.Len(t, results, 3)

	ids := make(map[string]bool)
	for _, r := range results {
		ids[r.ID] = true
	}
	assert.True(t, ids["id1"])
	assert.True(t, ids["id2"])
	assert.True(t, ids["id3"])
}

func TestMemory_Delete(t *testing.T) {
	store := NewMemory()
	result := &AnalysisResult{ID: "test-123", Status: "pending"}
	store.Store(result)

	deleted := store.Delete("test-123")
	assert.True(t, deleted)

	_, ok := store.Get("test-123")
	assert.False(t, ok)
}

func TestMemory_Delete_NotFound(t *testing.T) {
	store := NewMemory()

	deleted := store.Delete("nonexistent")
	assert.False(t, deleted)
}

func TestMemory_Clear(t *testing.T) {
	store := NewMemory()

	store.Store(&AnalysisResult{ID: "id1", Status: "pending"})
	store.Store(&AnalysisResult{ID: "id2", Status: "running"})

	store.Clear()

	results := store.List()
	assert.Empty(t, results)
}

func TestMemory_ConcurrentAccess(t *testing.T) {
	store := NewMemory()
	var wg sync.WaitGroup
	iterations := 100

	// Concurrent writes
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			result := &AnalysisResult{
				ID:     string(rune(id)),
				Status: "pending",
			}
			store.Store(result)
		}(i)
	}

	// Concurrent reads
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			store.Get(string(rune(id)))
		}(i)
	}

	wg.Wait()

	// Verify no race conditions occurred
	results := store.List()
	assert.NotNil(t, results)
}

func TestMemory_UpdateExisting(t *testing.T) {
	store := NewMemory()

	initial := &AnalysisResult{ID: "test-123", Status: "pending"}
	store.Store(initial)

	updated := &AnalysisResult{ID: "test-123", Status: "completed"}
	store.Store(updated)

	got, ok := store.Get("test-123")
	require.True(t, ok)
	assert.Equal(t, "completed", got.Status)
}
