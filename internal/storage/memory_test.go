package storage

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/opd-ai/go-stats-generator/internal/metrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemoryStorage_StoreAndRetrieve(t *testing.T) {
	storage := NewMemoryStorage()
	defer storage.Close()

	ctx := context.Background()

	snapshot := metrics.Snapshot{
		ID: "test-123",
		Report: metrics.Report{
			Overview: metrics.OverviewMetrics{
				TotalFiles: 10,
			},
		},
	}

	metadata := metrics.SnapshotMetadata{
		Timestamp: time.Now(),
		GitCommit: "abc123",
		GitBranch: "main",
	}

	err := storage.Store(ctx, snapshot, metadata)
	require.NoError(t, err)

	retrieved, err := storage.Retrieve(ctx, "test-123")
	require.NoError(t, err)
	assert.Equal(t, snapshot.ID, retrieved.ID)
	assert.Equal(t, 10, retrieved.Report.Overview.TotalFiles)
}

func TestMemoryStorage_StoreEmptyID(t *testing.T) {
	storage := NewMemoryStorage()
	defer storage.Close()

	ctx := context.Background()

	snapshot := metrics.Snapshot{
		ID: "",
	}
	metadata := metrics.SnapshotMetadata{}

	err := storage.Store(ctx, snapshot, metadata)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "snapshot ID cannot be empty")
}

