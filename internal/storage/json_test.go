package storage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/opd-ai/go-stats-generator/internal/metrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewJSONStorageImpl(t *testing.T) {
	tempDir := t.TempDir()

	config := JSONConfig{
		Directory:   tempDir,
		Compression: false,
		Pretty:      false,
	}

	storage, err := NewJSONStorageImpl(config)
	require.NoError(t, err)
	require.NotNil(t, storage)
	defer storage.Close()

	// Verify directory was created
	_, err = os.Stat(tempDir)
	assert.NoError(t, err)
}

func TestJSONStorage_Store_Uncompressed(t *testing.T) {
	tempDir := t.TempDir()
	storage, err := NewJSONStorageImpl(JSONConfig{
		Directory:   tempDir,
		Compression: false,
		Pretty:      false,
	})
	require.NoError(t, err)
	defer storage.Close()

	ctx := context.Background()
	snapshot := createTestSnapshot("test-1")
	metadata := createTestMetadata()

	err = storage.Store(ctx, snapshot, metadata)
	require.NoError(t, err)

	// Verify file was created
	expectedFile := filepath.Join(tempDir, "test-1.json")
	_, err = os.Stat(expectedFile)
	assert.NoError(t, err)
}

func TestJSONStorage_Store_Compressed(t *testing.T) {
	tempDir := t.TempDir()
	storage, err := NewJSONStorageImpl(JSONConfig{
		Directory:   tempDir,
		Compression: true,
		Pretty:      false,
	})
	require.NoError(t, err)
	defer storage.Close()

	ctx := context.Background()
	snapshot := createTestSnapshot("test-2")
	metadata := createTestMetadata()

	err = storage.Store(ctx, snapshot, metadata)
	require.NoError(t, err)

	// Verify compressed file was created
	expectedFile := filepath.Join(tempDir, "test-2.json.gz")
	_, err = os.Stat(expectedFile)
	assert.NoError(t, err)
}

func TestJSONStorage_Store_Pretty(t *testing.T) {
	tempDir := t.TempDir()
	storage, err := NewJSONStorageImpl(JSONConfig{
		Directory:   tempDir,
		Compression: false,
		Pretty:      true,
	})
	require.NoError(t, err)
	defer storage.Close()

	ctx := context.Background()
	snapshot := createTestSnapshot("test-3")
	metadata := createTestMetadata()

	err = storage.Store(ctx, snapshot, metadata)
	require.NoError(t, err)

	// Read file and verify it's pretty-printed
	data, err := os.ReadFile(filepath.Join(tempDir, "test-3.json"))
	require.NoError(t, err)

	// Pretty JSON should contain newlines and indentation
	assert.Contains(t, string(data), "\n")
	assert.Contains(t, string(data), "  ")
}

func TestJSONStorage_Retrieve_Uncompressed(t *testing.T) {
	tempDir := t.TempDir()
	storage, err := NewJSONStorageImpl(JSONConfig{
		Directory:   tempDir,
		Compression: false,
		Pretty:      false,
	})
	require.NoError(t, err)
	defer storage.Close()

	ctx := context.Background()
	original := createTestSnapshot("test-4")
	metadata := createTestMetadata()

	// Store
	err = storage.Store(ctx, original, metadata)
	require.NoError(t, err)

	// Retrieve
	retrieved, err := storage.Retrieve(ctx, "test-4")
	require.NoError(t, err)

	// Verify data matches
	assert.Equal(t, original.ID, retrieved.ID)
	assert.Equal(t, original.Report.Overview.TotalFiles, retrieved.Report.Overview.TotalFiles)
	assert.Equal(t, metadata.GitCommit, retrieved.Metadata.GitCommit)
}

func TestJSONStorage_Retrieve_Compressed(t *testing.T) {
	tempDir := t.TempDir()
	storage, err := NewJSONStorageImpl(JSONConfig{
		Directory:   tempDir,
		Compression: true,
		Pretty:      false,
	})
	require.NoError(t, err)
	defer storage.Close()

	ctx := context.Background()
	original := createTestSnapshot("test-5")
	metadata := createTestMetadata()

	// Store
	err = storage.Store(ctx, original, metadata)
	require.NoError(t, err)

	// Retrieve
	retrieved, err := storage.Retrieve(ctx, "test-5")
	require.NoError(t, err)

	// Verify data matches
	assert.Equal(t, original.ID, retrieved.ID)
	assert.Equal(t, original.Report.Overview.TotalFiles, retrieved.Report.Overview.TotalFiles)
}

