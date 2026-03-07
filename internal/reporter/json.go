package reporter

import (
	"encoding/json"
	"io"

	"github.com/opd-ai/go-stats-generator/internal/metrics"
)

// JSONReporter generates JSON formatted output for analysis reports and diffs.
// JSONReporter implements the Reporter interface and produces properly indented JSON.
type JSONReporter struct {
	indent          bool
	firstSection    bool
	sectionsWritten int
}

// NewJSONReporter creates a new JSON reporter with pretty-printing enabled by default.
// The reporter generates machine-readable JSON output suitable for CI/CD integration, automated
// analysis, and programmatic consumption. JSON format enables easy parsing, filtering with jq,
// and integration with monitoring/visualization tools. Use --format json flag to activate.
func NewJSONReporter() *JSONReporter {
	return &JSONReporter{
		indent:       true,
		firstSection: true,
	}
}

// BeginReport writes the JSON opening brace and metadata section for streaming output.
// This is the first method called in streaming mode. It initializes the JSON structure
// and writes repository metadata. Must be called before WriteSection.
func (jr *JSONReporter) BeginReport(output io.Writer, metadata *metrics.ReportMetadata) error {
	jr.firstSection = true
	jr.sectionsWritten = 0

	// Write opening brace and metadata
	if _, err := output.Write([]byte("{\n")); err != nil {
		return err
	}

	// Write metadata section
	if _, err := output.Write([]byte("  \"metadata\": ")); err != nil {
		return err
	}

	encoder := json.NewEncoder(output)
	if jr.indent {
		encoder.SetIndent("  ", "  ")
	}
	if err := encoder.Encode(metadata); err != nil {
		return err
	}

	jr.firstSection = false
	jr.sectionsWritten++
	return nil
}

// WriteSection writes an individual report section to the output stream.
// Sections are written as top-level JSON fields with proper comma separation.
// The sectionData must be JSON-serializable. Can be called multiple times.
func (jr *JSONReporter) WriteSection(output io.Writer, sectionName string, sectionData interface{}) error {
	// Write comma separator (except after metadata which was already written in BeginReport)
	if jr.sectionsWritten > 0 {
		if _, err := output.Write([]byte(",\n")); err != nil {
			return err
		}
	}

	// Write section name
	sectionJSON, err := json.Marshal(sectionName)
	if err != nil {
		return err
	}
	if _, err := output.Write([]byte("  ")); err != nil {
		return err
	}
	if _, err := output.Write(sectionJSON); err != nil {
		return err
	}
	if _, err := output.Write([]byte(": ")); err != nil {
		return err
	}

	// Write section data
	encoder := json.NewEncoder(output)
	if jr.indent {
		encoder.SetIndent("  ", "  ")
	}
	if err := encoder.Encode(sectionData); err != nil {
		return err
	}

	jr.sectionsWritten++
	return nil
}

// EndReport writes the JSON closing brace to finalize streaming output.
// This is the final method called in streaming mode. After this, the output
// represents a complete, valid JSON document. Must be called after all WriteSection calls.
func (jr *JSONReporter) EndReport(output io.Writer) error {
	if _, err := output.Write([]byte("}\n")); err != nil {
		return err
	}
	return nil
}

// Generate generates a JSON-formatted analysis report by encoding the metrics.Report structure
// with proper indentation for human readability. The output includes all analysis sections
// (functions, structs, packages, patterns, etc.) in a structured format. Errors occur only if
// the writer fails; the Report structure is always valid JSON-serializable.
func (jr *JSONReporter) Generate(report *metrics.Report, output io.Writer) error {
	encoder := json.NewEncoder(output)
	if jr.indent {
		encoder.SetIndent("", "  ")
	}

	return encoder.Encode(report)
}

// WriteDiff generates a JSON-formatted differential analysis report comparing two analysis snapshots.
// The output includes detailed change categories (improvements, regressions, new functions, removed
// functions) with line-level deltas and complexity changes. Essential for CI/CD quality gates and
// pull request reviews. Format enables programmatic comparison and trend tracking.
func (jr *JSONReporter) WriteDiff(output io.Writer, diff *metrics.ComplexityDiff) error {
	encoder := json.NewEncoder(output)
	if jr.indent {
		encoder.SetIndent("", "  ")
	}

	return encoder.Encode(diff)
}

// NewHTMLReporter creates a new HTML reporter for generating interactive, visually-rich analysis reports.
// HTML output includes embedded charts (via Chart.js), sortable tables, and collapsible sections for
// easy navigation. Best suited for human consumption and sharing with stakeholders via web browsers.
func NewHTMLReporter() Reporter {
	return NewHTMLReporterWithConfig(nil)
}
