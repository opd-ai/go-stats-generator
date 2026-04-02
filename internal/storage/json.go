package storage

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/opd-ai/go-stats-generator/internal/metrics"
)

// JSONStorage implements MetricsStorage using JSON files
type JSONStorage struct {
	config JSONConfig
}

// snapshotFile represents the JSON file structure for a stored snapshot
type snapshotFile struct {
	ID       string                   `json:"id"`
	Metadata metrics.SnapshotMetadata `json:"metadata"`
	Report   metrics.Report           `json:"report"`
}

// NewJSONStorageImpl creates a JSON file-based storage backend for baseline snapshots with automatic directory initialization.
// It establishes a file system storage structure using the configured directory path, creating necessary directories with
// appropriate permissions (0755). Each snapshot is persisted as an individual JSON file (optionally gzip-compressed) enabling
// human-readable inspection and version control integration. Returns initialized storage instance or error if directory creation fails.
func NewJSONStorageImpl(config JSONConfig) (*JSONStorage, error) {
	// Create directory if it doesn't exist
	if err := os.MkdirAll(config.Directory, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	return &JSONStorage{
		config: config,
	}, nil
}

// Store saves a metrics snapshot as a JSON file
func (j *JSONStorage) Store(ctx context.Context, snapshot metrics.Snapshot, metadata metrics.SnapshotMetadata) error {
	// Create snapshot file structure
	fileData := snapshotFile{
		ID:       snapshot.ID,
		Metadata: metadata,
		Report:   snapshot.Report,
	}

	// Marshal to JSON
	var data []byte
	var err error
	if j.config.Pretty {
		data, err = json.MarshalIndent(fileData, "", "  ")
	} else {
		data, err = json.Marshal(fileData)
	}
	if err != nil {
		return fmt.Errorf("failed to marshal snapshot: %w", err)
	}

	// Determine filename
	filename := j.getFilename(snapshot.ID)
	filepath := filepath.Join(j.config.Directory, filename)

	// Write to file (with optional compression)
	if j.config.Compression {
		return j.writeCompressed(filepath, data)
	}
	return j.writeUncompressed(filepath, data)
}

// Retrieve reads a snapshot from a JSON file by ID
func (j *JSONStorage) Retrieve(ctx context.Context, id string) (metrics.Snapshot, error) {
	// Try compressed version first, then uncompressed
	filename := j.getFilename(id)
	filepath := filepath.Join(j.config.Directory, filename)

	data, err := j.readFile(filepath)
	if err != nil {
		return metrics.Snapshot{}, fmt.Errorf("snapshot not found: %s", id)
	}

	// Parse JSON
	var fileData snapshotFile
	if err := json.Unmarshal(data, &fileData); err != nil {
		return metrics.Snapshot{}, fmt.Errorf("failed to parse snapshot: %w", err)
	}

	return metrics.Snapshot{
		ID:       fileData.ID,
		Report:   fileData.Report,
		Metadata: fileData.Metadata,
	}, nil
}

// List returns available snapshots with optional filtering
func (j *JSONStorage) List(ctx context.Context, filter SnapshotFilter) ([]SnapshotInfo, error) {
	files, err := j.readStorageDirectory()
	if err != nil {
		return nil, err
	}

	snapshots := j.collectSnapshots(files, filter)
	j.sortSnapshotsByTimestamp(snapshots)
	return j.applyPagination(snapshots, filter), nil
}

// readStorageDirectory reads snapshot files from the storage directory
func (j *JSONStorage) readStorageDirectory() ([]os.DirEntry, error) {
	files, err := os.ReadDir(j.config.Directory)
	if err != nil {
		if os.IsNotExist(err) {
			return []os.DirEntry{}, nil
		}
		return nil, fmt.Errorf("failed to read storage directory: %w", err)
	}
	return files, nil
}

// collectSnapshots processes directory entries and builds the snapshot list
func (j *JSONStorage) collectSnapshots(files []os.DirEntry, filter SnapshotFilter) []SnapshotInfo {
	var snapshots []SnapshotInfo

	for _, file := range files {
		if snapshot, ok := j.processSnapshotFile(file, filter); ok {
			snapshots = append(snapshots, snapshot)
		}
	}
	return snapshots
}

// processSnapshotFile processes a single file and returns snapshot info if valid
func (j *JSONStorage) processSnapshotFile(file os.DirEntry, filter SnapshotFilter) (SnapshotInfo, bool) {
	if file.IsDir() || !j.isSnapshotFile(file.Name()) {
		return SnapshotInfo{}, false
	}

	id := j.extractIDFromFilename(file.Name())
	if id == "" {
		return SnapshotInfo{}, false
	}

	fileData, err := j.loadSnapshotMetadata(file.Name())
	if err != nil {
		return SnapshotInfo{}, false
	}

	if !j.matchesFilter(fileData, filter) {
		return SnapshotInfo{}, false
	}

	return j.buildSnapshotInfo(fileData, file.Name()), true
}

// isSnapshotFile checks if a filename represents a snapshot file
func (j *JSONStorage) isSnapshotFile(name string) bool {
	return strings.HasSuffix(name, ".json") || strings.HasSuffix(name, ".json.gz")
}

// loadSnapshotMetadata reads and parses snapshot metadata from file
func (j *JSONStorage) loadSnapshotMetadata(filename string) (snapshotFile, error) {
	filepath := filepath.Join(j.config.Directory, filename)
	data, err := j.readFile(filepath)
	if err != nil {
		return snapshotFile{}, err
	}

	var fileData snapshotFile
	if err := json.Unmarshal(data, &fileData); err != nil {
		return snapshotFile{}, err
	}
	return fileData, nil
}

// buildSnapshotInfo constructs a SnapshotInfo from file data
func (j *JSONStorage) buildSnapshotInfo(fileData snapshotFile, filename string) SnapshotInfo {
	filepath := filepath.Join(j.config.Directory, filename)
	fileInfo, _ := os.Stat(filepath)

	var size int64
	if fileInfo != nil {
		size = fileInfo.Size()
	}

	return SnapshotInfo{
		ID:          fileData.ID,
		Timestamp:   fileData.Metadata.Timestamp,
		GitCommit:   fileData.Metadata.GitCommit,
		GitBranch:   fileData.Metadata.GitBranch,
		GitTag:      fileData.Metadata.GitTag,
		Version:     fileData.Metadata.Version,
		Author:      fileData.Metadata.Author,
		Description: fileData.Metadata.Description,
		Tags:        fileData.Metadata.Tags,
		Size:        size,
	}
}

// sortSnapshotsByTimestamp sorts snapshots by timestamp descending (newest first)
func (j *JSONStorage) sortSnapshotsByTimestamp(snapshots []SnapshotInfo) {
	sort.Slice(snapshots, func(i, j int) bool {
		return snapshots[i].Timestamp.After(snapshots[j].Timestamp)
	})
}

// applyPagination applies offset and limit to snapshot list
func (j *JSONStorage) applyPagination(snapshots []SnapshotInfo, filter SnapshotFilter) []SnapshotInfo {
	if filter.Offset > 0 {
		if filter.Offset >= len(snapshots) {
			return []SnapshotInfo{}
		}
		snapshots = snapshots[filter.Offset:]
	}

	if filter.Limit > 0 && filter.Limit < len(snapshots) {
		snapshots = snapshots[:filter.Limit]
	}

	return snapshots
}

// Delete removes a baseline snapshot file from the JSON file-based storage directory.
// It attempts deletion of both compressed (.json.gz) and uncompressed (.json) versions of the snapshot file,
// returning error only if the file doesn't exist or deletion fails due to permissions. Automatically handles
// cleanup of associated metadata files. Used by "baseline delete" when JSON storage backend is configured.
func (j *JSONStorage) Delete(ctx context.Context, id string) error {
	filename := j.getFilename(id)
	filepath := filepath.Join(j.config.Directory, filename)

	// Try both compressed and uncompressed versions
	if err := os.Remove(filepath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("snapshot not found: %s", id)
		}
		return fmt.Errorf("failed to delete snapshot: %w", err)
	}

	return nil
}

