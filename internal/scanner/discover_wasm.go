//go:build js && wasm

package scanner

import (
	"fmt"
	"go/ast"
	"go/parser"
	"path/filepath"
	"strings"
)

// MemoryFile represents a file stored in memory for WASM analysis
type MemoryFile struct {
	Path    string
	Content string
}

// DiscoverFilesFromMemory processes files from memory instead of filesystem
func (d *Discoverer) DiscoverFilesFromMemory(files []MemoryFile, rootDir string) ([]FileInfo, error) {
	var result []FileInfo

	for _, memFile := range files {
		fileInfo, err := d.analyzeMemoryFile(memFile, rootDir)
		if err != nil {
			continue
		}

		if d.shouldIncludeFile(fileInfo) {
			result = append(result, fileInfo)
		}
	}

	return result, nil
}

// analyzeMemoryFile extracts info from in-memory file
func (d *Discoverer) analyzeMemoryFile(memFile MemoryFile, rootDir string) (FileInfo, error) {
	path := memFile.Path
	relPath := path
	if rootDir != "" && strings.HasPrefix(path, rootDir) {
		relPath = strings.TrimPrefix(path, rootDir)
		relPath = strings.TrimPrefix(relPath, "/")
	}

	fileInfo := FileInfo{
		Path:        path,
		RelPath:     relPath,
		Size:        int64(len(memFile.Content)),
		IsTestFile:  strings.HasSuffix(path, "_test.go"),
		IsGenerated: isGeneratedFile(memFile.Content),
	}

	file, err := parser.ParseFile(d.fset, path, memFile.Content, parser.PackageClauseOnly)
	if err != nil {
		return fileInfo, nil
	}

	if file.Name != nil {
		fileInfo.Package = file.Name.Name
	}

	return fileInfo, nil
}

// ParseMemoryFile parses an in-memory Go file
func (d *Discoverer) ParseMemoryFile(path, content string) (*ast.File, error) {
	file, err := parser.ParseFile(d.fset, path, content, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file %s: %w", path, err)
	}
	return file, nil
}

// shouldSkipMemoryDirectory checks if path should be skipped
func (d *Discoverer) shouldSkipMemoryDirectory(path string) bool {
	if d.config.SkipVendor && containsPathSegment(path, "vendor") {
		return true
	}

	segments := strings.Split(filepath.ToSlash(path), "/")
	for _, segment := range segments {
		if strings.HasPrefix(segment, ".") && segment != "." {
			return true
		}
	}

	return false
}

// ParseFile is not supported in WASM (use ParseMemoryFile instead)
func (d *Discoverer) ParseFile(path string) (*ast.File, error) {
	return nil, fmt.Errorf("ParseFile not supported in WASM; use ParseMemoryFile")
}
