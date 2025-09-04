package cmd

import (
	"context"
	"fmt"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/opd-ai/go-stats-generator/internal/analyzer"
	"github.com/opd-ai/go-stats-generator/internal/config"
	"github.com/opd-ai/go-stats-generator/internal/metrics"
	"github.com/opd-ai/go-stats-generator/internal/reporter"
	"github.com/opd-ai/go-stats-generator/internal/scanner"
)

var (
	outputFormat string
	outputFile   string
	workers      int
	timeout      time.Duration
)

// analyzeCmd represents the analyze command
var analyzeCmd = &cobra.Command{
	Use:   "analyze [directory|file]",
	Short: "Analyze Go source code in a directory or single file",
	Long: `Analyze Go source code in the specified directory or file and generate comprehensive
statistics about code structure, complexity, and patterns.

The analyze command can operate in two modes:
  • Directory mode: recursively scans for Go source files and processes them concurrently
  • File mode: analyzes a single Go source file

Both modes generate detailed metrics including:

  • Function and method length analysis
  • Struct complexity and member categorization
  • Cyclomatic complexity calculations
  • Design pattern detection
  • Code smell identification
  • Documentation quality assessment
  • Generic usage analysis (Go 1.18+)

Examples:
  # Analyze current directory with console output
  go-stats-generator analyze .

  # Analyze specific directory with JSON output
  go-stats-generator analyze ./src --format json --output report.json

  # Analyze a single file
  go-stats-generator analyze ./main.go

  # Analyze a single file with detailed output
  go-stats-generator analyze ./internal/analyzer/function.go --format json --verbose

  # Analyze with custom worker count and timeout
  go-stats-generator analyze . --workers 8 --timeout 5m

  # Analyze excluding test files
  go-stats-generator analyze . --skip-tests

  # Analyze with verbose output
  go-stats-generator analyze . --verbose`,

	Args: cobra.MaximumNArgs(1),
	RunE: runAnalyze,
}

func init() {
	rootCmd.AddCommand(analyzeCmd)

	// Output flags
	analyzeCmd.Flags().StringVarP(&outputFormat, "format", "f", "console",
		"output format (console, json, csv, html, markdown)")
	analyzeCmd.Flags().StringVarP(&outputFile, "output", "o", "",
		"output file (default: stdout)")
	analyzeCmd.Flags().Bool("verbose", false,
		"enable verbose output")

	// Performance flags
	analyzeCmd.Flags().IntVarP(&workers, "workers", "w", 0,
		"number of worker goroutines (default: number of CPU cores)")
	analyzeCmd.Flags().DurationVar(&timeout, "timeout", 10*time.Minute,
		"analysis timeout")

	// Filter flags
	analyzeCmd.Flags().Bool("skip-vendor", true,
		"skip vendor directories")
	analyzeCmd.Flags().Bool("skip-tests", false,
		"skip test files (*_test.go)")
	analyzeCmd.Flags().Bool("skip-generated", true,
		"skip generated files")
	analyzeCmd.Flags().StringSlice("exclude", []string{},
		"exclude patterns (glob)")
	analyzeCmd.Flags().StringSlice("include", []string{"**/*.go"},
		"include patterns (glob)")

	// Analysis flags
	analyzeCmd.Flags().Bool("include-patterns", true,
		"include design pattern detection")
	analyzeCmd.Flags().Bool("include-complexity", true,
		"include complexity analysis")
	analyzeCmd.Flags().Bool("include-documentation", true,
		"include documentation analysis")
	analyzeCmd.Flags().Bool("include-generics", true,
		"include generic usage analysis")

	// Threshold flags
	analyzeCmd.Flags().Int("max-function-length", 30,
		"maximum function length warning threshold")
	analyzeCmd.Flags().Int("max-complexity", 10,
		"maximum cyclomatic complexity warning threshold")
	analyzeCmd.Flags().Float64("min-doc-coverage", 0.7,
		"minimum documentation coverage warning threshold")

	// Bind flags to viper
	viper.BindPFlag("output.format", analyzeCmd.Flags().Lookup("format"))
	viper.BindPFlag("output.destination", analyzeCmd.Flags().Lookup("output"))
	viper.BindPFlag("output.verbose", analyzeCmd.Flags().Lookup("verbose"))
	viper.BindPFlag("performance.worker_count", analyzeCmd.Flags().Lookup("workers"))
	viper.BindPFlag("performance.timeout", analyzeCmd.Flags().Lookup("timeout"))
	viper.BindPFlag("filters.skip_vendor", analyzeCmd.Flags().Lookup("skip-vendor"))
	viper.BindPFlag("filters.skip_test_files", analyzeCmd.Flags().Lookup("skip-tests"))
	viper.BindPFlag("filters.skip_generated", analyzeCmd.Flags().Lookup("skip-generated"))
	viper.BindPFlag("filters.exclude_patterns", analyzeCmd.Flags().Lookup("exclude"))
	viper.BindPFlag("filters.include_patterns", analyzeCmd.Flags().Lookup("include"))
	viper.BindPFlag("analysis.include_patterns", analyzeCmd.Flags().Lookup("include-patterns"))
	viper.BindPFlag("analysis.include_complexity", analyzeCmd.Flags().Lookup("include-complexity"))
	viper.BindPFlag("analysis.include_documentation", analyzeCmd.Flags().Lookup("include-documentation"))
	viper.BindPFlag("analysis.include_generics", analyzeCmd.Flags().Lookup("include-generics"))
	viper.BindPFlag("analysis.max_function_length", analyzeCmd.Flags().Lookup("max-function-length"))
	viper.BindPFlag("analysis.max_cyclomatic_complexity", analyzeCmd.Flags().Lookup("max-complexity"))
	viper.BindPFlag("analysis.min_documentation_coverage", analyzeCmd.Flags().Lookup("min-doc-coverage"))
}

