package reporter

import (
	_ "embed"
	"fmt"
	"io"
	"strings"
	"text/template"
	"time"

	"github.com/opd-ai/go-stats-generator/internal/metrics"
)

//go:embed templates/markdown/report.md
var markdownTemplate string

//go:embed templates/markdown/diff.md
var markdownDiffTemplate string

// MarkdownReporter generates Markdown reports for Git workflows
type MarkdownReporter struct {
	includeOverview bool
	includeDetails  bool
	maxItems        int
}

// NewMarkdownReporter creates a new Markdown reporter with default settings
func NewMarkdownReporter() Reporter {
	return &MarkdownReporter{
		includeOverview: true,
		includeDetails:  true,
		maxItems:        50, // Limit to prevent extremely long reports
	}
}

// NewMarkdownReporterWithOptions creates a Markdown reporter with custom options
func NewMarkdownReporterWithOptions(includeOverview, includeDetails bool, maxItems int) *MarkdownReporter {
	return &MarkdownReporter{
		includeOverview: includeOverview,
		includeDetails:  includeDetails,
		maxItems:        maxItems,
	}
}

// Generate generates a comprehensive Markdown report
func (mr *MarkdownReporter) Generate(report *metrics.Report, output io.Writer) error {
	// Create template with helper functions
	tmpl, err := template.New("markdown-report").Funcs(template.FuncMap{
		"formatDuration": mr.formatDuration,
		"formatFloat":    mr.formatFloat,
		"formatPercent":  mr.formatPercent,
		"truncateList":   mr.truncateList,
		"escapeMarkdown": mr.escapeMarkdown,
	}).Parse(markdownTemplate)

	if err != nil {
		return fmt.Errorf("failed to parse embedded markdown template: %w", err)
	}

	// Execute template with report data
	return tmpl.Execute(output, map[string]interface{}{
		"Report":          report,
		"IncludeOverview": mr.includeOverview,
		"IncludeDetails":  mr.includeDetails,
		"MaxItems":        mr.maxItems,
	})
}

// WriteDiff generates a Markdown diff report comparing two snapshots
func (mr *MarkdownReporter) WriteDiff(output io.Writer, diff *metrics.ComplexityDiff) error {
	tmpl, err := template.New("markdown-diff").Funcs(template.FuncMap{
		"formatDuration":   mr.formatDuration,
		"formatFloat":      mr.formatFloat,
		"formatPercent":    mr.formatPercent,
		"formatChange":     mr.formatChange,
		"formatChangeSign": mr.formatChangeSign,
		"escapeMarkdown":   mr.escapeMarkdown,
	}).Parse(markdownDiffTemplate)

	if err != nil {
		return fmt.Errorf("failed to parse embedded markdown diff template: %w", err)
	}

	return tmpl.Execute(output, diff)
}

// Template helper functions

func (mr *MarkdownReporter) formatDuration(d time.Duration) string {
	if d < time.Millisecond {
		return fmt.Sprintf("%.2fμs", float64(d.Nanoseconds())/1000)
	}
	if d < time.Second {
		return fmt.Sprintf("%.2fms", float64(d.Nanoseconds())/1000000)
	}
	return fmt.Sprintf("%.2fs", d.Seconds())
}

func (mr *MarkdownReporter) formatFloat(f float64) string {
	if f == float64(int(f)) {
		return fmt.Sprintf("%.0f", f)
	}
	return fmt.Sprintf("%.2f", f)
}

func (mr *MarkdownReporter) formatPercent(f float64) string {
	return fmt.Sprintf("%.1f%%", f*100)
}

func (mr *MarkdownReporter) formatChange(oldVal, newVal float64) string {
	if oldVal == 0 {
		if newVal == 0 {
			return "no change"
		}
		return "new"
	}

	change := ((newVal - oldVal) / oldVal) * 100
	if change > 0 {
		return fmt.Sprintf("+%.1f%%", change)
	} else if change < 0 {
		return fmt.Sprintf("%.1f%%", change)
	}
	return "no change"
}

func (mr *MarkdownReporter) formatChangeSign(change float64) string {
	if change > 0 {
		return "📈"
	} else if change < 0 {
		return "📉"
	}
	return "➡️"
}

func (mr *MarkdownReporter) truncateList(items interface{}, limit int) interface{} {
	switch v := items.(type) {
	case []metrics.FunctionMetrics:
		if len(v) <= limit {
			return v
		}
		return v[:limit]
	case []metrics.StructMetrics:
		if len(v) <= limit {
			return v
		}
		return v[:limit]
	case []metrics.InterfaceMetrics:
		if len(v) <= limit {
			return v
		}
		return v[:limit]
	case []metrics.PackageMetrics:
		if len(v) <= limit {
			return v
		}
		return v[:limit]
	default:
		return items
	}
}

func (mr *MarkdownReporter) escapeMarkdown(s string) string {
	// Escape special Markdown characters that could break formatting
	replacer := strings.NewReplacer(
		"*", "\\*",
		"_", "\\_",
		"`", "\\`",
		"#", "\\#",
		"[", "\\[",
		"]", "\\]",
		"(", "\\(",
		")", "\\)",
		"|", "\\|",
	)
	return replacer.Replace(s)
}
