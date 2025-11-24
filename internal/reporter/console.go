package reporter

import (
	"fmt"
	"io"
	"sort"
	"time"

	"github.com/opd-ai/go-stats-generator/internal/config"
	"github.com/opd-ai/go-stats-generator/internal/metrics"
)

// ConsoleReporter generates human-readable console output
type ConsoleReporter struct {
	config    *config.OutputConfig
	useColors bool
}

// NewConsoleReporter creates a new console reporter
func NewConsoleReporter(cfg *config.OutputConfig) *ConsoleReporter {
	if cfg == nil {
		cfg = &config.OutputConfig{
			UseColors:       true,
			IncludeOverview: true,
			IncludeDetails:  true,
			Limit:           20,
		}
	}

	return &ConsoleReporter{
		config:    cfg,
		useColors: cfg.UseColors,
	}
}

// Generate generates a console report
func (cr *ConsoleReporter) Generate(report *metrics.Report, output io.Writer) error {
	// Header
	cr.writeHeader(output, report)

	// Overview section
	if cr.config.IncludeOverview {
		cr.writeOverview(output, report)
	}

	// Function analysis section
	if cr.config.IncludeDetails && len(report.Functions) > 0 {
		cr.writeFunctionAnalysis(output, report)
	}

	// Complexity analysis
	if cr.config.IncludeDetails {
		cr.writeComplexityAnalysis(output, report)
	}

	// Package analysis section
	if cr.config.IncludeDetails && len(report.Packages) > 0 {
		cr.writePackageAnalysis(output, report)
	}

	// Footer
	cr.writeFooter(output, report)

	return nil
}

// WriteDiff generates a console diff report
func (cr *ConsoleReporter) WriteDiff(output io.Writer, diff *metrics.ComplexityDiff) error {
	// Header
	cr.writeDiffHeader(output, diff)

	// Summary section
	cr.writeDiffSummary(output, diff)

	// Regressions section
	if len(diff.Regressions) > 0 {
		cr.writeDiffRegressions(output, diff.Regressions)
	}

	// Improvements section
	if len(diff.Improvements) > 0 {
		cr.writeDiffImprovements(output, diff.Improvements)
	}

	// Changes section (if requested)
	if len(diff.Changes) > 0 && cr.config.IncludeDetails {
		cr.writeDiffChanges(output, diff.Changes)
	}

	return nil
}

func (cr *ConsoleReporter) writeHeader(output io.Writer, report *metrics.Report) {
	fmt.Fprintln(output, "=== GO SOURCE CODE STATISTICS REPORT ===")
	fmt.Fprintf(output, "Repository: %s\n", report.Metadata.Repository)
	fmt.Fprintf(output, "Generated: %s\n", report.Metadata.GeneratedAt.Format("2006-01-02 15:04:05"))
	fmt.Fprintf(output, "Analysis Time: %v\n", report.Metadata.AnalysisTime.Round(time.Millisecond))
	fmt.Fprintf(output, "Files Processed: %d\n", report.Metadata.FilesProcessed)
	fmt.Fprintln(output)
}

func (cr *ConsoleReporter) writeOverview(output io.Writer, report *metrics.Report) {
	fmt.Fprintln(output, "=== OVERVIEW ===")

	overview := report.Overview
	fmt.Fprintf(output, "Total Lines of Code: %d\n", overview.TotalLinesOfCode)
	fmt.Fprintf(output, "Total Functions: %d\n", overview.TotalFunctions)
	fmt.Fprintf(output, "Total Methods: %d\n", overview.TotalMethods)
	fmt.Fprintf(output, "Total Structs: %d\n", overview.TotalStructs)
	fmt.Fprintf(output, "Total Interfaces: %d\n", overview.TotalInterfaces)
	fmt.Fprintf(output, "Total Packages: %d\n", overview.TotalPackages)
	fmt.Fprintf(output, "Total Files: %d\n", overview.TotalFiles)
	fmt.Fprintln(output)
}

