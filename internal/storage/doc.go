// Package storage provides historical metrics persistence for trend analysis.
//
// The storage package defines the MetricsStorage interface and provides multiple
// implementations including SQLite for persistent storage, JSON file storage,
// and in-memory storage for testing. It supports storing, retrieving, listing,
// and cleaning up metrics snapshots with metadata tagging.
package storage
