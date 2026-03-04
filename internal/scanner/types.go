package scanner

import (
	"go/token"

	"github.com/opd-ai/go-stats-generator/internal/config"
)

// FileInfo represents information about a Go source file
type FileInfo struct {
	Path        string
	RelPath     string
	Package     string
	Size        int64
	IsTestFile  bool
	IsGenerated bool
}

// Discoverer handles file discovery and filtering
type Discoverer struct {
	config *config.FilterConfig
	fset   *token.FileSet
}

// NewDiscoverer creates a new file discoverer
func NewDiscoverer(cfg *config.FilterConfig) *Discoverer {
	return &Discoverer{
		config: cfg,
		fset:   token.NewFileSet(),
	}
}

// GetFileSet returns the token file set used by this discoverer
func (d *Discoverer) GetFileSet() *token.FileSet {
	return d.fset
}
