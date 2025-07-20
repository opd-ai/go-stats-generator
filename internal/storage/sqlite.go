package storage

import (
	"compress/gzip"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/opd-ai/go-stats-generator/internal/metrics"
	_ "modernc.org/sqlite"
)

// SQLiteStorage implements MetricsStorage using SQLite
type SQLiteStorage struct {
	db     *sql.DB
	config SQLiteConfig
}

// NewSQLiteStorageImpl creates a new SQLite storage instance
func NewSQLiteStorageImpl(config SQLiteConfig) (*SQLiteStorage, error) {
	// Create directory if it doesn't exist
	if err := createDirIfNotExists(config.Path); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	// Open database connection
	db, err := sql.Open("sqlite", config.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure database
	db.SetMaxOpenConns(config.MaxConnections)
	db.SetMaxIdleConns(config.MaxConnections / 2)

	storage := &SQLiteStorage{
		db:     db,
		config: config,
	}

	// Apply configuration
	if err := storage.configure(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to configure database: %w", err)
	}

	// Initialize schema
	if err := storage.initSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return storage, nil
}

// configure applies SQLite-specific configuration
func (s *SQLiteStorage) configure() error {
	ctx := context.Background()

	// Enable WAL mode for better concurrency
	if s.config.EnableWAL {
		if _, err := s.db.ExecContext(ctx, "PRAGMA journal_mode=WAL"); err != nil {
			return fmt.Errorf("failed to enable WAL mode: %w", err)
		}
	}

	// Enable foreign keys
	if s.config.EnableFK {
		if _, err := s.db.ExecContext(ctx, "PRAGMA foreign_keys=ON"); err != nil {
			return fmt.Errorf("failed to enable foreign keys: %w", err)
		}
	}

	// Set synchronous mode for performance
	if _, err := s.db.ExecContext(ctx, "PRAGMA synchronous=NORMAL"); err != nil {
		return fmt.Errorf("failed to set synchronous mode: %w", err)
	}

	return nil
}

// initSchema creates the necessary tables if they don't exist
func (s *SQLiteStorage) initSchema() error {
	ctx := context.Background()

	// Create snapshots table
	createSnapshotsTable := `
	CREATE TABLE IF NOT EXISTS snapshots (
		id TEXT PRIMARY KEY,
		timestamp DATETIME NOT NULL,
		git_commit TEXT,
		git_branch TEXT,
		git_tag TEXT,
		version TEXT,
		author TEXT,
		description TEXT,
		size_bytes INTEGER NOT NULL,
		data_compressed BLOB NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	if _, err := s.db.ExecContext(ctx, createSnapshotsTable); err != nil {
		return fmt.Errorf("failed to create snapshots table: %w", err)
	}

	// Create tags table for key-value metadata
	createTagsTable := `
	CREATE TABLE IF NOT EXISTS snapshot_tags (
		snapshot_id TEXT NOT NULL,
		key TEXT NOT NULL,
		value TEXT NOT NULL,
		PRIMARY KEY (snapshot_id, key),
		FOREIGN KEY (snapshot_id) REFERENCES snapshots(id) ON DELETE CASCADE
	);`

	if _, err := s.db.ExecContext(ctx, createTagsTable); err != nil {
		return fmt.Errorf("failed to create tags table: %w", err)
	}

	// Create indexes for performance
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_snapshots_timestamp ON snapshots(timestamp)",
		"CREATE INDEX IF NOT EXISTS idx_snapshots_branch ON snapshots(git_branch)",
		"CREATE INDEX IF NOT EXISTS idx_snapshots_tag ON snapshots(git_tag)",
		"CREATE INDEX IF NOT EXISTS idx_tags_key ON snapshot_tags(key)",
		"CREATE INDEX IF NOT EXISTS idx_tags_value ON snapshot_tags(value)",
	}

	for _, indexSQL := range indexes {
		if _, err := s.db.ExecContext(ctx, indexSQL); err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}

	return nil
}

// Store saves a metrics snapshot with metadata
func (s *SQLiteStorage) Store(ctx context.Context, snapshot metrics.MetricsSnapshot, metadata metrics.SnapshotMetadata) error {
	// Serialize the snapshot data
	data, err := json.Marshal(snapshot.Report)
	if err != nil {
		return fmt.Errorf("failed to marshal snapshot data: %w", err)
	}

	// Compress if enabled
	var compressedData []byte
	if s.config.EnableCompression {
		compressedData, err = compress(data)
		if err != nil {
			return fmt.Errorf("failed to compress data: %w", err)
		}
	} else {
		compressedData = data
	}

	// Start transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert snapshot
	insertSnapshot := `
	INSERT INTO snapshots (
		id, timestamp, git_commit, git_branch, git_tag, version, 
		author, description, size_bytes, data_compressed
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err = tx.ExecContext(ctx, insertSnapshot,
		snapshot.ID,
		metadata.Timestamp,
		metadata.GitCommit,
		metadata.GitBranch,
		metadata.GitTag,
		metadata.Version,
		metadata.Author,
		metadata.Description,
		len(compressedData),
		compressedData,
	)
	if err != nil {
		return fmt.Errorf("failed to insert snapshot: %w", err)
	}

	// Insert tags
	if len(metadata.Tags) > 0 {
		insertTag := "INSERT INTO snapshot_tags (snapshot_id, key, value) VALUES (?, ?, ?)"
		for key, value := range metadata.Tags {
			_, err = tx.ExecContext(ctx, insertTag, snapshot.ID, key, value)
			if err != nil {
				return fmt.Errorf("failed to insert tag %s: %w", key, err)
			}
		}
	}

	return tx.Commit()
}

// Retrieve gets a specific snapshot by ID
func (s *SQLiteStorage) Retrieve(ctx context.Context, id string) (metrics.MetricsSnapshot, error) {
	var snapshot metrics.MetricsSnapshot
	var metadata metrics.SnapshotMetadata
	var compressedData []byte
	var gitCommit, gitBranch, gitTag, version, author, description sql.NullString

	// Get snapshot data
	query := `
	SELECT id, timestamp, git_commit, git_branch, git_tag, version, 
		   author, description, data_compressed
	FROM snapshots WHERE id = ?`

	row := s.db.QueryRowContext(ctx, query, id)
	err := row.Scan(
		&snapshot.ID,
		&metadata.Timestamp,
		&gitCommit,
		&gitBranch,
		&gitTag,
		&version,
		&author,
		&description,
		&compressedData,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return snapshot, fmt.Errorf("snapshot not found: %s", id)
		}
		return snapshot, fmt.Errorf("failed to retrieve snapshot: %w", err)
	}

	// Handle nullable fields
	metadata.GitCommit = gitCommit.String
	metadata.GitBranch = gitBranch.String
	metadata.GitTag = gitTag.String
	metadata.Version = version.String
	metadata.Author = author.String
	metadata.Description = description.String

	// Get tags
	tags := make(map[string]string)
	tagQuery := "SELECT key, value FROM snapshot_tags WHERE snapshot_id = ?"
	rows, err := s.db.QueryContext(ctx, tagQuery, id)
	if err != nil {
		return snapshot, fmt.Errorf("failed to retrieve tags: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return snapshot, fmt.Errorf("failed to scan tag: %w", err)
		}
		tags[key] = value
	}
	metadata.Tags = tags

	// Decompress data
	var data []byte
	if s.config.EnableCompression {
		data, err = decompress(compressedData)
		if err != nil {
			return snapshot, fmt.Errorf("failed to decompress data: %w", err)
		}
	} else {
		data = compressedData
	}

	// Unmarshal report data
	if err := json.Unmarshal(data, &snapshot.Report); err != nil {
		return snapshot, fmt.Errorf("failed to unmarshal report data: %w", err)
	}

	snapshot.Metadata = metadata
	return snapshot, nil
}

// List returns available snapshots with optional filtering
func (s *SQLiteStorage) List(ctx context.Context, filter SnapshotFilter) ([]SnapshotInfo, error) {
	var snapshots []SnapshotInfo

	// Build query with filters
	query := "SELECT id, timestamp, git_commit, git_branch, git_tag, version, author, description, size_bytes FROM snapshots WHERE 1=1"
	var args []interface{}
	argIndex := 0

	if filter.After != nil {
		query += " AND timestamp > ?"
		args = append(args, *filter.After)
		argIndex++
	}

	if filter.Before != nil {
		query += " AND timestamp < ?"
		args = append(args, *filter.Before)
		argIndex++
	}

	if filter.Branch != "" {
		query += " AND git_branch = ?"
		args = append(args, filter.Branch)
		argIndex++
	}

	if filter.Tag != "" {
		query += " AND git_tag = ?"
		args = append(args, filter.Tag)
		argIndex++
	}

	if filter.Author != "" {
		query += " AND author = ?"
		args = append(args, filter.Author)
		argIndex++
	}

	// Handle custom tags
	if len(filter.Tags) > 0 {
		for key, value := range filter.Tags {
			query += " AND id IN (SELECT snapshot_id FROM snapshot_tags WHERE key = ? AND value = ?)"
			args = append(args, key, value)
			argIndex += 2
		}
	}

	// Add ordering and limits
	query += " ORDER BY timestamp DESC"
	if filter.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filter.Limit)
		if filter.Offset > 0 {
			query += " OFFSET ?"
			args = append(args, filter.Offset)
		}
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list snapshots: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var info SnapshotInfo
		var gitCommit, gitBranch, gitTag, version, author, description sql.NullString

		err := rows.Scan(
			&info.ID,
			&info.Timestamp,
			&gitCommit,
			&gitBranch,
			&gitTag,
			&version,
			&author,
			&description,
			&info.Size,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan snapshot: %w", err)
		}

		// Handle nullable fields
		info.GitCommit = gitCommit.String
		info.GitBranch = gitBranch.String
		info.GitTag = gitTag.String
		info.Version = version.String
		info.Author = author.String
		info.Description = description.String

		// Get tags for this snapshot
		info.Tags = make(map[string]string)
		tagQuery := "SELECT key, value FROM snapshot_tags WHERE snapshot_id = ?"
		tagRows, err := s.db.QueryContext(ctx, tagQuery, info.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve tags for snapshot %s: %w", info.ID, err)
		}

		for tagRows.Next() {
			var key, value string
			if err := tagRows.Scan(&key, &value); err != nil {
				tagRows.Close()
				return nil, fmt.Errorf("failed to scan tag: %w", err)
			}
			info.Tags[key] = value
		}
		tagRows.Close()

		snapshots = append(snapshots, info)
	}

	return snapshots, nil
}

