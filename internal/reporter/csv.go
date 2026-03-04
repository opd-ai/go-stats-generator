package reporter

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"

	"github.com/opd-ai/go-stats-generator/internal/metrics"
)

// CSVReporter generates analysis reports in CSV format.
type CSVReporter struct{}

// NewCSVReporter creates a new CSV reporter for generating analysis reports in comma-separated values format.
// CSV output is ideal for importing into spreadsheet applications, business intelligence tools, or data pipelines.
// Each section (functions, structs, packages) is written as a separate CSV table with appropriate headers.
func NewCSVReporter() Reporter {
	return &CSVReporter{}
}

// Generate writes the analysis report to the output writer in CSV format.
func (r *CSVReporter) Generate(report *metrics.Report, output io.Writer) error {
	writer := csv.NewWriter(output)
	defer writer.Flush()

	sections := []func(*csv.Writer, *metrics.Report) error{
		r.writeMetadataSection,
		r.writeOverviewSection,
		r.writeFunctionsSection,
		r.writeStructsSection,
		r.writePackagesSection,
		r.writeNamingSection,
	}

	for _, writeSection := range sections {
		if err := writeSection(writer, report); err != nil {
			return err
		}
	}

	return nil
}

// writeMetadataSection writes the metadata section to CSV output.
func (r *CSVReporter) writeMetadataSection(writer *csv.Writer, report *metrics.Report) error {
	if err := writer.Write([]string{"# METADATA"}); err != nil {
		return fmt.Errorf("failed to write metadata header: %w", err)
	}

	metadataRows := [][]string{
		{"Repository", report.Metadata.Repository},
		{"Generated At", report.Metadata.GeneratedAt.Format("2006-01-02 15:04:05")},
		{"Analysis Time", report.Metadata.AnalysisTime.String()},
		{"Files Processed", strconv.Itoa(report.Metadata.FilesProcessed)},
		{"Tool Version", report.Metadata.ToolVersion},
		{"Go Version", report.Metadata.GoVersion},
	}

	for _, row := range metadataRows {
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write metadata row: %w", err)
		}
	}

	return nil
}

// writeOverviewSection writes the overview statistics section to CSV output.
func (r *CSVReporter) writeOverviewSection(writer *csv.Writer, report *metrics.Report) error {
	if err := writer.Write([]string{""}); err != nil {
		return err
	}
	if err := writer.Write([]string{"# OVERVIEW"}); err != nil {
		return fmt.Errorf("failed to write overview header: %w", err)
	}

	overviewRows := [][]string{
		{"Total Lines of Code", strconv.Itoa(report.Overview.TotalLinesOfCode)},
		{"Total Functions", strconv.Itoa(report.Overview.TotalFunctions)},
		{"Total Methods", strconv.Itoa(report.Overview.TotalMethods)},
		{"Total Structs", strconv.Itoa(report.Overview.TotalStructs)},
		{"Total Interfaces", strconv.Itoa(report.Overview.TotalInterfaces)},
		{"Total Packages", strconv.Itoa(report.Overview.TotalPackages)},
		{"Total Files", strconv.Itoa(report.Overview.TotalFiles)},
	}

	for _, row := range overviewRows {
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write overview row: %w", err)
		}
	}

	return nil
}

// writeFunctionsSection outputs the functions analysis section to CSV format,
// writing headers (name, package, file, metrics) and detailed rows for each
// function with complexity, line counts, signature details, and documentation status.
func (r *CSVReporter) writeFunctionsSection(writer *csv.Writer, report *metrics.Report) error {
	headers := []string{
		"Name", "Package", "File", "Line", "Is Exported", "Is Method",
		"Lines Total", "Lines Code", "Lines Comments", "Lines Blank",
		"Cyclomatic Complexity", "Cognitive Complexity", "Nesting Depth", "Overall Complexity",
		"Parameter Count", "Return Count", "Has Variadic", "Returns Error",
		"Has Documentation", "Documentation Quality",
	}

	formatter := func(fn metrics.FunctionMetrics) []string {
		return []string{
			fn.Name,
			fn.Package,
			fn.File,
			strconv.Itoa(fn.Line),
			formatBool(fn.IsExported),
			formatBool(fn.IsMethod),
			strconv.Itoa(fn.Lines.Total),
			strconv.Itoa(fn.Lines.Code),
			strconv.Itoa(fn.Lines.Comments),
			strconv.Itoa(fn.Lines.Blank),
			strconv.Itoa(fn.Complexity.Cyclomatic),
			strconv.Itoa(fn.Complexity.Cognitive),
			strconv.Itoa(fn.Complexity.NestingDepth),
			formatFloat(fn.Complexity.Overall),
			strconv.Itoa(fn.Signature.ParameterCount),
			strconv.Itoa(fn.Signature.ReturnCount),
			formatBool(fn.Signature.VariadicUsage),
			formatBool(fn.Signature.ErrorReturn),
			formatBool(fn.Documentation.HasComment),
			formatFloat(fn.Documentation.QualityScore),
		}
	}

	return writeSectionData(writer, "# FUNCTIONS", headers, report.Functions, formatter)
}

