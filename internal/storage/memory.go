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
	snapshot metrics.MetricsSnapshot
	metadata metrics.SnapshotMetadata
	stored   time.Time
}

// NewMemoryStorage creates a new in-memory storage instance
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		snapshots: make(map[string]*storedSnapshot),
	}
}

// Store saves a metrics snapshot in memory
func (m *MemoryStorage) Store(ctx context.Context, snapshot metrics.MetricsSnapshot, metadata metrics.SnapshotMetadata) error {
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

// Retrieve gets a specific snapshot by ID
func (m *MemoryStorage) Retrieve(ctx context.Context, id string) (metrics.MetricsSnapshot, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stored, ok := m.snapshots[id]
	if !ok {
		return metrics.MetricsSnapshot{}, fmt.Errorf("snapshot not found: %s", id)
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

// Delete removes a snapshot from memory
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
func (m *MemoryStorage) GetLatest(ctx context.Context) (metrics.MetricsSnapshot, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.snapshots) == 0 {
		return metrics.MetricsSnapshot{}, fmt.Errorf("no snapshots available")
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
func (m *MemoryStorage) GetByTag(ctx context.Context, key, value string) ([]metrics.MetricsSnapshot, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []metrics.MetricsSnapshot

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
	if filter.After != nil && stored.metadata.Timestamp.Before(*filter.After) {
		return false
	}
	if filter.Before != nil && stored.metadata.Timestamp.After(*filter.Before) {
		return false
	}
	if filter.Branch != "" && stored.metadata.GitBranch != filter.Branch {
		return false
	}
	if filter.Tag != "" && stored.metadata.GitTag != filter.Tag {
		return false
	}
	if filter.Author != "" && stored.metadata.Author != filter.Author {
		return false
	}
	if filter.Tags != nil {
		for key, value := range filter.Tags {
			if tagValue, ok := stored.metadata.Tags[key]; !ok || tagValue != value {
				return false
			}
		}
	}
	return true
}