// Delete removes a snapshot
func (s *SQLiteStorage) Delete(ctx context.Context, id string) error {
	// Delete will cascade to tags due to foreign key constraint
	result, err := s.db.ExecContext(ctx, "DELETE FROM snapshots WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete snapshot: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("snapshot not found: %s", id)
	}

	return nil
}

// Cleanup removes old snapshots based on retention policy
func (s *SQLiteStorage) Cleanup(ctx context.Context, policy RetentionPolicy) error {
	var deletedCount int64

	// Delete by age
	if policy.MaxAge > 0 {
		cutoff := time.Now().Add(-policy.MaxAge)
		query := "DELETE FROM snapshots WHERE timestamp < ?"

		if policy.KeepTagged {
			query += " AND id NOT IN (SELECT DISTINCT snapshot_id FROM snapshot_tags)"
		}

		if policy.KeepReleases {
			query += " AND (git_tag IS NULL OR git_tag = '')"
		}

		result, err := s.db.ExecContext(ctx, query, cutoff)
		if err != nil {
			return fmt.Errorf("failed to delete old snapshots: %w", err)
		}

		if affected, err := result.RowsAffected(); err == nil {
			deletedCount += affected
		}
	}

	// Delete by count (keep only the most recent N)
	if policy.MaxCount > 0 {
		countQuery := "SELECT COUNT(*) FROM snapshots"
		var currentCount int
		if err := s.db.QueryRowContext(ctx, countQuery).Scan(&currentCount); err != nil {
			return fmt.Errorf("failed to count snapshots: %w", err)
		}

		if currentCount > policy.MaxCount {
			toDelete := currentCount - policy.MaxCount
			query := `DELETE FROM snapshots WHERE id IN (
				SELECT id FROM snapshots 
				ORDER BY timestamp ASC 
				LIMIT ?
			)`

			if policy.KeepTagged {
				query = `DELETE FROM snapshots WHERE id IN (
					SELECT id FROM snapshots 
					WHERE id NOT IN (SELECT DISTINCT snapshot_id FROM snapshot_tags)
					ORDER BY timestamp ASC 
					LIMIT ?
				)`
			}

			result, err := s.db.ExecContext(ctx, query, toDelete)
			if err != nil {
				return fmt.Errorf("failed to delete excess snapshots: %w", err)
			}

			if affected, err := result.RowsAffected(); err == nil {
				deletedCount += affected
			}
		}
	}

	// Log cleanup results (would normally use a logger)
	if deletedCount > 0 {
		fmt.Printf("Cleaned up %d old snapshots\n", deletedCount)
	}

	return nil
}

