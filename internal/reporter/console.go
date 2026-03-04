package reporter

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/opd-ai/go-stats-generator/internal/config"
	"github.com/opd-ai/go-stats-generator/internal/metrics"
)

// ConsoleReporter generates human-readable console output
type ConsoleReporter struct {
	config    *config.OutputConfig
	useColors bool
}

// sectionContent holds information for printing a standardized analysis section.
type sectionContent struct {
	header        string
	summaryLines  []string
	detailWriters []func()
}

// writeSectionWithDetails prints a section header, summary lines, and optional detail subsections.
func (cr *ConsoleReporter) writeSectionWithDetails(output io.Writer, content sectionContent) {
	fmt.Fprintln(output, content.header)
	for _, line := range content.summaryLines {
		fmt.Fprintln(output, line)
	}
	fmt.Fprintln(output)
	for _, writer := range content.detailWriters {
		writer()
	}
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
	cr.writeHeader(output, report)
	cr.writeReportSections(report, output)
	cr.writeFooter(output, report)
	return nil
}

type sectionWriter struct {
	shouldWrite func(*metrics.Report) bool
	write       func(io.Writer, *metrics.Report)
}

func (cr *ConsoleReporter) writeReportSections(report *metrics.Report, output io.Writer) {
	sections := []sectionWriter{
		{cr.shouldWriteOverview, cr.writeOverview},
		{cr.shouldWriteFunctionAnalysis, cr.writeFunctionAnalysis},
		{cr.shouldWriteComplexityAnalysis, cr.writeComplexityAnalysis},
		{cr.shouldWritePackageAnalysis, cr.writePackageAnalysis},
		{cr.shouldWriteCircularDependencies, cr.writeCircularDependencies},
		{cr.shouldWriteDuplicationAnalysis, cr.writeDuplicationAnalysis},
		{cr.shouldWriteNamingAnalysis, cr.writeNamingAnalysis},
		{cr.shouldWritePlacementAnalysis, cr.writePlacementAnalysis},
		{cr.shouldWriteDocumentationAnalysis, cr.writeDocumentationAnalysis},
		{cr.shouldWriteBurdenAnalysis, cr.writeBurdenAnalysis},
		{cr.shouldWriteOrganizationAnalysis, cr.writeOrganizationAnalysis},
		{cr.shouldWriteRefactoringSuggestions, cr.writeRefactoringSuggestions},
	}

	for _, section := range sections {
		if section.shouldWrite(report) {
			section.write(output, report)
		}
	}
}

func (cr *ConsoleReporter) shouldWriteOverview(report *metrics.Report) bool {
	return cr.config.IncludeOverview
}

func (cr *ConsoleReporter) shouldWriteFunctionAnalysis(report *metrics.Report) bool {
	return cr.config.IncludeDetails && len(report.Functions) > 0
}

func (cr *ConsoleReporter) shouldWriteComplexityAnalysis(report *metrics.Report) bool {
	return cr.config.IncludeDetails
}

func (cr *ConsoleReporter) shouldWritePackageAnalysis(report *metrics.Report) bool {
	return cr.config.IncludeDetails && len(report.Packages) > 0
}

func (cr *ConsoleReporter) shouldWriteCircularDependencies(report *metrics.Report) bool {
	return cr.config.IncludeDetails && len(report.Packages) > 0
}

func (cr *ConsoleReporter) shouldWriteDuplicationAnalysis(report *metrics.Report) bool {
	return cr.config.IncludeDetails && report.Duplication.ClonePairs > 0
}

func (cr *ConsoleReporter) shouldWriteNamingAnalysis(report *metrics.Report) bool {
	totalNamingViolations := report.Naming.FileNameViolations + report.Naming.IdentifierViolations + report.Naming.PackageNameViolations
	return cr.config.IncludeDetails && totalNamingViolations > 0
}

func (cr *ConsoleReporter) shouldWritePlacementAnalysis(report *metrics.Report) bool {
	totalPlacementIssues := report.Placement.MisplacedFunctions + report.Placement.MisplacedMethods + report.Placement.LowCohesionFiles
	return cr.config.IncludeDetails && totalPlacementIssues > 0
}

func (cr *ConsoleReporter) shouldWriteDocumentationAnalysis(report *metrics.Report) bool {
	totalAnnotations := len(report.Documentation.TODOComments) + len(report.Documentation.FIXMEComments) + len(report.Documentation.HACKComments) + len(report.Documentation.BUGComments)
	return cr.config.IncludeDetails && (report.Documentation.Coverage.Overall > 0 || totalAnnotations > 0)
}

func (cr *ConsoleReporter) shouldWriteBurdenAnalysis(report *metrics.Report) bool {
	totalBurdenIssues := len(report.Burden.MagicNumbers) + len(report.Burden.DeadCode.UnreferencedFunctions) + len(report.Burden.DeadCode.UnreachableCode) + len(report.Burden.ComplexSignatures) + len(report.Burden.DeeplyNestedFunctions) + len(report.Burden.FeatureEnvyMethods)
	return cr.config.IncludeDetails && totalBurdenIssues > 0
}

func (cr *ConsoleReporter) shouldWriteOrganizationAnalysis(report *metrics.Report) bool {
	totalOrgIssues := len(report.Organization.OversizedFiles) + len(report.Organization.OversizedPackages) + len(report.Organization.DeepDirectories) + len(report.Organization.HighFanInPackages) + len(report.Organization.HighFanOutPackages)
	return cr.config.IncludeDetails && totalOrgIssues > 0
}

