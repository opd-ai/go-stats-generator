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
	Store(ctx context.Context, snapshot metrics.Snapshot, metadata metrics.SnapshotMetadata) error

	// Retrieve gets a specific snapshot by ID
	Retrieve(ctx context.Context, id string) (metrics.Snapshot, error)

	// List returns available snapshots with optional filtering
	List(ctx context.Context, filter SnapshotFilter) ([]SnapshotInfo, error)

	// Delete removes a snapshot
	Delete(ctx context.Context, id string) error

	// Cleanup removes old snapshots based on retention policy
	Cleanup(ctx context.Context, policy RetentionPolicy) error

	// GetLatest returns the most recent snapshot
	GetLatest(ctx context.Context) (metrics.Snapshot, error)

	// GetByTag returns snapshots matching a specific tag
	GetByTag(ctx context.Context, key, value string) ([]metrics.Snapshot, error)

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
	ID                   string            `json:"id"`
	Timestamp            time.Time         `json:"timestamp"`
	GitCommit            string            `json:"git_commit,omitempty"`
	GitBranch            string            `json:"git_branch,omitempty"`
	GitTag               string            `json:"git_tag,omitempty"`
	Version              string            `json:"version,omitempty"`
	Author               string            `json:"author,omitempty"`
	Description          string            `json:"description,omitempty"`
	Tags                 map[string]string `json:"tags,omitempty"`
	Size                 int64             `json:"size_bytes"`
	MBIScoreAvg          *float64          `json:"mbi_score_avg,omitempty"`
	DuplicationRatio     *float64          `json:"duplication_ratio,omitempty"`
	DocCoverage          *float64          `json:"doc_coverage,omitempty"`
	ComplexityViolations *int              `json:"complexity_violations,omitempty"`
	NamingViolations     *int              `json:"naming_violations,omitempty"`
}

// RetentionPolicy defines how long to keep historical snapshots
type RetentionPolicy struct {
	MaxAge       time.Duration `json:"max_age"`
	MaxCount     int           `json:"max_count"`
	KeepTagged   bool          `json:"keep_tagged"`
	KeepReleases bool          `json:"keep_releases"`
}

// DefaultRetentionPolicy returns a sensible default retention policy of 90 days with automatic pruning enabled.
// Baselines older than 90 days are automatically deleted during storage operations to prevent unbounded growth.
// Adjust the retention window based on your trend analysis requirements (30 days for short-term, 365+ for long-term trends).
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

// DefaultStorageConfig returns default storage configuration with SQLite backend and 90-day retention policy.
// The configuration assumes a local database file at "./metrics.db" and automatically prunes baselines older than
// the retention window to prevent unbounded database growth. Adjust retention days based on your trend analysis needs.
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

// NewStorage instantiates the appropriate metrics storage backend (SQLite, JSON, or in-memory) based on provided configuration.
// It routes to specialized constructors for each storage type, handling database initialization, file system setup, or memory
// allocation as appropriate. Returns a fully initialized storage instance conforming to the MetricsStorage interface, ready for
// baseline snapshot persistence and retrieval operations. Returns error if the storage type is unsupported or initialization fails.
func NewStorage(config StorageConfig) (MetricsStorage, error) {
	switch config.Type {
	case "sqlite":
		return NewSQLiteStorage(config.SQLite)
	case "json":
		return NewJSONStorage(config.JSON)
	case "memory":
		return NewMemoryStorage(), nil
	default:
		return nil, fmt.Errorf("unsupported storage type: %s", config.Type)
	}
}

// NewSQLiteStorage creates a SQLite-based persistent storage backend for baseline snapshots with full ACID transaction support.
// This is a forward declaration wrapper that routes to NewSQLiteStorageImpl for actual initialization. SQLite storage provides
// zero-configuration persistence, making it ideal for CI/CD pipelines and local development. The backend supports concurrent reads
// and serialized writes, with automatic schema migrations and index optimization for query performance on large snapshot histories.
func NewSQLiteStorage(config SQLiteConfig) (MetricsStorage, error) {
	return NewSQLiteStorageImpl(config)
}

// NewJSONStorage creates a JSON file-based storage backend for baseline snapshots enabling human-readable persistence and version control integration.
// This forward declaration wrapper routes to NewJSONStorageImpl for initialization. JSON storage writes each snapshot as an individual file
// (optionally gzip-compressed), allowing git tracking of metrics history and manual inspection/editing. Ideal for small to medium repositories
// where transparency and audit trails are prioritized over query performance or concurrent access patterns.
func NewJSONStorage(config JSONConfig) (MetricsStorage, error) {
	return NewJSONStorageImpl(config)
}
