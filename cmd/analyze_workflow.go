package cmd

import (
	"context"
	"fmt"
	"go/ast"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/opd-ai/go-stats-generator/internal/analyzer"
	"github.com/opd-ai/go-stats-generator/internal/config"
	"github.com/opd-ai/go-stats-generator/internal/metrics"
	"github.com/opd-ai/go-stats-generator/internal/scanner"
)

// runDirectoryAnalysis performs comprehensive code analysis on a directory,
// discovering Go files, processing them through a worker pool, and generating
// a detailed metrics report including functions, structs, interfaces, patterns, etc.
func runDirectoryAnalysis(ctx context.Context, targetDir string, cfg *config.Config) (*metrics.Report, error) {
	return runAnalysisWorkflow(ctx, targetDir, cfg)
}

// runFileAnalysis performs comprehensive code analysis on a single Go source file,
// validating it's a .go file, parsing the AST, and running all analyzers (function,
// struct, interface, package, concurrency, duplication, etc.) on the parsed content.
func runFileAnalysis(ctx context.Context, filePath string, cfg *config.Config) (*metrics.Report, error) {
	startTime := time.Now()

	if !isGoSourceFile(filePath) {
		return nil, fmt.Errorf("file %s is not a Go source file", filePath)
	}

	logVerboseFileAnalysis(filePath, cfg)

	projectRoot := findProjectRoot(filePath)
	result, discoverer, err := parseAndPrepareFile(filePath, projectRoot, cfg)
	if err != nil {
		return nil, err
	}

	report, collectedMetrics, analyzers := runSingleFileAnalysis(result, discoverer, filePath, startTime, cfg)
	finalizeAllMetrics(report, collectedMetrics, analyzers, projectRoot, cfg)

	logVerboseFileResults(collectedMetrics, cfg)

	return report, nil
}

// logVerboseFileAnalysis prints file analysis progress if verbose mode is enabled.
func logVerboseFileAnalysis(filePath string, cfg *config.Config) {
	if cfg.Output.Verbose {
		fmt.Fprintf(os.Stderr, "Analyzing file: %s\n", filePath)
	}
}

// parseAndPrepareFile parses a single file and creates its scanner result with metadata.
func parseAndPrepareFile(filePath, projectRoot string, cfg *config.Config) (scanner.Result, *scanner.Discoverer, error) {
	discoverer := scanner.NewDiscoverer(&cfg.Filters)

	file, err := discoverer.ParseFile(filePath)
	if err != nil {
		return scanner.Result{}, nil, fmt.Errorf("failed to parse file %s: %w", filePath, err)
	}

	fileInfo, err := createFileInfoForSingleFile(filePath, projectRoot, file)
	if err != nil {
		return scanner.Result{}, nil, err
	}

	result := scanner.Result{
		FileInfo: fileInfo,
		File:     file,
		Error:    nil,
	}

	return result, discoverer, nil
}

// createFileInfoForSingleFile builds scanner file metadata from file system and AST information.
func createFileInfoForSingleFile(filePath, projectRoot string, file *ast.File) (scanner.FileInfo, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return scanner.FileInfo{}, fmt.Errorf("failed to get file info: %w", err)
	}

	relPath := calculateRelativePath(filePath, projectRoot)

	scannerFileInfo := scanner.FileInfo{
		Path:        filePath,
		RelPath:     relPath,
		Size:        fileInfo.Size(),
		IsTestFile:  strings.HasSuffix(filePath, "_test.go"),
		IsGenerated: false,
	}

	if file.Name != nil {
		scannerFileInfo.Package = file.Name.Name
	}

	return scannerFileInfo, nil
}

// calculateRelativePath computes the relative path from project root, falling back to basename.
func calculateRelativePath(filePath, projectRoot string) string {
	if projectRoot != "" {
		if rel, err := filepath.Rel(projectRoot, filePath); err == nil {
			return rel
		}
	}
	return filepath.Base(filePath)
}

