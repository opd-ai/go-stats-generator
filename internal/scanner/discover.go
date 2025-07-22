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
	if !d.passesFileConstraints(fileInfo) {
		return false
	}

	if !d.matchesIncludePatterns(fileInfo) {
		return false
	}

	if d.matchesExcludePatterns(fileInfo) {
		return false
	}

	if !d.passesPackageFilter(fileInfo) {
		return false
	}

	return true
}

// passesFileConstraints checks basic file constraints like size limits and file types
func (d *Discoverer) passesFileConstraints(fileInfo FileInfo) bool {
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

	return true
}

// matchesIncludePatterns checks if file matches any include patterns
func (d *Discoverer) matchesIncludePatterns(fileInfo FileInfo) bool {
	if len(d.config.IncludePatterns) == 0 {
		return true
	}

	for _, pattern := range d.config.IncludePatterns {
		if d.patternMatches(pattern, fileInfo.RelPath) {
			return true
		}
	}

	return false
}

// matchesExcludePatterns checks if file matches any exclude patterns
func (d *Discoverer) matchesExcludePatterns(fileInfo FileInfo) bool {
	for _, pattern := range d.config.ExcludePatterns {
		if matched, _ := filepath.Match(pattern, fileInfo.RelPath); matched {
			return true
		}
	}
	return false
}

// passesPackageFilter checks package inclusion/exclusion rules
func (d *Discoverer) passesPackageFilter(fileInfo FileInfo) bool {
	// Check package inclusion
	if len(d.config.IncludePackages) > 0 {
		for _, pkg := range d.config.IncludePackages {
			if fileInfo.Package == pkg {
				return true
			}
		}
		return false
	}

	// Check package exclusion
	for _, pkg := range d.config.ExcludePackages {
		if fileInfo.Package == pkg {
			return false
		}
	}

	return true
}

// patternMatches checks if a file path matches a given pattern
func (d *Discoverer) patternMatches(pattern, relPath string) bool {
	// Handle recursive patterns like **/*.go
	if strings.Contains(pattern, "**") {
		return strings.HasSuffix(relPath, ".go")
	}

	matched, _ := filepath.Match(pattern, relPath)
	return matched
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
	// Normalize both forward slashes and backslashes to forward slashes
	// This handles both Unix and Windows path formats
	normalizedPath := strings.ReplaceAll(path, "\\", "/")
	segments := strings.Split(normalizedPath, "/")
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