// Cleanup removes old snapshots based on retention policy
func (j *JSONStorage) Cleanup(ctx context.Context, policy RetentionPolicy) error {
	snapshots, err := j.List(ctx, SnapshotFilter{})
	if err != nil {
		return fmt.Errorf("failed to list snapshots: %w", err)
	}

	toDelete := j.identifySnapshotsToDelete(snapshots, policy)
	j.executeCleanupDeletions(ctx, toDelete)

	return nil
}

// identifySnapshotsToDelete determines which snapshots should be deleted based on retention policy
func (j *JSONStorage) identifySnapshotsToDelete(snapshots []SnapshotInfo, policy RetentionPolicy) []string {
	toDelete := j.findSnapshotsOlderThanMaxAge(snapshots, policy)
	toDelete = j.addExcessSnapshotsOverMaxCount(snapshots, policy, toDelete)
	return toDelete
}

// findSnapshotsOlderThanMaxAge returns snapshot IDs older than the policy's maximum age
func (j *JSONStorage) findSnapshotsOlderThanMaxAge(snapshots []SnapshotInfo, policy RetentionPolicy) []string {
	if policy.MaxAge == 0 {
		return nil
	}

	var toDelete []string
	cutoff := time.Now().Add(-policy.MaxAge)
	for _, snapshot := range snapshots {
		if snapshot.Timestamp.Before(cutoff) && !j.shouldKeepSnapshot(snapshot, policy) {
			toDelete = append(toDelete, snapshot.ID)
		}
	}
	return toDelete
}