// runSingleFileAnalysis orchestrates analysis for a single file and returns results with collected metrics.
func runSingleFileAnalysis(result scanner.Result, discoverer *scanner.Discoverer, filePath string, startTime time.Time, cfg *config.Config) (*metrics.Report, *CollectedMetrics, *AnalyzerSet) {
	analyzers := createAnalyzers(discoverer.GetFileSet(), cfg)
	report := createInitialReport(filepath.Dir(filePath), startTime, 1)
	collectedMetrics := &CollectedMetrics{}

	processFileAnalysis(result, analyzers, collectedMetrics, report, cfg)
	report.Metadata.AnalysisTime = time.Since(startTime)

	return report, collectedMetrics, analyzers
}

// finalizeAllMetrics runs all post-processing steps to complete the analysis report.
func finalizeAllMetrics(report *metrics.Report, collectedMetrics *CollectedMetrics, analyzers *AnalyzerSet, projectRoot string, cfg *config.Config) {
	finalizeReport(report, collectedMetrics, analyzers.Package, cfg)
	finalizeDuplicationMetrics(report, analyzers.Duplication, collectedMetrics, cfg)
	finalizeNamingMetrics(report, analyzers, collectedMetrics, cfg)
	finalizePlacementMetrics(report, analyzers, collectedMetrics, cfg)
	finalizeDocumentationMetrics(report, analyzers, collectedMetrics, cfg)
	finalizeOrganizationMetrics(report, analyzers, collectedMetrics, cfg, projectRoot)
	finalizeTeamMetrics(report, projectRoot, cfg)
	finalizeRefactoringSuggestions(report, cfg)
}

func logVerboseFileResults(collectedMetrics *CollectedMetrics, cfg *config.Config) {
	if cfg.Output.Verbose {
		fmt.Fprintf(os.Stderr, "Analyzed 1 file, found %d functions, %d structs, %d interfaces\n",
			len(collectedMetrics.Functions), len(collectedMetrics.Structs), len(collectedMetrics.Interfaces))
	}
}

// isGoSourceFile checks if a file is a Go source file
func isGoSourceFile(filePath string) bool {
	return strings.HasSuffix(filePath, ".go")
}

// findProjectRoot attempts to find the project root by looking for go.mod, .git, or other indicators
func findProjectRoot(filePath string) string {
	dir := filepath.Dir(filePath)

	for {
		// Check for go.mod (Go module root)
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}

		// Check for .git directory (Git repository root)
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir
		}

		// Move up one directory
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root
			break
		}
		dir = parent
	}

	// If no project root found, return empty string
	return ""
}

// runAnalysisWorkflow orchestrates the complete analysis workflow: file discovery,
// concurrent processing, metric aggregation, and report finalization.
func runAnalysisWorkflow(ctx context.Context, targetDir string, cfg *config.Config) (*metrics.Report, error) {
	startTime := time.Now()

	// Step 1: Discover and validate files
	discoverer, files, err := discoverAndValidateFiles(targetDir, cfg)
	if err != nil {
		return nil, err
	}

	// Step 2: Process files through worker pool
	results, err := processFilesWithWorkerPool(ctx, files, discoverer, cfg)
	if err != nil {
		return nil, err
	}

	// Step 3: Create analyzers and initial report structure
	analyzers := createAnalyzers(discoverer.GetFileSet(), cfg)
	report := createInitialReport(targetDir, startTime, len(files))

	// Step 4: Process analysis results from worker pool
	metrics, packageAnalyzer, err := processAnalysisResults(ctx, results, analyzers, report, cfg)
	if err != nil {
		return nil, err
	}

	// Step 5: Finalize report with all collected metrics
	finalizeReport(report, metrics, packageAnalyzer, cfg)
	finalizeDuplicationMetrics(report, analyzers.Duplication, metrics, cfg)
	finalizeNamingMetrics(report, analyzers, metrics, cfg)
	finalizePlacementMetrics(report, analyzers, metrics, cfg)
	finalizeDocumentationMetrics(report, analyzers, metrics, cfg)
	finalizeOrganizationMetrics(report, analyzers, metrics, cfg, targetDir)
	finalizeTeamMetrics(report, targetDir, cfg)

	// Step 6: Generate refactoring suggestions after all metrics are finalized
	finalizeRefactoringSuggestions(report, cfg)

	report.Metadata.AnalysisTime = time.Since(startTime)

	return report, nil
}