func (cr *ConsoleReporter) writeFunctionAnalysis(output io.Writer, report *metrics.Report) {
	fmt.Fprintln(output, "=== FUNCTION ANALYSIS ===")

	functions := report.Functions
	if len(functions) == 0 {
		fmt.Fprintln(output, "No functions found.")
		fmt.Fprintln(output)
		return
	}

	// Calculate statistics
	stats := cr.calculateFunctionStats(functions)

	// Statistics
	fmt.Fprintln(output, "Function Statistics:")
	fmt.Fprintf(output, "  Average Function Length: %.1f lines\n", stats.AvgLength)
	fmt.Fprintf(output, "  Longest Function: %s (%d lines)\n", stats.LongestName, stats.LongestLength)
	fmt.Fprintf(output, "  Functions > 50 lines: %d (%.1f%%)\n", stats.LongFunctions, stats.LongFunctionsPct)
	fmt.Fprintf(output, "  Functions > 100 lines: %d (%.1f%%)\n", stats.VeryLongFunctions, stats.VeryLongFunctionsPct)
	fmt.Fprintf(output, "  Average Complexity: %.1f\n", stats.AvgComplexity)
	fmt.Fprintf(output, "  High Complexity (>10): %d functions\n", stats.HighComplexity)
	fmt.Fprintln(output)

	// Top complex functions
	if cr.config.IncludeDetails {
		cr.writeTopComplexFunctions(output, functions)
	}
}

func (cr *ConsoleReporter) writeComplexityAnalysis(output io.Writer, report *metrics.Report) {
	if len(report.Functions) == 0 {
		return
	}

	fmt.Fprintln(output, "=== COMPLEXITY ANALYSIS ===")

	// Sort functions by complexity
	sortedFunctions := make([]metrics.FunctionMetrics, len(report.Functions))
	copy(sortedFunctions, report.Functions)

	sort.Slice(sortedFunctions, func(i, j int) bool {
		// Primary sort: by complexity (descending)
		if sortedFunctions[i].Complexity.Overall != sortedFunctions[j].Complexity.Overall {
			return sortedFunctions[i].Complexity.Overall > sortedFunctions[j].Complexity.Overall
		}
		// Secondary sort: by function length (descending) when complexity is tied
		return sortedFunctions[i].Lines.Total > sortedFunctions[j].Lines.Total
	})

	// Show top complex functions
	limit := cr.config.Limit
	if limit > len(sortedFunctions) {
		limit = len(sortedFunctions)
	}

	fmt.Fprintf(output, "Top %d Most Complex Functions:\n", limit)
	fmt.Fprintf(output, "%-30s %-20s %8s %10s %10s\n", "Function", "Package", "Lines", "Cyclomatic", "Overall")
	fmt.Fprintln(output, "--------------------------------------------------------------------------------")

	for i := 0; i < limit; i++ {
		fn := sortedFunctions[i]
		fmt.Fprintf(output, "%-30s %-20s %8d %10d %10.1f\n",
			cr.truncate(fn.Name, 30),
			cr.truncate(fn.Package, 20),
			fn.Lines.Total,
			fn.Complexity.Cyclomatic,
			fn.Complexity.Overall,
		)
	}
	fmt.Fprintln(output)
}

func (cr *ConsoleReporter) writeTopComplexFunctions(output io.Writer, functions []metrics.FunctionMetrics) {
	// Sort by overall complexity
	sorted := make([]metrics.FunctionMetrics, len(functions))
	copy(sorted, functions)

	sort.Slice(sorted, func(i, j int) bool {
		// Primary sort: by complexity (descending)
		if sorted[i].Complexity.Overall != sorted[j].Complexity.Overall {
			return sorted[i].Complexity.Overall > sorted[j].Complexity.Overall
		}
		// Secondary sort: by function length (descending) when complexity is tied
		return sorted[i].Lines.Total > sorted[j].Lines.Total
	})

	limit := 10
	if limit > len(sorted) {
		limit = len(sorted)
	}

	fmt.Fprintln(output, "Top Complex Functions:")
	fmt.Fprintf(output, "%4s %-25s %-20s %8s %10s\n", "Rank", "Function", "File", "Lines", "Complexity")
	fmt.Fprintln(output, "-----------------------------------------------------------------------")

	for i := 0; i < limit; i++ {
		fn := sorted[i]
		fmt.Fprintf(output, "%4d %-25s %-20s %8d %10.1f\n",
			i+1,
			cr.truncate(fn.Name, 25),
			cr.truncate(fn.File, 20),
			fn.Lines.Total,
			fn.Complexity.Overall,
		)
	}
	fmt.Fprintln(output)
}