func (cr *ConsoleReporter) shouldWriteRefactoringSuggestions(report *metrics.Report) bool {
	return cr.config.IncludeDetails && len(report.Suggestions) > 0
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

// writeHeader outputs the report header section with metadata.
func (cr *ConsoleReporter) writeHeader(output io.Writer, report *metrics.Report) {
	fmt.Fprintln(output, "=== GO SOURCE CODE STATISTICS REPORT ===")
	fmt.Fprintf(output, "Repository: %s\n", report.Metadata.Repository)
	fmt.Fprintf(output, "Generated: %s\n", report.Metadata.GeneratedAt.Format("2006-01-02 15:04:05"))
	fmt.Fprintf(output, "Analysis Time: %v\n", report.Metadata.AnalysisTime.Round(time.Millisecond))
	fmt.Fprintf(output, "Files Processed: %d\n", report.Metadata.FilesProcessed)
	fmt.Fprintln(output)
}

// writeOverview outputs the overview statistics section.
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

// writeFunctionAnalysis outputs the function analysis section with statistics.
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

// writeComplexityAnalysis outputs the complexity analysis section with rankings.
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

// writeTopComplexFunctions outputs the most complex functions in a ranked table.
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

// writeRefactoringSuggestions outputs the prioritized refactoring suggestions.
func (cr *ConsoleReporter) writeRefactoringSuggestions(output io.Writer, report *metrics.Report) {
	fmt.Fprintln(output, "=== REFACTORING SUGGESTIONS ===")

	suggestions := report.Suggestions
	if len(suggestions) == 0 {
		fmt.Fprintln(output, "No refactoring suggestions generated.")
		fmt.Fprintln(output)
		return
	}

	fmt.Fprintf(output, "Total Suggestions: %d (sorted by impact/effort ratio)\n", len(suggestions))
	fmt.Fprintln(output)

	// Display top 20 suggestions (or all if fewer)
	limit := 20
	if len(suggestions) < limit {
		limit = len(suggestions)
	}

	for i := 0; i < limit; i++ {
		cr.writeSingleSuggestion(output, i+1, &suggestions[i])
	}

	if len(suggestions) > limit {
		fmt.Fprintf(output, "... and %d more suggestions (use JSON output for full list)\n", len(suggestions)-limit)
		fmt.Fprintln(output)
	}
}

// writeSingleSuggestion outputs a single refactoring suggestion with details.
func (cr *ConsoleReporter) writeSingleSuggestion(output io.Writer, index int, s *metrics.SuggestionInfo) {
	fmt.Fprintf(output, "%d. [%s] %s\n", index, s.Category, s.Description)
	fmt.Fprintf(output, "   Target: %s\n", s.Target)
	fmt.Fprintf(output, "   Location: %s\n", s.Location)
	fmt.Fprintf(output, "   Effort: %s (%d lines affected)\n", s.Effort, s.AffectedLines)
	fmt.Fprintf(output, "   MBI Impact: %.1f points\n", s.MBIImpact)
	fmt.Fprintf(output, "   ROI Score: %.2f (higher = better)\n", s.ImpactEffort)
	fmt.Fprintln(output)
}

// writeFooter outputs the report footer with tool version information.
func (cr *ConsoleReporter) writeFooter(output io.Writer, report *metrics.Report) {
	fmt.Fprintln(output, "=== ANALYSIS COMPLETE ===")
	fmt.Fprintf(output, "Report generated by go-stats-generator v%s\n", report.Metadata.ToolVersion)
}

// truncate shortens a string to the maximum length with ellipsis.
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

// calculateFunctionStats aggregates statistics across all functions, computing
// average length, average complexity, counts of long/very long/high complexity
// functions, and identifying the longest function by line count.
func (cr *ConsoleReporter) calculateFunctionStats(functions []metrics.FunctionMetrics) functionStats {
	if len(functions) == 0 {
		return functionStats{}
	}

	stats := functionStats{
		LongestName:   functions[0].Name,
		LongestLength: functions[0].Lines.Total,
	}

	var totalLength int
	var totalComplexity float64

	for _, fn := range functions {
		totalLength += fn.Lines.Total
		totalComplexity += fn.Complexity.Overall

		updateLongestFunction(&stats, fn)
		incrementLengthCounters(&stats, fn.Lines.Total)
		incrementComplexityCounters(&stats, fn.Complexity.Cyclomatic)
	}

	count := float64(len(functions))
	stats.AvgLength = float64(totalLength) / count
	stats.LongFunctionsPct = float64(stats.LongFunctions) / count * 100
	stats.VeryLongFunctionsPct = float64(stats.VeryLongFunctions) / count * 100
	stats.AvgComplexity = totalComplexity / count

	return stats
}

func updateLongestFunction(stats *functionStats, fn metrics.FunctionMetrics) {
	if fn.Lines.Total > stats.LongestLength {
		stats.LongestLength = fn.Lines.Total
		stats.LongestName = fn.Name
	}
}

func incrementLengthCounters(stats *functionStats, lines int) {
	if lines > 50 {
		stats.LongFunctions++
	}
	if lines > 100 {
		stats.VeryLongFunctions++
	}
}

func incrementComplexityCounters(stats *functionStats, complexity int) {
	if complexity > 10 {
		stats.HighComplexity++
	}
}

// writeDiffHeader outputs the diff report header with baseline and current snapshot info.
func (cr *ConsoleReporter) writeDiffHeader(output io.Writer, diff *metrics.ComplexityDiff) {
	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "Complexity Diff Report")
	fmt.Fprintln(output, "======================")
	fmt.Fprintf(output, "Baseline: %s (%s)\n", diff.Baseline.ID, diff.Baseline.Metadata.Timestamp.Format(time.RFC3339))
	fmt.Fprintf(output, "Current:  %s (%s)\n", diff.Current.ID, diff.Current.Metadata.Timestamp.Format(time.RFC3339))
	fmt.Fprintln(output, "")
}

// writeDiffSummary outputs the diff summary section with counts and scores.
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

// writeDiffRegressions outputs regression findings to the console with severity
// icons (🚨 critical, ❌ error, ⚠️ warning), displaying type, location, file,
// function, old/new values, and change percentage for each regression.
func (cr *ConsoleReporter) writeDiffRegressions(output io.Writer, regressions []metrics.Regression) {
	fmt.Fprintln(output, "=== REGRESSIONS ===")
	for _, regression := range regressions {
		cr.writeRegressionEntry(output, regression)
	}
}

func (cr *ConsoleReporter) writeRegressionEntry(output io.Writer, regression metrics.Regression) {
	icon := cr.getSeverityIcon(regression.Severity)
	fmt.Fprintf(output, "%s %s: %s\n", icon, regression.Type, regression.Location)
	cr.writeRegressionFile(output, regression)
	cr.writeRegressionChange(output, regression)
	cr.writeRegressionSuggestion(output, regression)
	fmt.Fprintln(output, "")
}

func (cr *ConsoleReporter) getSeverityIcon(severity metrics.SeverityLevel) string {
	switch severity {
	case metrics.SeverityLevelCritical:
		return "🚨"
	case metrics.SeverityLevelError:
		return "❌"
	case metrics.SeverityLevelWarning:
		return "⚠️"
	default:
		return "ℹ️"
	}
}

