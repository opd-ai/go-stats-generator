package storage

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/opd-ai/go-stats-generator/internal/metrics"
)

// MemoryStorage implements MetricsStorage using in-memory data structures.
// MemoryStorage is ideal for temporary analysis runs, CI/CD pipelines, and testing.
type MemoryStorage struct {
	mu        sync.RWMutex
	snapshots map[string]*storedSnapshot
}

// storedSnapshot wraps snapshot data with metadata
type storedSnapshot struct {
	snapshot metrics.Snapshot
	metadata metrics.SnapshotMetadata
	stored   time.Time
}

// NewMemoryStorage creates a new in-memory storage instance for temporary baseline retention without persistent storage.
// Ideal for CI/CD pipelines, ephemeral environments, or testing where baseline history doesn't need to survive restarts.
// Provides fast reads/writes with RWMutex synchronization for concurrent access, but all data is lost on process termination.
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		snapshots: make(map[string]*storedSnapshot),
	}
}

// Store saves a metrics snapshot in memory
func (m *MemoryStorage) Store(ctx context.Context, snapshot metrics.Snapshot, metadata metrics.SnapshotMetadata) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if snapshot.ID == "" {
		return fmt.Errorf("snapshot ID cannot be empty")
	}

	m.snapshots[snapshot.ID] = &storedSnapshot{
		snapshot: snapshot,
		metadata: metadata,
		stored:   time.Now(),
	}

	return nil
}

// Retrieve fetches a baseline snapshot from in-memory storage by ID with read-lock protection for safe concurrent access.
// It returns a direct reference to the stored snapshot (not a deep copy), so callers should treat returned data as read-only.
// Primarily used in testing scenarios and development workflows where baseline persistence is not required.
// Returns error if the snapshot ID doesn't exist in the memory map.
func (m *MemoryStorage) Retrieve(ctx context.Context, id string) (metrics.Snapshot, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stored, ok := m.snapshots[id]
	if !ok {
		return metrics.Snapshot{}, fmt.Errorf("snapshot not found: %s", id)
	}

	return stored.snapshot, nil
}

// List returns available snapshots with optional filtering
func (m *MemoryStorage) List(ctx context.Context, filter SnapshotFilter) ([]SnapshotInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := m.collectMatchingSnapshots(filter)
	m.sortByTimestamp(result)
	return m.applyPagination(result, filter), nil
}

// collectMatchingSnapshots gathers snapshots matching the filter
func (m *MemoryStorage) collectMatchingSnapshots(filter SnapshotFilter) []SnapshotInfo {
	var result []SnapshotInfo
	for id, stored := range m.snapshots {
		if !matchesFilter(stored, filter) {
			continue
		}
		result = append(result, m.createSnapshotInfo(id, stored))
	}
	return result
}

// createSnapshotInfo builds a SnapshotInfo from stored snapshot
func (m *MemoryStorage) createSnapshotInfo(id string, stored *storedSnapshot) SnapshotInfo {
	return SnapshotInfo{
		ID:          id,
		Timestamp:   stored.metadata.Timestamp,
		GitCommit:   stored.metadata.GitCommit,
		GitBranch:   stored.metadata.GitBranch,
		GitTag:      stored.metadata.GitTag,
		Version:     stored.metadata.Version,
		Author:      stored.metadata.Author,
		Description: stored.metadata.Description,
		Tags:        stored.metadata.Tags,
		Size:        0,
	}
}

// sortByTimestamp sorts snapshots by timestamp descending
func (m *MemoryStorage) sortByTimestamp(result []SnapshotInfo) {
	sort.Slice(result, func(i, j int) bool {
		return result[i].Timestamp.After(result[j].Timestamp)
	})
}

// applyPagination applies limit and offset to results
func (m *MemoryStorage) applyPagination(result []SnapshotInfo, filter SnapshotFilter) []SnapshotInfo {
	if filter.Limit <= 0 || len(result) <= filter.Limit {
		return result
	}
	if filter.Offset >= len(result) {
		return []SnapshotInfo{}
	}
	end := filter.Offset + filter.Limit
	if end > len(result) {
		end = len(result)
	}
	return result[filter.Offset:end]
}

// Delete removes a baseline snapshot from the in-memory storage map with thread-safe synchronization.
// It acquires an exclusive write lock before deletion to prevent concurrent access issues, returning error
// if the snapshot ID doesn't exist in memory. This implementation is primarily used for testing scenarios
// where persistence is not required and fast iteration is prioritized over durability.
func (m *MemoryStorage) Delete(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.snapshots[id]; !ok {
		return fmt.Errorf("snapshot not found: %s", id)
	}

	delete(m.snapshots, id)
	return nil
}