// AnalyzerSet holds all the different analyzers used in the workflow
type AnalyzerSet struct {
	Function      *analyzer.FunctionAnalyzer
	Struct        *analyzer.StructAnalyzer
	Interface     *analyzer.InterfaceAnalyzer
	Package       *analyzer.PackageAnalyzer
	Concurrency   *analyzer.ConcurrencyAnalyzer
	Duplication   *analyzer.DuplicationAnalyzer
	Naming        *analyzer.NamingAnalyzer
	Placement     *analyzer.PlacementAnalyzer
	Documentation *analyzer.DocumentationAnalyzer
	Organization  *analyzer.OrganizationAnalyzer
	Burden        *analyzer.BurdenAnalyzer
	Generic       *analyzer.GenericAnalyzer
	fileSet       *token.FileSet
}

// CollectedMetrics holds all metrics collected during analysis
type CollectedMetrics struct {
	Functions  []metrics.FunctionMetrics
	Structs    []metrics.StructMetrics
	Interfaces []metrics.InterfaceMetrics
	Generics   []metrics.GenericMetrics
	TotalLines int
	Files      map[string]*ast.File
}

// discoverAndValidateFiles discovers Go files in the target directory and validates the results
func discoverAndValidateFiles(targetDir string, cfg *config.Config) (*scanner.Discoverer, []scanner.FileInfo, error) {
	if cfg.Output.Verbose {
		fmt.Fprintf(os.Stderr, "Analyzing directory: %s\n", targetDir)
	}

	discoverer := scanner.NewDiscoverer(&cfg.Filters)
	files, err := discoverer.DiscoverFiles(targetDir)
	if err != nil {
		return nil, nil, fmt.Errorf("file discovery failed: %w", err)
	}

	if len(files) == 0 {
		return nil, nil, fmt.Errorf("no Go files found in %s", targetDir)
	}

	if cfg.Output.Verbose {
		fmt.Fprintf(os.Stderr, "Found %d Go files\n", len(files))
	}

	return discoverer, files, nil
}

// processFilesWithWorkerPool processes files using the worker pool with optional progress reporting
func processFilesWithWorkerPool(ctx context.Context, files []scanner.FileInfo, discoverer *scanner.Discoverer, cfg *config.Config) (<-chan scanner.Result, error) {
	workerPool := scanner.NewWorkerPool(&cfg.Performance, discoverer)

	var progressCallback scanner.ProgressCallback
	if cfg.Output.ShowProgress {
		progressCallback = func(completed, total int) {
			fmt.Fprintf(os.Stderr, "\rProcessing files: %d/%d (%.1f%%)",
				completed, total, float64(completed)/float64(total)*100)
		}
	}

	results, err := workerPool.ProcessFiles(ctx, files, progressCallback)
	if err != nil {
		return nil, fmt.Errorf("file processing failed: %w", err)
	}

	if cfg.Output.ShowProgress {
		fmt.Fprintf(os.Stderr, "\n")
	}

	return results, nil
}

