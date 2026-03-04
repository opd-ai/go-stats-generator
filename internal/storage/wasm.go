//go:build js && wasm

// Package storage provides stub implementations for WASM builds.
// Storage operations (SQLite, JSON files, and in-memory) are not supported
// in WASM environments since they rely on filesystem and database access
// that is unavailable in browser contexts.
//
// For browser-based analysis, use one-shot analysis without baseline/diff/trend
// features that require persistent storage.
package storage

import (
	"errors"
)

// ErrNotSupported is returned when storage operations are attempted in WASM builds.
// WASM builds perform one-shot analysis in the browser and do not support
// baseline management, differential analysis, or trend tracking.
var ErrNotSupported = errors.New("storage operations not supported in WASM builds")

// SQLiteStorage stub type for WASM builds
type SQLiteStorage struct{}

// JSONStorage stub type for WASM builds
type JSONStorage struct{}

// MemoryStorage stub type for WASM builds
type MemoryStorage struct{}

// NewSQLiteStorageImpl returns ErrNotSupported in WASM builds.
// SQLite database operations require filesystem access not available in browsers.
func NewSQLiteStorageImpl(config SQLiteConfig) (*SQLiteStorage, error) {
	return nil, ErrNotSupported
}

// NewJSONStorageImpl returns ErrNotSupported in WASM builds.
// JSON file storage requires filesystem access not available in browsers.
func NewJSONStorageImpl(config JSONConfig) (*JSONStorage, error) {
	return nil, ErrNotSupported
}

// NewMemoryStorage returns nil for WASM builds.
// In-memory storage is not needed for one-shot browser analysis.
func NewMemoryStorage() *MemoryStorage {
	return nil
}