func runAnalyze(cmd *cobra.Command, args []string) error {
	// Determine target path
	targetPath := "."
	if len(args) > 0 {
		targetPath = args[0]
	}

	// Convert to absolute path
	absPath, err := filepath.Abs(targetPath)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}

	// Check if path exists
	fileInfo, err := os.Stat(absPath)
	if os.IsNotExist(err) {
		return fmt.Errorf("path does not exist: %s", absPath)
	}

	// Load configuration
	cfg := loadConfiguration()

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Performance.Timeout)
	defer cancel()

	// Run analysis based on whether target is file or directory
	var report *metrics.Report
	if fileInfo.IsDir() {
		report, err = runDirectoryAnalysis(ctx, absPath, cfg)
	} else {
		report, err = runFileAnalysis(ctx, absPath, cfg)
	}

	if err != nil {
		return fmt.Errorf("analysis failed: %w", err)
	}

	// Generate output
	err = generateOutput(report, cfg)
	if err != nil {
		return fmt.Errorf("failed to generate output: %w", err)
	}

	return nil
}

func loadConfiguration() *config.Config {
	cfg := config.DefaultConfig()

	// Load configuration sections
	loadOutputConfiguration(cfg)
	loadPerformanceConfiguration(cfg)
	loadFilterConfiguration(cfg)
	loadAnalysisConfiguration(cfg)

	return cfg
}

// loadOutputConfiguration loads output-related configuration from viper
func loadOutputConfiguration(cfg *config.Config) {
	if viper.IsSet("output.format") {
		cfg.Output.Format = config.OutputFormat(viper.GetString("output.format"))
	}
	if viper.IsSet("output.destination") {
		cfg.Output.Destination = viper.GetString("output.destination")
	}
	if viper.IsSet("output.verbose") {
		cfg.Output.Verbose = viper.GetBool("output.verbose")
	}
}