// createAnalyzers creates and returns all analyzers needed for the workflow
func createAnalyzers(fileSet *token.FileSet, cfg *config.Config) *AnalyzerSet {
	docConfig := &analyzer.DocumentationConfig{
		RequireExportedDoc:  cfg.Analysis.Documentation.RequireExportedDoc,
		RequirePackageDoc:   cfg.Analysis.Documentation.RequirePackageDoc,
		StaleAnnotationDays: cfg.Analysis.Documentation.StaleAnnotationDays,
		MinCommentWords:     cfg.Analysis.Documentation.MinCommentWords,
	}

	return &AnalyzerSet{
		Function:      analyzer.NewFunctionAnalyzer(fileSet),
		Struct:        analyzer.NewStructAnalyzer(fileSet),
		Interface:     analyzer.NewInterfaceAnalyzer(fileSet),
		Package:       analyzer.NewPackageAnalyzer(fileSet),
		Concurrency:   analyzer.NewConcurrencyAnalyzer(fileSet),
		Duplication:   analyzer.NewDuplicationAnalyzer(fileSet),
		Naming:        analyzer.NewNamingAnalyzer(),
		Placement:     analyzer.NewPlacementAnalyzer(cfg.Analysis.Placement.AffinityMargin, cfg.Analysis.Placement.MinCohesion),
		Documentation: analyzer.NewDocumentationAnalyzer(fileSet, docConfig),
		Organization:  analyzer.NewOrganizationAnalyzer(fileSet),
		Burden:        analyzer.NewBurdenAnalyzer(fileSet),
		Generic:       analyzer.NewGenericAnalyzer(fileSet),
		fileSet:       fileSet,
	}
}

// createInitialReport creates the initial report structure with metadata and empty pattern metrics containers.
// This function constructs the foundational Report object that will be populated during analysis phases, including
// metadata fields (repository path, timestamp, analysis duration), and initializes empty data structures for all
// metric categories (functions, structs, patterns, concurrency, documentation, etc.). The report structure is then
// progressively enriched by analyzer components throughout the workflow execution.
func createInitialReport(targetDir string, startTime time.Time, fileCount int) *metrics.Report {
	return &metrics.Report{
		Metadata: metrics.ReportMetadata{
			Repository:     targetDir,
			GeneratedAt:    time.Now(),
			AnalysisTime:   time.Since(startTime),
			FilesProcessed: fileCount,
			ToolVersion:    "1.0.0",
		},
		Patterns: metrics.PatternMetrics{
			ConcurrencyPatterns: metrics.ConcurrencyPatternMetrics{
				WorkerPools: []metrics.PatternInstance{},
				Pipelines:   []metrics.PatternInstance{},
				FanOut:      []metrics.PatternInstance{},
				FanIn:       []metrics.PatternInstance{},
				Semaphores:  []metrics.PatternInstance{},
				Goroutines: metrics.GoroutineMetrics{
					Instances:      []metrics.GoroutineInstance{},
					GoroutineLeaks: []metrics.GoroutineLeakWarning{},
				},
				Channels: metrics.ChannelMetrics{
					Instances: []metrics.ChannelInstance{},
				},
				SyncPrims: metrics.SyncPrimitives{
					Mutexes:    []metrics.SyncPrimitiveInstance{},
					RWMutexes:  []metrics.SyncPrimitiveInstance{},
					WaitGroups: []metrics.SyncPrimitiveInstance{},
					Once:       []metrics.SyncPrimitiveInstance{},
					Cond:       []metrics.SyncPrimitiveInstance{},
					Atomic:     []metrics.SyncPrimitiveInstance{},
				},
			},
		},
		Burden: metrics.BurdenMetrics{
			MagicNumbers:          []metrics.MagicNumber{},
			ComplexSignatures:     []metrics.SignatureIssue{},
			DeeplyNestedFunctions: []metrics.NestingIssue{},
			FeatureEnvyMethods:    []metrics.FeatureEnvyIssue{},
			DeadCode: metrics.DeadCodeMetrics{
				UnreferencedFunctions: []metrics.UnreferencedSymbol{},
				UnreachableCode:       []metrics.UnreachableBlock{},
			},
		},
		Scores: metrics.ScoringMetrics{
			FileScores:    []metrics.FileScore{},
			PackageScores: []metrics.PackageScore{},
		},
	}
}