// writeStructsSection outputs the structs analysis section to CSV format,
// writing headers (name, package, file, fields, methods) and detailed rows
// for each struct with complexity metrics and field categorization.
func (r *CSVReporter) writeStructsSection(writer *csv.Writer, report *metrics.Report) error {
	headers := []string{
		"Name", "Package", "File", "Line", "Is Exported", "Total Fields",
		"Methods Count", "Cyclomatic Complexity", "Overall Complexity",
		"Has Documentation", "Documentation Quality",
	}

	formatter := func(st metrics.StructMetrics) []string {
		return []string{
			st.Name,
			st.Package,
			st.File,
			strconv.Itoa(st.Line),
			formatBool(st.IsExported),
			strconv.Itoa(st.TotalFields),
			strconv.Itoa(len(st.Methods)),
			strconv.Itoa(st.Complexity.Cyclomatic),
			formatFloat(st.Complexity.Overall),
			formatBool(st.Documentation.HasComment),
			formatFloat(st.Documentation.QualityScore),
		}
	}

	return writeSectionData(writer, "# STRUCTS", headers, report.Structs, formatter)
}

// writePackagesSection outputs the packages analysis section to CSV format,
// writing headers (name, path, files, functions, structs) and detailed rows
// for each package with dependency metrics, cohesion, and coupling scores.
func (r *CSVReporter) writePackagesSection(writer *csv.Writer, report *metrics.Report) error {
	headers := []string{
		"Name", "Path", "Files", "Functions", "Structs", "Interfaces",
		"Lines of Code", "Dependencies", "Dependents", "Cohesion", "Coupling",
		"Has Documentation", "Documentation Quality",
	}

	formatter := func(pkg metrics.PackageMetrics) []string {
		return []string{
			pkg.Name,
			pkg.Path,
			strconv.Itoa(len(pkg.Files)),
			strconv.Itoa(pkg.Functions),
			strconv.Itoa(pkg.Structs),
			strconv.Itoa(pkg.Interfaces),
			strconv.Itoa(pkg.Lines.Code),
			strconv.Itoa(len(pkg.Dependencies)),
			strconv.Itoa(len(pkg.Dependents)),
			formatFloat(pkg.CohesionScore),
			formatFloat(pkg.CouplingScore),
			formatBool(pkg.Documentation.HasComment),
			formatFloat(pkg.Documentation.QualityScore),
		}
	}

	return writeSectionData(writer, "# PACKAGES", headers, report.Packages, formatter)
}

// writeNamingSection outputs the naming convention analysis section to CSV format,
// including file name violations, identifier violations, package name violations,
// and overall naming score with detailed violation listings.
func (r *CSVReporter) writeNamingSection(writer *csv.Writer, report *metrics.Report) error {
	if !hasNamingViolations(report) {
		return nil
	}

	if err := writeNamingHeader(writer); err != nil {
		return err
	}

	if err := writeNamingSummaryRows(writer, report); err != nil {
		return err
	}

	return writeNamingSubsections(r, writer, report)
}

func hasNamingViolations(report *metrics.Report) bool {
	totalNamingViolations := report.Naming.FileNameViolations + report.Naming.IdentifierViolations + report.Naming.PackageNameViolations
	return totalNamingViolations > 0
}

func writeNamingHeader(writer *csv.Writer) error {
	if err := writer.Write([]string{""}); err != nil {
		return err
	}
	if err := writer.Write([]string{"# NAMING CONVENTION ANALYSIS"}); err != nil {
		return fmt.Errorf("failed to write naming header: %w", err)
	}
	return nil
}