func (cr *ConsoleReporter) writeFooter(output io.Writer, report *metrics.Report) {
	fmt.Fprintln(output, "=== ANALYSIS COMPLETE ===")
	fmt.Fprintf(output, "Report generated by go-stats-generator v%s\n", report.Metadata.ToolVersion)
}

func (cr *ConsoleReporter) truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// functionStats holds calculated function statistics
type functionStats struct {
	AvgLength            float64
	LongestName          string
	LongestLength        int
	LongFunctions        int
	LongFunctionsPct     float64
	VeryLongFunctions    int
	VeryLongFunctionsPct float64
	AvgComplexity        float64
	HighComplexity       int
}

func (cr *ConsoleReporter) calculateFunctionStats(functions []metrics.FunctionMetrics) functionStats {
	if len(functions) == 0 {
		return functionStats{}
	}

	var totalLength int
	var totalComplexity float64
	var longFunctions int
	var veryLongFunctions int
	var highComplexity int

	longestName := functions[0].Name
	longestLength := functions[0].Lines.Total

	for _, fn := range functions {
		totalLength += fn.Lines.Total
		totalComplexity += fn.Complexity.Overall

		if fn.Lines.Total > longestLength {
			longestLength = fn.Lines.Total
			longestName = fn.Name
		}

		if fn.Lines.Total > 50 {
			longFunctions++
		}
		if fn.Lines.Total > 100 {
			veryLongFunctions++
		}
		if fn.Complexity.Cyclomatic > 10 {
			highComplexity++
		}
	}

	count := float64(len(functions))

	return functionStats{
		AvgLength:            float64(totalLength) / count,
		LongestName:          longestName,
		LongestLength:        longestLength,
		LongFunctions:        longFunctions,
		LongFunctionsPct:     float64(longFunctions) / count * 100,
		VeryLongFunctions:    veryLongFunctions,
		VeryLongFunctionsPct: float64(veryLongFunctions) / count * 100,
		AvgComplexity:        totalComplexity / count,
		HighComplexity:       highComplexity,
	}
}

func (cr *ConsoleReporter) writeDiffHeader(output io.Writer, diff *metrics.ComplexityDiff) {
	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "Complexity Diff Report")
	fmt.Fprintln(output, "======================")
	fmt.Fprintf(output, "Baseline: %s (%s)\n", diff.Baseline.ID, diff.Baseline.Metadata.Timestamp.Format(time.RFC3339))
	fmt.Fprintf(output, "Current:  %s (%s)\n", diff.Current.ID, diff.Current.Metadata.Timestamp.Format(time.RFC3339))
	fmt.Fprintln(output, "")
}

func (cr *ConsoleReporter) writeDiffSummary(output io.Writer, diff *metrics.ComplexityDiff) {
	fmt.Fprintln(output, "=== SUMMARY ===")

	summary := diff.Summary

	if summary.ImprovementCount > 0 {
		fmt.Fprintf(output, "✅ Improvements: %d\n", summary.ImprovementCount)
	}

	if summary.NeutralChangeCount > 0 {
		fmt.Fprintf(output, "⚠️  Neutral Changes: %d\n", summary.NeutralChangeCount)
	}

	if summary.RegressionCount > 0 {
		fmt.Fprintf(output, "❌ Regressions: %d\n", summary.RegressionCount)
	}

	if summary.CriticalIssues > 0 {
		fmt.Fprintf(output, "🚨 Critical Issues: %d\n", summary.CriticalIssues)
	}

	fmt.Fprintf(output, "Overall Trend: %s\n", string(summary.OverallTrend))
	fmt.Fprintf(output, "Quality Score: %.1f/100\n", summary.QualityScore)
	fmt.Fprintln(output, "")
}