// Cleanup removes old snapshots based on retention policy
func (m *MemoryStorage) Cleanup(ctx context.Context, policy RetentionPolicy) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	toDelete := m.findExpiredSnapshots(policy)
	toDelete = append(toDelete, m.findExcessSnapshots(policy, toDelete)...)

	for _, id := range toDelete {
		delete(m.snapshots, id)
	}
	return nil
}

// findExpiredSnapshots identifies snapshots exceeding max age
func (m *MemoryStorage) findExpiredSnapshots(policy RetentionPolicy) []string {
	if policy.MaxAge <= 0 {
		return nil
	}
	var toDelete []string
	now := time.Now()
	for id, stored := range m.snapshots {
		if m.isExpired(stored, policy, now) {
			toDelete = append(toDelete, id)
		}
	}
	return toDelete
}

// isExpired checks if a snapshot has exceeded retention age
func (m *MemoryStorage) isExpired(stored *storedSnapshot, policy RetentionPolicy, now time.Time) bool {
	if now.Sub(stored.stored) <= policy.MaxAge {
		return false
	}
	if policy.KeepTagged && len(stored.metadata.Tags) > 0 {
		return false
	}
	if policy.KeepReleases && stored.metadata.GitTag != "" {
		return false
	}
	return true
}

// findExcessSnapshots identifies snapshots exceeding max count
func (m *MemoryStorage) findExcessSnapshots(policy RetentionPolicy, excluded []string) []string {
	if policy.MaxCount <= 0 || len(m.snapshots)-len(excluded) <= policy.MaxCount {
		return nil
	}
	sorted := m.sortSnapshotsByAge(excluded)
	excess := len(sorted) - policy.MaxCount
	var toDelete []string
	for i := 0; i < excess; i++ {
		toDelete = append(toDelete, sorted[i].snapshot.ID)
	}
	return toDelete
}

// sortSnapshotsByAge returns snapshots sorted by age, excluding specified IDs
func (m *MemoryStorage) sortSnapshotsByAge(excluded []string) []*storedSnapshot {
	var sorted []*storedSnapshot
	for id, s := range m.snapshots {
		if !contains(excluded, id) {
			sorted = append(sorted, s)
		}
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].stored.Before(sorted[j].stored)
	})
	return sorted
}

// contains checks if a slice contains a string
func contains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

// GetLatest returns the most recent snapshot
func (m *MemoryStorage) GetLatest(ctx context.Context) (metrics.Snapshot, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.snapshots) == 0 {
		return metrics.Snapshot{}, fmt.Errorf("no snapshots available")
	}

	var latest *storedSnapshot
	for _, stored := range m.snapshots {
		if latest == nil || stored.metadata.Timestamp.After(latest.metadata.Timestamp) {
			latest = stored
		}
	}

	return latest.snapshot, nil
}

// GetByTag returns snapshots matching a specific tag
func (m *MemoryStorage) GetByTag(ctx context.Context, key, value string) ([]metrics.Snapshot, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []metrics.Snapshot

	for _, stored := range m.snapshots {
		if tagValue, ok := stored.metadata.Tags[key]; ok && tagValue == value {
			result = append(result, stored.snapshot)
		}
	}

	return result, nil
}

// Close releases storage resources (no-op for memory storage)
func (m *MemoryStorage) Close() error {
	return nil
}

// matchesFilter checks if a snapshot matches the filter criteria
func matchesFilter(stored *storedSnapshot, filter SnapshotFilter) bool {
	return matchesTimeRange(stored.metadata.Timestamp, filter) &&
		matchesBranch(stored.metadata.GitBranch, filter) &&
		matchesTag(stored.metadata.GitTag, filter) &&
		matchesAuthor(stored.metadata.Author, filter) &&
		matchesTags(stored.metadata.Tags, filter)
}

// matchesTimeRange checks if timestamp falls within the filter's time range
func matchesTimeRange(timestamp time.Time, filter SnapshotFilter) bool {
	if filter.After != nil && timestamp.Before(*filter.After) {
		return false
	}
	if filter.Before != nil && timestamp.After(*filter.Before) {
		return false
	}
	return true
}

// matchesBranch checks if branch matches the filter
func matchesBranch(branch string, filter SnapshotFilter) bool {
	return filter.Branch == "" || branch == filter.Branch
}

// matchesTag checks if tag matches the filter
func matchesTag(tag string, filter SnapshotFilter) bool {
	return filter.Tag == "" || tag == filter.Tag
}

// matchesAuthor checks if author matches the filter
func matchesAuthor(author string, filter SnapshotFilter) bool {
	return filter.Author == "" || author == filter.Author
}

// matchesTags checks if custom tags match the filter
func matchesTags(tags map[string]string, filter SnapshotFilter) bool {
	if filter.Tags == nil {
		return true
	}
	for key, value := range filter.Tags {
		if tagValue, ok := tags[key]; !ok || tagValue != value {
			return false
		}
	}
	return true
}