// writeNamingSummaryRows writes naming convention summary statistics to CSV including file name
// violations, identifier violations, package name violations, and overall naming score. This provides
// a high-level overview of naming convention compliance before detailed violation listings.
func writeNamingSummaryRows(writer *csv.Writer, report *metrics.Report) error {
	namingSummary := [][]string{
		{"File Name Violations", strconv.Itoa(report.Naming.FileNameViolations)},
		{"Identifier Violations", strconv.Itoa(report.Naming.IdentifierViolations)},
		{"Package Name Violations", strconv.Itoa(report.Naming.PackageNameViolations)},
		{"Overall Naming Score", formatFloat(report.Naming.OverallNamingScore)},
	}

	for _, row := range namingSummary {
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write naming summary row: %w", err)
		}
	}
	return nil
}

func writeNamingSubsections(r *CSVReporter, writer *csv.Writer, report *metrics.Report) error {
	if err := r.writeFileNameIssues(writer, report); err != nil {
		return err
	}
	if err := r.writeIdentifierIssues(writer, report); err != nil {
		return err
	}
	return r.writePackageNameIssues(writer, report)
}

// writeFileNameIssues outputs file naming violations to CSV format, listing
// each file with naming issues, violation type, suggested name, and severity.
func (r *CSVReporter) writeFileNameIssues(writer *csv.Writer, report *metrics.Report) error {
	headers := []string{"File", "Violation Type", "Description", "Suggested Name", "Severity"}

	formatter := func(issue metrics.FileNameViolation) []string {
		return []string{
			issue.File,
			issue.ViolationType,
			issue.Description,
			issue.SuggestedName,
			issue.Severity,
		}
	}

	return writeSectionData(writer, "## FILE NAME ISSUES", headers, report.Naming.FileNameIssues, formatter)
}

// writeIdentifierIssues outputs identifier naming violations to CSV format,
// listing each identifier with naming issues (underscores, wrong acronyms),
// violation type, suggested name, and severity.
func (r *CSVReporter) writeIdentifierIssues(writer *csv.Writer, report *metrics.Report) error {
	headers := []string{"Name", "File", "Line", "Type", "Violation Type", "Description", "Suggested Name", "Severity"}

	formatter := func(issue metrics.IdentifierViolation) []string {
		return []string{
			issue.Name,
			issue.File,
			strconv.Itoa(issue.Line),
			issue.Type,
			issue.ViolationType,
			issue.Description,
			issue.SuggestedName,
			issue.Severity,
		}
	}

	return writeSectionData(writer, "## IDENTIFIER ISSUES", headers, report.Naming.IdentifierIssues, formatter)
}

// writePackageNameIssues outputs package naming violations to CSV format,
// listing each package with naming issues (underscores, non-conventional),
// violation description, and severity.
func (r *CSVReporter) writePackageNameIssues(writer *csv.Writer, report *metrics.Report) error {
	headers := []string{"Package", "Directory", "Violation Type", "Description", "Suggested Name", "Severity"}

	formatter := func(issue metrics.PackageNameViolation) []string {
		return []string{
			issue.Package,
			issue.Directory,
			issue.ViolationType,
			issue.Description,
			issue.SuggestedName,
			issue.Severity,
		}
	}

	return writeSectionData(writer, "## PACKAGE NAME ISSUES", headers, report.Naming.PackageNameIssues, formatter)
}

// WriteDiff writes a metrics comparison report to the output writer in CSV format.
func (r *CSVReporter) WriteDiff(output io.Writer, diff *metrics.ComplexityDiff) error {
	writer := csv.NewWriter(output)
	defer writer.Flush()

	if err := writer.Write([]string{"# METRICS COMPARISON REPORT"}); err != nil {
		return fmt.Errorf("failed to write diff header: %w", err)
	}

	if err := r.writeDiffSummary(writer, diff); err != nil {
		return err
	}

	if err := r.writeDiffRegressions(writer, diff); err != nil {
		return err
	}

	if err := r.writeDiffImprovements(writer, diff); err != nil {
		return err
	}

	return nil
}

// writeDiffSummary writes the diff summary section to CSV output.
func (r *CSVReporter) writeDiffSummary(writer *csv.Writer, diff *metrics.ComplexityDiff) error {
	if err := writer.Write([]string{""}); err != nil {
		return err
	}
	if err := writer.Write([]string{"# SUMMARY"}); err != nil {
		return fmt.Errorf("failed to write summary header: %w", err)
	}

	summaryRows := [][]string{
		{"Total Changes", strconv.Itoa(diff.Summary.TotalChanges)},
		{"Significant Changes", strconv.Itoa(diff.Summary.SignificantChanges)},
		{"Regressions", strconv.Itoa(diff.Summary.RegressionCount)},
		{"Improvements", strconv.Itoa(diff.Summary.ImprovementCount)},
		{"Analysis Date", diff.Timestamp.Format("2006-01-02 15:04:05")},
	}

	for _, row := range summaryRows {
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write summary row: %w", err)
		}
	}

	return nil
}