func (cr *ConsoleReporter) writeDiffRegressions(output io.Writer, regressions []metrics.Regression) {
	fmt.Fprintln(output, "=== REGRESSIONS ===")

	for _, regression := range regressions {
		var icon string
		switch regression.Severity {
		case metrics.SeverityLevelCritical:
			icon = "🚨"
		case metrics.SeverityLevelError:
			icon = "❌"
		case metrics.SeverityLevelWarning:
			icon = "⚠️"
		default:
			icon = "ℹ️"
		}

		fmt.Fprintf(output, "%s %s: %s\n", icon, regression.Type, regression.Location)
		if regression.File != "" {
			fmt.Fprintf(output, "   File: %s", regression.File)
			if regression.Line > 0 {
				fmt.Fprintf(output, ":%d", regression.Line)
			}
			fmt.Fprintln(output, "")
		}
		fmt.Fprintf(output, "   Change: %v → %v", regression.OldValue, regression.NewValue)
		if regression.Delta.Percentage > 0 {
			fmt.Fprintf(output, " (%+.1f%%)", regression.Delta.Percentage)
		}
		fmt.Fprintln(output, "")

		if regression.Suggestion != "" {
			fmt.Fprintf(output, "   Suggestion: %s\n", regression.Suggestion)
		}
		fmt.Fprintln(output, "")
	}
}

func (cr *ConsoleReporter) writeDiffImprovements(output io.Writer, improvements []metrics.Improvement) {
	fmt.Fprintln(output, "=== IMPROVEMENTS ===")

	for _, improvement := range improvements {
		fmt.Fprintf(output, "✅ %s: %s\n", improvement.Type, improvement.Location)
		if improvement.File != "" {
			fmt.Fprintf(output, "   File: %s", improvement.File)
			if improvement.Line > 0 {
				fmt.Fprintf(output, ":%d", improvement.Line)
			}
			fmt.Fprintln(output, "")
		}
		fmt.Fprintf(output, "   Change: %v → %v", improvement.OldValue, improvement.NewValue)
		if improvement.Delta.Percentage > 0 {
			fmt.Fprintf(output, " (%.1f%% improvement)", improvement.Delta.Percentage)
		}
		fmt.Fprintln(output, "")

		if improvement.Benefit != "" {
			fmt.Fprintf(output, "   Benefit: %s\n", improvement.Benefit)
		}
		fmt.Fprintln(output, "")
	}
}

func (cr *ConsoleReporter) writeDiffChanges(output io.Writer, changes []metrics.MetricChange) {
	fmt.Fprintln(output, "=== DETAILED CHANGES ===")

	// Group changes by category
	changesByCategory := make(map[string][]metrics.MetricChange)
	for _, change := range changes {
		changesByCategory[change.Category] = append(changesByCategory[change.Category], change)
	}

	for category, categoryChanges := range changesByCategory {
		fmt.Fprintf(output, "## %s\n", category)

		for _, change := range categoryChanges {
			fmt.Fprintf(output, "- %s: %v → %v", change.Name, change.OldValue, change.NewValue)
			if change.Delta.Percentage > 0 {
				direction := "+"
				if change.Delta.Direction == metrics.ChangeDirectionDecrease {
					direction = "-"
				}
				fmt.Fprintf(output, " (%s%.1f%%)", direction, change.Delta.Percentage)
			}
			fmt.Fprintln(output, "")
		}
		fmt.Fprintln(output, "")
	}
}

// writePackageAnalysis generates comprehensive package analysis output
func (cr *ConsoleReporter) writePackageAnalysis(output io.Writer, report *metrics.Report) {
	fmt.Fprintln(output, "=== PACKAGE ANALYSIS ===")

	packages := report.Packages
	if len(packages) == 0 {
		fmt.Fprintln(output, "No packages found.")
		fmt.Fprintln(output)
		return
	}

	// Sort packages by name for consistent output
	cr.sortPackagesByName(packages)

	// Write summary statistics
	cr.writePackageSummaryStats(output, packages)

	// Write quality issue analysis
	cr.writePackageQualityIssues(output, packages)

	// Write largest packages ranking
	cr.writeLargestPackages(output, packages)

	// Write detailed dependencies (if verbose)
	cr.writePackageDependencies(output, packages)
}

// sortPackagesByName sorts packages alphabetically by name
func (cr *ConsoleReporter) sortPackagesByName(packages []metrics.PackageMetrics) {
	sort.Slice(packages, func(i, j int) bool {
		return packages[i].Name < packages[j].Name
	})
}

// writePackageSummaryStats calculates and writes package summary statistics
func (cr *ConsoleReporter) writePackageSummaryStats(output io.Writer, packages []metrics.PackageMetrics) {
	totalDeps, totalFiles := cr.calculatePackageTotals(packages)
	avgDepsPerPkg := float64(totalDeps) / float64(len(packages))
	avgFilesPerPkg := float64(totalFiles) / float64(len(packages))

	fmt.Fprintf(output, "Total Packages: %d\n", len(packages))
	fmt.Fprintf(output, "Average Dependencies per Package: %.1f\n", avgDepsPerPkg)
	fmt.Fprintf(output, "Average Files per Package: %.1f\n", avgFilesPerPkg)
	fmt.Fprintln(output)
}