// addExcessSnapshotsOverMaxCount adds snapshots exceeding the maximum count to deletion list
func (j *JSONStorage) addExcessSnapshotsOverMaxCount(snapshots []SnapshotInfo, policy RetentionPolicy, toDelete []string) []string {
	if policy.MaxCount == 0 || len(snapshots) <= policy.MaxCount {
		return toDelete
	}

	excess := snapshots[policy.MaxCount:]
	for _, snapshot := range excess {
		if !j.shouldKeepSnapshot(snapshot, policy) && !j.isDuplicate(snapshot.ID, toDelete) {
			toDelete = append(toDelete, snapshot.ID)
		}
	}
	return toDelete
}

// shouldKeepSnapshot checks if a snapshot should be kept based on tags or releases
func (j *JSONStorage) shouldKeepSnapshot(snapshot SnapshotInfo, policy RetentionPolicy) bool {
	if policy.KeepTagged && len(snapshot.Tags) > 0 {
		return true
	}
	if policy.KeepReleases && snapshot.GitTag != "" {
		return true
	}
	return false
}

// isDuplicate checks if an ID already exists in the deletion list
func (j *JSONStorage) isDuplicate(id string, toDelete []string) bool {
	for _, existing := range toDelete {
		if existing == id {
			return true
		}
	}
	return false
}

// executeCleanupDeletions performs the actual deletion of snapshots
func (j *JSONStorage) executeCleanupDeletions(ctx context.Context, toDelete []string) {
	for _, id := range toDelete {
		if err := j.Delete(ctx, id); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to delete snapshot %s: %v\n", id, err)
		}
	}

	if len(toDelete) > 0 {
		fmt.Fprintf(os.Stderr, "Cleaned up %d old snapshots\n", len(toDelete))
	}
}

// GetLatest returns the most recent snapshot
func (j *JSONStorage) GetLatest(ctx context.Context) (metrics.Snapshot, error) {
	snapshots, err := j.List(ctx, SnapshotFilter{Limit: 1})
	if err != nil {
		return metrics.Snapshot{}, err
	}

	if len(snapshots) == 0 {
		return metrics.Snapshot{}, fmt.Errorf("no snapshots found")
	}

	return j.Retrieve(ctx, snapshots[0].ID)
}

// GetByTag returns snapshots matching a specific tag
func (j *JSONStorage) GetByTag(ctx context.Context, key, value string) ([]metrics.Snapshot, error) {
	filter := SnapshotFilter{
		Tags: map[string]string{key: value},
	}

	infos, err := j.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	var snapshots []metrics.Snapshot
	for _, info := range infos {
		snapshot, err := j.Retrieve(ctx, info.ID)
		if err != nil {
			continue // Skip snapshots we can't retrieve
		}
		snapshots = append(snapshots, snapshot)
	}

	return snapshots, nil
}

// Close releases storage resources (no-op for JSON storage)
func (j *JSONStorage) Close() error {
	return nil
}

// Helper functions