func (cr *ConsoleReporter) writeRegressionFile(output io.Writer, regression metrics.Regression) {
	if regression.File == "" {
		return
	}
	fmt.Fprintf(output, "   File: %s", regression.File)
	if regression.Line > 0 {
		fmt.Fprintf(output, ":%d", regression.Line)
	}
	fmt.Fprintln(output, "")
}

func (cr *ConsoleReporter) writeRegressionChange(output io.Writer, regression metrics.Regression) {
	fmt.Fprintf(output, "   Change: %v → %v", regression.OldValue, regression.NewValue)
	if regression.Delta.Percentage > 0 {
		fmt.Fprintf(output, " (%+.1f%%)", regression.Delta.Percentage)
	}
	fmt.Fprintln(output, "")
}

func (cr *ConsoleReporter) writeRegressionSuggestion(output io.Writer, regression metrics.Regression) {
	if regression.Suggestion != "" {
		fmt.Fprintf(output, "   Suggestion: %s\n", regression.Suggestion)
	}
}

// writeDiffImprovements outputs the list of improvements with their details.
func (cr *ConsoleReporter) writeDiffImprovements(output io.Writer, improvements []metrics.Improvement) {
	fmt.Fprintln(output, "=== IMPROVEMENTS ===")

	for _, improvement := range improvements {
		writeImprovementEntry(output, improvement)
	}
}

func writeImprovementEntry(output io.Writer, improvement metrics.Improvement) {
	fmt.Fprintf(output, "✅ %s: %s\n", improvement.Type, improvement.Location)
	writeImprovementFile(output, improvement)
	writeImprovementChange(output, improvement)
	writeImprovementBenefit(output, improvement)
	fmt.Fprintln(output, "")
}

func writeImprovementFile(output io.Writer, improvement metrics.Improvement) {
	if improvement.File == "" {
		return
	}
	fmt.Fprintf(output, "   File: %s", improvement.File)
	if improvement.Line > 0 {
		fmt.Fprintf(output, ":%d", improvement.Line)
	}
	fmt.Fprintln(output, "")
}

func writeImprovementChange(output io.Writer, improvement metrics.Improvement) {
	fmt.Fprintf(output, "   Change: %v → %v", improvement.OldValue, improvement.NewValue)
	if improvement.Delta.Percentage > 0 {
		fmt.Fprintf(output, " (%.1f%% improvement)", improvement.Delta.Percentage)
	}
	fmt.Fprintln(output, "")
}

func writeImprovementBenefit(output io.Writer, improvement metrics.Improvement) {
	if improvement.Benefit != "" {
		fmt.Fprintf(output, "   Benefit: %s\n", improvement.Benefit)
	}
}

// writeDiffChanges outputs all metric changes to the console, grouped by category
// (functions, structs, packages), displaying name, old value, new value, and
// change percentage/direction for each metric that changed.
func (cr *ConsoleReporter) writeDiffChanges(output io.Writer, changes []metrics.MetricChange) {
	fmt.Fprintln(output, "=== DETAILED CHANGES ===")

	changesByCategory := groupChangesByCategory(changes)

	for category, categoryChanges := range changesByCategory {
		fmt.Fprintf(output, "## %s\n", category)
		writeCategoryChanges(output, categoryChanges)
		fmt.Fprintln(output, "")
	}
}

func groupChangesByCategory(changes []metrics.MetricChange) map[string][]metrics.MetricChange {
	changesByCategory := make(map[string][]metrics.MetricChange)
	for _, change := range changes {
		changesByCategory[change.Category] = append(changesByCategory[change.Category], change)
	}
	return changesByCategory
}

func writeCategoryChanges(output io.Writer, changes []metrics.MetricChange) {
	for _, change := range changes {
		fmt.Fprintf(output, "- %s: %v → %v", change.Name, change.OldValue, change.NewValue)
		if change.Delta.Percentage > 0 {
			writeChangePercentage(output, change.Delta)
		}
		fmt.Fprintln(output, "")
	}
}

func writeChangePercentage(output io.Writer, delta metrics.Delta) {
	direction := "+"
	if delta.Direction == metrics.ChangeDirectionDecrease {
		direction = "-"
	}
	fmt.Fprintf(output, " (%s%.1f%%)", direction, delta.Percentage)
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

// writeCircularDependencies displays circular dependency detection results
func (cr *ConsoleReporter) writeCircularDependencies(output io.Writer, report *metrics.Report) {
	fmt.Fprintln(output, "=== CIRCULAR DEPENDENCIES ===")
	if cr.writeCircularDepsEmpty(output, report) {
		return
	}
	fmt.Fprintf(output, "Found %d circular dependency chain(s):\n\n", len(report.CircularDependencies))
	cr.writeCircularDepsList(output, report.CircularDependencies)
}

func (cr *ConsoleReporter) writeCircularDepsEmpty(output io.Writer, report *metrics.Report) bool {
	if len(report.CircularDependencies) == 0 {
		fmt.Fprintln(output, "No circular dependencies detected.")
		fmt.Fprintln(output)
		return true
	}
	return false
}

func (cr *ConsoleReporter) writeCircularDepsList(output io.Writer, cycles []metrics.CircularDependency) {
	for i, cycle := range cycles {
		cr.writeCircularDepsEntry(output, i+1, cycle)
	}
}

func (cr *ConsoleReporter) writeCircularDepsEntry(output io.Writer, index int, cycle metrics.CircularDependency) {
	severity := cycle.Severity
	if severity == "" {
		severity = "unknown"
	}
	fmt.Fprintf(output, "%d. [%s SEVERITY] ", index, toUpperCase(severity))
	cr.writeCircularDepsChain(output, cycle.Packages)
	fmt.Fprintln(output)
}

func (cr *ConsoleReporter) writeCircularDepsChain(output io.Writer, packages []string) {
	for j, pkg := range packages {
		if j > 0 {
			fmt.Fprint(output, " → ")
		}
		fmt.Fprint(output, pkg)
	}
	if len(packages) > 0 {
		fmt.Fprintf(output, " → %s\n", packages[0])
	}
}

// toUpperCase converts a string to uppercase
func toUpperCase(s string) string {
	return strings.ToUpper(s)
}

// writeDuplicationAnalysis generates duplication analysis output
func (cr *ConsoleReporter) writeDuplicationAnalysis(output io.Writer, report *metrics.Report) {
	fmt.Fprintln(output, "=== DUPLICATION ANALYSIS ===")
	cr.writeDuplicationSummary(output, report.Duplication)
	if len(report.Duplication.Clones) == 0 {
		return
	}
	cr.writeDuplicationTable(output, report.Duplication.Clones)
}

func (cr *ConsoleReporter) writeDuplicationSummary(output io.Writer, dup metrics.DuplicationMetrics) {
	fmt.Fprintf(output, "Clone Pairs Detected: %d\n", dup.ClonePairs)
	fmt.Fprintf(output, "Duplicated Lines: %d\n", dup.DuplicatedLines)
	fmt.Fprintf(output, "Duplication Ratio: %.2f%%\n", dup.DuplicationRatio*100)
	fmt.Fprintf(output, "Largest Clone Size: %d lines\n", dup.LargestCloneSize)
	fmt.Fprintln(output)
}

func (cr *ConsoleReporter) writeDuplicationTable(output io.Writer, clones []metrics.ClonePair) {
	sortedClones := cr.getSortedClones(clones)
	limit := cr.calculateCloneLimit(len(sortedClones))
	cr.writeDuplicationHeader(output, limit)
	cr.writeDuplicationRows(output, sortedClones, limit)
}

func (cr *ConsoleReporter) getSortedClones(clones []metrics.ClonePair) []metrics.ClonePair {
	sorted := make([]metrics.ClonePair, len(clones))
	copy(sorted, clones)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].LineCount > sorted[j].LineCount
	})
	return sorted
}