// processAnalysisResults coordinates the analysis of all scanner results
func processAnalysisResults(ctx context.Context, results <-chan scanner.Result, analyzers *AnalyzerSet, report *metrics.Report, cfg *config.Config) (*CollectedMetrics, *analyzer.PackageAnalyzer, error) {
	collectedMetrics := &CollectedMetrics{}
	processedFiles := 0

	for {
		done, err := processNextResult(ctx, results, &processedFiles, analyzers, collectedMetrics, report, cfg)
		if done {
			return collectedMetrics, analyzers.Package, err
		}
	}
}

// processNextResult processes a single result from the channel and returns completion status
func processNextResult(ctx context.Context, results <-chan scanner.Result, processedFiles *int, analyzers *AnalyzerSet, collectedMetrics *CollectedMetrics, report *metrics.Report, cfg *config.Config) (bool, error) {
	select {
	case result, ok := <-results:
		if !ok {
			logProcessingSummary(*processedFiles, collectedMetrics, cfg)
			return true, nil
		}
		return processValidResult(result, processedFiles, analyzers, collectedMetrics, report, cfg), nil

	case <-ctx.Done():
		return true, fmt.Errorf("analysis cancelled: %w", ctx.Err())
	}
}

// processValidResult handles a single valid result from the scanner
func processValidResult(result scanner.Result, processedFiles *int, analyzers *AnalyzerSet, collectedMetrics *CollectedMetrics, report *metrics.Report, cfg *config.Config) bool {
	*processedFiles++

	if !handleScannerError(result.Error, cfg) {
		return false
	}

	processFileAnalysis(result, analyzers, collectedMetrics, report, cfg)
	return false
}

// handleScannerError processes scanner errors and returns whether to continue processing
func handleScannerError(err error, cfg *config.Config) bool {
	if err != nil {
		if cfg.Output.Verbose {
			fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
		}
		return false
	}
	return true
}

// processFileAnalysis performs all analysis types on a single file
func processFileAnalysis(result scanner.Result, analyzers *AnalyzerSet, collectedMetrics *CollectedMetrics, report *metrics.Report, cfg *config.Config) {
	// Store the parsed file for duplication analysis
	if collectedMetrics.Files == nil {
		collectedMetrics.Files = make(map[string]*ast.File)
	}
	collectedMetrics.Files[result.FileInfo.RelPath] = result.File

	collectStructuralMetrics(result, analyzers, collectedMetrics, cfg)
	analyzePackageStructure(result, analyzers, cfg)
	analyzeConcurrencyPatterns(result, analyzers, report, cfg)
	analyzeBurdenIndicators(result, analyzers, report, cfg)
}

// collectStructuralMetrics analyzes functions, structs, and interfaces in a file
func collectStructuralMetrics(result scanner.Result, analyzers *AnalyzerSet, collectedMetrics *CollectedMetrics, cfg *config.Config) {
	if functions, err := analyzeFunctionsInFile(analyzers.Function, result, cfg); err == nil {
		collectedMetrics.Functions = append(collectedMetrics.Functions, functions...)
	}

	if structs, err := analyzeStructsInFile(analyzers.Struct, result, cfg); err == nil {
		collectedMetrics.Structs = append(collectedMetrics.Structs, structs...)
	}

	if interfaces, err := analyzeInterfacesInFile(analyzers.Interface, result, cfg); err == nil {
		collectedMetrics.Interfaces = append(collectedMetrics.Interfaces, interfaces...)
	}

	if generics, err := analyzeGenericsInFile(analyzers.Generic, result, cfg); err == nil {
		collectedMetrics.Generics = append(collectedMetrics.Generics, generics)
	}
}

// analyzePackageStructure analyzes package information for a file
func analyzePackageStructure(result scanner.Result, analyzers *AnalyzerSet, cfg *config.Config) {
	if err := analyzePackageInFile(analyzers.Package, result, cfg); err != nil && cfg.Output.Verbose {
		fmt.Fprintf(os.Stderr, "Warning: failed to analyze package in %s: %v\n",
			result.FileInfo.Path, err)
	}
}

