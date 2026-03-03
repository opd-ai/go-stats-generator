package reporter

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strconv"

	"github.com/opd-ai/go-stats-generator/internal/metrics"
)

// JSONReporter generates JSON output
type JSONReporter struct {
	indent bool
}

// NewJSONReporter creates a new JSON reporter
func NewJSONReporter() *JSONReporter {
	return &JSONReporter{
		indent: true,
	}
}

// Generate generates a JSON report
func (jr *JSONReporter) Generate(report *metrics.Report, output io.Writer) error {
	encoder := json.NewEncoder(output)
	if jr.indent {
		encoder.SetIndent("", "  ")
	}

	return encoder.Encode(report)
}

// WriteDiff generates a JSON diff report
func (jr *JSONReporter) WriteDiff(output io.Writer, diff *metrics.ComplexityDiff) error {
	encoder := json.NewEncoder(output)
	if jr.indent {
		encoder.SetIndent("", "  ")
	}

	return encoder.Encode(diff)
}

// NewCSVReporter creates a new CSV reporter (placeholder)
func NewCSVReporter() Reporter {
	return &CSVReporter{}
}

// NewHTMLReporter creates a new HTML reporter
func NewHTMLReporter() Reporter {
	return NewHTMLReporterWithConfig(nil)
}

// Placeholder implementations for future reporters

// CSVReporter generates analysis reports in CSV format.
type CSVReporter struct{}

// Generate writes the analysis report to the output writer in CSV format.
func (r *CSVReporter) Generate(report *metrics.Report, output io.Writer) error {
	writer := csv.NewWriter(output)
	defer writer.Flush()

	// Write metadata header
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

	// Write overview header
	if err := writer.Write([]string{""}); err != nil { // Empty row
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

	// Write functions header
	if len(report.Functions) > 0 {
		if err := writer.Write([]string{""}); err != nil { // Empty row
			return err
		}
		if err := writer.Write([]string{"# FUNCTIONS"}); err != nil {
			return fmt.Errorf("failed to write functions header: %w", err)
		}

		// Function column headers
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

		// Function data rows
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
	}

	// Write structs section
	if len(report.Structs) > 0 {
		if err := writer.Write([]string{""}); err != nil { // Empty row
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
	}

	// Write packages section
	if len(report.Packages) > 0 {
		if err := writer.Write([]string{""}); err != nil { // Empty row
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
	}

	// Write naming metrics section
	totalNamingViolations := report.Naming.FileNameViolations + report.Naming.IdentifierViolations + report.Naming.PackageNameViolations
	if totalNamingViolations > 0 {
		if err := writer.Write([]string{""}); err != nil { // Empty row
			return err
		}
		if err := writer.Write([]string{"# NAMING CONVENTION ANALYSIS"}); err != nil {
			return fmt.Errorf("failed to write naming header: %w", err)
		}

		// Naming summary
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

		// File name violations
		if len(report.Naming.FileNameIssues) > 0 {
			if err := writer.Write([]string{""}); err != nil { // Empty row
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
		}

		// Identifier violations
		if len(report.Naming.IdentifierIssues) > 0 {
			if err := writer.Write([]string{""}); err != nil { // Empty row
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
		}

		// Package name violations
		if len(report.Naming.PackageNameIssues) > 0 {
			if err := writer.Write([]string{""}); err != nil { // Empty row
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
		}
	}

	return nil
}

// WriteDiff writes a metrics comparison report to the output writer in CSV format.
func (r *CSVReporter) WriteDiff(output io.Writer, diff *metrics.ComplexityDiff) error {
	writer := csv.NewWriter(output)
	defer writer.Flush()

	// Write header
	if err := writer.Write([]string{"# METRICS COMPARISON REPORT"}); err != nil {
		return fmt.Errorf("failed to write diff header: %w", err)
	}

	// Write summary
	if err := writer.Write([]string{""}); err != nil { // Empty row
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

	// Write regressions if any
	if len(diff.Regressions) > 0 {
		if err := writer.Write([]string{""}); err != nil { // Empty row
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
	}

	// Write improvements if any
	if len(diff.Improvements) > 0 {
		if err := writer.Write([]string{""}); err != nil { // Empty row
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
	}

	return nil
}

// Helper functions for CSV formatting
func formatBool(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