func (cr *ConsoleReporter) calculateCloneLimit(count int) int {
	limit := cr.config.Limit
	if limit > count {
		limit = count
	}
	if limit > 10 {
		limit = 10
	}
	return limit
}

func (cr *ConsoleReporter) writeDuplicationHeader(output io.Writer, limit int) {
	fmt.Fprintf(output, "Top %d Clone Pairs (by size):\n", limit)
	fmt.Fprintf(output, "%-15s %8s %8s %s\n", "Type", "Lines", "Instances", "Locations")
	fmt.Fprintln(output, "--------------------------------------------------------------------------------")
}

func (cr *ConsoleReporter) writeDuplicationRows(output io.Writer, clones []metrics.ClonePair, limit int) {
	for i := 0; i < limit; i++ {
		cr.writeDuplicationRow(output, clones[i])
	}
	fmt.Fprintln(output)
}

func (cr *ConsoleReporter) writeDuplicationRow(output io.Writer, clone metrics.ClonePair) {
	locations := cr.formatCloneLocations(clone)
	fmt.Fprintf(output, "%-15s %8d %8d %s\n",
		string(clone.Type),
		clone.LineCount,
		len(clone.Instances),
		locations,
	)
}

func (cr *ConsoleReporter) formatCloneLocations(clone metrics.ClonePair) string {
	if len(clone.Instances) == 0 {
		return ""
	}
	inst := clone.Instances[0]
	location := fmt.Sprintf("%s:%d-%d", cr.truncate(inst.File, 40), inst.StartLine, inst.EndLine)
	if len(clone.Instances) > 1 {
		location += fmt.Sprintf(" (+%d more)", len(clone.Instances)-1)
	}
	return location
}

// writeNamingAnalysis generates naming convention analysis output
func (cr *ConsoleReporter) writeNamingAnalysis(output io.Writer, report *metrics.Report) {
	naming := report.Naming
	content := sectionContent{
		header: "=== NAMING CONVENTION ANALYSIS ===",
		summaryLines: []string{
			fmt.Sprintf("File Name Violations: %d", naming.FileNameViolations),
			fmt.Sprintf("Identifier Violations: %d", naming.IdentifierViolations),
			fmt.Sprintf("Package Name Violations: %d", naming.PackageNameViolations),
			fmt.Sprintf("Overall Naming Score: %.2f", naming.OverallNamingScore),
		},
	}
	if len(naming.IdentifierIssues) > 0 {
		content.detailWriters = append(content.detailWriters, func() {
			cr.writeIdentifierViolations(output, naming.IdentifierIssues)
		})
	}
	if len(naming.PackageNameIssues) > 0 {
		content.detailWriters = append(content.detailWriters, func() {
			cr.writePackageNameViolations(output, naming.PackageNameIssues)
		})
	}
	if len(naming.FileNameIssues) > 0 {
		content.detailWriters = append(content.detailWriters, func() {
			cr.writeFileNameViolations(output, naming.FileNameIssues)
		})
	}
	cr.writeSectionWithDetails(output, content)
}

// writeIdentifierViolations displays identifier naming violations
func (cr *ConsoleReporter) writeIdentifierViolations(output io.Writer, violations []metrics.IdentifierViolation) {
	// Sort by severity (high > medium > low) then by file
	sorted := make([]metrics.IdentifierViolation, len(violations))
	copy(sorted, violations)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Severity != sorted[j].Severity {
			return severityWeight(sorted[i].Severity) > severityWeight(sorted[j].Severity)
		}
		return sorted[i].File < sorted[j].File
	})

	limit := cr.config.Limit
	if limit > len(sorted) {
		limit = len(sorted)
	}
	if limit > 10 {
		limit = 10
	}

	fmt.Fprintf(output, "Top %d Identifier Violations:\n", limit)
	fmt.Fprintf(output, "%-25s %-10s %-12s %-40s\n", "Name", "Type", "Violation", "File:Line")
	fmt.Fprintln(output, "--------------------------------------------------------------------------------")

	for i := 0; i < limit; i++ {
		v := sorted[i]
		location := fmt.Sprintf("%s:%d", cr.truncate(v.File, 30), v.Line)
		fmt.Fprintf(output, "%-25s %-10s %-12s %-40s\n",
			cr.truncate(v.Name, 25),
			v.Type,
			cr.truncate(v.ViolationType, 12),
			location,
		)
	}
	fmt.Fprintln(output)
}

