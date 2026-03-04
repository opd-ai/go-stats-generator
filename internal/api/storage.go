// Deprecated: This file is kept for backward compatibility.
// New code should use internal/api/storage package instead.
package api

import "github.com/opd-ai/go-stats-generator/internal/api/storage"

// Storage is deprecated. Use storage.Memory or storage.ResultStore interface instead.
type Storage = storage.Memory

// NewStorage is deprecated. Use storage.NewMemory() instead.
func NewStorage() *Storage {
	return storage.NewMemory()
}
