//go:build !js || !wasm

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

// NewSQLiteStorageImpl creates a new SQLite storage instance with persistent database file for baseline retention.
// Automatically creates the database directory if it doesn't exist and initializes schema with compression enabled.
// Provides ACID guarantees for metric snapshots, making it suitable for production trend analysis and baseline management.
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
		mbi_score_avg REAL,
		duplication_ratio REAL,
		doc_coverage REAL,
		complexity_violations INTEGER,
		naming_violations INTEGER,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	if _, err := s.db.ExecContext(ctx, createSnapshotsTable); err != nil {
		return fmt.Errorf("failed to create snapshots table: %w", err)
	}

	// Migrate existing schema if burden columns don't exist
	if err := s.migrateSchema(ctx); err != nil {
		return fmt.Errorf("failed to migrate schema: %w", err)
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

// migrateSchema adds burden metric columns to existing databases
func (s *SQLiteStorage) migrateSchema(ctx context.Context) error {
	// Check if snapshots table exists
	var tableExists int
	err := s.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='snapshots'").Scan(&tableExists)
	if err != nil || tableExists == 0 {
		// Table doesn't exist yet, no migration needed
		return nil
	}

	columns := []string{
		"ALTER TABLE snapshots ADD COLUMN mbi_score_avg REAL",
		"ALTER TABLE snapshots ADD COLUMN duplication_ratio REAL",
		"ALTER TABLE snapshots ADD COLUMN doc_coverage REAL",
		"ALTER TABLE snapshots ADD COLUMN complexity_violations INTEGER",
		"ALTER TABLE snapshots ADD COLUMN naming_violations INTEGER",
	}

	for _, columnSQL := range columns {
		if _, err := s.db.ExecContext(ctx, columnSQL); err != nil {
			// Ignore "duplicate column" errors (already migrated)
			if !strings.Contains(err.Error(), "duplicate column") {
				return err
			}
		}
	}

	return nil
}

// extractBurdenMetrics calculates summary metrics from a Report for storage
func extractBurdenMetrics(report metrics.Report) (mbiAvg, dupRatio, docCov float64, complexViolations, namingViolations int) {
	// Calculate average MBI score across all files
	if len(report.Scores.FileScores) > 0 {
		var totalMBI float64
		for _, fs := range report.Scores.FileScores {
			totalMBI += fs.Score
		}
		mbiAvg = totalMBI / float64(len(report.Scores.FileScores))
	}

	// Extract duplication ratio
	dupRatio = report.Duplication.DuplicationRatio

	// Extract documentation coverage
	docCov = report.Documentation.Coverage.Overall

	// Count complexity violations (functions with overall complexity > 10)
	for _, fn := range report.Functions {
		if fn.Complexity.Overall > 10.0 {
			complexViolations++
		}
	}

	// Count naming violations
	namingViolations = report.Naming.FileNameViolations +
		report.Naming.IdentifierViolations +
		report.Naming.PackageNameViolations

	return mbiAvg, dupRatio, docCov, complexViolations, namingViolations
}

// Store saves a metrics snapshot with metadata
func (s *SQLiteStorage) Store(ctx context.Context, snapshot metrics.MetricsSnapshot, metadata metrics.SnapshotMetadata) error {
	compressedData, err := s.prepareSnapshotData(snapshot.Report)
	if err != nil {
		return err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	if err := s.insertSnapshotRecord(ctx, tx, snapshot, metadata, compressedData); err != nil {
		return err
	}

	if err := s.insertSnapshotTags(ctx, tx, snapshot.ID, metadata.Tags); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *SQLiteStorage) prepareSnapshotData(report metrics.Report) ([]byte, error) {
	data, err := json.Marshal(report)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal snapshot data: %w", err)
	}

	if s.config.EnableCompression {
		compressedData, err := compress(data)
		if err != nil {
			return nil, fmt.Errorf("failed to compress data: %w", err)
		}
		return compressedData, nil
	}

	return data, nil
}

func (s *SQLiteStorage) insertSnapshotRecord(ctx context.Context, tx *sql.Tx, snapshot metrics.MetricsSnapshot,
	metadata metrics.SnapshotMetadata, compressedData []byte,
) error {
	mbiAvg, dupRatio, docCov, complexViolations, namingViolations := extractBurdenMetrics(snapshot.Report)

	insertSnapshot := `
	INSERT INTO snapshots (
		id, timestamp, git_commit, git_branch, git_tag, version, 
		author, description, size_bytes, data_compressed,
		mbi_score_avg, duplication_ratio, doc_coverage, 
		complexity_violations, naming_violations
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := tx.ExecContext(ctx, insertSnapshot,
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
		mbiAvg,
		dupRatio,
		docCov,
		complexViolations,
		namingViolations,
	)
	if err != nil {
		return fmt.Errorf("failed to insert snapshot: %w", err)
	}
	return nil
}

func (s *SQLiteStorage) insertSnapshotTags(ctx context.Context, tx *sql.Tx, snapshotID string, tags map[string]string) error {
	if len(tags) == 0 {
		return nil
	}

	insertTag := "INSERT INTO snapshot_tags (snapshot_id, key, value) VALUES (?, ?, ?)"
	for key, value := range tags {
		_, err := tx.ExecContext(ctx, insertTag, snapshotID, key, value)
		if err != nil {
			return fmt.Errorf("failed to insert tag %s: %w", key, err)
		}
	}
	return nil
}

// Retrieve gets a specific snapshot by ID
func (s *SQLiteStorage) Retrieve(ctx context.Context, id string) (metrics.MetricsSnapshot, error) {
	var snapshot metrics.MetricsSnapshot

	compressedData, metadata, err := s.fetchSnapshotData(ctx, id, &snapshot)
	if err != nil {
		return snapshot, err
	}

	if err := s.loadSnapshotTags(ctx, id, &metadata); err != nil {
		return snapshot, err
	}

	data, err := s.decompressIfNeeded(compressedData)
	if err != nil {
		return snapshot, err
	}

	if err := json.Unmarshal(data, &snapshot.Report); err != nil {
		return snapshot, fmt.Errorf("failed to unmarshal report data: %w", err)
	}

	snapshot.Metadata = metadata
	return snapshot, nil
}

// fetchSnapshotData retrieves snapshot data and metadata from the database
func (s *SQLiteStorage) fetchSnapshotData(ctx context.Context, id string, snapshot *metrics.MetricsSnapshot) ([]byte, metrics.SnapshotMetadata, error) {
	var metadata metrics.SnapshotMetadata
	var compressedData []byte
	var gitCommit, gitBranch, gitTag, version, author, description sql.NullString

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
			return nil, metadata, fmt.Errorf("snapshot not found: %s", id)
		}
		return nil, metadata, fmt.Errorf("failed to retrieve snapshot: %w", err)
	}

	s.populateNullableFields(&metadata, gitCommit, gitBranch, gitTag, version, author, description)
	return compressedData, metadata, nil
}

// populateNullableFields converts SQL nullable fields to metadata fields
func (s *SQLiteStorage) populateNullableFields(metadata *metrics.SnapshotMetadata, gitCommit, gitBranch, gitTag, version, author, description sql.NullString) {
	metadata.GitCommit = gitCommit.String
	metadata.GitBranch = gitBranch.String
	metadata.GitTag = gitTag.String
	metadata.Version = version.String
	metadata.Author = author.String
	metadata.Description = description.String
}

// loadSnapshotTags retrieves tags associated with the snapshot
func (s *SQLiteStorage) loadSnapshotTags(ctx context.Context, id string, metadata *metrics.SnapshotMetadata) error {
	tags := make(map[string]string)
	tagQuery := "SELECT key, value FROM snapshot_tags WHERE snapshot_id = ?"
	rows, err := s.db.QueryContext(ctx, tagQuery, id)
	if err != nil {
		return fmt.Errorf("failed to retrieve tags: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return fmt.Errorf("failed to scan tag: %w", err)
		}
		tags[key] = value
	}
	metadata.Tags = tags
	return nil
}

// decompressIfNeeded decompresses data if compression is enabled
func (s *SQLiteStorage) decompressIfNeeded(compressedData []byte) ([]byte, error) {
	if s.config.EnableCompression {
		data, err := decompress(compressedData)
		if err != nil {
			return nil, fmt.Errorf("failed to decompress data: %w", err)
		}
		return data, nil
	}
	return compressedData, nil
}

// List returns available snapshots with optional filtering
// List retrieves snapshots matching the specified filter criteria
func (s *SQLiteStorage) List(ctx context.Context, filter SnapshotFilter) ([]SnapshotInfo, error) {
	var snapshots []SnapshotInfo

	// Build and execute query
	query, args := s.buildListQuery(filter)
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list snapshots: %w", err)
	}
	defer rows.Close()

	// Process query results
	for rows.Next() {
		info, err := s.scanSnapshotInfo(rows)
		if err != nil {
			return nil, err
		}

		// Retrieve and attach tags
		if err := s.attachSnapshotTags(ctx, &info); err != nil {
			return nil, err
		}

		snapshots = append(snapshots, info)
	}

	return snapshots, nil
}

// buildListQuery constructs the SQL query and parameters based on filter criteria
func (s *SQLiteStorage) buildListQuery(filter SnapshotFilter) (string, []interface{}) {
	query := "SELECT id, timestamp, git_commit, git_branch, git_tag, version, author, description, size_bytes, mbi_score_avg, duplication_ratio, doc_coverage, complexity_violations, naming_violations FROM snapshots WHERE 1=1"
	var args []interface{}

	query, args = s.addTimeFilters(query, args, filter)
	query, args = s.addMetadataFilters(query, args, filter)
	query, args = s.addTagFilters(query, args, filter)
	query = s.addOrderingAndLimits(query, &args, filter)

	return query, args
}

// addTimeFilters adds timestamp-based filtering to the query
func (s *SQLiteStorage) addTimeFilters(query string, args []interface{}, filter SnapshotFilter) (string, []interface{}) {
	if filter.After != nil {
		query += " AND timestamp > ?"
		args = append(args, *filter.After)
	}

	if filter.Before != nil {
		query += " AND timestamp < ?"
		args = append(args, *filter.Before)
	}

	return query, args
}

// addMetadataFilters adds git metadata filtering to the query
func (s *SQLiteStorage) addMetadataFilters(query string, args []interface{}, filter SnapshotFilter) (string, []interface{}) {
	if filter.Branch != "" {
		query += " AND git_branch = ?"
		args = append(args, filter.Branch)
	}

	if filter.Tag != "" {
		query += " AND git_tag = ?"
		args = append(args, filter.Tag)
	}

	if filter.Author != "" {
		query += " AND author = ?"
		args = append(args, filter.Author)
	}

	return query, args
}

// addTagFilters adds custom tag filtering to the query
func (s *SQLiteStorage) addTagFilters(query string, args []interface{}, filter SnapshotFilter) (string, []interface{}) {
	for key, value := range filter.Tags {
		query += " AND id IN (SELECT snapshot_id FROM snapshot_tags WHERE key = ? AND value = ?)"
		args = append(args, key, value)
	}
	return query, args
}

// addOrderingAndLimits adds sorting and pagination to the query
func (s *SQLiteStorage) addOrderingAndLimits(query string, args *[]interface{}, filter SnapshotFilter) string {
	query += " ORDER BY timestamp DESC"
	if filter.Limit > 0 {
		query += " LIMIT ?"
		*args = append(*args, filter.Limit)
		if filter.Offset > 0 {
			query += " OFFSET ?"
			*args = append(*args, filter.Offset)
		}
	}
	return query
}

// scanSnapshotInfo scans a database row into a SnapshotInfo struct
func (s *SQLiteStorage) scanSnapshotInfo(rows *sql.Rows) (SnapshotInfo, error) {
	var info SnapshotInfo
	var gitCommit, gitBranch, gitTag, version, author, description sql.NullString
	var mbiScoreAvg, duplicationRatio, docCoverage sql.NullFloat64
	var complexityViolations, namingViolations sql.NullInt64

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
		&mbiScoreAvg,
		&duplicationRatio,
		&docCoverage,
		&complexityViolations,
		&namingViolations,
	)
	if err != nil {
		return info, fmt.Errorf("failed to scan snapshot: %w", err)
	}

	populateNullableStrings(&info, gitCommit, gitBranch, gitTag, version, author, description)
	populateBurdenMetrics(&info, mbiScoreAvg, duplicationRatio, docCoverage, complexityViolations, namingViolations)

	return info, nil
}

// populateNullableStrings sets string fields from nullable SQL values.
func populateNullableStrings(info *SnapshotInfo, gitCommit, gitBranch, gitTag, version, author, description sql.NullString) {
	info.GitCommit = gitCommit.String
	info.GitBranch = gitBranch.String
	info.GitTag = gitTag.String
	info.Version = version.String
	info.Author = author.String
	info.Description = description.String
}

// populateBurdenMetrics sets burden metric fields from nullable SQL values.
func populateBurdenMetrics(info *SnapshotInfo, mbiScoreAvg, duplicationRatio, docCoverage sql.NullFloat64, complexityViolations, namingViolations sql.NullInt64) {
	if mbiScoreAvg.Valid {
		val := mbiScoreAvg.Float64
		info.MBIScoreAvg = &val
	}
	if duplicationRatio.Valid {
		val := duplicationRatio.Float64
		info.DuplicationRatio = &val
	}
	if docCoverage.Valid {
		val := docCoverage.Float64
		info.DocCoverage = &val
	}
	if complexityViolations.Valid {
		val := int(complexityViolations.Int64)
		info.ComplexityViolations = &val
	}
	if namingViolations.Valid {
		val := int(namingViolations.Int64)
		info.NamingViolations = &val
	}
}

// attachSnapshotTags retrieves and attaches tags for a specific snapshot
func (s *SQLiteStorage) attachSnapshotTags(ctx context.Context, info *SnapshotInfo) error {
	info.Tags = make(map[string]string)
	tagQuery := "SELECT key, value FROM snapshot_tags WHERE snapshot_id = ?"
	tagRows, err := s.db.QueryContext(ctx, tagQuery, info.ID)
	if err != nil {
		return fmt.Errorf("failed to retrieve tags for snapshot %s: %w", info.ID, err)
	}
	defer tagRows.Close()

	for tagRows.Next() {
		var key, value string
		if err := tagRows.Scan(&key, &value); err != nil {
			return fmt.Errorf("failed to scan tag: %w", err)
		}
		info.Tags[key] = value
	}

	return nil
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
	deletedCount, err := s.deleteByAge(ctx, policy)
	if err != nil {
		return err
	}

	countDeleted, err := s.deleteByCount(ctx, policy)
	if err != nil {
		return err
	}

	s.reportCleanupResults(deletedCount + countDeleted)
	return nil
}

// deleteByAge removes snapshots older than the maximum age specified in the policy
func (s *SQLiteStorage) deleteByAge(ctx context.Context, policy RetentionPolicy) (int64, error) {
	if policy.MaxAge == 0 {
		return 0, nil
	}

	cutoff := time.Now().Add(-policy.MaxAge)
	query := s.buildAgeBasedDeleteQuery(policy)

	result, err := s.db.ExecContext(ctx, query, cutoff)
	if err != nil {
		return 0, fmt.Errorf("failed to delete old snapshots: %w", err)
	}

	affected, _ := result.RowsAffected()
	return affected, nil
}

// deleteByCount removes excess snapshots beyond the maximum count specified in the policy
func (s *SQLiteStorage) deleteByCount(ctx context.Context, policy RetentionPolicy) (int64, error) {
	if policy.MaxCount == 0 {
		return 0, nil
	}

	currentCount, err := s.countSnapshots(ctx)
	if err != nil {
		return 0, err
	}

	if currentCount <= policy.MaxCount {
		return 0, nil
	}

	toDelete := currentCount - policy.MaxCount
	query := s.buildCountBasedDeleteQuery(policy)

	result, err := s.db.ExecContext(ctx, query, toDelete)
	if err != nil {
		return 0, fmt.Errorf("failed to delete excess snapshots: %w", err)
	}

	affected, _ := result.RowsAffected()
	return affected, nil
}

// buildAgeBasedDeleteQuery constructs the SQL query for age-based deletion
func (s *SQLiteStorage) buildAgeBasedDeleteQuery(policy RetentionPolicy) string {
	query := "DELETE FROM snapshots WHERE timestamp < ?"

	if policy.KeepTagged {
		query += " AND id NOT IN (SELECT DISTINCT snapshot_id FROM snapshot_tags)"
	}

	if policy.KeepReleases {
		query += " AND (git_tag IS NULL OR git_tag = '')"
	}

	return query
}

// buildCountBasedDeleteQuery constructs the SQL query for count-based deletion
func (s *SQLiteStorage) buildCountBasedDeleteQuery(policy RetentionPolicy) string {
	if policy.KeepTagged {
		return `DELETE FROM snapshots WHERE id IN (
			SELECT id FROM snapshots 
			WHERE id NOT IN (SELECT DISTINCT snapshot_id FROM snapshot_tags)
			ORDER BY timestamp ASC 
			LIMIT ?
		)`
	}

	return `DELETE FROM snapshots WHERE id IN (
		SELECT id FROM snapshots 
		ORDER BY timestamp ASC 
		LIMIT ?
	)`
}

// countSnapshots returns the total number of snapshots in the database
func (s *SQLiteStorage) countSnapshots(ctx context.Context) (int, error) {
	var count int
	err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM snapshots").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count snapshots: %w", err)
	}
	return count, nil
}

// reportCleanupResults logs the number of deleted snapshots
func (s *SQLiteStorage) reportCleanupResults(deletedCount int64) {
	if deletedCount > 0 {
		fmt.Printf("Cleaned up %d old snapshots\n", deletedCount)
	}
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

// createDirIfNotExists creates the directory for a file path if it doesn't exist.
func createDirIfNotExists(path string) error {
	dir := filepath.Dir(path)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return os.MkdirAll(dir, 0o755)
	}
	return nil
}

// compress compresses data using gzip encoding.
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

// decompress decompresses gzip-encoded data.
func decompress(data []byte) ([]byte, error) {
	reader, err := gzip.NewReader(strings.NewReader(string(data)))
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	return io.ReadAll(reader)
}