// calculatePackageTotals computes total dependencies and files across all packages
func (cr *ConsoleReporter) calculatePackageTotals(packages []metrics.PackageMetrics) (int, int) {
	totalDeps := 0
	totalFiles := 0
	for _, pkg := range packages {
		totalDeps += len(pkg.Dependencies)
		totalFiles += len(pkg.Files)
	}
	return totalDeps, totalFiles
}

// writePackageQualityIssues identifies and reports high coupling and low cohesion packages
func (cr *ConsoleReporter) writePackageQualityIssues(output io.Writer, packages []metrics.PackageMetrics) {
	cr.writeHighCouplingPackages(output, packages)
	cr.writeLowCohesionPackages(output, packages)
}

// writeHighCouplingPackages reports packages with excessive dependencies (>3)
func (cr *ConsoleReporter) writeHighCouplingPackages(output io.Writer, packages []metrics.PackageMetrics) {
	var highCouplingPkgs []metrics.PackageMetrics
	for _, pkg := range packages {
		if len(pkg.Dependencies) > 3 {
			highCouplingPkgs = append(highCouplingPkgs, pkg)
		}
	}

	if len(highCouplingPkgs) > 0 {
		fmt.Fprintln(output, "High Coupling Packages (>3 dependencies):")
		for _, pkg := range highCouplingPkgs {
			fmt.Fprintf(output, "  %s: %d dependencies (coupling: %.1f)\n",
				pkg.Name, len(pkg.Dependencies), pkg.CouplingScore)
		}
		fmt.Fprintln(output)
	}
}

// writeLowCohesionPackages reports packages with poor internal cohesion (<2.0)
func (cr *ConsoleReporter) writeLowCohesionPackages(output io.Writer, packages []metrics.PackageMetrics) {
	var lowCohesionPkgs []metrics.PackageMetrics
	for _, pkg := range packages {
		if pkg.CohesionScore < 2.0 {
			lowCohesionPkgs = append(lowCohesionPkgs, pkg)
		}
	}

	if len(lowCohesionPkgs) > 0 {
		fmt.Fprintln(output, "Low Cohesion Packages (<2.0 cohesion score):")
		for _, pkg := range lowCohesionPkgs {
			fmt.Fprintf(output, "  %s: %.1f cohesion, %d files, %d functions\n",
				pkg.Name, pkg.CohesionScore, len(pkg.Files), pkg.Functions)
		}
		fmt.Fprintln(output)
	}
}

// writeLargestPackages reports the largest packages ranked by function count
func (cr *ConsoleReporter) writeLargestPackages(output io.Writer, packages []metrics.PackageMetrics) {
	// Sort by function count descending
	sort.Slice(packages, func(i, j int) bool {
		return packages[i].Functions > packages[j].Functions
	})

	limit := len(packages)
	if limit > 10 {
		limit = 10
	}

	fmt.Fprintln(output, "Largest Packages (by function count):")
	for i := 0; i < limit; i++ {
		pkg := packages[i]
		fmt.Fprintf(output, "  %s: %d functions, %d structs, %d interfaces, %d files\n",
			pkg.Name, pkg.Functions, pkg.Structs, pkg.Interfaces, len(pkg.Files))
	}
	fmt.Fprintln(output)
}

// writePackageDependencies writes detailed dependency information in verbose mode
func (cr *ConsoleReporter) writePackageDependencies(output io.Writer, packages []metrics.PackageMetrics) {
	if !cr.config.Verbose || len(packages) > 5 {
		return
	}

	fmt.Fprintln(output, "Package Dependencies:")
	for _, pkg := range packages {
		fmt.Fprintf(output, "  %s:\n", pkg.Name)
		if len(pkg.Dependencies) == 0 {
			fmt.Fprintln(output, "    (no internal dependencies)")
		} else {
			for _, dep := range pkg.Dependencies {
				fmt.Fprintf(output, "    → %s\n", dep)
			}
		}
	}
	fmt.Fprintln(output)
}