// writePackageNameViolations displays package naming violations
func (cr *ConsoleReporter) writePackageNameViolations(output io.Writer, violations []metrics.PackageNameViolation) {
	// Sort by severity
	sorted := make([]metrics.PackageNameViolation, len(violations))
	copy(sorted, violations)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Severity != sorted[j].Severity {
			return severityWeight(sorted[i].Severity) > severityWeight(sorted[j].Severity)
		}
		return sorted[i].Package < sorted[j].Package
	})

	fmt.Fprintln(output, "Package Name Violations:")
	fmt.Fprintf(output, "%-20s %-20s %-40s\n", "Package", "Violation", "Description")
	fmt.Fprintln(output, "--------------------------------------------------------------------------------")

	for _, v := range sorted {
		fmt.Fprintf(output, "%-20s %-20s %-40s\n",
			cr.truncate(v.Package, 20),
			cr.truncate(v.ViolationType, 20),
			cr.truncate(v.Description, 40),
		)
	}
	fmt.Fprintln(output)
}

// writeFileNameViolations displays file naming violations
func (cr *ConsoleReporter) writeFileNameViolations(output io.Writer, violations []metrics.FileNameViolation) {
	// Sort by severity
	sorted := make([]metrics.FileNameViolation, len(violations))
	copy(sorted, violations)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Severity != sorted[j].Severity {
			return severityWeight(sorted[i].Severity) > severityWeight(sorted[j].Severity)
		}
		return sorted[i].File < sorted[j].File
	})

	limit := cr.config.Limit
	if limit > len(sorted) {
		limit = len(sorted)
	}
	if limit > 10 {
		limit = 10
	}

	fmt.Fprintf(output, "Top %d File Name Violations:\n", limit)
	fmt.Fprintf(output, "%-40s %-20s %-30s\n", "File", "Violation", "Suggested Name")
	fmt.Fprintln(output, "--------------------------------------------------------------------------------")

	for i := 0; i < limit; i++ {
		v := sorted[i]
		fmt.Fprintf(output, "%-40s %-20s %-30s\n",
			cr.truncate(v.File, 40),
			cr.truncate(v.ViolationType, 20),
			cr.truncate(v.SuggestedName, 30),
		)
	}
	fmt.Fprintln(output)
}

// severityWeight returns numeric weight for severity sorting
func severityWeight(severity string) int {
	switch severity {
	case "high":
		return 3
	case "medium":
		return 2
	case "low":
		return 1
	default:
		return 0
	}
}

// writePlacementAnalysis generates placement analysis output
func (cr *ConsoleReporter) writePlacementAnalysis(output io.Writer, report *metrics.Report) {
	placement := report.Placement
	content := sectionContent{
		header: "=== PLACEMENT ANALYSIS ===",
		summaryLines: []string{
			fmt.Sprintf("Misplaced Functions: %d", placement.MisplacedFunctions),
			fmt.Sprintf("Misplaced Methods: %d", placement.MisplacedMethods),
			fmt.Sprintf("Low Cohesion Files: %d", placement.LowCohesionFiles),
			fmt.Sprintf("Average File Cohesion: %.2f", placement.AvgFileCohesion),
		},
	}
	if len(placement.FunctionIssues) > 0 {
		content.detailWriters = append(content.detailWriters, func() {
			cr.writeMisplacedFunctions(output, placement.FunctionIssues)
		})
	}
	if len(placement.MethodIssues) > 0 {
		content.detailWriters = append(content.detailWriters, func() {
			cr.writeMisplacedMethods(output, placement.MethodIssues)
		})
	}
	if len(placement.CohesionIssues) > 0 {
		content.detailWriters = append(content.detailWriters, func() {
			cr.writeFileCohesionIssues(output, placement.CohesionIssues)
		})
	}
	cr.writeSectionWithDetails(output, content)
}

// writeMisplacedFunctions displays misplaced function issues
func (cr *ConsoleReporter) writeMisplacedFunctions(output io.Writer, issues []metrics.MisplacedFunctionIssue) {
	// Sort by severity (high > medium > low) then by suggested affinity (descending)
	sorted := make([]metrics.MisplacedFunctionIssue, len(issues))
	copy(sorted, issues)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Severity != sorted[j].Severity {
			return severityWeight(sorted[i].Severity) > severityWeight(sorted[j].Severity)
		}
		return sorted[i].SuggestedAffinity > sorted[j].SuggestedAffinity
	})

	limit := cr.config.Limit
	if limit > len(sorted) {
		limit = len(sorted)
	}
	if limit > 10 {
		limit = 10
	}

	fmt.Fprintf(output, "Top %d Misplaced Functions:\n", limit)
	fmt.Fprintf(output, "%-30s %-25s %-25s %s\n", "Function", "Current File", "Suggested File", "Affinity Gain")
	fmt.Fprintln(output, "--------------------------------------------------------------------------------")

	for i := 0; i < limit; i++ {
		issue := sorted[i]
		affinityGain := issue.SuggestedAffinity - issue.CurrentAffinity
		fmt.Fprintf(output, "%-30s %-25s %-25s +%.2f\n",
			cr.truncate(issue.Name, 30),
			cr.truncate(issue.CurrentFile, 25),
			cr.truncate(issue.SuggestedFile, 25),
			affinityGain,
		)
	}
	fmt.Fprintln(output)
}

// writeMisplacedMethods displays misplaced method issues
func (cr *ConsoleReporter) writeMisplacedMethods(output io.Writer, issues []metrics.MisplacedMethodIssue) {
	// Sort by severity (high > medium > low) then by distance
	sorted := make([]metrics.MisplacedMethodIssue, len(issues))
	copy(sorted, issues)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Severity != sorted[j].Severity {
			return severityWeight(sorted[i].Severity) > severityWeight(sorted[j].Severity)
		}
		// "different_package" before "same_package"
		if sorted[i].Distance != sorted[j].Distance {
			return sorted[i].Distance > sorted[j].Distance
		}
		return sorted[i].MethodName < sorted[j].MethodName
	})

	limit := cr.config.Limit
	if limit > len(sorted) {
		limit = len(sorted)
	}
	if limit > 10 {
		limit = 10
	}

	fmt.Fprintf(output, "Top %d Misplaced Methods:\n", limit)
	fmt.Fprintf(output, "%-30s %-20s %-25s %-25s\n", "Method", "Receiver Type", "Current File", "Receiver File")
	fmt.Fprintln(output, "--------------------------------------------------------------------------------")

	for i := 0; i < limit; i++ {
		issue := sorted[i]
		fmt.Fprintf(output, "%-30s %-20s %-25s %-25s\n",
			cr.truncate(issue.MethodName, 30),
			cr.truncate(issue.ReceiverType, 20),
			cr.truncate(issue.CurrentFile, 25),
			cr.truncate(issue.ReceiverFile, 25),
		)
	}
	fmt.Fprintln(output)
}

