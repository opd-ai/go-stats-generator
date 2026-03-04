package reporter

import (
	"encoding/json"
	"io"

	"github.com/opd-ai/go-stats-generator/internal/metrics"
)

// JSONReporter generates JSON formatted output for analysis reports and diffs.
// JSONReporter implements the Reporter interface and produces properly indented JSON.
type JSONReporter struct {
	indent bool
}

// NewJSONReporter creates a new JSON reporter with pretty-printing enabled by default.
// The reporter generates machine-readable JSON output suitable for CI/CD integration, automated
// analysis, and programmatic consumption. JSON format enables easy parsing, filtering with jq,
// and integration with monitoring/visualization tools. Use --format json flag to activate.
func NewJSONReporter() *JSONReporter {
	return &JSONReporter{
		indent: true,
	}
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

// NewHTMLReporter creates a new HTML reporter
func NewHTMLReporter() Reporter {
	return NewHTMLReporterWithConfig(nil)
}
