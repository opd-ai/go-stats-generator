package reporter

import (
	"encoding/json"
	"io"

	"github.com/opd-ai/go-stats-generator/internal/metrics"
)

// JSONReporter generates JSON formatted output for analysis reports and diffs.
// It implements the Reporter interface and produces properly indented JSON.
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

// NewHTMLReporter creates a new HTML reporter
func NewHTMLReporter() Reporter {
	return NewHTMLReporterWithConfig(nil)
}