// writeFileCohesionIssues displays file cohesion issues
func (cr *ConsoleReporter) writeFileCohesionIssues(output io.Writer, issues []metrics.FileCohesionIssue) {
	sorted := cr.sortCohesionIssues(issues)
	limit := cr.calculateCohesionLimit(len(sorted))
	cr.writeCohesionHeader(output, limit)
	cr.writeCohesionRows(output, sorted, limit)
}

func (cr *ConsoleReporter) sortCohesionIssues(issues []metrics.FileCohesionIssue) []metrics.FileCohesionIssue {
	sorted := make([]metrics.FileCohesionIssue, len(issues))
	copy(sorted, issues)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Severity != sorted[j].Severity {
			return severityWeight(sorted[i].Severity) > severityWeight(sorted[j].Severity)
		}
		return sorted[i].CohesionScore < sorted[j].CohesionScore
	})
	return sorted
}

func (cr *ConsoleReporter) calculateCohesionLimit(count int) int {
	limit := cr.config.Limit
	if limit > count {
		limit = count
	}
	if limit > 10 {
		limit = 10
	}
	return limit
}

func (cr *ConsoleReporter) writeCohesionHeader(output io.Writer, limit int) {
	fmt.Fprintf(output, "Top %d Low Cohesion Files:\n", limit)
	fmt.Fprintf(output, "%-40s %-12s %s\n", "File", "Cohesion", "Suggested Splits")
	fmt.Fprintln(output, "--------------------------------------------------------------------------------")
}

func (cr *ConsoleReporter) writeCohesionRows(output io.Writer, sorted []metrics.FileCohesionIssue, limit int) {
	for i := 0; i < limit; i++ {
		cr.writeCohesionRow(output, sorted[i])
	}
	fmt.Fprintln(output)
}

func (cr *ConsoleReporter) writeCohesionRow(output io.Writer, issue metrics.FileCohesionIssue) {
	splits := cr.formatSuggestedSplits(issue.SuggestedSplits)
	fmt.Fprintf(output, "%-40s %-12.2f %s\n",
		cr.truncate(issue.File, 40),
		issue.CohesionScore,
		splits,
	)
}

func (cr *ConsoleReporter) formatSuggestedSplits(splits []string) string {
	if len(splits) == 0 {
		return ""
	}
	result := splits[0]
	if len(splits) > 1 {
		result += fmt.Sprintf(" (+%d more)", len(splits)-1)
	}
	return result
}

// writeDocumentationAnalysis generates documentation analysis output
func (cr *ConsoleReporter) writeDocumentationAnalysis(output io.Writer, report *metrics.Report) {
	fmt.Fprintln(output, "=== DOCUMENTATION ANALYSIS ===")

	doc := report.Documentation

	// Coverage summary
	fmt.Fprintf(output, "Overall Coverage: %.1f%%\n", doc.Coverage.Overall)
	fmt.Fprintf(output, "Package Coverage: %.1f%%\n", doc.Coverage.Packages)
	fmt.Fprintf(output, "Function Coverage: %.1f%%\n", doc.Coverage.Functions)
	fmt.Fprintf(output, "Type Coverage: %.1f%%\n", doc.Coverage.Types)
	fmt.Fprintf(output, "Method Coverage: %.1f%%\n", doc.Coverage.Methods)
	fmt.Fprintln(output)

	// Annotation summary
	totalAnnotations := len(doc.TODOComments) + len(doc.FIXMEComments) + len(doc.HACKComments) + len(doc.BUGComments) + len(doc.XXXComments) + len(doc.DEPRECATEDComments) + len(doc.NOTEComments)
	if totalAnnotations > 0 {
		fmt.Fprintln(output, "Annotation Summary:")
		fmt.Fprintf(output, "  TODO: %d\n", len(doc.TODOComments))
		fmt.Fprintf(output, "  FIXME: %d (critical)\n", len(doc.FIXMEComments))
		fmt.Fprintf(output, "  HACK: %d\n", len(doc.HACKComments))
		fmt.Fprintf(output, "  BUG: %d (critical)\n", len(doc.BUGComments))
		fmt.Fprintf(output, "  XXX: %d\n", len(doc.XXXComments))
		fmt.Fprintf(output, "  DEPRECATED: %d\n", len(doc.DEPRECATEDComments))
		fmt.Fprintf(output, "  NOTE: %d\n", len(doc.NOTEComments))
		fmt.Fprintf(output, "  Total: %d\n", totalAnnotations)
		fmt.Fprintln(output)

		// Show top annotations by severity
		cr.writeTopAnnotations(output, doc)
	}

	fmt.Fprintln(output)
}

// annotationItem represents a code annotation with its metadata for console display.
type annotationItem struct {
	category string
	file     string
	line     int
	desc     string
	severity string
}

// collectAnnotations gathers all annotations from documentation metrics
func collectAnnotations(doc metrics.DocumentationMetrics) []annotationItem {
	var annotations []annotationItem

	for _, c := range doc.FIXMEComments {
		annotations = append(annotations, annotationItem{"FIXME", c.File, c.Line, c.Description, "critical"})
	}
	for _, c := range doc.BUGComments {
		annotations = append(annotations, annotationItem{"BUG", c.File, c.Line, c.Description, "critical"})
	}
	for _, c := range doc.HACKComments {
		annotations = append(annotations, annotationItem{"HACK", c.File, c.Line, c.Reason, "high"})
	}
	for _, c := range doc.TODOComments {
		annotations = append(annotations, annotationItem{"TODO", c.File, c.Line, c.Description, "medium"})
	}
	for _, c := range doc.XXXComments {
		annotations = append(annotations, annotationItem{"XXX", c.File, c.Line, c.Description, "medium"})
	}

	return annotations
}

