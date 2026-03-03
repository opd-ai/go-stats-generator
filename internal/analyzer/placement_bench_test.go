package analyzer

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"testing"
)

// BenchmarkPlacementAnalysis_SmallPackage benchmarks analysis of a small package
func BenchmarkPlacementAnalysis_SmallPackage(b *testing.B) {
	testDir := filepath.Join("..", "..", "testdata", "placement", "misplaced_function")
	fset := token.NewFileSet()

	pkgs, err := parser.ParseDir(fset, testDir, nil, parser.ParseComments)
	if err != nil {
		b.Fatalf("Failed to parse test directory: %v", err)
	}

	var files []*ast.File
	for _, pkg := range pkgs {
		for _, file := range pkg.Files {
			files = append(files, file)
		}
	}

	if len(files) == 0 {
		b.Fatal("No files parsed")
	}

	analyzer := NewPlacementAnalyzer(0.25, 0.3)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = analyzer.Analyze(files, fset)
	}
}

// BenchmarkPlacementAnalysis_SymbolIndexing benchmarks symbol index building
func BenchmarkPlacementAnalysis_SymbolIndexing(b *testing.B) {
	testDir := filepath.Join("..", "..", "testdata", "placement", "low_cohesion")
	fset := token.NewFileSet()

	pkgs, err := parser.ParseDir(fset, testDir, nil, parser.ParseComments)
	if err != nil {
		b.Fatalf("Failed to parse test directory: %v", err)
	}

	var files []*ast.File
	for _, pkg := range pkgs {
		for _, file := range pkg.Files {
			files = append(files, file)
		}
	}

	if len(files) == 0 {
		b.Fatal("No files parsed")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		analyzer := NewPlacementAnalyzer(0.25, 0.3)
		analyzer.buildSymbolIndex(files, fset)
	}
}

// BenchmarkPlacementAnalysis_FunctionAffinity benchmarks function affinity analysis
func BenchmarkPlacementAnalysis_FunctionAffinity(b *testing.B) {
	testDir := filepath.Join("..", "..", "testdata", "placement", "misplaced_function")
	fset := token.NewFileSet()

	pkgs, err := parser.ParseDir(fset, testDir, nil, parser.ParseComments)
	if err != nil {
		b.Fatalf("Failed to parse test directory: %v", err)
	}

	var files []*ast.File
	for _, pkg := range pkgs {
		for _, file := range pkg.Files {
			files = append(files, file)
		}
	}

	if len(files) == 0 {
		b.Fatal("No files parsed")
	}

	analyzer := NewPlacementAnalyzer(0.25, 0.3)
	analyzer.buildSymbolIndex(files, fset)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = analyzer.AnalyzeFunctionAffinity()
	}
}

// BenchmarkPlacementAnalysis_MethodPlacement benchmarks method placement analysis
func BenchmarkPlacementAnalysis_MethodPlacement(b *testing.B) {
	testDir := filepath.Join("..", "..", "testdata", "placement", "misplaced_method")
	fset := token.NewFileSet()

	pkgs, err := parser.ParseDir(fset, testDir, nil, parser.ParseComments)
	if err != nil {
		b.Fatalf("Failed to parse test directory: %v", err)
	}

	var files []*ast.File
	for _, pkg := range pkgs {
		for _, file := range pkg.Files {
			files = append(files, file)
		}
	}

	if len(files) == 0 {
		b.Fatal("No files parsed")
	}

	analyzer := NewPlacementAnalyzer(0.25, 0.3)
	analyzer.buildSymbolIndex(files, fset)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = analyzer.AnalyzeMethodPlacement()
	}
}

// BenchmarkPlacementAnalysis_FileCohesion benchmarks file cohesion analysis
func BenchmarkPlacementAnalysis_FileCohesion(b *testing.B) {
	testDir := filepath.Join("..", "..", "testdata", "placement", "low_cohesion")
	fset := token.NewFileSet()

	pkgs, err := parser.ParseDir(fset, testDir, nil, parser.ParseComments)
	if err != nil {
		b.Fatalf("Failed to parse test directory: %v", err)
	}

	var files []*ast.File
	for _, pkg := range pkgs {
		for _, file := range pkg.Files {
			files = append(files, file)
		}
	}

	if len(files) == 0 {
		b.Fatal("No files parsed")
	}

	analyzer := NewPlacementAnalyzer(0.25, 0.3)
	analyzer.buildSymbolIndex(files, fset)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = analyzer.AnalyzeFileCohesion()
	}
}

// BenchmarkPlacementAnalysis_FullPipeline benchmarks the complete analysis pipeline
func BenchmarkPlacementAnalysis_FullPipeline(b *testing.B) {
	testDir := filepath.Join("..", "..", "testdata", "placement", "low_cohesion")
	fset := token.NewFileSet()

	pkgs, err := parser.ParseDir(fset, testDir, nil, parser.ParseComments)
	if err != nil {
		b.Fatalf("Failed to parse test directory: %v", err)
	}

	var files []*ast.File
	for _, pkg := range pkgs {
		for _, file := range pkg.Files {
			files = append(files, file)
		}
	}

	if len(files) == 0 {
		b.Fatal("No files parsed")
	}

	analyzer := NewPlacementAnalyzer(0.25, 0.3)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		analyzer.buildSymbolIndex(files, fset)
		_ = analyzer.AnalyzeFunctionAffinity()
		_ = analyzer.AnalyzeMethodPlacement()
		_ = analyzer.AnalyzeFileCohesion()
	}
}