// analyzeConcurrencyPatterns analyzes concurrency patterns in a file
func analyzeConcurrencyPatterns(result scanner.Result, analyzers *AnalyzerSet, report *metrics.Report, cfg *config.Config) {
	if err := analyzeConcurrencyInFile(analyzers.Concurrency, result, report, cfg); err != nil && cfg.Output.Verbose {
		fmt.Fprintf(os.Stderr, "Warning: failed to analyze concurrency in %s: %v\n",
			result.FileInfo.Path, err)
	}
}

// analyzeBurdenIndicators analyzes maintenance burden indicators in a file
func analyzeBurdenIndicators(result scanner.Result, analyzers *AnalyzerSet, report *metrics.Report, cfg *config.Config) {
	if err := analyzeBurdenInFile(analyzers.Burden, result, report, cfg); err != nil && cfg.Output.Verbose {
		fmt.Fprintf(os.Stderr, "Warning: failed to analyze burden in %s: %v\n",
			result.FileInfo.Path, err)
	}
}

// logProcessingSummary logs a summary of the processing results
func logProcessingSummary(processedFiles int, collectedMetrics *CollectedMetrics, cfg *config.Config) {
	if cfg.Output.Verbose {
		fmt.Fprintf(os.Stderr, "Processed %d files, found %d functions, %d structs, %d interfaces\n",
			processedFiles, len(collectedMetrics.Functions), len(collectedMetrics.Structs), len(collectedMetrics.Interfaces))
	}
}

// analyzeFunctionsInFile analyzes functions in a single file result
func analyzeFunctionsInFile(functionAnalyzer *analyzer.FunctionAnalyzer, result scanner.Result, cfg *config.Config) ([]metrics.FunctionMetrics, error) {
	functions, err := functionAnalyzer.AnalyzeFunctionsWithPath(result.File, result.FileInfo.Package, result.FileInfo.RelPath)
	if err != nil && cfg.Output.Verbose {
		fmt.Fprintf(os.Stderr, "Warning: failed to analyze functions in %s: %v\n",
			result.FileInfo.Path, err)
		return nil, err
	}
	return functions, nil
}

// analyzeStructsInFile analyzes structs in a single file result
func analyzeStructsInFile(structAnalyzer *analyzer.StructAnalyzer, result scanner.Result, cfg *config.Config) ([]metrics.StructMetrics, error) {
	structs, err := structAnalyzer.AnalyzeStructsWithPath(result.File, result.FileInfo.Package, result.FileInfo.RelPath)
	if err != nil && cfg.Output.Verbose {
		fmt.Fprintf(os.Stderr, "Warning: failed to analyze structs in %s: %v\n",
			result.FileInfo.Path, err)
		return nil, err
	}
	return structs, nil
}

// analyzeInterfacesInFile analyzes interfaces in a single file result
func analyzeInterfacesInFile(interfaceAnalyzer *analyzer.InterfaceAnalyzer, result scanner.Result, cfg *config.Config) ([]metrics.InterfaceMetrics, error) {
	interfaces, err := interfaceAnalyzer.AnalyzeInterfacesWithPath(result.File, result.FileInfo.Package, result.FileInfo.RelPath)
	if err != nil && cfg.Output.Verbose {
		fmt.Fprintf(os.Stderr, "Warning: failed to analyze interfaces in %s: %v\n",
			result.FileInfo.Path, err)
		return nil, err
	}
	return interfaces, nil
}

// analyzeGenericsInFile analyzes generic types and functions in a single file result
func analyzeGenericsInFile(genericAnalyzer *analyzer.GenericAnalyzer, result scanner.Result, cfg *config.Config) (metrics.GenericMetrics, error) {
	generics, err := genericAnalyzer.AnalyzeGenerics(result.File, result.FileInfo.Package, result.FileInfo.RelPath)
	if err != nil && cfg.Output.Verbose {
		fmt.Fprintf(os.Stderr, "Warning: failed to analyze generics in %s: %v\n",
			result.FileInfo.Path, err)
		return metrics.GenericMetrics{}, err
	}
	return generics, nil
}

