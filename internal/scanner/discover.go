package scanner

import (
	"fmt"
	"go/ast"
	"go/parser"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// DiscoverFiles finds all Go source files in the given root directory
func (d *Discoverer) DiscoverFiles(rootDir string) ([]FileInfo, error) {
	var files []FileInfo
	walkFunc := d.createWalkDirFunction(rootDir, &files)
	err := filepath.WalkDir(rootDir, walkFunc)
	if err != nil {
		return nil, fmt.Errorf("failed to discover files in %s: %w", rootDir, err)
	}
	return files, nil
}

// createWalkDirFunction creates the fs.WalkDirFunc for discovering Go files.
// Using filepath.WalkDir (vs the older filepath.Walk) avoids the extra os.Stat syscall per entry
// because fs.DirEntry already carries the lstat result.
func (d *Discoverer) createWalkDirFunction(rootDir string, files *[]FileInfo) fs.WalkDirFunc {
	return func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return d.shouldSkipDirectory(path, rootDir)
		}
		if strings.HasSuffix(path, ".go") {
			// DirEntry.Info() returns the cached lstat result, avoiding an extra os.Stat syscall.
			info, infoErr := entry.Info()
			if infoErr != nil {
				return nil // skip unreadable entries silently
			}
			d.processGoFile(path, rootDir, info, files)
		}
		return nil
	}
}

// processGoFile analyzes and adds a Go file to the file list if it should be included
func (d *Discoverer) processGoFile(path, rootDir string, info fs.FileInfo, files *[]FileInfo) {
	fileInfo, err := d.analyzeFile(path, rootDir, info)
	if err != nil {
		return
	}
	if d.shouldIncludeFile(fileInfo) {
		*files = append(*files, fileInfo)
	}
}

// analyzeFile extracts information about a Go source file.
// It reads the file once, caches the bytes in FileInfo.Src, and parses the package clause.
// The cached bytes are later reused by the worker to avoid a second os.ReadFile during AST parsing.
func (d *Discoverer) analyzeFile(path, rootDir string, info fs.FileInfo) (FileInfo, error) {
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

	// Read file once; cache bytes in FileInfo.Src so the worker can reuse them
	// instead of re-reading the file during full AST parsing.
	src, err := os.ReadFile(path)
	if err != nil {
		return fileInfo, fmt.Errorf("failed to read file %s: %w", path, err)
	}
	fileInfo.Src = src

	// Check for generated file markers
	fileInfo.IsGenerated = isGeneratedFile(string(src))

	// Parse to get package name (PackageClauseOnly is fast; reuses src to avoid another read)
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

// ParseFile parses a Go source file and returns the AST.
func (d *Discoverer) ParseFile(path string) (*ast.File, error) {
	src, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", path, err)
	}

	return d.parseFileWithSrc(path, src)
}

// parseFileWithSrc parses a Go source file using already-read bytes, eliminating a redundant
// os.ReadFile call for files whose content was cached during discovery.
func (d *Discoverer) parseFileWithSrc(path string, src []byte) (*ast.File, error) {
	file, err := parser.ParseFile(d.fset, path, src, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file %s: %w", path, err)
	}

	return file, nil
}
