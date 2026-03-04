// Deprecated: This file is kept for backward compatibility.
// New code should use internal/api/storage package instead.
package api

import "github.com/opd-ai/go-stats-generator/internal/api/storage"

// Storage is deprecated. Use storage.Memory or storage.ResultStore interface instead.
type Storage = storage.Memory

// NewStorage is deprecated. Use storage.NewMemory() instead for in-memory storage or storage.NewSQLite() for persistence.
// This function remains for backward compatibility with existing integrations but will be removed in v2.0.
// Migrate to the new storage package constructors to access additional storage backends and configuration options.
func NewStorage() *Storage {
	return storage.NewMemory()
}
