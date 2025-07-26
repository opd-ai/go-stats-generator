package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/opd-ai/go-stats-generator/internal/metrics"
)

// MetricsStorage defines the interface for storing and retrieving historical metrics
type MetricsStorage interface {
	// Store saves a metrics snapshot with metadata
	Store(ctx context.Context, snapshot metrics.MetricsSnapshot, metadata metrics.SnapshotMetadata) error

	// Retrieve gets a specific snapshot by ID
	Retrieve(ctx context.Context, id string) (metrics.MetricsSnapshot, error)

	// List returns available snapshots with optional filtering
	List(ctx context.Context, filter SnapshotFilter) ([]SnapshotInfo, error)

	// Delete removes a snapshot
	Delete(ctx context.Context, id string) error

	// Cleanup removes old snapshots based on retention policy
	Cleanup(ctx context.Context, policy RetentionPolicy) error

	// GetLatest returns the most recent snapshot
	GetLatest(ctx context.Context) (metrics.MetricsSnapshot, error)

	// GetByTag returns snapshots matching a specific tag
	GetByTag(ctx context.Context, key, value string) ([]metrics.MetricsSnapshot, error)

	// Close releases storage resources
	Close() error
}

// SnapshotFilter defines filtering criteria for listing snapshots
type SnapshotFilter struct {
	After  *time.Time        `json:"after,omitempty"`
	Before *time.Time        `json:"before,omitempty"`
	Branch string            `json:"branch,omitempty"`
	Tag    string            `json:"tag,omitempty"`
	Author string            `json:"author,omitempty"`
	Tags   map[string]string `json:"tags,omitempty"`
	Limit  int               `json:"limit,omitempty"`
	Offset int               `json:"offset,omitempty"`
}

// SnapshotInfo provides summary information about a stored snapshot
type SnapshotInfo struct {
	ID          string            `json:"id"`
	Timestamp   time.Time         `json:"timestamp"`
	GitCommit   string            `json:"git_commit,omitempty"`
	GitBranch   string            `json:"git_branch,omitempty"`
	GitTag      string            `json:"git_tag,omitempty"`
	Version     string            `json:"version,omitempty"`
	Author      string            `json:"author,omitempty"`
	Description string            `json:"description,omitempty"`
	Tags        map[string]string `json:"tags,omitempty"`
	Size        int64             `json:"size_bytes"`
}

// RetentionPolicy defines how long to keep historical snapshots
type RetentionPolicy struct {
	MaxAge       time.Duration `json:"max_age"`
	MaxCount     int           `json:"max_count"`
	KeepTagged   bool          `json:"keep_tagged"`
	KeepReleases bool          `json:"keep_releases"`
}

// DefaultRetentionPolicy returns a sensible default retention policy
func DefaultRetentionPolicy() RetentionPolicy {
	return RetentionPolicy{
		MaxAge:       30 * 24 * time.Hour, // 30 days
		MaxCount:     100,                 // 100 snapshots
		KeepTagged:   true,                // Keep tagged snapshots
		KeepReleases: true,                // Keep release snapshots
	}
}

// StorageConfig defines configuration for different storage backends
type StorageConfig struct {
	Type   string       `yaml:"type" json:"type"`
	SQLite SQLiteConfig `yaml:"sqlite" json:"sqlite"`
	JSON   JSONConfig   `yaml:"json" json:"json"`
}

// SQLiteConfig defines SQLite-specific configuration
type SQLiteConfig struct {
	Path              string `yaml:"path" json:"path"`
	MaxConnections    int    `yaml:"max_connections" json:"max_connections"`
	EnableWAL         bool   `yaml:"enable_wal" json:"enable_wal"`
	EnableFK          bool   `yaml:"enable_foreign_keys" json:"enable_foreign_keys"`
	EnableCompression bool   `yaml:"enable_compression" json:"enable_compression"`
}

// JSONConfig defines JSON file storage configuration
type JSONConfig struct {
	Directory   string `yaml:"directory" json:"directory"`
	Compression bool   `yaml:"compression" json:"compression"`
	Pretty      bool   `yaml:"pretty" json:"pretty"`
}

// DefaultStorageConfig returns default storage configuration
func DefaultStorageConfig() StorageConfig {
	return StorageConfig{
		Type: "sqlite",
		SQLite: SQLiteConfig{
			Path:              ".go-stats-generator/metrics.db",
			MaxConnections:    10,
			EnableWAL:         true,
			EnableFK:          true,
			EnableCompression: true,
		},
		JSON: JSONConfig{
			Directory:   ".go-stats-generator/snapshots",
			Compression: true,
			Pretty:      false,
		},
	}
}

// NewStorage creates a new storage instance based on configuration
func NewStorage(config StorageConfig) (MetricsStorage, error) {
	switch config.Type {
	case "sqlite":
		return NewSQLiteStorage(config.SQLite)
	case "json":
		return NewJSONStorage(config.JSON)
	default:
		return nil, fmt.Errorf("unsupported storage type: %s", config.Type)
	}
}

// NewSQLiteStorage creates a new SQLite storage backend (forward declaration)
func NewSQLiteStorage(config SQLiteConfig) (MetricsStorage, error) {
	return NewSQLiteStorageImpl(config)
}

// NewJSONStorage creates a new JSON file storage backend (forward declaration)
func NewJSONStorage(config JSONConfig) (MetricsStorage, error) {
	// Implementation will be in json.go
	return nil, fmt.Errorf("JSON storage not yet implemented")
}
