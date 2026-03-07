package reporter

import (
	"fmt"
	"io"
	"sort"
	"time"

	"github.com/opd-ai/go-stats-generator/internal/metrics"
)

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
	limit := cr.calculateDisplayLimit(len(sortedFunctions))

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

// updateLongestFunction updates the longest function record if current function exceeds previous maximum.
func updateLongestFunction(stats *functionStats, fn metrics.FunctionMetrics) {
	if fn.Lines.Total > stats.LongestLength {
		stats.LongestLength = fn.Lines.Total
		stats.LongestName = fn.Name
	}
}

// incrementLengthCounters updates long and very-long function counters based on line count thresholds.
func incrementLengthCounters(stats *functionStats, lines int) {
	if lines > 50 {
		stats.LongFunctions++
	}
	if lines > 100 {
		stats.VeryLongFunctions++
	}
}

// incrementComplexityCounters updates high complexity function counter based on cyclomatic complexity threshold.
func incrementComplexityCounters(stats *functionStats, complexity int) {
	if complexity > 10 {
		stats.HighComplexity++
	}
}
