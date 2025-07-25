package reporter

import (
	"encoding/json"
	"fmt"
	"io"

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

type CSVReporter struct{}

func (r *CSVReporter) Generate(report *metrics.Report, output io.Writer) error {
	// Placeholder implementation
	return fmt.Errorf("CSV reporter not yet implemented")
}

func (r *CSVReporter) WriteDiff(output io.Writer, diff *metrics.ComplexityDiff) error {
	// Placeholder implementation
	return fmt.Errorf("CSV diff reporter not yet implemented")
}
