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

// NewCSVReporter creates a new CSV reporter
func NewCSVReporter() Reporter {
	return &CSVReporter{}
}

// Generate writes the analysis report to the output writer in CSV format.
func (r *CSVReporter) Generate(report *metrics.Report, output io.Writer) error {
	writer := csv.NewWriter(output)
	defer writer.Flush()

	if err := r.writeMetadataSection(writer, report); err != nil {
		return err
	}

	if err := r.writeOverviewSection(writer, report); err != nil {
		return err
	}

	if err := r.writeFunctionsSection(writer, report); err != nil {
		return err
	}

	if err := r.writeStructsSection(writer, report); err != nil {
		return err
	}

	if err := r.writePackagesSection(writer, report); err != nil {
		return err
	}

	if err := r.writeNamingSection(writer, report); err != nil {
		return err
	}

	return nil
}

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
	if len(report.Functions) == 0 {
		return nil
	}

	if err := writer.Write([]string{""}); err != nil {
		return err
	}
	if err := writer.Write([]string{"# FUNCTIONS"}); err != nil {
		return fmt.Errorf("failed to write functions header: %w", err)
	}

	functionHeaders := []string{
		"Name", "Package", "File", "Line", "Is Exported", "Is Method",
		"Lines Total", "Lines Code", "Lines Comments", "Lines Blank",
		"Cyclomatic Complexity", "Cognitive Complexity", "Nesting Depth", "Overall Complexity",
		"Parameter Count", "Return Count", "Has Variadic", "Returns Error",
		"Has Documentation", "Documentation Quality",
	}

	if err := writer.Write(functionHeaders); err != nil {
		return fmt.Errorf("failed to write function headers: %w", err)
	}

	for _, fn := range report.Functions {
		row := []string{
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

		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write function row: %w", err)
		}
	}

	return nil
}

// writeStructsSection outputs the structs analysis section to CSV format,
// writing headers (name, package, file, fields, methods) and detailed rows
// for each struct with complexity metrics and field categorization.
func (r *CSVReporter) writeStructsSection(writer *csv.Writer, report *metrics.Report) error {
	if len(report.Structs) == 0 {
		return nil
	}

	if err := writer.Write([]string{""}); err != nil {
		return err
	}
	if err := writer.Write([]string{"# STRUCTS"}); err != nil {
		return fmt.Errorf("failed to write structs header: %w", err)
	}

	structHeaders := []string{
		"Name", "Package", "File", "Line", "Is Exported", "Total Fields",
		"Methods Count", "Cyclomatic Complexity", "Overall Complexity",
		"Has Documentation", "Documentation Quality",
	}

	if err := writer.Write(structHeaders); err != nil {
		return fmt.Errorf("failed to write struct headers: %w", err)
	}

	for _, st := range report.Structs {
		row := []string{
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

		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write struct row: %w", err)
		}
	}

	return nil
}

// writePackagesSection outputs the packages analysis section to CSV format,
// writing headers (name, path, files, functions, structs) and detailed rows
// for each package with dependency metrics, cohesion, and coupling scores.
func (r *CSVReporter) writePackagesSection(writer *csv.Writer, report *metrics.Report) error {
	if len(report.Packages) == 0 {
		return nil
	}

	if err := writer.Write([]string{""}); err != nil {
		return err
	}
	if err := writer.Write([]string{"# PACKAGES"}); err != nil {
		return fmt.Errorf("failed to write packages header: %w", err)
	}

	packageHeaders := []string{
		"Name", "Path", "Files", "Functions", "Structs", "Interfaces",
		"Lines of Code", "Dependencies", "Dependents", "Cohesion", "Coupling",
		"Has Documentation", "Documentation Quality",
	}

	if err := writer.Write(packageHeaders); err != nil {
		return fmt.Errorf("failed to write package headers: %w", err)
	}

	for _, pkg := range report.Packages {
		row := []string{
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

		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write package row: %w", err)
		}
	}

	return nil
}

// writeNamingSection outputs the naming convention analysis section to CSV format,
// including file name violations, identifier violations, package name violations,
// and overall naming score with detailed violation listings.
func (r *CSVReporter) writeNamingSection(writer *csv.Writer, report *metrics.Report) error {
	totalNamingViolations := report.Naming.FileNameViolations + report.Naming.IdentifierViolations + report.Naming.PackageNameViolations
	if totalNamingViolations == 0 {
		return nil
	}

	if err := writer.Write([]string{""}); err != nil {
		return err
	}
	if err := writer.Write([]string{"# NAMING CONVENTION ANALYSIS"}); err != nil {
		return fmt.Errorf("failed to write naming header: %w", err)
	}

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

	if err := r.writeFileNameIssues(writer, report); err != nil {
		return err
	}

	if err := r.writeIdentifierIssues(writer, report); err != nil {
		return err
	}

	if err := r.writePackageNameIssues(writer, report); err != nil {
		return err
	}

	return nil
}

