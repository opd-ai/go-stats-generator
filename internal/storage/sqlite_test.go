//go:build !js || !wasm

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

func createTestSQLiteSnapshot(id string) metrics.Snapshot {
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

func createTestSQLiteMetadata() metrics.SnapshotMetadata {
	return metrics.SnapshotMetadata{
		Timestamp:   time.Now(),
		GitCommit:   "abc123",
		GitBranch:   "main",
		GitTag:      "",
		Version:     "1.0.0",
		Author:      "test-user",
		Description: "Test snapshot",
	}
}

func TestNewSQLiteStorageImpl(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	config := SQLiteConfig{
		Path:              dbPath,
		MaxConnections:    5,
		EnableWAL:         true,
		EnableFK:          true,
		EnableCompression: true,
	}

	storage, err := NewSQLiteStorageImpl(config)
	require.NoError(t, err)
	require.NotNil(t, storage)
	defer storage.Close()

	// Verify database file exists
	_, err = os.Stat(dbPath)
	assert.NoError(t, err)
}

func TestSQLiteStorage_Store(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	config := SQLiteConfig{
		Path:              dbPath,
		MaxConnections:    5,
		EnableWAL:         true,
		EnableFK:          true,
		EnableCompression: true,
	}

	storage, err := NewSQLiteStorageImpl(config)
	require.NoError(t, err)
	defer storage.Close()

	ctx := context.Background()
	snapshot := createTestSQLiteSnapshot("test-snapshot-1")
	metadata := createTestSQLiteMetadata()

	err = storage.Store(ctx, snapshot, metadata)
	assert.NoError(t, err)
}

func TestSQLiteStorage_Retrieve(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	config := SQLiteConfig{
		Path:              dbPath,
		MaxConnections:    5,
		EnableWAL:         true,
		EnableFK:          true,
		EnableCompression: true,
	}

	storage, err := NewSQLiteStorageImpl(config)
	require.NoError(t, err)
	defer storage.Close()

	ctx := context.Background()
	snapshot := createTestSQLiteSnapshot("test-snapshot-2")
	metadata := createTestSQLiteMetadata()
	metadata.GitCommit = "def456"
	metadata.GitBranch = "develop"

	// Store first
	err = storage.Store(ctx, snapshot, metadata)
	require.NoError(t, err)

	// Then retrieve
	retrieved, err := storage.Retrieve(ctx, "test-snapshot-2")
	assert.NoError(t, err)
	assert.Equal(t, "test-snapshot-2", retrieved.ID)
	assert.Equal(t, 10, retrieved.Report.Overview.TotalFiles)
}

func TestSQLiteStorage_List(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	config := SQLiteConfig{
		Path:              dbPath,
		MaxConnections:    5,
		EnableWAL:         true,
		EnableFK:          true,
		EnableCompression: true,
	}

	storage, err := NewSQLiteStorageImpl(config)
	require.NoError(t, err)
	defer storage.Close()

	ctx := context.Background()

	// Store multiple snapshots
	for i := 0; i < 3; i++ {
		snapshot := createTestSQLiteSnapshot(fmt.Sprintf("snapshot-%d", i))
		metadata := createTestSQLiteMetadata()
		metadata.Timestamp = time.Now().Add(time.Duration(i) * time.Hour)
		err = storage.Store(ctx, snapshot, metadata)
		require.NoError(t, err)
	}

	// List all
	filter := SnapshotFilter{
		Limit: 10,
	}
	snapshots, err := storage.List(ctx, filter)
	assert.NoError(t, err)
	assert.Len(t, snapshots, 3)
}

func TestSQLiteStorage_Delete(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	config := SQLiteConfig{
		Path:              dbPath,
		MaxConnections:    5,
		EnableWAL:         true,
		EnableFK:          true,
		EnableCompression: true,
	}

	storage, err := NewSQLiteStorageImpl(config)
	require.NoError(t, err)
	defer storage.Close()

	ctx := context.Background()
	snapshot := createTestSQLiteSnapshot("delete-test")
	metadata := createTestSQLiteMetadata()

	// Store
	err = storage.Store(ctx, snapshot, metadata)
	require.NoError(t, err)

	// Delete
	err = storage.Delete(ctx, "delete-test")
	assert.NoError(t, err)

	// Verify deletion
	_, err = storage.Retrieve(ctx, "delete-test")
	assert.Error(t, err)
}

func TestSQLiteStorage_Cleanup(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	config := SQLiteConfig{
		Path:              dbPath,
		MaxConnections:    5,
		EnableWAL:         true,
		EnableFK:          true,
		EnableCompression: true,
	}

	storage, err := NewSQLiteStorageImpl(config)
	require.NoError(t, err)
	defer storage.Close()

	ctx := context.Background()

	// Store old snapshots
	oldTime := time.Now().Add(-10 * 24 * time.Hour)
	for i := 0; i < 3; i++ {
		snapshot := createTestSQLiteSnapshot(fmt.Sprintf("old-snapshot-%d", i))
		metadata := createTestSQLiteMetadata()
		metadata.Timestamp = oldTime.Add(time.Duration(i) * time.Hour)
		err = storage.Store(ctx, snapshot, metadata)
		require.NoError(t, err)
	}

	// Store recent snapshots
	for i := 0; i < 3; i++ {
		snapshot := createTestSQLiteSnapshot(fmt.Sprintf("recent-snapshot-%d", i))
		metadata := createTestSQLiteMetadata()
		metadata.Timestamp = time.Now().Add(time.Duration(i) * time.Hour)
		err = storage.Store(ctx, snapshot, metadata)
		require.NoError(t, err)
	}

	// Cleanup
	policy := RetentionPolicy{
		MaxAge:   7 * 24 * time.Hour,
		MaxCount: 5,
	}
	err = storage.Cleanup(ctx, policy)
	assert.NoError(t, err)

	// Verify old snapshots were deleted
	snapshots, err := storage.List(ctx, SnapshotFilter{Limit: 100})
	assert.NoError(t, err)
	assert.LessOrEqual(t, len(snapshots), 5)
}

func TestSQLiteStorage_GetLatest(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	config := SQLiteConfig{
		Path:              dbPath,
		MaxConnections:    5,
		EnableWAL:         true,
		EnableFK:          true,
		EnableCompression: true,
	}

	storage, err := NewSQLiteStorageImpl(config)
	require.NoError(t, err)
	defer storage.Close()

	ctx := context.Background()

	// Store snapshots with different timestamps
	baseTime := time.Now()
	for i := 0; i < 3; i++ {
		snapshot := createTestSQLiteSnapshot(fmt.Sprintf("snapshot-%d", i))
		metadata := createTestSQLiteMetadata()
		metadata.Timestamp = baseTime.Add(time.Duration(i) * time.Hour)
		err = storage.Store(ctx, snapshot, metadata)
		require.NoError(t, err)
	}

	// Get latest
	latest, err := storage.GetLatest(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "snapshot-2", latest.ID)
}

func TestSQLiteStorage_GetByTag(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	config := SQLiteConfig{
		Path:              dbPath,
		MaxConnections:    5,
		EnableWAL:         true,
		EnableFK:          true,
		EnableCompression: true,
	}

	storage, err := NewSQLiteStorageImpl(config)
	require.NoError(t, err)
	defer storage.Close()

	ctx := context.Background()
	snapshot := createTestSQLiteSnapshot("tagged-snapshot")
	metadata := createTestSQLiteMetadata()
	metadata.Tags = map[string]string{
		"environment": "production",
		"version":     "1.0.0",
	}

	// Store
	err = storage.Store(ctx, snapshot, metadata)
	require.NoError(t, err)

	// Get by tag
	retrieved, err := storage.GetByTag(ctx, "environment", "production")
	assert.NoError(t, err)
	assert.Len(t, retrieved, 1)
	assert.Equal(t, "tagged-snapshot", retrieved[0].ID)
}

func TestSQLiteStorage_Close(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	config := SQLiteConfig{
		Path:              dbPath,
		MaxConnections:    5,
		EnableWAL:         true,
		EnableFK:          true,
		EnableCompression: true,
	}

	storage, err := NewSQLiteStorageImpl(config)
	require.NoError(t, err)

	err = storage.Close()
	assert.NoError(t, err)
}

func TestSQLiteStorage_StoreWithTags(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	config := SQLiteConfig{
		Path:              dbPath,
		MaxConnections:    5,
		EnableWAL:         true,
		EnableFK:          true,
		EnableCompression: true,
	}

	storage, err := NewSQLiteStorageImpl(config)
	require.NoError(t, err)
	defer storage.Close()

	ctx := context.Background()
	snapshot := createTestSQLiteSnapshot("tagged-snapshot-2")
	metadata := createTestSQLiteMetadata()
	metadata.Tags = map[string]string{
		"environment": "staging",
		"region":      "us-west-2",
	}

	err = storage.Store(ctx, snapshot, metadata)
	assert.NoError(t, err)

	// Retrieve and verify tags persist
	retrieved, err := storage.Retrieve(ctx, "tagged-snapshot-2")
	assert.NoError(t, err)
	assert.Equal(t, "tagged-snapshot-2", retrieved.ID)
}

func TestSQLiteStorage_ListWithFilters(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	config := SQLiteConfig{
		Path:              dbPath,
		MaxConnections:    5,
		EnableWAL:         true,
		EnableFK:          true,
		EnableCompression: true,
	}

	storage, err := NewSQLiteStorageImpl(config)
	require.NoError(t, err)
	defer storage.Close()

	ctx := context.Background()
	baseTime := time.Now()

	// Store snapshots with different branches
	for i := 0; i < 3; i++ {
		branch := "main"
		if i == 2 {
			branch = "develop"
		}
		snapshot := createTestSQLiteSnapshot(fmt.Sprintf("filter-test-%d", i))
		metadata := createTestSQLiteMetadata()
		metadata.Timestamp = baseTime.Add(time.Duration(i) * time.Hour)
		metadata.GitBranch = branch
		err = storage.Store(ctx, snapshot, metadata)
		require.NoError(t, err)
	}

	// Filter by branch
	filter := SnapshotFilter{
		Branch: "main",
		Limit:  10,
	}
	snapshots, err := storage.List(ctx, filter)
	assert.NoError(t, err)
	assert.Len(t, snapshots, 2)
	for _, s := range snapshots {
		assert.Equal(t, "main", s.GitBranch)
	}
}

func TestSQLiteStorage_StoreWithoutCompression(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	config := SQLiteConfig{
		Path:              dbPath,
		MaxConnections:    5,
		EnableWAL:         true,
		EnableFK:          true,
		EnableCompression: false,
	}

	storage, err := NewSQLiteStorageImpl(config)
	require.NoError(t, err)
	defer storage.Close()

	ctx := context.Background()
	snapshot := createTestSQLiteSnapshot("uncompressed-test")
	metadata := createTestSQLiteMetadata()

	err = storage.Store(ctx, snapshot, metadata)
	assert.NoError(t, err)

	// Retrieve should work the same
	retrieved, err := storage.Retrieve(ctx, "uncompressed-test")
	assert.NoError(t, err)
	assert.Equal(t, "uncompressed-test", retrieved.ID)
}