// analyzePackageInFile analyzes package information for a single file result
func analyzePackageInFile(packageAnalyzer *analyzer.PackageAnalyzer, result scanner.Result, cfg *config.Config) error {
	return packageAnalyzer.AnalyzePackage(result.File, result.FileInfo.Path)
}

// analyzeConcurrencyInFile analyzes concurrency patterns in a single file and aggregates to report
func analyzeConcurrencyInFile(concurrencyAnalyzer *analyzer.ConcurrencyAnalyzer, result scanner.Result, report *metrics.Report, cfg *config.Config) error {
	concurrencyMetrics, err := concurrencyAnalyzer.AnalyzeConcurrency(result.File, result.FileInfo.Package)
	if err != nil {
		return err
	}

	aggregateConcurrencyMetrics(report, &concurrencyMetrics)
	return nil
}

// aggregateConcurrencyMetrics aggregates concurrency metrics into the report
func aggregateConcurrencyMetrics(report *metrics.Report, concurrencyMetrics *metrics.ConcurrencyPatternMetrics) {
	report.Patterns.ConcurrencyPatterns.Goroutines.Instances = append(report.Patterns.ConcurrencyPatterns.Goroutines.Instances, concurrencyMetrics.Goroutines.Instances...)
	report.Patterns.ConcurrencyPatterns.Goroutines.GoroutineLeaks = append(report.Patterns.ConcurrencyPatterns.Goroutines.GoroutineLeaks, concurrencyMetrics.Goroutines.GoroutineLeaks...)
	report.Patterns.ConcurrencyPatterns.Channels.Instances = append(report.Patterns.ConcurrencyPatterns.Channels.Instances, concurrencyMetrics.Channels.Instances...)
	report.Patterns.ConcurrencyPatterns.SyncPrims.Mutexes = append(report.Patterns.ConcurrencyPatterns.SyncPrims.Mutexes, concurrencyMetrics.SyncPrims.Mutexes...)
	report.Patterns.ConcurrencyPatterns.SyncPrims.RWMutexes = append(report.Patterns.ConcurrencyPatterns.SyncPrims.RWMutexes, concurrencyMetrics.SyncPrims.RWMutexes...)
	report.Patterns.ConcurrencyPatterns.SyncPrims.WaitGroups = append(report.Patterns.ConcurrencyPatterns.SyncPrims.WaitGroups, concurrencyMetrics.SyncPrims.WaitGroups...)
	report.Patterns.ConcurrencyPatterns.SyncPrims.Once = append(report.Patterns.ConcurrencyPatterns.SyncPrims.Once, concurrencyMetrics.SyncPrims.Once...)
	report.Patterns.ConcurrencyPatterns.SyncPrims.Cond = append(report.Patterns.ConcurrencyPatterns.SyncPrims.Cond, concurrencyMetrics.SyncPrims.Cond...)
	report.Patterns.ConcurrencyPatterns.SyncPrims.Atomic = append(report.Patterns.ConcurrencyPatterns.SyncPrims.Atomic, concurrencyMetrics.SyncPrims.Atomic...)
	report.Patterns.ConcurrencyPatterns.WorkerPools = append(report.Patterns.ConcurrencyPatterns.WorkerPools, concurrencyMetrics.WorkerPools...)
	report.Patterns.ConcurrencyPatterns.Pipelines = append(report.Patterns.ConcurrencyPatterns.Pipelines, concurrencyMetrics.Pipelines...)
	report.Patterns.ConcurrencyPatterns.FanOut = append(report.Patterns.ConcurrencyPatterns.FanOut, concurrencyMetrics.FanOut...)
	report.Patterns.ConcurrencyPatterns.FanIn = append(report.Patterns.ConcurrencyPatterns.FanIn, concurrencyMetrics.FanIn...)
	report.Patterns.ConcurrencyPatterns.Semaphores = append(report.Patterns.ConcurrencyPatterns.Semaphores, concurrencyMetrics.Semaphores...)
}