func TestMemoryStorage_RetrieveNotFound(t *testing.T) {
	storage := NewMemoryStorage()
	defer storage.Close()

	ctx := context.Background()

	_, err := storage.Retrieve(ctx, "nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "snapshot not found")
}

func TestMemoryStorage_List(t *testing.T) {
	storage := NewMemoryStorage()
	defer storage.Close()

	ctx := context.Background()

	for i := 0; i < 5; i++ {
		snapshot := metrics.Snapshot{
			ID: fmt.Sprintf("snap-%d", i),
		}
		metadata := metrics.SnapshotMetadata{
			Timestamp: time.Now().Add(time.Duration(i) * time.Hour),
			GitBranch: "main",
		}
		err := storage.Store(ctx, snapshot, metadata)
		require.NoError(t, err)
	}

	list, err := storage.List(ctx, SnapshotFilter{})
	require.NoError(t, err)
	assert.Len(t, list, 5)
	assert.True(t, list[0].Timestamp.After(list[1].Timestamp))
}

func TestMemoryStorage_ListWithFilter(t *testing.T) {
	storage := NewMemoryStorage()
	defer storage.Close()

	ctx := context.Background()
	now := time.Now()

	snapshots := []struct {
		id     string
		branch string
		when   time.Time
	}{
		{"snap-1", "main", now.Add(-2 * time.Hour)},
		{"snap-2", "feature", now.Add(-1 * time.Hour)},
		{"snap-3", "main", now},
	}

	for _, s := range snapshots {
		snapshot := metrics.Snapshot{ID: s.id}
		metadata := metrics.SnapshotMetadata{
			Timestamp: s.when,
			GitBranch: s.branch,
		}
		err := storage.Store(ctx, snapshot, metadata)
		require.NoError(t, err)
	}

	filter := SnapshotFilter{Branch: "main"}
	list, err := storage.List(ctx, filter)
	require.NoError(t, err)
	assert.Len(t, list, 2)
	assert.Equal(t, "main", list[0].GitBranch)
}

func TestMemoryStorage_ListWithLimit(t *testing.T) {
	storage := NewMemoryStorage()
	defer storage.Close()

	ctx := context.Background()

	for i := 0; i < 10; i++ {
		snapshot := metrics.Snapshot{
			ID: fmt.Sprintf("snap-%d", i),
		}
		metadata := metrics.SnapshotMetadata{
			Timestamp: time.Now().Add(time.Duration(i) * time.Minute),
		}
		err := storage.Store(ctx, snapshot, metadata)
		require.NoError(t, err)
	}

	filter := SnapshotFilter{Limit: 3}
	list, err := storage.List(ctx, filter)
	require.NoError(t, err)
	assert.Len(t, list, 3)
}

func TestMemoryStorage_Delete(t *testing.T) {
	storage := NewMemoryStorage()
	defer storage.Close()

	ctx := context.Background()

	snapshot := metrics.Snapshot{ID: "test-delete"}
	metadata := metrics.SnapshotMetadata{Timestamp: time.Now()}

	err := storage.Store(ctx, snapshot, metadata)
	require.NoError(t, err)

	err = storage.Delete(ctx, "test-delete")
	require.NoError(t, err)

	_, err = storage.Retrieve(ctx, "test-delete")
	assert.Error(t, err)
}

func TestMemoryStorage_DeleteNotFound(t *testing.T) {
	storage := NewMemoryStorage()
	defer storage.Close()

	ctx := context.Background()

	err := storage.Delete(ctx, "nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "snapshot not found")
}

func TestMemoryStorage_Cleanup(t *testing.T) {
	storage := NewMemoryStorage()
	defer storage.Close()

	ctx := context.Background()
	now := time.Now()

	oldSnapshot := metrics.Snapshot{ID: "old"}
	oldMetadata := metrics.SnapshotMetadata{
		Timestamp: now.Add(-48 * time.Hour),
	}
	storage.snapshots["old"] = &storedSnapshot{
		snapshot: oldSnapshot,
		metadata: oldMetadata,
		stored:   now.Add(-48 * time.Hour),
	}

	newSnapshot := metrics.Snapshot{ID: "new"}
	newMetadata := metrics.SnapshotMetadata{
		Timestamp: now,
	}
	err := storage.Store(ctx, newSnapshot, newMetadata)
	require.NoError(t, err)

	policy := RetentionPolicy{
		MaxAge:   24 * time.Hour,
		MaxCount: 0,
	}

	err = storage.Cleanup(ctx, policy)
	require.NoError(t, err)

	list, err := storage.List(ctx, SnapshotFilter{})
	require.NoError(t, err)
	assert.Len(t, list, 1)
	assert.Equal(t, "new", list[0].ID)
}

func TestMemoryStorage_CleanupMaxCount(t *testing.T) {
	storage := NewMemoryStorage()
	defer storage.Close()

	ctx := context.Background()

	for i := 0; i < 5; i++ {
		snapshot := metrics.Snapshot{
			ID: fmt.Sprintf("snap-%d", i),
		}
		storage.snapshots[snapshot.ID] = &storedSnapshot{
			snapshot: snapshot,
			metadata: metrics.SnapshotMetadata{
				Timestamp: time.Now(),
			},
			stored: time.Now().Add(-time.Duration(i) * time.Hour),
		}
	}

	policy := RetentionPolicy{
		MaxCount: 2,
	}

	err := storage.Cleanup(ctx, policy)
	require.NoError(t, err)

	list, err := storage.List(ctx, SnapshotFilter{})
	require.NoError(t, err)
	assert.Len(t, list, 2)
}

func TestMemoryStorage_GetLatest(t *testing.T) {
	storage := NewMemoryStorage()
	defer storage.Close()

	ctx := context.Background()
	now := time.Now()

	for i := 0; i < 3; i++ {
		snapshot := metrics.Snapshot{
			ID: fmt.Sprintf("snap-%d", i),
		}
		metadata := metrics.SnapshotMetadata{
			Timestamp: now.Add(time.Duration(i) * time.Hour),
		}
		err := storage.Store(ctx, snapshot, metadata)
		require.NoError(t, err)
	}

	latest, err := storage.GetLatest(ctx)
	require.NoError(t, err)
	assert.Equal(t, "snap-2", latest.ID)
}

func TestMemoryStorage_GetLatestEmpty(t *testing.T) {
	storage := NewMemoryStorage()
	defer storage.Close()

	ctx := context.Background()

	_, err := storage.GetLatest(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no snapshots available")
}

func TestMemoryStorage_GetByTag(t *testing.T) {
	storage := NewMemoryStorage()
	defer storage.Close()

	ctx := context.Background()

	snap1 := metrics.Snapshot{ID: "snap-1"}
	meta1 := metrics.SnapshotMetadata{
		Timestamp: time.Now(),
		Tags:      map[string]string{"env": "prod"},
	}
	err := storage.Store(ctx, snap1, meta1)
	require.NoError(t, err)

	snap2 := metrics.Snapshot{ID: "snap-2"}
	meta2 := metrics.SnapshotMetadata{
		Timestamp: time.Now(),
		Tags:      map[string]string{"env": "dev"},
	}
	err = storage.Store(ctx, snap2, meta2)
	require.NoError(t, err)

	results, err := storage.GetByTag(ctx, "env", "prod")
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "snap-1", results[0].ID)
}

func TestMemoryStorage_ConcurrentAccess(t *testing.T) {
	storage := NewMemoryStorage()
	defer storage.Close()

	ctx := context.Background()

	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			snapshot := metrics.Snapshot{
				ID: fmt.Sprintf("snap-%d", id),
			}
			metadata := metrics.SnapshotMetadata{
				Timestamp: time.Now(),
			}
			storage.Store(ctx, snapshot, metadata)
			storage.Retrieve(ctx, snapshot.ID)
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	list, err := storage.List(ctx, SnapshotFilter{})
	require.NoError(t, err)
	assert.Len(t, list, 10)
}
