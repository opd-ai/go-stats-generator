// Package reporter provides output formatting for code statistics analysis reports.
//
// The reporter package implements multiple output formats for presenting code metrics
// and analysis results. Each format is optimized for different use cases:
//
//   - Console: Rich terminal output with color-coded tables and symbols for human reading
//   - JSON: Structured data for programmatic consumption and API integration
//   - HTML: Interactive web reports with sortable tables and visual charts
//   - CSV: Spreadsheet-compatible output for data analysis in Excel/LibreOffice
//   - Markdown: GitHub-flavored markdown for README files and documentation
//
// # Reporter Interface
//
// All reporters implement the Reporter interface which defines two core methods:
//
//   - Write: Generates a full analysis report for a single metrics snapshot
//   - WriteDiff: Generates a differential comparison report between baseline and current snapshots
//
// # Usage
//
// Creating a reporter:
//
//	reporter := reporter.NewConsoleReporter(config.DefaultConfig())
//	reporter.Write(os.Stdout, metricsReport)
//
// Generating differential reports:
//
//	diff := metrics.ComplexityDiff{
//	    Baseline: baselineSnapshot,
//	    Current:  currentSnapshot,
//	    Changes:  detectedChanges,
//	}
//	reporter.WriteDiff(os.Stdout, diff)
//
// # Format Selection
//
// Use CreateReporter() or NewReporter() factory functions to dynamically select
// the output format based on configuration:
//
//	cfg := config.Config{Format: "json"}
//	reporter, err := reporter.CreateReporter(cfg)
//
// # Section Filtering
//
// All reporters support section filtering via the config.Sections field to include
// only specific analysis sections (functions, structs, packages, documentation, etc.)
// in the generated output.
package reporter
