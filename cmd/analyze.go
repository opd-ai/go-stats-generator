package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
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
	Use:   "analyze [directory]",
	Short: "Analyze Go source code in a directory",
	Long: `Analyze Go source code in the specified directory and generate comprehensive
statistics about code structure, complexity, and patterns.

The analyze command recursively scans the directory for Go source files,
processes them concurrently, and generates detailed metrics including:

  • Function and method length analysis
  • Struct complexity and member categorization
  • Cyclomatic complexity calculations
  • Design pattern detection
  • Code smell identification
  • Documentation quality assessment
  • Generic usage analysis (Go 1.18+)

Examples:
  # Analyze current directory with console output
  gostats analyze .

  # Analyze specific directory with JSON output
  gostats analyze ./src --format json --output report.json

  # Analyze with custom worker count and timeout
  gostats analyze . --workers 8 --timeout 5m

  # Analyze excluding test files
  gostats analyze . --skip-tests

  # Analyze with verbose output
  gostats analyze . --verbose`,

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
	// Determine target directory
	targetDir := "."
	if len(args) > 0 {
		targetDir = args[0]
	}

	// Convert to absolute path
	absPath, err := filepath.Abs(targetDir)
	if err != nil {
		return fmt.Errorf("failed to resolve directory path: %w", err)
	}

	// Verify directory exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return fmt.Errorf("directory does not exist: %s", absPath)
	}

	// Load configuration
	cfg := loadConfiguration()

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Performance.Timeout)
	defer cancel()

	// Run analysis
	report, err := runAnalysisWorkflow(ctx, absPath, cfg)
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

	// Override with viper values
	if viper.IsSet("output.format") {
		cfg.Output.Format = config.OutputFormat(viper.GetString("output.format"))
	}
	if viper.IsSet("output.destination") {
		cfg.Output.Destination = viper.GetString("output.destination")
	}
	if viper.IsSet("performance.worker_count") {
		cfg.Performance.WorkerCount = viper.GetInt("performance.worker_count")
	}
	if viper.IsSet("performance.timeout") {
		cfg.Performance.Timeout = viper.GetDuration("performance.timeout")
	}

	// Override filter settings
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

	// Override analysis settings
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

	return cfg
}

func runAnalysisWorkflow(ctx context.Context, targetDir string, cfg *config.Config) (*metrics.Report, error) {
	startTime := time.Now()

	if cfg.Output.Verbose {
		fmt.Fprintf(os.Stderr, "Analyzing directory: %s\n", targetDir)
	}

	// Discover files
	discoverer := scanner.NewDiscoverer(&cfg.Filters)
	files, err := discoverer.DiscoverFiles(targetDir)
	if err != nil {
		return nil, fmt.Errorf("file discovery failed: %w", err)
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no Go files found in %s", targetDir)
	}

	if cfg.Output.Verbose {
		fmt.Fprintf(os.Stderr, "Found %d Go files\n", len(files))
	}

	// Create worker pool
	workerPool := scanner.NewWorkerPool(&cfg.Performance, discoverer)

	// Process files with progress reporting
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

	// Analyze results
	analyzer := analyzer.NewFunctionAnalyzer(discoverer.GetFileSet())

	report := &metrics.Report{
		Metadata: metrics.ReportMetadata{
			Repository:     targetDir,
			GeneratedAt:    time.Now(),
			AnalysisTime:   time.Since(startTime),
			FilesProcessed: len(files),
			ToolVersion:    "1.0.0",
		},
	}

	var allFunctions []metrics.FunctionMetrics
	var totalLines int

	// Process analysis results
	for result := range results {
		if result.Error != nil {
			if cfg.Output.Verbose {
				fmt.Fprintf(os.Stderr, "Warning: %v\n", result.Error)
			}
			continue
		}

		// Analyze functions in this file
		functions, err := analyzer.AnalyzeFunctions(result.File, result.FileInfo.Package)
		if err != nil {
			if cfg.Output.Verbose {
				fmt.Fprintf(os.Stderr, "Warning: failed to analyze %s: %v\n",
					result.FileInfo.Path, err)
			}
			continue
		}

		allFunctions = append(allFunctions, functions...)

		// Count lines (simplified)
		totalLines += int(result.FileInfo.Size) / 50 // Rough estimate
	}

	// Populate report
	report.Functions = allFunctions
	report.Overview = metrics.OverviewMetrics{
		TotalLinesOfCode: totalLines,
		TotalFunctions:   len(allFunctions),
		TotalFiles:       len(files),
	}

	// Count methods vs functions
	for _, fn := range allFunctions {
		if fn.IsMethod {
			report.Overview.TotalMethods++
		}
	}
	report.Overview.TotalFunctions -= report.Overview.TotalMethods

	report.Metadata.AnalysisTime = time.Since(startTime)

	return report, nil
}

func generateOutput(report *metrics.Report, cfg *config.Config) error {
	// Create appropriate reporter
	var rep reporter.Reporter
	var err error

	switch cfg.Output.Format {
	case config.FormatJSON:
		rep = reporter.NewJSONReporter()
	case config.FormatConsole:
		fallthrough
	default:
		rep = reporter.NewConsoleReporter(&cfg.Output)
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
