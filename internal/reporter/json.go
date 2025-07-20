package reporter

import (
	"encoding/json"
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

// NewCSVReporter creates a new CSV reporter (placeholder)
func NewCSVReporter() Reporter {
	return &JSONReporter{} // Placeholder
}

// NewHTMLReporter creates a new HTML reporter (placeholder)
func NewHTMLReporter() Reporter {
	return &JSONReporter{} // Placeholder
}

// NewMarkdownReporter creates a new Markdown reporter (placeholder)
func NewMarkdownReporter() Reporter {
	return &JSONReporter{} // Placeholder
}
