package api

import (
	"context"
	"os"

	"github.com/opd-ai/go-stats-generator/internal/config"
	"github.com/opd-ai/go-stats-generator/internal/metrics"
	"github.com/opd-ai/go-stats-generator/pkg/generator"
)

// executeAnalysisWithPath runs the analysis on the specified path.
func executeAnalysisWithPath(cfg *config.Config, path string) (*metrics.Report, error) {
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Performance.Timeout)
	defer cancel()

	analyzer := generator.NewAnalyzerWithConfig(cfg)

	fileInfo, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	var report *metrics.Report
	if fileInfo.IsDir() {
		report, err = analyzer.AnalyzeDirectory(ctx, path)
	} else {
		report, err = analyzer.AnalyzeFile(ctx, path)
	}

	return report, err
}