// getFilename generates the filename for a snapshot with the appropriate extension.
func (j *JSONStorage) getFilename(id string) string {
	if j.config.Compression {
		return id + ".json.gz"
	}
	return id + ".json"
}

// extractIDFromFilename removes the file extension to get the snapshot ID.
func (j *JSONStorage) extractIDFromFilename(filename string) string {
	// Remove .json or .json.gz extension
	id := strings.TrimSuffix(filename, ".json.gz")
	id = strings.TrimSuffix(id, ".json")
	return id
}

// writeCompressed writes data to a gzip-compressed file.
func (j *JSONStorage) writeCompressed(filepath string, data []byte) error {
	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	gzWriter := gzip.NewWriter(file)
	defer gzWriter.Close()

	if _, err := gzWriter.Write(data); err != nil {
		return fmt.Errorf("failed to write compressed data: %w", err)
	}

	return nil
}

// writeUncompressed writes data to a plain JSON file.
func (j *JSONStorage) writeUncompressed(filepath string, data []byte) error {
	if err := os.WriteFile(filepath, data, 0o644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}
	return nil
}

// readFile reads data from a file, handling both compressed and uncompressed formats.
func (j *JSONStorage) readFile(filepath string) ([]byte, error) {
	// Try as-is first
	data, err := j.tryReadFile(filepath)
	if err == nil {
		return data, nil
	}

	// If compressed config but file doesn't exist, try uncompressed
	if j.config.Compression && strings.HasSuffix(filepath, ".json.gz") {
		uncompressedPath := strings.TrimSuffix(filepath, ".gz")
		return j.tryReadFile(uncompressedPath)
	}

	// If uncompressed config but file doesn't exist, try compressed
	if !j.config.Compression && strings.HasSuffix(filepath, ".json") {
		compressedPath := filepath + ".gz"
		return j.tryReadFile(compressedPath)
	}

	return nil, err
}

// tryReadFile attempts to read a file, decompressing if the path ends in .gz.
func (j *JSONStorage) tryReadFile(filepath string) ([]byte, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Check if file is gzipped
	if strings.HasSuffix(filepath, ".gz") {
		gzReader, err := gzip.NewReader(file)
		if err != nil {
			return nil, fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer gzReader.Close()

		return io.ReadAll(gzReader)
	}

	return io.ReadAll(file)
}

// matchesFilter evaluates whether a snapshot file matches the provided filter criteria,
// checking time range (After/Before), Git branch, tag, and author constraints.
func (j *JSONStorage) matchesFilter(fileData snapshotFile, filter SnapshotFilter) bool {
	return j.matchesTimeRange(fileData.Metadata.Timestamp, filter) &&
		j.matchesBranch(fileData.Metadata.GitBranch, filter) &&
		j.matchesGitTag(fileData.Metadata.GitTag, filter) &&
		j.matchesAuthorFilter(fileData.Metadata.Author, filter) &&
		j.matchesCustomTags(fileData.Metadata.Tags, filter)
}

// matchesTimeRange checks if timestamp falls within the filter's time range
func (j *JSONStorage) matchesTimeRange(timestamp time.Time, filter SnapshotFilter) bool {
	if filter.After != nil && !timestamp.After(*filter.After) {
		return false
	}
	if filter.Before != nil && !timestamp.Before(*filter.Before) {
		return false
	}
	return true
}

// matchesBranch checks if branch matches the filter
func (j *JSONStorage) matchesBranch(branch string, filter SnapshotFilter) bool {
	return filter.Branch == "" || branch == filter.Branch
}

// matchesGitTag checks if git tag matches the filter
func (j *JSONStorage) matchesGitTag(tag string, filter SnapshotFilter) bool {
	return filter.Tag == "" || tag == filter.Tag
}

// matchesAuthorFilter checks if author matches the filter
func (j *JSONStorage) matchesAuthorFilter(author string, filter SnapshotFilter) bool {
	return filter.Author == "" || author == filter.Author
}

// matchesCustomTags checks if custom tags match the filter
func (j *JSONStorage) matchesCustomTags(tags map[string]string, filter SnapshotFilter) bool {
	if len(filter.Tags) == 0 {
		return true
	}
	for key, value := range filter.Tags {
		if tags[key] != value {
			return false
		}
	}
	return true
}