func TestJSONStorage_Retrieve_NotFound(t *testing.T) {
	tempDir := t.TempDir()
	storage, err := NewJSONStorageImpl(JSONConfig{
		Directory:   tempDir,
		Compression: false,
		Pretty:      false,
	})
	require.NoError(t, err)
	defer storage.Close()

	ctx := context.Background()
	_, err = storage.Retrieve(ctx, "nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "snapshot not found")
}

func TestJSONStorage_List_Empty(t *testing.T) {
	tempDir := t.TempDir()
	storage, err := NewJSONStorageImpl(JSONConfig{
		Directory:   tempDir,
		Compression: false,
		Pretty:      false,
	})
	require.NoError(t, err)
	defer storage.Close()

	ctx := context.Background()
	snapshots, err := storage.List(ctx, SnapshotFilter{})
	require.NoError(t, err)
	assert.Empty(t, snapshots)
}

func TestJSONStorage_List_MultipleSnapshots(t *testing.T) {
	tempDir := t.TempDir()
	storage, err := NewJSONStorageImpl(JSONConfig{
		Directory:   tempDir,
		Compression: false,
		Pretty:      false,
	})
	require.NoError(t, err)
	defer storage.Close()

	ctx := context.Background()

	// Store multiple snapshots
	for i := 1; i <= 5; i++ {
		snapshot := createTestSnapshot(fmt.Sprintf("snapshot-%d", i))
		metadata := createTestMetadata()
		metadata.Timestamp = time.Now().Add(time.Duration(i) * time.Hour)
		err = storage.Store(ctx, snapshot, metadata)
		require.NoError(t, err)
	}

	// List all
	snapshots, err := storage.List(ctx, SnapshotFilter{})
	require.NoError(t, err)
	assert.Len(t, snapshots, 5)

	// Verify sorted by timestamp descending (newest first)
	for i := 0; i < len(snapshots)-1; i++ {
		assert.True(t, snapshots[i].Timestamp.After(snapshots[i+1].Timestamp))
	}
}

func TestJSONStorage_List_WithLimit(t *testing.T) {
	tempDir := t.TempDir()
	storage, err := NewJSONStorageImpl(JSONConfig{
		Directory:   tempDir,
		Compression: false,
		Pretty:      false,
	})
	require.NoError(t, err)
	defer storage.Close()

	ctx := context.Background()

	// Store 5 snapshots
	for i := 1; i <= 5; i++ {
		snapshot := createTestSnapshot(fmt.Sprintf("snapshot-%d", i))
		metadata := createTestMetadata()
		err = storage.Store(ctx, snapshot, metadata)
		require.NoError(t, err)
	}

	// List with limit
	snapshots, err := storage.List(ctx, SnapshotFilter{Limit: 3})
	require.NoError(t, err)
	assert.Len(t, snapshots, 3)
}

func TestJSONStorage_List_WithOffset(t *testing.T) {
	tempDir := t.TempDir()
	storage, err := NewJSONStorageImpl(JSONConfig{
		Directory:   tempDir,
		Compression: false,
		Pretty:      false,
	})
	require.NoError(t, err)
	defer storage.Close()

	ctx := context.Background()

	// Store 5 snapshots
	for i := 1; i <= 5; i++ {
		snapshot := createTestSnapshot(fmt.Sprintf("snapshot-%d", i))
		metadata := createTestMetadata()
		err = storage.Store(ctx, snapshot, metadata)
		require.NoError(t, err)
	}

	// List with offset
	snapshots, err := storage.List(ctx, SnapshotFilter{Offset: 2})
	require.NoError(t, err)
	assert.Len(t, snapshots, 3)
}

func TestJSONStorage_List_FilterByBranch(t *testing.T) {
	tempDir := t.TempDir()
	storage, err := NewJSONStorageImpl(JSONConfig{
		Directory:   tempDir,
		Compression: false,
		Pretty:      false,
	})
	require.NoError(t, err)
	defer storage.Close()

	ctx := context.Background()

	// Store snapshots on different branches
	for i := 1; i <= 3; i++ {
		snapshot := createTestSnapshot(fmt.Sprintf("main-%d", i))
		metadata := createTestMetadata()
		metadata.GitBranch = "main"
		err = storage.Store(ctx, snapshot, metadata)
		require.NoError(t, err)
	}

	for i := 1; i <= 2; i++ {
		snapshot := createTestSnapshot(fmt.Sprintf("dev-%d", i))
		metadata := createTestMetadata()
		metadata.GitBranch = "develop"
		err = storage.Store(ctx, snapshot, metadata)
		require.NoError(t, err)
	}

	// Filter by branch
	snapshots, err := storage.List(ctx, SnapshotFilter{Branch: "main"})
	require.NoError(t, err)
	assert.Len(t, snapshots, 3)
	for _, s := range snapshots {
		assert.Equal(t, "main", s.GitBranch)
	}
}

func TestJSONStorage_List_FilterByTags(t *testing.T) {
	tempDir := t.TempDir()
	storage, err := NewJSONStorageImpl(JSONConfig{
		Directory:   tempDir,
		Compression: false,
		Pretty:      false,
	})
	require.NoError(t, err)
	defer storage.Close()

	ctx := context.Background()

	// Store snapshots with different tags
	snapshot1 := createTestSnapshot("tagged-1")
	metadata1 := createTestMetadata()
	metadata1.Tags = map[string]string{"env": "prod", "version": "1.0"}
	err = storage.Store(ctx, snapshot1, metadata1)
	require.NoError(t, err)

	snapshot2 := createTestSnapshot("tagged-2")
	metadata2 := createTestMetadata()
	metadata2.Tags = map[string]string{"env": "dev", "version": "1.0"}
	err = storage.Store(ctx, snapshot2, metadata2)
	require.NoError(t, err)

	// Filter by tag
	snapshots, err := storage.List(ctx, SnapshotFilter{
		Tags: map[string]string{"env": "prod"},
	})
	require.NoError(t, err)
	assert.Len(t, snapshots, 1)
	assert.Equal(t, "tagged-1", snapshots[0].ID)
}

func TestJSONStorage_List_FilterByTimeRange(t *testing.T) {
	tempDir := t.TempDir()
	storage, err := NewJSONStorageImpl(JSONConfig{
		Directory:   tempDir,
		Compression: false,
		Pretty:      false,
	})
	require.NoError(t, err)
	defer storage.Close()

	ctx := context.Background()
	baseTime := time.Now()

	// Store snapshots at different times
	for i := 0; i < 5; i++ {
		snapshot := createTestSnapshot(fmt.Sprintf("time-%d", i))
		metadata := createTestMetadata()
		metadata.Timestamp = baseTime.Add(time.Duration(i) * time.Hour)
		err = storage.Store(ctx, snapshot, metadata)
		require.NoError(t, err)
	}

	// Filter by time range
	after := baseTime.Add(2 * time.Hour)
	snapshots, err := storage.List(ctx, SnapshotFilter{After: &after})
	require.NoError(t, err)
	assert.Len(t, snapshots, 2) // snapshots 3 and 4
}

func TestJSONStorage_Delete(t *testing.T) {
	tempDir := t.TempDir()
	storage, err := NewJSONStorageImpl(JSONConfig{
		Directory:   tempDir,
		Compression: false,
		Pretty:      false,
	})
	require.NoError(t, err)
	defer storage.Close()

	ctx := context.Background()
	snapshot := createTestSnapshot("delete-me")
	metadata := createTestMetadata()

	// Store
	err = storage.Store(ctx, snapshot, metadata)
	require.NoError(t, err)

	// Verify exists
	_, err = storage.Retrieve(ctx, "delete-me")
	require.NoError(t, err)

	// Delete
	err = storage.Delete(ctx, "delete-me")
	require.NoError(t, err)

	// Verify deleted
	_, err = storage.Retrieve(ctx, "delete-me")
	assert.Error(t, err)
}

func TestJSONStorage_Delete_NotFound(t *testing.T) {
	tempDir := t.TempDir()
	storage, err := NewJSONStorageImpl(JSONConfig{
		Directory:   tempDir,
		Compression: false,
		Pretty:      false,
	})
	require.NoError(t, err)
	defer storage.Close()

	ctx := context.Background()
	err = storage.Delete(ctx, "nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "snapshot not found")
}

func TestJSONStorage_GetLatest(t *testing.T) {
	tempDir := t.TempDir()
	storage, err := NewJSONStorageImpl(JSONConfig{
		Directory:   tempDir,
		Compression: false,
		Pretty:      false,
	})
	require.NoError(t, err)
	defer storage.Close()

	ctx := context.Background()

	// Store snapshots with different timestamps
	oldest := createTestSnapshot("oldest")
	oldestMeta := createTestMetadata()
	oldestMeta.Timestamp = time.Now().Add(-2 * time.Hour)
	err = storage.Store(ctx, oldest, oldestMeta)
	require.NoError(t, err)

	newest := createTestSnapshot("newest")
	newestMeta := createTestMetadata()
	newestMeta.Timestamp = time.Now()
	err = storage.Store(ctx, newest, newestMeta)
	require.NoError(t, err)

	// Get latest
	latest, err := storage.GetLatest(ctx)
	require.NoError(t, err)
	assert.Equal(t, "newest", latest.ID)
}

func TestJSONStorage_GetLatest_Empty(t *testing.T) {
	tempDir := t.TempDir()
	storage, err := NewJSONStorageImpl(JSONConfig{
		Directory:   tempDir,
		Compression: false,
		Pretty:      false,
	})
	require.NoError(t, err)
	defer storage.Close()

	ctx := context.Background()
	_, err = storage.GetLatest(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no snapshots found")
}

func TestJSONStorage_GetByTag(t *testing.T) {
	tempDir := t.TempDir()
	storage, err := NewJSONStorageImpl(JSONConfig{
		Directory:   tempDir,
		Compression: false,
		Pretty:      false,
	})
	require.NoError(t, err)
	defer storage.Close()

	ctx := context.Background()

	// Store snapshots with different tags
	for i := 1; i <= 3; i++ {
		snapshot := createTestSnapshot(fmt.Sprintf("release-%d", i))
		metadata := createTestMetadata()
		metadata.Tags = map[string]string{"type": "release"}
		err = storage.Store(ctx, snapshot, metadata)
		require.NoError(t, err)
	}

	snapshot := createTestSnapshot("dev-1")
	metadata := createTestMetadata()
	metadata.Tags = map[string]string{"type": "dev"}
	err = storage.Store(ctx, snapshot, metadata)
	require.NoError(t, err)

	// Get by tag
	snapshots, err := storage.GetByTag(ctx, "type", "release")
	require.NoError(t, err)
	assert.Len(t, snapshots, 3)
}

func TestJSONStorage_Cleanup_ByAge(t *testing.T) {
	tempDir := t.TempDir()
	storage, err := NewJSONStorageImpl(JSONConfig{
		Directory:   tempDir,
		Compression: false,
		Pretty:      false,
	})
	require.NoError(t, err)
	defer storage.Close()

	ctx := context.Background()

	// Store old snapshots
	old := createTestSnapshot("old")
	oldMeta := createTestMetadata()
	oldMeta.Timestamp = time.Now().Add(-48 * time.Hour)
	err = storage.Store(ctx, old, oldMeta)
	require.NoError(t, err)

	// Store recent snapshot
	recent := createTestSnapshot("recent")
	recentMeta := createTestMetadata()
	recentMeta.Timestamp = time.Now()
	err = storage.Store(ctx, recent, recentMeta)
	require.NoError(t, err)

	// Cleanup (delete older than 24 hours)
	policy := RetentionPolicy{
		MaxAge:   24 * time.Hour,
		MaxCount: 0,
	}
	err = storage.Cleanup(ctx, policy)
	require.NoError(t, err)

	// Verify old deleted, recent kept
	_, err = storage.Retrieve(ctx, "old")
	assert.Error(t, err)

	_, err = storage.Retrieve(ctx, "recent")
	assert.NoError(t, err)
}

func TestJSONStorage_Cleanup_ByCount(t *testing.T) {
	tempDir := t.TempDir()
	storage, err := NewJSONStorageImpl(JSONConfig{
		Directory:   tempDir,
		Compression: false,
		Pretty:      false,
	})
	require.NoError(t, err)
	defer storage.Close()

	ctx := context.Background()

	// Store 5 snapshots
	for i := 1; i <= 5; i++ {
		snapshot := createTestSnapshot(fmt.Sprintf("snap-%d", i))
		metadata := createTestMetadata()
		metadata.Timestamp = time.Now().Add(time.Duration(i) * time.Minute)
		err = storage.Store(ctx, snapshot, metadata)
		require.NoError(t, err)
	}

	// Cleanup (keep only 3)
	policy := RetentionPolicy{
		MaxAge:   0,
		MaxCount: 3,
	}
	err = storage.Cleanup(ctx, policy)
	require.NoError(t, err)

	// Verify only 3 remain
	snapshots, err := storage.List(ctx, SnapshotFilter{})
	require.NoError(t, err)
	assert.Len(t, snapshots, 3)
}

func TestJSONStorage_Cleanup_KeepTagged(t *testing.T) {
	tempDir := t.TempDir()
	storage, err := NewJSONStorageImpl(JSONConfig{
		Directory:   tempDir,
		Compression: false,
		Pretty:      false,
	})
	require.NoError(t, err)
	defer storage.Close()

	ctx := context.Background()

	// Store old tagged snapshot
	tagged := createTestSnapshot("tagged")
	taggedMeta := createTestMetadata()
	taggedMeta.Timestamp = time.Now().Add(-48 * time.Hour)
	taggedMeta.Tags = map[string]string{"important": "true"}
	err = storage.Store(ctx, tagged, taggedMeta)
	require.NoError(t, err)

	// Store old untagged snapshot
	untagged := createTestSnapshot("untagged")
	untaggedMeta := createTestMetadata()
	untaggedMeta.Timestamp = time.Now().Add(-48 * time.Hour)
	err = storage.Store(ctx, untagged, untaggedMeta)
	require.NoError(t, err)

	// Cleanup with KeepTagged
	policy := RetentionPolicy{
		MaxAge:     24 * time.Hour,
		KeepTagged: true,
	}
	err = storage.Cleanup(ctx, policy)
	require.NoError(t, err)

	// Verify tagged kept, untagged deleted
	_, err = storage.Retrieve(ctx, "tagged")
	assert.NoError(t, err)

	_, err = storage.Retrieve(ctx, "untagged")
	assert.Error(t, err)
}

func TestJSONStorage_Cleanup_KeepReleases(t *testing.T) {
	tempDir := t.TempDir()
	storage, err := NewJSONStorageImpl(JSONConfig{
		Directory:   tempDir,
		Compression: false,
		Pretty:      false,
	})
	require.NoError(t, err)
	defer storage.Close()

	ctx := context.Background()

	// Store old release snapshot
	release := createTestSnapshot("release")
	releaseMeta := createTestMetadata()
	releaseMeta.Timestamp = time.Now().Add(-48 * time.Hour)
	releaseMeta.GitTag = "v1.0.0"
	err = storage.Store(ctx, release, releaseMeta)
	require.NoError(t, err)

	// Store old non-release snapshot
	nonRelease := createTestSnapshot("non-release")
	nonReleaseMeta := createTestMetadata()
	nonReleaseMeta.Timestamp = time.Now().Add(-48 * time.Hour)
	err = storage.Store(ctx, nonRelease, nonReleaseMeta)
	require.NoError(t, err)

	// Cleanup with KeepReleases
	policy := RetentionPolicy{
		MaxAge:       24 * time.Hour,
		KeepReleases: true,
	}
	err = storage.Cleanup(ctx, policy)
	require.NoError(t, err)

	// Verify release kept, non-release deleted
	_, err = storage.Retrieve(ctx, "release")
	assert.NoError(t, err)

	_, err = storage.Retrieve(ctx, "non-release")
	assert.Error(t, err)
}

func TestJSONStorage_Close(t *testing.T) {
	tempDir := t.TempDir()
	storage, err := NewJSONStorageImpl(JSONConfig{
		Directory:   tempDir,
		Compression: false,
		Pretty:      false,
	})
	require.NoError(t, err)

	// Close should succeed (no-op)
	err = storage.Close()
	assert.NoError(t, err)
}

// Helper functions

func createTestSnapshot(id string) metrics.Snapshot {
	return metrics.Snapshot{
		ID: id,
		Report: metrics.Report{
			Overview: metrics.OverviewMetrics{
				TotalFiles:       10,
				TotalLinesOfCode: 1000,
				TotalFunctions:   50,
			},
		},
	}
}

func createTestMetadata() metrics.SnapshotMetadata {
	return metrics.SnapshotMetadata{
		Timestamp:   time.Now(),
		GitCommit:   "abc123",
		GitBranch:   "main",
		GitTag:      "",
		Version:     "1.0.0",
		Author:      "test-user",
		Description: "Test snapshot",
		Tags:        make(map[string]string),
	}
}
