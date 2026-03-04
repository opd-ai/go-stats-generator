//go:build !js || !wasm

package scanner

import (
	"fmt"
	"go/ast"
	"go/parser"
	"os"
	"path/filepath"
	"strings"
)

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
