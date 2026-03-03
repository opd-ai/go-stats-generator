package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/opd-ai/go-stats-generator/internal/config"
	"github.com/opd-ai/go-stats-generator/internal/metrics"
	"github.com/opd-ai/go-stats-generator/internal/reporter"
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
	analyzeCmd.Flags().Bool("enforce-thresholds", false,
		"exit with non-zero code if quality thresholds are violated (for CI/CD integration)")

	// Duplication detection flags
	analyzeCmd.Flags().Int("min-block-lines", 6,
		"minimum block size to consider for duplication detection")
	analyzeCmd.Flags().Float64("similarity-threshold", 0.80,
		"similarity threshold for near-duplicate detection (0.0-1.0)")
	analyzeCmd.Flags().Bool("ignore-test-duplication", false,
		"exclude test files from duplication analysis")

	// Organization analysis flags
	analyzeCmd.Flags().Int("max-file-lines", 500,
		"maximum lines per file before flagging")
	analyzeCmd.Flags().Int("max-file-functions", 20,
		"maximum functions/methods per file")
	analyzeCmd.Flags().Int("max-file-types", 5,
		"maximum type declarations per file")
	analyzeCmd.Flags().Int("max-package-files", 20,
		"maximum files per package")
	analyzeCmd.Flags().Int("max-exported-symbols", 50,
		"maximum exported symbols per package")
	analyzeCmd.Flags().Int("max-directory-depth", 5,
		"maximum directory nesting depth")
	analyzeCmd.Flags().Int("max-file-imports", 15,
		"maximum import statements per file")

	// Maintenance burden analysis flags
	analyzeCmd.Flags().Int("max-params", 5,
		"maximum function parameters before flagging high signature complexity")
	analyzeCmd.Flags().Int("max-returns", 3,
		"maximum return values before flagging high signature complexity")
	analyzeCmd.Flags().Int("max-nesting", 4,
		"maximum nesting depth before flagging deeply nested code")
	analyzeCmd.Flags().Float64("feature-envy-ratio", 2.0,
		"threshold ratio for detecting feature envy (external references / self references)")

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
	viper.BindPFlag("analysis.enforce_thresholds", analyzeCmd.Flags().Lookup("enforce-thresholds"))
	viper.BindPFlag("analysis.duplication.min_block_lines", analyzeCmd.Flags().Lookup("min-block-lines"))
	viper.BindPFlag("analysis.duplication.similarity_threshold", analyzeCmd.Flags().Lookup("similarity-threshold"))
	viper.BindPFlag("analysis.duplication.ignore_test_files", analyzeCmd.Flags().Lookup("ignore-test-duplication"))
	viper.BindPFlag("analysis.organization.max_file_lines", analyzeCmd.Flags().Lookup("max-file-lines"))
	viper.BindPFlag("analysis.organization.max_file_functions", analyzeCmd.Flags().Lookup("max-file-functions"))
	viper.BindPFlag("analysis.organization.max_file_types", analyzeCmd.Flags().Lookup("max-file-types"))
	viper.BindPFlag("analysis.organization.max_package_files", analyzeCmd.Flags().Lookup("max-package-files"))
	viper.BindPFlag("analysis.organization.max_exported_symbols", analyzeCmd.Flags().Lookup("max-exported-symbols"))
	viper.BindPFlag("analysis.organization.max_directory_depth", analyzeCmd.Flags().Lookup("max-directory-depth"))
	viper.BindPFlag("analysis.organization.max_file_imports", analyzeCmd.Flags().Lookup("max-file-imports"))
	viper.BindPFlag("analysis.burden.max_params", analyzeCmd.Flags().Lookup("max-params"))
	viper.BindPFlag("analysis.burden.max_returns", analyzeCmd.Flags().Lookup("max-returns"))
	viper.BindPFlag("analysis.burden.max_nesting", analyzeCmd.Flags().Lookup("max-nesting"))
	viper.BindPFlag("analysis.burden.feature_envy_ratio", analyzeCmd.Flags().Lookup("feature-envy-ratio"))
}

// runAnalyze is the main entry point for the analyze command. Validates the target
// path (file or directory), loads configuration from flags and config file, performs
// analysis, enforces quality thresholds, saves baseline if requested, and outputs
// results in the specified format (console/JSON/HTML/CSV/Markdown).
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

	// Check quality gates if enforcement is enabled
	if err := checkQualityGates(report, cfg); err != nil {
		return err
	}

	return nil
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

// checkQualityGates validates that the report meets configured quality thresholds
func checkQualityGates(report *metrics.Report, cfg *config.Config) error {
	if !cfg.Analysis.EnforceThresholds {
		return nil
	}

	var violations []string

	// Check documentation coverage threshold (coverage is stored as percentage, threshold as fraction)
	thresholdPercent := cfg.Analysis.MinDocumentationCoverage * 100
	if report.Documentation.Coverage.Overall < thresholdPercent {
		violations = append(violations, fmt.Sprintf(
			"Documentation coverage (%.2f%%) is below threshold (%.2f%%)",
			report.Documentation.Coverage.Overall,
			thresholdPercent,
		))
	}

	// If violations exist, print them and return error
	if len(violations) > 0 {
		fmt.Fprintf(os.Stderr, "\n=== QUALITY GATE FAILURES ===\n")
		for _, violation := range violations {
			fmt.Fprintf(os.Stderr, "❌ %s\n", violation)
		}
		fmt.Fprintf(os.Stderr, "\nUse --enforce-thresholds=false to disable quality gate enforcement.\n")
		return fmt.Errorf("quality gates failed: %d violation(s)", len(violations))
	}

	return nil
}