// writeTopAnnotations displays top annotations by severity
func (cr *ConsoleReporter) writeTopAnnotations(output io.Writer, doc metrics.DocumentationMetrics) {
	annotations := collectAnnotations(doc)
	if len(annotations) == 0 {
		return
	}

	limit := 10
	if limit > len(annotations) {
		limit = len(annotations)
	}

	fmt.Fprintf(output, "Top %d Annotations by Severity:\n", limit)
	fmt.Fprintf(output, "%-10s %-50s %-6s %s\n", "Category", "File", "Line", "Description")
	fmt.Fprintln(output, "--------------------------------------------------------------------------------")

	for i := 0; i < limit; i++ {
		a := annotations[i]
		fmt.Fprintf(output, "%-10s %-50s %-6d %s\n",
			a.category,
			cr.truncate(a.file, 50),
			a.line,
			cr.truncate(a.desc, 40),
		)
	}
	fmt.Fprintln(output)
}

// writeBurdenAnalysis generates maintenance burden analysis output
func (cr *ConsoleReporter) writeBurdenAnalysis(output io.Writer, report *metrics.Report) {
	fmt.Fprintln(output, "=== MAINTENANCE BURDEN ===")

	burden := report.Burden

	// Summary metrics
	fmt.Fprintf(output, "Magic Numbers: %d\n", len(burden.MagicNumbers))
	fmt.Fprintf(output, "Dead Code (Unreferenced): %d functions\n", len(burden.DeadCode.UnreferencedFunctions))
	fmt.Fprintf(output, "Dead Code (Unreachable): %d blocks\n", len(burden.DeadCode.UnreachableCode))
	fmt.Fprintf(output, "Dead Code Percentage: %.2f%%\n", burden.DeadCode.DeadCodePercent)
	fmt.Fprintf(output, "Complex Signatures: %d\n", len(burden.ComplexSignatures))
	fmt.Fprintf(output, "Deeply Nested Functions: %d\n", len(burden.DeeplyNestedFunctions))
	fmt.Fprintf(output, "Feature Envy Methods: %d\n", len(burden.FeatureEnvyMethods))
	fmt.Fprintln(output)

	cr.writeTopBurdenIssues(output, burden)

	fmt.Fprintln(output)
}

// writeTopBurdenIssues displays top burden violations
func (cr *ConsoleReporter) writeTopBurdenIssues(output io.Writer, burden metrics.BurdenMetrics) {
	cr.writeTopComplexSignatures(output, burden.ComplexSignatures)
	cr.writeTopDeeplyNestedFunctions(output, burden.DeeplyNestedFunctions)
	cr.writeTopMagicNumbers(output, burden.MagicNumbers)
}

// writeTopComplexSignatures displays functions with complex signatures
func (cr *ConsoleReporter) writeTopComplexSignatures(output io.Writer, signatures []metrics.SignatureIssue) {
	if len(signatures) == 0 {
		return
	}

	limit := cr.getDisplayLimit(len(signatures))
	fmt.Fprintf(output, "Top %d Complex Signatures:\n", limit)
	fmt.Fprintf(output, "%-40s %-20s %8s %8s\n", "Function", "File", "Params", "Returns")
	fmt.Fprintln(output, "--------------------------------------------------------------------------------")

	for i := 0; i < limit; i++ {
		s := signatures[i]
		fmt.Fprintf(output, "%-40s %-20s %8d %8d\n",
			cr.truncate(s.Function, 40),
			cr.truncate(s.File, 20),
			s.ParameterCount,
			s.ReturnCount,
		)
	}
	fmt.Fprintln(output)
}

// writeTopDeeplyNestedFunctions displays functions with deep nesting
func (cr *ConsoleReporter) writeTopDeeplyNestedFunctions(output io.Writer, nesting []metrics.NestingIssue) {
	if len(nesting) == 0 {
		return
	}

	limit := cr.getDisplayLimit(len(nesting))
	fmt.Fprintf(output, "Top %d Deeply Nested Functions:\n", limit)
	fmt.Fprintf(output, "%-40s %-20s %8s\n", "Function", "File", "Max Depth")
	fmt.Fprintln(output, "--------------------------------------------------------------------------------")

	for i := 0; i < limit; i++ {
		n := nesting[i]
		fmt.Fprintf(output, "%-40s %-20s %8d\n",
			cr.truncate(n.Function, 40),
			cr.truncate(n.File, 20),
			n.MaxDepth,
		)
	}
	fmt.Fprintln(output)
}

// writeTopMagicNumbers displays top magic number occurrences
func (cr *ConsoleReporter) writeTopMagicNumbers(output io.Writer, numbers []metrics.MagicNumber) {
	if len(numbers) == 0 {
		return
	}

	limit := cr.getDisplayLimit(len(numbers))
	fmt.Fprintf(output, "Top %d Magic Numbers:\n", limit)
	fmt.Fprintf(output, "%-20s %-30s %8s %s\n", "Value", "File", "Line", "Context")
	fmt.Fprintln(output, "--------------------------------------------------------------------------------")

	for i := 0; i < limit; i++ {
		m := numbers[i]
		fmt.Fprintf(output, "%-20s %-30s %8d %s\n",
			cr.truncate(m.Value, 20),
			cr.truncate(m.File, 30),
			m.Line,
			cr.truncate(m.Context, 30),
		)
	}
	fmt.Fprintln(output)
}

// writeOrganizationAnalysis generates organization health analysis output
func (cr *ConsoleReporter) writeOrganizationAnalysis(output io.Writer, report *metrics.Report) {
	org := report.Organization
	content := sectionContent{
		header: "=== ORGANIZATION HEALTH ===",
		summaryLines: []string{
			fmt.Sprintf("Oversized Files: %d", len(org.OversizedFiles)),
			fmt.Sprintf("Oversized Packages: %d", len(org.OversizedPackages)),
			fmt.Sprintf("Deep Directories: %d", len(org.DeepDirectories)),
			fmt.Sprintf("High Fan-In Packages: %d", len(org.HighFanInPackages)),
			fmt.Sprintf("High Fan-Out Packages: %d", len(org.HighFanOutPackages)),
			fmt.Sprintf("Avg Package Instability: %.2f", org.AvgPackageStability),
		},
		detailWriters: []func(){
			func() { cr.writeOversizedFiles(output, org.OversizedFiles) },
			func() { cr.writeOversizedPackages(output, org.OversizedPackages) },
			func() { cr.writeDeepDirectories(output, org.DeepDirectories) },
			func() { cr.writeHighFanInPackages(output, org.HighFanInPackages) },
			func() { cr.writeHighFanOutPackages(output, org.HighFanOutPackages) },
		},
	}
	cr.writeSectionWithDetails(output, content)
}