// loadPerformanceConfiguration loads performance-related configuration from viper
func loadPerformanceConfiguration(cfg *config.Config) {
	if viper.IsSet("performance.worker_count") {
		cfg.Performance.WorkerCount = viper.GetInt("performance.worker_count")
	}
	if viper.IsSet("performance.timeout") {
		cfg.Performance.Timeout = viper.GetDuration("performance.timeout")
	}
}

// loadFilterConfiguration loads filter-related configuration from viper
func loadFilterConfiguration(cfg *config.Config) {
	if viper.IsSet("filters.skip_vendor") {
		cfg.Filters.SkipVendor = viper.GetBool("filters.skip_vendor")
	}
	if viper.IsSet("filters.skip_test_files") {
		cfg.Filters.SkipTestFiles = viper.GetBool("filters.skip_test_files")
	}
	if viper.IsSet("filters.skip_generated") {
		cfg.Filters.SkipGenerated = viper.GetBool("filters.skip_generated")
	}
	if viper.IsSet("filters.exclude_patterns") {
		cfg.Filters.ExcludePatterns = viper.GetStringSlice("filters.exclude_patterns")
	}
	if viper.IsSet("filters.include_patterns") {
		cfg.Filters.IncludePatterns = viper.GetStringSlice("filters.include_patterns")
	}
}

// loadAnalysisConfiguration loads analysis-related configuration from viper
func loadAnalysisConfiguration(cfg *config.Config) {
	if viper.IsSet("analysis.include_patterns") {
		cfg.Analysis.IncludePatterns = viper.GetBool("analysis.include_patterns")
	}
	if viper.IsSet("analysis.include_complexity") {
		cfg.Analysis.IncludeComplexity = viper.GetBool("analysis.include_complexity")
	}
	if viper.IsSet("analysis.include_documentation") {
		cfg.Analysis.IncludeDocumentation = viper.GetBool("analysis.include_documentation")
	}
	if viper.IsSet("analysis.include_generics") {
		cfg.Analysis.IncludeGenerics = viper.GetBool("analysis.include_generics")
	}
	if viper.IsSet("analysis.max_function_length") {
		cfg.Analysis.MaxFunctionLength = viper.GetInt("analysis.max_function_length")
	}
	if viper.IsSet("analysis.max_cyclomatic_complexity") {
		cfg.Analysis.MaxCyclomaticComplexity = viper.GetInt("analysis.max_cyclomatic_complexity")
	}
	if viper.IsSet("analysis.min_documentation_coverage") {
		cfg.Analysis.MinDocumentationCoverage = viper.GetFloat64("analysis.min_documentation_coverage")
	}
}

func runDirectoryAnalysis(ctx context.Context, targetDir string, cfg *config.Config) (*metrics.Report, error) {
	return runAnalysisWorkflow(ctx, targetDir, cfg)
}

