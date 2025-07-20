package reporter

import (
	"io"

	"github.com/opd-ai/go-stats-generator/internal/metrics"
)

// Reporter interface defines the contract for generating reports
type Reporter interface {
	Generate(report *metrics.Report, output io.Writer) error
}

// ReporterType represents the type of reporter
type ReporterType string

const (
	TypeConsole  ReporterType = "console"
	TypeJSON     ReporterType = "json"
	TypeCSV      ReporterType = "csv"
	TypeHTML     ReporterType = "html"
	TypeMarkdown ReporterType = "markdown"
)

// CreateReporter creates a new reporter of the specified type
func CreateReporter(reporterType ReporterType, options interface{}) Reporter {
	switch reporterType {
	case TypeJSON:
		return NewJSONReporter()
	case TypeCSV:
		return NewCSVReporter()
	case TypeHTML:
		return NewHTMLReporter()
	case TypeMarkdown:
		return NewMarkdownReporter()
	case TypeConsole:
		fallthrough
	default:
		return NewConsoleReporter(nil)
	}
}
