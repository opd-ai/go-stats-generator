package scanner

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

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

// DiscoverFiles finds all Go source files in the given root directory
func (d *Discoverer) DiscoverFiles(rootDir string) ([]FileInfo, error) {
	var files []FileInfo

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return d.shouldSkipDirectory(path, rootDir)
		}

		// Only process .go files
		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		fileInfo, err := d.analyzeFile(path, rootDir, info)
		if err != nil {
			// Log warning but continue processing
			return nil
		}

		if d.shouldIncludeFile(fileInfo) {
			files = append(files, fileInfo)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to discover files in %s: %w", rootDir, err)
	}

	return files, nil
}

// analyzeFile extracts information about a Go source file
func (d *Discoverer) analyzeFile(path, rootDir string, info os.FileInfo) (FileInfo, error) {
	relPath, err := filepath.Rel(rootDir, path)
	if err != nil {
		return FileInfo{}, fmt.Errorf("failed to get relative path: %w", err)
	}

	fileInfo := FileInfo{
		Path:        path,
		RelPath:     relPath,
		Size:        info.Size(),
		IsTestFile:  strings.HasSuffix(path, "_test.go"),
		IsGenerated: false,
	}

	// Parse the file to get package name and check if generated
	src, err := os.ReadFile(path)
	if err != nil {
		return fileInfo, fmt.Errorf("failed to read file %s: %w", path, err)
	}

	// Check for generated file markers
	content := string(src)
	fileInfo.IsGenerated = isGeneratedFile(content)

	// Parse to get package name
	file, err := parser.ParseFile(d.fset, path, src, parser.PackageClauseOnly)
	if err != nil {
		// Return what we have even if parsing fails
		return fileInfo, nil
	}

	if file.Name != nil {
		fileInfo.Package = file.Name.Name
	}

	return fileInfo, nil
}

// shouldSkipDirectory determines if a directory should be skipped
func (d *Discoverer) shouldSkipDirectory(dirPath, rootDir string) error {
	relPath, err := filepath.Rel(rootDir, dirPath)
	if err != nil {
		return nil
	}

	// Skip vendor directories if configured
	if d.config.SkipVendor && containsPathSegment(relPath, "vendor") {
		return filepath.SkipDir
	}

	// Skip hidden directories
	if strings.HasPrefix(filepath.Base(dirPath), ".") && relPath != "." {
		return filepath.SkipDir
	}

	return nil
}

// shouldIncludeFile determines if a file should be included in analysis
func (d *Discoverer) shouldIncludeFile(fileInfo FileInfo) bool {
	// Check file size limit
	if d.config.MaxFileSizeKB > 0 && fileInfo.Size > int64(d.config.MaxFileSizeKB*1024) {
		return false
	}

	// Check test file exclusion
	if d.config.SkipTestFiles && fileInfo.IsTestFile {
		return false
	}

	// Check generated file exclusion
	if d.config.SkipGenerated && fileInfo.IsGenerated {
		return false
	}

	// Check include patterns
	if len(d.config.IncludePatterns) > 0 {
		included := false
		for _, pattern := range d.config.IncludePatterns {
			// Handle recursive patterns like **/*.go
			if strings.Contains(pattern, "**") {
				if strings.HasSuffix(fileInfo.RelPath, ".go") {
					included = true
					break
				}
			} else {
				if matched, _ := filepath.Match(pattern, fileInfo.RelPath); matched {
					included = true
					break
				}
			}
		}
		if !included {
			return false
		}
	}

	// Check exclude patterns
	for _, pattern := range d.config.ExcludePatterns {
		if matched, _ := filepath.Match(pattern, fileInfo.RelPath); matched {
			return false
		}
	}

	// Check package inclusion/exclusion
	if len(d.config.IncludePackages) > 0 {
		included := false
		for _, pkg := range d.config.IncludePackages {
			if fileInfo.Package == pkg {
				included = true
				break
			}
		}
		if !included {
			return false
		}
	}

	for _, pkg := range d.config.ExcludePackages {
		if fileInfo.Package == pkg {
			return false
		}
	}

	return true
}

// isGeneratedFile checks if a file appears to be generated
func isGeneratedFile(content string) bool {
	// Check first few lines for common generated file markers
	lines := strings.Split(content, "\n")
	checkLines := len(lines)
	if checkLines > 10 {
		checkLines = 10
	}

	for i := 0; i < checkLines; i++ {
		line := strings.ToLower(strings.TrimSpace(lines[i]))

		// Common generated file markers
		if strings.Contains(line, "code generated") ||
			strings.Contains(line, "do not edit") ||
			strings.Contains(line, "autogenerated") ||
			strings.Contains(line, "auto-generated") ||
			strings.Contains(line, "generated automatically") ||
			strings.Contains(line, "this file was generated") {
			return true
		}
	}

	return false
}

// containsPathSegment checks if a path contains a specific segment
func containsPathSegment(path, segment string) bool {
	segments := strings.Split(filepath.ToSlash(path), "/")
	for _, s := range segments {
		if s == segment {
			return true
		}
	}
	return false
}

// ParseFile parses a Go source file and returns the AST
func (d *Discoverer) ParseFile(path string) (*ast.File, error) {
	src, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", path, err)
	}

	file, err := parser.ParseFile(d.fset, path, src, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file %s: %w", path, err)
	}

	return file, nil
}

// GetFileSet returns the token file set used by this discoverer
func (d *Discoverer) GetFileSet() *token.FileSet {
	return d.fset
}