// GetLatest returns the most recent snapshot
func (s *SQLiteStorage) GetLatest(ctx context.Context) (metrics.MetricsSnapshot, error) {
	var id string
	row := s.db.QueryRowContext(ctx, "SELECT id FROM snapshots ORDER BY timestamp DESC LIMIT 1")
	if err := row.Scan(&id); err != nil {
		if err == sql.ErrNoRows {
			return metrics.MetricsSnapshot{}, fmt.Errorf("no snapshots found")
		}
		return metrics.MetricsSnapshot{}, fmt.Errorf("failed to get latest snapshot: %w", err)
	}

	return s.Retrieve(ctx, id)
}

// GetByTag returns snapshots matching a specific tag
func (s *SQLiteStorage) GetByTag(ctx context.Context, key, value string) ([]metrics.MetricsSnapshot, error) {
	filter := SnapshotFilter{
		Tags: map[string]string{key: value},
	}

	infos, err := s.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	var snapshots []metrics.MetricsSnapshot
	for _, info := range infos {
		snapshot, err := s.Retrieve(ctx, info.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve snapshot %s: %w", info.ID, err)
		}
		snapshots = append(snapshots, snapshot)
	}

	return snapshots, nil
}

// Close releases storage resources
func (s *SQLiteStorage) Close() error {
	return s.db.Close()
}

// Helper functions

func createDirIfNotExists(path string) error {
	dir := filepath.Dir(path)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return os.MkdirAll(dir, 0755)
	}
	return nil
}

func compress(data []byte) ([]byte, error) {
	var buf strings.Builder
	gz := gzip.NewWriter(&buf)
	if _, err := gz.Write(data); err != nil {
		return nil, err
	}
	if err := gz.Close(); err != nil {
		return nil, err
	}
	return []byte(buf.String()), nil
}

func decompress(data []byte) ([]byte, error) {
	reader, err := gzip.NewReader(strings.NewReader(string(data)))
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	return io.ReadAll(reader)
}
