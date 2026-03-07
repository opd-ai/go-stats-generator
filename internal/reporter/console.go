package reporter

import (
	"fmt"
	"io"

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

// NewConsoleReporter creates a new console reporter for generating rich terminal output with tables and colors.
// If cfg is nil, uses sensible defaults (colors enabled, overview included, 20-item limit per section).
// The reporter formats metrics using ANSI escape codes and Unicode box-drawing characters for enhanced readability.
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

// calculateDisplayLimit returns the effective limit for displaying items,
// capped at both the configured limit and a maximum of 10 items
func (cr *ConsoleReporter) calculateDisplayLimit(itemCount int) int {
	limit := cr.config.Limit
	if limit > itemCount {
		limit = itemCount
	}
	if limit > 10 {
		limit = 10
	}
	return limit
}

// Generate produces a human-readable console report with formatted tables, colors, and section headers for terminal display.
// It orchestrates writing of all report sections (functions, complexity, duplication, documentation) with appropriate
// formatting, pagination limits, and visual separators. Output is optimized for 80-120 column terminal widths with
// ANSI color codes for improved readability. This is the default output format when no --format flag is specified.
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

// writeReportSections iterates through all configured sections and writes those that should be included.
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

// shouldWriteOverview returns true if the overview section should be included in the report.
func (cr *ConsoleReporter) shouldWriteOverview(report *metrics.Report) bool {
	return cr.config.IncludeOverview
}

// shouldWriteFunctionAnalysis returns true if function details should be included.
func (cr *ConsoleReporter) shouldWriteFunctionAnalysis(report *metrics.Report) bool {
	return cr.config.IncludeDetails && len(report.Functions) > 0
}

// shouldWriteComplexityAnalysis returns true if complexity analysis should be included.
func (cr *ConsoleReporter) shouldWriteComplexityAnalysis(report *metrics.Report) bool {
	return cr.config.IncludeDetails
}

// shouldWritePackageAnalysis returns true if package metrics should be included.
func (cr *ConsoleReporter) shouldWritePackageAnalysis(report *metrics.Report) bool {
	return cr.config.IncludeDetails && len(report.Packages) > 0
}

// shouldWriteCircularDependencies returns true if circular dependency analysis should be included.
func (cr *ConsoleReporter) shouldWriteCircularDependencies(report *metrics.Report) bool {
	return cr.config.IncludeDetails && len(report.Packages) > 0
}

// shouldWriteDuplicationAnalysis returns true if duplication metrics should be included.
func (cr *ConsoleReporter) shouldWriteDuplicationAnalysis(report *metrics.Report) bool {
	return cr.config.IncludeDetails && report.Duplication.ClonePairs > 0
}

// shouldWriteNamingAnalysis returns true if naming violation analysis should be included.
func (cr *ConsoleReporter) shouldWriteNamingAnalysis(report *metrics.Report) bool {
	totalNamingViolations := report.Naming.FileNameViolations + report.Naming.IdentifierViolations + report.Naming.PackageNameViolations
	return cr.config.IncludeDetails && totalNamingViolations > 0
}

// shouldWritePlacementAnalysis returns true if placement issue analysis should be included.
func (cr *ConsoleReporter) shouldWritePlacementAnalysis(report *metrics.Report) bool {
	totalPlacementIssues := report.Placement.MisplacedFunctions + report.Placement.MisplacedMethods + report.Placement.LowCohesionFiles
	return cr.config.IncludeDetails && totalPlacementIssues > 0
}

// shouldWriteDocumentationAnalysis returns true if documentation coverage and annotation metrics should be included.
func (cr *ConsoleReporter) shouldWriteDocumentationAnalysis(report *metrics.Report) bool {
	totalAnnotations := len(report.Documentation.TODOComments) + len(report.Documentation.FIXMEComments) + len(report.Documentation.HACKComments) + len(report.Documentation.BUGComments)
	return cr.config.IncludeDetails && (report.Documentation.Coverage.Overall > 0 || totalAnnotations > 0)
}

// shouldWriteBurdenAnalysis returns true if code burden metrics should be included.
func (cr *ConsoleReporter) shouldWriteBurdenAnalysis(report *metrics.Report) bool {
	totalBurdenIssues := len(report.Burden.MagicNumbers) + len(report.Burden.DeadCode.UnreferencedFunctions) + len(report.Burden.DeadCode.UnreachableCode) + len(report.Burden.ComplexSignatures) + len(report.Burden.DeeplyNestedFunctions) + len(report.Burden.FeatureEnvyMethods)
	return cr.config.IncludeDetails && totalBurdenIssues > 0
}

// shouldWriteOrganizationAnalysis returns true if code organization metrics should be included.
func (cr *ConsoleReporter) shouldWriteOrganizationAnalysis(report *metrics.Report) bool {
	totalOrgIssues := len(report.Organization.OversizedFiles) + len(report.Organization.OversizedPackages) + len(report.Organization.DeepDirectories) + len(report.Organization.HighFanInPackages) + len(report.Organization.HighFanOutPackages)
	return cr.config.IncludeDetails && totalOrgIssues > 0
}

// shouldWriteRefactoringSuggestions returns true if refactoring suggestions should be included.
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
