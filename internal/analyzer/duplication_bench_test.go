package analyzer

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"testing"
)

// BenchmarkDuplicationAnalysis_SmallFile benchmarks analysis of a small file
func BenchmarkDuplicationAnalysis_SmallFile(b *testing.B) {
	fset := token.NewFileSet()
	testFile := filepath.Join("..", "..", "testdata", "duplication", "exact_clone.go")

	file, err := parser.ParseFile(fset, testFile, nil, parser.ParseComments)
	if err != nil {
		b.Fatalf("Failed to parse test file: %v", err)
	}

	analyzer := NewDuplicationAnalyzer(fset)
	minBlockLines := 6

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = analyzer.ExtractBlocks(file, testFile, minBlockLines)
	}
}

// BenchmarkDuplicationAnalysis_MediumFile benchmarks analysis of a medium-sized file
func BenchmarkDuplicationAnalysis_MediumFile(b *testing.B) {
	fset := token.NewFileSet()
	// Use a real file from the codebase
	testFile := filepath.Join("..", "..", "internal", "analyzer", "duplication.go")

	file, err := parser.ParseFile(fset, testFile, nil, parser.ParseComments)
	if err != nil {
		b.Fatalf("Failed to parse test file: %v", err)
	}

	analyzer := NewDuplicationAnalyzer(fset)
	minBlockLines := 6

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = analyzer.ExtractBlocks(file, testFile, minBlockLines)
	}
}

// BenchmarkDuplicationAnalysis_BlockNormalization benchmarks block normalization
func BenchmarkDuplicationAnalysis_BlockNormalization(b *testing.B) {
	fset := token.NewFileSet()
	// Use duplication.go which definitely has large functions
	testFile := filepath.Join("..", "..", "internal", "analyzer", "duplication.go")

	file, err := parser.ParseFile(fset, testFile, nil, parser.ParseComments)
	if err != nil {
		b.Fatalf("Failed to parse test file: %v", err)
	}

	analyzer := NewDuplicationAnalyzer(fset)
	blocks := analyzer.ExtractBlocks(file, testFile, 6)

	if len(blocks) == 0 {
		b.Skip("No blocks extracted - file may have small functions")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, block := range blocks {
			_ = analyzer.NormalizeBlock(block)
		}
	}
}

// BenchmarkDuplicationAnalysis_HashComputation benchmarks hash computation
func BenchmarkDuplicationAnalysis_HashComputation(b *testing.B) {
	fset := token.NewFileSet()
	// Use duplication.go which definitely has large functions
	testFile := filepath.Join("..", "..", "internal", "analyzer", "duplication.go")

	file, err := parser.ParseFile(fset, testFile, nil, parser.ParseComments)
	if err != nil {
		b.Fatalf("Failed to parse test file: %v", err)
	}

	analyzer := NewDuplicationAnalyzer(fset)
	blocks := analyzer.ExtractBlocks(file, testFile, 6)

	if len(blocks) == 0 {
		b.Skip("No blocks extracted - file may have small functions")
	}

	normalized := make([]NormalizedBlock, len(blocks))
	for i, block := range blocks {
		normalized[i] = analyzer.NormalizeBlock(block)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, norm := range normalized {
			_ = analyzer.ComputeHash(norm)
		}
	}
}

// BenchmarkDuplicationAnalysis_FullPipeline benchmarks the complete analysis pipeline
func BenchmarkDuplicationAnalysis_FullPipeline(b *testing.B) {
	fset := token.NewFileSet()
	testFile := filepath.Join("..", "..", "testdata", "duplication", "exact_clone.go")

	file, err := parser.ParseFile(fset, testFile, nil, parser.ParseComments)
	if err != nil {
		b.Fatalf("Failed to parse test file: %v", err)
	}

	analyzer := NewDuplicationAnalyzer(fset)
	minBlockLines := 6

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		blocks := analyzer.ExtractBlocks(file, testFile, minBlockLines)
		fingerprints := analyzer.FingerprintBlocks(blocks)
		_ = analyzer.DetectClonePairs(fingerprints, 0.80)
	}
}