func runFileAnalysis(ctx context.Context, filePath string, cfg *config.Config) (*metrics.Report, error) {
	startTime := time.Now()

	// Validate the file is a Go source file
	if !isGoSourceFile(filePath) {
		return nil, fmt.Errorf("file %s is not a Go source file", filePath)
	}

	if cfg.Output.Verbose {
		fmt.Fprintf(os.Stderr, "Analyzing file: %s\n", filePath)
	}

	// Find project root for relative path calculation
	projectRoot := findProjectRoot(filePath)

	// Create discoverer and parse the file
	discoverer := scanner.NewDiscoverer(&cfg.Filters)

	// Parse the single file
	file, err := discoverer.ParseFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file %s: %w", filePath, err)
	}

	// Get file info for the single file
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	// Calculate relative path from project root
	relPath := filePath
	if projectRoot != "" {
		if rel, err := filepath.Rel(projectRoot, filePath); err == nil {
			relPath = rel
		}
	} else {
		relPath = filepath.Base(filePath)
	}

	// Create a FileInfo struct for the file
	scannerFileInfo := scanner.FileInfo{
		Path:        filePath,
		RelPath:     relPath,
		Size:        fileInfo.Size(),
		IsTestFile:  strings.HasSuffix(filePath, "_test.go"),
		IsGenerated: false, // Will be determined during analysis
	}

	// Get package name from the parsed file
	if file.Name != nil {
		scannerFileInfo.Package = file.Name.Name
	}

	// Create result for worker-like processing
	result := scanner.Result{
		FileInfo: scannerFileInfo,
		File:     file,
		Error:    nil,
	}

	// Create analyzers
	analyzers := createAnalyzers(discoverer.GetFileSet())
	report := createInitialReport(filepath.Dir(filePath), startTime, 1)

	// Process the single file
	collectedMetrics := &CollectedMetrics{}
	processFileAnalysis(result, analyzers, collectedMetrics, report, cfg)

	// Finalize report
	finalizeReport(report, collectedMetrics, analyzers.Package, cfg)
	report.Metadata.AnalysisTime = time.Since(startTime)

	if cfg.Output.Verbose {
		fmt.Fprintf(os.Stderr, "Analyzed 1 file, found %d functions, %d structs, %d interfaces\n",
			len(collectedMetrics.Functions), len(collectedMetrics.Structs), len(collectedMetrics.Interfaces))
	}

	return report, nil
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
	analyzers := createAnalyzers(discoverer.GetFileSet())
	report := createInitialReport(targetDir, startTime, len(files))

	// Step 4: Process analysis results from worker pool
	metrics, packageAnalyzer, err := processAnalysisResults(ctx, results, analyzers, report, cfg)
	if err != nil {
		return nil, err
	}

	// Step 5: Finalize report with all collected metrics
	finalizeReport(report, metrics, packageAnalyzer, cfg)
	report.Metadata.AnalysisTime = time.Since(startTime)

	return report, nil
}

// AnalyzerSet holds all the different analyzers used in the workflow
type AnalyzerSet struct {
	Function    *analyzer.FunctionAnalyzer
	Struct      *analyzer.StructAnalyzer
	Interface   *analyzer.InterfaceAnalyzer
	Package     *analyzer.PackageAnalyzer
	Concurrency *analyzer.ConcurrencyAnalyzer
}

// CollectedMetrics holds all metrics collected during analysis
type CollectedMetrics struct {
	Functions  []metrics.FunctionMetrics
	Structs    []metrics.StructMetrics
	Interfaces []metrics.InterfaceMetrics
	TotalLines int
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
func createAnalyzers(fileSet *token.FileSet) *AnalyzerSet {
	return &AnalyzerSet{
		Function:    analyzer.NewFunctionAnalyzer(fileSet),
		Struct:      analyzer.NewStructAnalyzer(fileSet),
		Interface:   analyzer.NewInterfaceAnalyzer(fileSet),
		Package:     analyzer.NewPackageAnalyzer(fileSet),
		Concurrency: analyzer.NewConcurrencyAnalyzer(fileSet),
	}
}

// createInitialReport creates the initial report structure with metadata and empty pattern metrics
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
	}
}