// BenchmarkPlacementAnalysis_MultiplePackages benchmarks analysis of multiple test packages
func BenchmarkPlacementAnalysis_MultiplePackages(b *testing.B) {
	testDirs := []string{
		filepath.Join("..", "..", "testdata", "placement", "misplaced_function"),
		filepath.Join("..", "..", "testdata", "placement", "misplaced_method"),
		filepath.Join("..", "..", "testdata", "placement", "low_cohesion"),
		filepath.Join("..", "..", "testdata", "placement", "high_cohesion"),
	}

	fset := token.NewFileSet()
	allFiles := make([]*ast.File, 0)

	for _, testDir := range testDirs {
		pkgs, err := parser.ParseDir(fset, testDir, nil, parser.ParseComments)
		if err != nil {
			b.Logf("Warning: Failed to parse %s: %v", testDir, err)
			continue
		}

		for _, pkg := range pkgs {
			for _, file := range pkg.Files {
				allFiles = append(allFiles, file)
			}
		}
	}

	if len(allFiles) == 0 {
		b.Fatal("No files parsed")
	}

	analyzer := NewPlacementAnalyzer(0.25, 0.3)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = analyzer.Analyze(allFiles, fset)
	}
}

// BenchmarkPlacementAnalysis_LargeCodebase benchmarks analysis of the entire codebase
func BenchmarkPlacementAnalysis_LargeCodebase(b *testing.B) {
	// Collect all Go files in internal/analyzer
	var goFiles []string
	analyzerDir := filepath.Join("..", "..", "internal", "analyzer")

	err := filepath.Walk(analyzerDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".go" && filepath.Base(path) != "placement_bench_test.go" {
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

	analyzer := NewPlacementAnalyzer(0.25, 0.3)

	b.Logf("Benchmarking with %d files from internal/analyzer", len(files))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = analyzer.Analyze(files, fset)
	}
}

// BenchmarkPlacementAnalysis_SymbolReferenceCollection benchmarks reference collection
func BenchmarkPlacementAnalysis_SymbolReferenceCollection(b *testing.B) {
	testDir := filepath.Join("..", "..", "testdata", "placement", "low_cohesion")
	fset := token.NewFileSet()

	pkgs, err := parser.ParseDir(fset, testDir, nil, parser.ParseComments)
	if err != nil {
		b.Fatalf("Failed to parse test directory: %v", err)
	}

	var files []*ast.File
	for _, pkg := range pkgs {
		for _, file := range pkg.Files {
			files = append(files, file)
		}
	}

	if len(files) == 0 {
		b.Fatal("No files parsed")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		analyzer := NewPlacementAnalyzer(0.25, 0.3)
		// Just build the index to benchmark reference collection
		analyzer.buildSymbolIndex(files, fset)
	}
}

// BenchmarkPlacementAnalysis_AffinityCalculation benchmarks affinity score computation
func BenchmarkPlacementAnalysis_AffinityCalculation(b *testing.B) {
	testDir := filepath.Join("..", "..", "testdata", "placement", "misplaced_function")
	fset := token.NewFileSet()

	pkgs, err := parser.ParseDir(fset, testDir, nil, parser.ParseComments)
	if err != nil {
		b.Fatalf("Failed to parse test directory: %v", err)
	}

	var files []*ast.File
	for _, pkg := range pkgs {
		for _, file := range pkg.Files {
			files = append(files, file)
		}
	}

	if len(files) == 0 {
		b.Fatal("No files parsed")
	}

	analyzer := NewPlacementAnalyzer(0.25, 0.3)
	analyzer.buildSymbolIndex(files, fset)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Run affinity analysis which includes affinity calculations
		_ = analyzer.AnalyzeFunctionAffinity()
	}
}

// BenchmarkPlacementAnalysis_CohesionScoring benchmarks cohesion score computation
func BenchmarkPlacementAnalysis_CohesionScoring(b *testing.B) {
	testDir := filepath.Join("..", "..", "testdata", "placement", "low_cohesion")
	fset := token.NewFileSet()

	pkgs, err := parser.ParseDir(fset, testDir, nil, parser.ParseComments)
	if err != nil {
		b.Fatalf("Failed to parse test directory: %v", err)
	}

	var files []*ast.File
	for _, pkg := range pkgs {
		for _, file := range pkg.Files {
			files = append(files, file)
		}
	}

	if len(files) == 0 {
		b.Fatal("No files parsed")
	}

	analyzer := NewPlacementAnalyzer(0.25, 0.3)
	analyzer.buildSymbolIndex(files, fset)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Run cohesion analysis which includes cohesion scoring
		_ = analyzer.AnalyzeFileCohesion()
	}
}
