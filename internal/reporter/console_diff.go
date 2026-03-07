package reporter

import (
	"fmt"
	"io"
	"time"

	"github.com/opd-ai/go-stats-generator/internal/metrics"
)

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

// writeRegressionEntry formats and outputs a single regression with icon, details, and suggestion.
func (cr *ConsoleReporter) writeRegressionEntry(output io.Writer, regression metrics.Regression) {
	icon := cr.getSeverityIcon(regression.Severity)
	fmt.Fprintf(output, "%s %s: %s\n", icon, regression.Type, regression.Location)
	cr.writeRegressionFile(output, regression)
	cr.writeRegressionChange(output, regression)
	cr.writeRegressionSuggestion(output, regression)
	fmt.Fprintln(output, "")
}

// writeRegressionFile outputs the file and line number information for a regression.
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

// writeRegressionChange displays the old and new values with percentage change for a regression.
func (cr *ConsoleReporter) writeRegressionChange(output io.Writer, regression metrics.Regression) {
	fmt.Fprintf(output, "   Change: %v → %v", regression.OldValue, regression.NewValue)
	if regression.Delta.Percentage > 0 {
		fmt.Fprintf(output, " (%+.1f%%)", regression.Delta.Percentage)
	}
	fmt.Fprintln(output, "")
}

// writeRegressionSuggestion outputs the remediation suggestion if available.
func (cr *ConsoleReporter) writeRegressionSuggestion(output io.Writer, regression metrics.Regression) {
	if regression.Suggestion != "" {
		fmt.Fprintf(output, "   Suggestion: %s\n", regression.Suggestion)
	}
}

// getSeverityIcon returns an emoji icon corresponding to the regression severity level.
func (cr *ConsoleReporter) getSeverityIcon(severity metrics.SeverityLevel) string {
	switch severity {
	case metrics.SeverityLevelCritical:
		return "🚨"
	case metrics.SeverityLevelViolation:
		return "❌"
	case metrics.SeverityLevelWarning:
		return "⚠️"
	default:
		return "ℹ️"
	}
}

// writeDiffImprovements outputs the list of improvements with their details.
func (cr *ConsoleReporter) writeDiffImprovements(output io.Writer, improvements []metrics.Improvement) {
	fmt.Fprintln(output, "=== IMPROVEMENTS ===")

	for _, improvement := range improvements {
		writeImprovementEntry(output, improvement)
	}
}

// writeImprovementEntry formats and outputs a single improvement with file, change details, and benefit.
func writeImprovementEntry(output io.Writer, improvement metrics.Improvement) {
	fmt.Fprintf(output, "✅ %s: %s\n", improvement.Type, improvement.Location)
	writeImprovementFile(output, improvement)
	writeImprovementChange(output, improvement)
	writeImprovementBenefit(output, improvement)
	fmt.Fprintln(output, "")
}

// writeImprovementFile outputs the file and line number for an improvement.
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

// writeImprovementChange displays the old and new values with improvement percentage.
func writeImprovementChange(output io.Writer, improvement metrics.Improvement) {
	fmt.Fprintf(output, "   Change: %v → %v", improvement.OldValue, improvement.NewValue)
	if improvement.Delta.Percentage > 0 {
		fmt.Fprintf(output, " (%.1f%% improvement)", improvement.Delta.Percentage)
	}
	fmt.Fprintln(output, "")
}

// writeImprovementBenefit outputs the benefit description if available.
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

// groupChangesByCategory organizes metric changes into a map indexed by category.
func groupChangesByCategory(changes []metrics.MetricChange) map[string][]metrics.MetricChange {
	changesByCategory := make(map[string][]metrics.MetricChange)
	for _, change := range changes {
		changesByCategory[change.Category] = append(changesByCategory[change.Category], change)
	}
	return changesByCategory
}

// writeCategoryChanges outputs all changes in a category with values and percentage changes.
func writeCategoryChanges(output io.Writer, changes []metrics.MetricChange) {
	for _, change := range changes {
		fmt.Fprintf(output, "- %s: %v → %v", change.Name, change.OldValue, change.NewValue)
		if change.Delta.Percentage > 0 {
			writeChangePercentage(output, change.Delta)
		}
		fmt.Fprintln(output, "")
	}
}

// writeChangePercentage formats and outputs the percentage change with direction indicator.
func writeChangePercentage(output io.Writer, delta metrics.Delta) {
	direction := "+"
	if delta.Direction == metrics.ChangeDirectionDecrease {
		direction = "-"
	}
	fmt.Fprintf(output, " (%s%.1f%%)", direction, delta.Percentage)
}