// writeDiffRegressions writes the regressions section to CSV output.
func (r *CSVReporter) writeDiffRegressions(writer *csv.Writer, diff *metrics.ComplexityDiff) error {
	if len(diff.Regressions) == 0 {
		return nil
	}

	if err := writeDiffSectionHeader(writer, "# REGRESSIONS"); err != nil {
		return err
	}

	headers := []string{"Location", "Function", "Old Value", "New Value", "Change", "Severity"}
	if err := writer.Write(headers); err != nil {
		return fmt.Errorf("failed to write regression headers: %w", err)
	}

	return writeRegressionRows(writer, diff.Regressions)
}

func writeDiffSectionHeader(writer *csv.Writer, title string) error {
	if err := writer.Write([]string{""}); err != nil {
		return err
	}
	if err := writer.Write([]string{title}); err != nil {
		return fmt.Errorf("failed to write section header: %w", err)
	}
	return nil
}

// writeRegressionRows writes regression details to CSV including function location, old/new metric
// values, absolute delta, and severity classification. Each regression is output as a separate row
// with formatted numeric values for analysis and filtering in spreadsheet tools.
func writeRegressionRows(writer *csv.Writer, regressions []metrics.Regression) error {
	for _, reg := range regressions {
		row := []string{
			reg.Location,
			reg.Function,
			formatValue(reg.OldValue),
			formatValue(reg.NewValue),
			formatFloat(reg.Delta.Absolute),
			string(reg.Severity),
		}

		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write regression row: %w", err)
		}
	}
	return nil
}

// writeDiffImprovements writes the improvements section to CSV output.
func (r *CSVReporter) writeDiffImprovements(writer *csv.Writer, diff *metrics.ComplexityDiff) error {
	if len(diff.Improvements) == 0 {
		return nil
	}

	if err := writeDiffSectionHeader(writer, "# IMPROVEMENTS"); err != nil {
		return err
	}

	headers := []string{"Location", "Function", "Old Value", "New Value", "Change", "Impact"}
	if err := writer.Write(headers); err != nil {
		return fmt.Errorf("failed to write improvement headers: %w", err)
	}

	return writeImprovementRows(writer, diff.Improvements)
}

// writeImprovementRows writes improvement details to CSV including function location, old/new metric
// values, absolute delta, and impact classification. Each improvement is output as a separate row
// with formatted numeric values for tracking code quality gains over time.
func writeImprovementRows(writer *csv.Writer, improvements []metrics.Improvement) error {
	for _, imp := range improvements {
		row := []string{
			imp.Location,
			imp.Function,
			formatValue(imp.OldValue),
			formatValue(imp.NewValue),
			formatFloat(imp.Delta.Absolute),
			string(imp.Impact),
		}

		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write improvement row: %w", err)
		}
	}
	return nil
}

// writeSectionData is a generic helper that reduces duplication in CSV section writing.
// It handles the common pattern: check if empty, write blank line, write header, write column headers, write rows.
func writeSectionData[T any](
	writer *csv.Writer,
	sectionTitle string,
	headers []string,
	data []T,
	rowFormatter func(T) []string,
) error {
	if len(data) == 0 {
		return nil
	}

	if err := writeCSVSectionHeader(writer, sectionTitle); err != nil {
		return err
	}

	if err := writer.Write(headers); err != nil {
		return fmt.Errorf("failed to write column headers: %w", err)
	}

	return writeCSVDataRows(writer, data, rowFormatter)
}

func writeCSVSectionHeader(writer *csv.Writer, title string) error {
	if err := writer.Write([]string{""}); err != nil {
		return err
	}
	if err := writer.Write([]string{title}); err != nil {
		return fmt.Errorf("failed to write section header: %w", err)
	}
	return nil
}

func writeCSVDataRows[T any](writer *csv.Writer, data []T, rowFormatter func(T) []string) error {
	for _, item := range data {
		row := rowFormatter(item)
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write data row: %w", err)
		}
	}
	return nil
}

// formatBool formats a boolean value as a string for CSV output
func formatBool(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