// BenchmarkDuplicationAnalysis_MultipleFiles benchmarks analysis of multiple files
func BenchmarkDuplicationAnalysis_MultipleFiles(b *testing.B) {
	testFiles := []string{
		filepath.Join("..", "..", "testdata", "duplication", "exact_clone.go"),
		filepath.Join("..", "..", "testdata", "duplication", "renamed_clone.go"),
		filepath.Join("..", "..", "testdata", "duplication", "near_clone.go"),
		filepath.Join("..", "..", "testdata", "duplication", "below_threshold.go"),
		filepath.Join("..", "..", "testdata", "duplication", "small_blocks.go"),
	}

	fset := token.NewFileSet()
	files := make([]*ast.File, 0, len(testFiles))

	for _, testFile := range testFiles {
		file, err := parser.ParseFile(fset, testFile, nil, parser.ParseComments)
		if err != nil {
			b.Logf("Warning: Failed to parse %s: %v", testFile, err)
			continue
		}
		files = append(files, file)
	}

	if len(files) == 0 {
		b.Fatal("No files parsed successfully")
	}

	analyzer := NewDuplicationAnalyzer(fset)
	minBlockLines := 6

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		allFingerprints := make([]BlockFingerprint, 0)
		for j, file := range files {
			blocks := analyzer.ExtractBlocks(file, testFiles[j], minBlockLines)
			fingerprints := analyzer.FingerprintBlocks(blocks)
			allFingerprints = append(allFingerprints, fingerprints...)
		}
		_ = analyzer.DetectClonePairs(allFingerprints, 0.80)
	}
}

// BenchmarkDuplicationAnalysis_LargeCodebase benchmarks analysis of the entire codebase
func BenchmarkDuplicationAnalysis_LargeCodebase(b *testing.B) {
	// Collect all Go files in internal/analyzer
	var goFiles []string
	analyzerDir := filepath.Join("..", "..", "internal", "analyzer")

	err := filepath.Walk(analyzerDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".go" && filepath.Base(path) != "duplication_bench_test.go" {
			goFiles = append(goFiles, path)
		}
		return nil
	})
	if err != nil {
		b.Fatalf("Failed to walk directory: %v", err)
	}

	if len(goFiles) == 0 {
		b.Skip("No Go files found in internal/analyzer")
	}

	fset := token.NewFileSet()
	files := make([]*ast.File, 0, len(goFiles))

	for _, goFile := range goFiles {
		file, err := parser.ParseFile(fset, goFile, nil, parser.ParseComments)
		if err != nil {
			b.Logf("Warning: Failed to parse %s: %v", goFile, err)
			continue
		}
		files = append(files, file)
	}

	if len(files) == 0 {
		b.Fatal("No files parsed successfully")
	}

	analyzer := NewDuplicationAnalyzer(fset)
	minBlockLines := 6

	b.Logf("Benchmarking with %d files from internal/analyzer", len(files))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		allFingerprints := make([]BlockFingerprint, 0)
		for j, file := range files {
			blocks := analyzer.ExtractBlocks(file, goFiles[j], minBlockLines)
			fingerprints := analyzer.FingerprintBlocks(blocks)
			allFingerprints = append(allFingerprints, fingerprints...)
		}
		_ = analyzer.DetectClonePairs(allFingerprints, 0.80)
	}
}

// BenchmarkDuplicationAnalysis_SimilarityComputation benchmarks Jaccard similarity
func BenchmarkDuplicationAnalysis_SimilarityComputation(b *testing.B) {
	fset := token.NewFileSet()
	testFile := filepath.Join("..", "..", "testdata", "duplication", "near_clone.go")

	file, err := parser.ParseFile(fset, testFile, nil, parser.ParseComments)
	if err != nil {
		b.Fatalf("Failed to parse test file: %v", err)
	}

	analyzer := NewDuplicationAnalyzer(fset)
	blocks := analyzer.ExtractBlocks(file, testFile, 6)

	if len(blocks) < 2 {
		b.Fatal("Need at least 2 blocks for similarity computation")
	}

	normalized1 := analyzer.NormalizeBlock(blocks[0])
	normalized2 := analyzer.NormalizeBlock(blocks[1])

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = analyzer.ComputeSimilarity(normalized1, normalized2)
	}
}