// processAnalysisResults processes the worker results and collects all metrics
// processAnalysisResults coordinates the analysis of all scanner results
func processAnalysisResults(ctx context.Context, results <-chan scanner.Result, analyzers *AnalyzerSet, report *metrics.Report, cfg *config.Config) (*CollectedMetrics, *analyzer.PackageAnalyzer, error) {
	collectedMetrics := &CollectedMetrics{}
	processedFiles := 0

	for {
		select {
		case result, ok := <-results:
			if !ok {
				// Channel is closed, all results processed
				logProcessingSummary(processedFiles, collectedMetrics, cfg)
				return collectedMetrics, analyzers.Package, nil
			}

			processedFiles++

			if !handleScannerError(result.Error, cfg) {
				continue
			}

			processFileAnalysis(result, analyzers, collectedMetrics, report, cfg)

		case <-ctx.Done():
			// Context cancelled, return with error
			return nil, nil, fmt.Errorf("analysis cancelled: %w", ctx.Err())
		}
	}
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
	collectStructuralMetrics(result, analyzers, collectedMetrics, cfg)
	analyzePackageStructure(result, analyzers, cfg)
	analyzeConcurrencyPatterns(result, analyzers, report, cfg)
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

// finalizeReport populates the report with collected metrics and generates final package report
func finalizeReport(report *metrics.Report, collectedMetrics *CollectedMetrics, packageAnalyzer *analyzer.PackageAnalyzer, cfg *config.Config) {
	// Generate package report
	packageReport, err := packageAnalyzer.GenerateReport()
	if err != nil {
		if cfg.Output.Verbose {
			fmt.Fprintf(os.Stderr, "Warning: failed to generate package report: %v\n", err)
		}
		packageReport = &metrics.PackageReport{
			Packages:      []metrics.PackageMetrics{},
			TotalPackages: 0,
		}
	}

	// Populate main metrics
	report.Functions = collectedMetrics.Functions
	report.Structs = collectedMetrics.Structs
	report.Interfaces = collectedMetrics.Interfaces
	report.Packages = packageReport.Packages

	// Calculate overview metrics
	calculateOverviewMetrics(report, collectedMetrics, packageReport)

	// Finalize concurrency metrics summary statistics
	finalizeConcurrencyMetrics(report)
}

// calculateOverviewMetrics calculates and sets the overview metrics in the report
func calculateOverviewMetrics(report *metrics.Report, collectedMetrics *CollectedMetrics, packageReport *metrics.PackageReport) {
	report.Overview = metrics.OverviewMetrics{
		TotalLinesOfCode: collectedMetrics.TotalLines,
		TotalFunctions:   len(collectedMetrics.Functions),
		TotalStructs:     len(collectedMetrics.Structs),
		TotalInterfaces:  len(collectedMetrics.Interfaces),
		TotalPackages:    packageReport.TotalPackages,
		TotalFiles:       report.Metadata.FilesProcessed,
	}

	// Count methods vs functions
	for _, fn := range collectedMetrics.Functions {
		if fn.IsMethod {
			report.Overview.TotalMethods++
		}
	}
	report.Overview.TotalFunctions -= report.Overview.TotalMethods
}

// finalizeConcurrencyMetrics calculates final concurrency metric summaries
func finalizeConcurrencyMetrics(report *metrics.Report) {
	report.Patterns.ConcurrencyPatterns.Goroutines.TotalCount = len(report.Patterns.ConcurrencyPatterns.Goroutines.Instances)
	for _, instance := range report.Patterns.ConcurrencyPatterns.Goroutines.Instances {
		if instance.IsAnonymous {
			report.Patterns.ConcurrencyPatterns.Goroutines.AnonymousCount++
		} else {
			report.Patterns.ConcurrencyPatterns.Goroutines.NamedCount++
		}
	}

	report.Patterns.ConcurrencyPatterns.Channels.TotalCount = len(report.Patterns.ConcurrencyPatterns.Channels.Instances)
	for _, instance := range report.Patterns.ConcurrencyPatterns.Channels.Instances {
		if instance.IsBuffered {
			report.Patterns.ConcurrencyPatterns.Channels.BufferedCount++
		} else {
			report.Patterns.ConcurrencyPatterns.Channels.UnbufferedCount++
		}
		if instance.IsDirectional {
			report.Patterns.ConcurrencyPatterns.Channels.DirectionalCount++
		}
	}
}

func generateOutput(report *metrics.Report, cfg *config.Config) error {
	// Create appropriate reporter using the factory
	rep, err := reporter.NewReporter(string(cfg.Output.Format))
	if err != nil {
		return fmt.Errorf("failed to create reporter: %w", err)
	}

	// Determine output destination
	var output *os.File
	if cfg.Output.Destination == "" || cfg.Output.Destination == "stdout" {
		output = os.Stdout
	} else {
		output, err = os.Create(cfg.Output.Destination)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer output.Close()
	}

	// Generate report
	err = rep.Generate(report, output)
	if err != nil {
		return fmt.Errorf("failed to generate report: %w", err)
	}

	return nil
}