// writeFileNameIssues outputs file naming violations to CSV format, listing
// each file with naming issues, violation type, suggested name, and severity.
func (r *CSVReporter) writeFileNameIssues(writer *csv.Writer, report *metrics.Report) error {
	if len(report.Naming.FileNameIssues) == 0 {
		return nil
	}

	if err := writer.Write([]string{""}); err != nil {
		return err
	}
	if err := writer.Write([]string{"## FILE NAME ISSUES"}); err != nil {
		return fmt.Errorf("failed to write file name issues header: %w", err)
	}

	fileNameHeaders := []string{"File", "Violation Type", "Description", "Suggested Name", "Severity"}
	if err := writer.Write(fileNameHeaders); err != nil {
		return fmt.Errorf("failed to write file name issue headers: %w", err)
	}

	for _, issue := range report.Naming.FileNameIssues {
		row := []string{
			issue.File,
			issue.ViolationType,
			issue.Description,
			issue.SuggestedName,
			issue.Severity,
		}
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write file name issue row: %w", err)
		}
	}

	return nil
}

// writeIdentifierIssues outputs identifier naming violations to CSV format,
// listing each identifier with naming issues (underscores, wrong acronyms),
// violation type, suggested name, and severity.
func (r *CSVReporter) writeIdentifierIssues(writer *csv.Writer, report *metrics.Report) error {
	if len(report.Naming.IdentifierIssues) == 0 {
		return nil
	}

	if err := writer.Write([]string{""}); err != nil {
		return err
	}
	if err := writer.Write([]string{"## IDENTIFIER ISSUES"}); err != nil {
		return fmt.Errorf("failed to write identifier issues header: %w", err)
	}

	identifierHeaders := []string{"Name", "File", "Line", "Type", "Violation Type", "Description", "Suggested Name", "Severity"}
	if err := writer.Write(identifierHeaders); err != nil {
		return fmt.Errorf("failed to write identifier issue headers: %w", err)
	}

	for _, issue := range report.Naming.IdentifierIssues {
		row := []string{
			issue.Name,
			issue.File,
			strconv.Itoa(issue.Line),
			issue.Type,
			issue.ViolationType,
			issue.Description,
			issue.SuggestedName,
			issue.Severity,
		}
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write identifier issue row: %w", err)
		}
	}

	return nil
}

// writePackageNameIssues outputs package naming violations to CSV format,
// listing each package with naming issues (underscores, non-conventional),
// violation description, and severity.
func (r *CSVReporter) writePackageNameIssues(writer *csv.Writer, report *metrics.Report) error {
	if len(report.Naming.PackageNameIssues) == 0 {
		return nil
	}

	if err := writer.Write([]string{""}); err != nil {
		return err
	}
	if err := writer.Write([]string{"## PACKAGE NAME ISSUES"}); err != nil {
		return fmt.Errorf("failed to write package name issues header: %w", err)
	}

	packageNameHeaders := []string{"Package", "Directory", "Violation Type", "Description", "Suggested Name", "Severity"}
	if err := writer.Write(packageNameHeaders); err != nil {
		return fmt.Errorf("failed to write package name issue headers: %w", err)
	}

	for _, issue := range report.Naming.PackageNameIssues {
		row := []string{
			issue.Package,
			issue.Directory,
			issue.ViolationType,
			issue.Description,
			issue.SuggestedName,
			issue.Severity,
		}
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write package name issue row: %w", err)
		}
	}

	return nil
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

func (r *CSVReporter) writeDiffRegressions(writer *csv.Writer, diff *metrics.ComplexityDiff) error {
	if len(diff.Regressions) == 0 {
		return nil
	}

	if err := writer.Write([]string{""}); err != nil {
		return err
	}
	if err := writer.Write([]string{"# REGRESSIONS"}); err != nil {
		return fmt.Errorf("failed to write regressions header: %w", err)
	}

	regressionHeaders := []string{"Location", "Function", "Old Value", "New Value", "Change", "Severity"}
	if err := writer.Write(regressionHeaders); err != nil {
		return fmt.Errorf("failed to write regression headers: %w", err)
	}

	for _, reg := range diff.Regressions {
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

func (r *CSVReporter) writeDiffImprovements(writer *csv.Writer, diff *metrics.ComplexityDiff) error {
	if len(diff.Improvements) == 0 {
		return nil
	}

	if err := writer.Write([]string{""}); err != nil {
		return err
	}
	if err := writer.Write([]string{"# IMPROVEMENTS"}); err != nil {
		return fmt.Errorf("failed to write improvements header: %w", err)
	}

	improvementHeaders := []string{"Location", "Function", "Old Value", "New Value", "Change", "Impact"}
	if err := writer.Write(improvementHeaders); err != nil {
		return fmt.Errorf("failed to write improvement headers: %w", err)
	}

	for _, imp := range diff.Improvements {
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

// formatBool formats a boolean value as a string for CSV output
func formatBool(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