// analyzeBurdenInFile analyzes maintenance burden indicators in a single file
func analyzeBurdenInFile(burdenAnalyzer *analyzer.BurdenAnalyzer, result scanner.Result, report *metrics.Report, cfg *config.Config) error {
	// File-level analysis: magic numbers and dead code
	magicNumbers := burdenAnalyzer.DetectMagicNumbers(result.File, result.FileInfo.Package)
	report.Burden.MagicNumbers = append(report.Burden.MagicNumbers, magicNumbers...)

	mergeDeadCodeMetrics(burdenAnalyzer, result, report)

	// Function-level analysis
	analyzeFunctionBurden(burdenAnalyzer, result, report, cfg)

	return nil
}

// mergeDeadCodeMetrics analyzes and merges dead code detection results
func mergeDeadCodeMetrics(burdenAnalyzer *analyzer.BurdenAnalyzer, result scanner.Result, report *metrics.Report) {
	deadCode := burdenAnalyzer.DetectDeadCode([]*ast.File{result.File}, result.FileInfo.Package)
	if deadCode == nil {
		return
	}
	report.Burden.DeadCode.UnreferencedFunctions = append(report.Burden.DeadCode.UnreferencedFunctions, deadCode.UnreferencedFunctions...)
	report.Burden.DeadCode.UnreachableCode = append(report.Burden.DeadCode.UnreachableCode, deadCode.UnreachableCode...)
	report.Burden.DeadCode.TotalDeadLines += deadCode.TotalDeadLines
}

// analyzeFunctionBurden analyzes function-level burden indicators
func analyzeFunctionBurden(burdenAnalyzer *analyzer.BurdenAnalyzer, result scanner.Result, report *metrics.Report, cfg *config.Config) {
	ast.Inspect(result.File, func(n ast.Node) bool {
		fn, ok := n.(*ast.FuncDecl)
		if !ok || fn.Body == nil {
			return true
		}

		analyzeSignatureAndNesting(burdenAnalyzer, fn, report, cfg)
		analyzeFeatureEnvy(burdenAnalyzer, fn, result.File, report, cfg)

		return true
	})
}

// analyzeSignatureAndNesting detects signature complexity and deep nesting
func analyzeSignatureAndNesting(burdenAnalyzer *analyzer.BurdenAnalyzer, fn *ast.FuncDecl, report *metrics.Report, cfg *config.Config) {
	if sigIssue := burdenAnalyzer.AnalyzeSignatureComplexity(fn, cfg.Analysis.Burden.MaxParams, cfg.Analysis.Burden.MaxReturns); sigIssue != nil {
		report.Burden.ComplexSignatures = append(report.Burden.ComplexSignatures, *sigIssue)
	}

	if nestingIssue := burdenAnalyzer.DetectDeepNesting(fn, cfg.Analysis.Burden.MaxNesting); nestingIssue != nil {
		report.Burden.DeeplyNestedFunctions = append(report.Burden.DeeplyNestedFunctions, *nestingIssue)
	}
}

// analyzeFeatureEnvy detects feature envy in methods
func analyzeFeatureEnvy(burdenAnalyzer *analyzer.BurdenAnalyzer, fn *ast.FuncDecl, file *ast.File, report *metrics.Report, cfg *config.Config) {
	if fn.Recv == nil {
		return
	}
	if envyIssue := burdenAnalyzer.DetectFeatureEnvy(fn, file, cfg.Analysis.Burden.FeatureEnvyRatio); envyIssue != nil {
		report.Burden.FeatureEnvyMethods = append(report.Burden.FeatureEnvyMethods, *envyIssue)
	}
}
