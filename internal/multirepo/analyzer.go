package multirepo

import (
	"context"
	"fmt"

	"github.com/opd-ai/go-stats-generator/internal/metrics"
	"github.com/opd-ai/go-stats-generator/pkg/generator"
)

// Analyzer orchestrates analysis across multiple repositories
type Analyzer struct {
	config *Config
}

// NewAnalyzer creates a new multi-repository analyzer for coordinated analysis across multiple Git repositories.
// Enables cross-project comparisons, aggregated metrics, and organization-wide trend tracking for microservices or monorepos.
// Each repository is analyzed independently, then combined for comparative reporting and benchmarking.
func NewAnalyzer(cfg *Config) *Analyzer {
	return &Analyzer{config: cfg}
}

// Analyze runs analysis on all configured repositories
func (a *Analyzer) Analyze() (*Report, error) {
	if a.config == nil {
		return nil, fmt.Errorf("config is nil")
	}

	report := &Report{
		Repositories: make([]RepoResult, 0, len(a.config.Repositories)),
	}

	ctx := context.Background()
	analyzer := generator.NewAnalyzer()

	for _, repo := range a.config.Repositories {
		repoResult := RepoResult{
			Name: repo.Name,
			Path: repo.Path,
		}

		repoReport, err := analyzer.AnalyzeDirectory(ctx, repo.Path)
		if err != nil {
			repoResult.Error = err.Error()
		} else {
			repoResult.Report = repoReport
		}

		report.Repositories = append(report.Repositories, repoResult)
	}

	return report, nil
}

// Report aggregates results from multiple repositories
type Report struct {
	Repositories []RepoResult `json:"repositories"`
}

// RepoResult holds analysis results for a single repository
type RepoResult struct {
	Name   string          `json:"name"`
	Path   string          `json:"path"`
	Report *metrics.Report `json:"report"`
	Error  string          `json:"error,omitempty"`
}