// writeOversizedFiles displays files exceeding size thresholds
func (cr *ConsoleReporter) writeOversizedFiles(output io.Writer, files []metrics.OversizedFile) {
	if len(files) == 0 {
		return
	}

	sorted := make([]metrics.OversizedFile, len(files))
	copy(sorted, files)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].MaintenanceBurden > sorted[j].MaintenanceBurden
	})

	limit := cr.config.Limit
	if limit > len(sorted) {
		limit = len(sorted)
	}
	if limit > 10 {
		limit = 10
	}

	fmt.Fprintf(output, "Top %d Oversized Files:\n", limit)
	fmt.Fprintf(output, "%-50s %8s %8s %8s %s\n", "File", "Lines", "Funcs", "Types", "Burden")
	fmt.Fprintln(output, "--------------------------------------------------------------------------------")

	for i := 0; i < limit; i++ {
		f := sorted[i]
		fmt.Fprintf(output, "%-50s %8d %8d %8d %.2f\n",
			cr.truncate(f.File, 50),
			f.Lines.Code,
			f.FunctionCount,
			f.TypeCount,
			f.MaintenanceBurden,
		)
	}
	fmt.Fprintln(output)
}

// writeOversizedPackages displays packages exceeding size thresholds
// writeOversizedPackages displays packages exceeding size thresholds
func (cr *ConsoleReporter) writeOversizedPackages(output io.Writer, pkgs []metrics.OversizedPackage) {
	if len(pkgs) == 0 {
		return
	}

	sorted := make([]metrics.OversizedPackage, len(pkgs))
	copy(sorted, pkgs)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].TotalFunctions > sorted[j].TotalFunctions
	})

	limit := cr.getDisplayLimit(len(sorted))
	fmt.Fprintf(output, "Top %d Oversized Packages:\n", limit)
	fmt.Fprintf(output, "%-30s %8s %8s %8s %s\n", "Package", "Files", "Exports", "Funcs", "Mega?")
	fmt.Fprintln(output, "--------------------------------------------------------------------------------")

	for i := 0; i < limit; i++ {
		p := sorted[i]
		mega := "No"
		if p.IsMegaPackage {
			mega = "Yes"
		}
		fmt.Fprintf(output, "%-30s %8d %8d %8d %s\n",
			cr.truncate(p.Package, 30),
			p.FileCount,
			p.ExportedSymbols,
			p.TotalFunctions,
			mega,
		)
	}
	fmt.Fprintln(output)
}

// getDisplayLimit returns the appropriate display limit for tables
func (cr *ConsoleReporter) getDisplayLimit(itemCount int) int {
	limit := cr.config.Limit
	if limit > itemCount {
		limit = itemCount
	}
	if limit > 10 {
		limit = 10
	}
	return limit
}

// writeDeepDirectories displays directory structures exceeding depth thresholds
func (cr *ConsoleReporter) writeDeepDirectories(output io.Writer, dirs []metrics.DeepDirectory) {
	if len(dirs) == 0 {
		return
	}

	sorted := make([]metrics.DeepDirectory, len(dirs))
	copy(sorted, dirs)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Depth > sorted[j].Depth
	})

	limit := cr.config.Limit
	if limit > len(sorted) {
		limit = len(sorted)
	}
	if limit > 10 {
		limit = 10
	}

	fmt.Fprintf(output, "Top %d Deep Directories:\n", limit)
	fmt.Fprintf(output, "%-60s %8s %8s\n", "Path", "Depth", "Files")
	fmt.Fprintln(output, "--------------------------------------------------------------------------------")

	for i := 0; i < limit; i++ {
		d := sorted[i]
		fmt.Fprintf(output, "%-60s %8d %8d\n",
			cr.truncate(d.Path, 60),
			d.Depth,
			d.FileCount,
		)
	}
	fmt.Fprintln(output)
}

// writeHighFanInPackages displays packages with high incoming dependencies
func (cr *ConsoleReporter) writeHighFanInPackages(output io.Writer, pkgs []metrics.FanInPackage) {
	if len(pkgs) == 0 {
		return
	}

	sorted := make([]metrics.FanInPackage, len(pkgs))
	copy(sorted, pkgs)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].FanIn > sorted[j].FanIn
	})

	limit := cr.config.Limit
	if limit > len(sorted) {
		limit = len(sorted)
	}
	if limit > 10 {
		limit = 10
	}

	fmt.Fprintf(output, "Top %d High Fan-In Packages (Bottlenecks):\n", limit)
	fmt.Fprintf(output, "%-40s %8s %s\n", "Package", "Fan-In", "Risk Level")
	fmt.Fprintln(output, "--------------------------------------------------------------------------------")

	for i := 0; i < limit; i++ {
		p := sorted[i]
		fmt.Fprintf(output, "%-40s %8d %s\n",
			cr.truncate(p.Package, 40),
			p.FanIn,
			p.RiskLevel,
		)
	}
	fmt.Fprintln(output)
}

// writeHighFanOutPackages displays packages with high outgoing dependencies
func (cr *ConsoleReporter) writeHighFanOutPackages(output io.Writer, pkgs []metrics.FanOutPackage) {
	if len(pkgs) == 0 {
		return
	}

	sorted := make([]metrics.FanOutPackage, len(pkgs))
	copy(sorted, pkgs)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].FanOut > sorted[j].FanOut
	})

	limit := cr.config.Limit
	if limit > len(sorted) {
		limit = len(sorted)
	}
	if limit > 10 {
		limit = 10
	}

	fmt.Fprintf(output, "Top %d High Fan-Out Packages (Authority):\n", limit)
	fmt.Fprintf(output, "%-40s %8s %12s %s\n", "Package", "Fan-Out", "Instability", "Risk")
	fmt.Fprintln(output, "--------------------------------------------------------------------------------")

	for i := 0; i < limit; i++ {
		p := sorted[i]
		fmt.Fprintf(output, "%-40s %8d %12.2f %s\n",
			cr.truncate(p.Package, 40),
			p.FanOut,
			p.Instability,
			p.CouplingRisk,
		)
	}
	fmt.Fprintln(output)
}
