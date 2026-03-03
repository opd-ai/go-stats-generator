package multirepo

import (
	"fmt"

	"github.com/opd-ai/go-stats-generator/internal/metrics"
)

// Analyzer orchestrates analysis across multiple repositories
type Analyzer struct {
	config *Config
}

// NewAnalyzer creates a new multi-repository analyzer
func NewAnalyzer(cfg *Config) *Analyzer {
	return &Analyzer{config: cfg}
}

// Analyze runs analysis on all configured repositories
func (a *Analyzer) Analyze() (*MultiRepoReport, error) {
	if a.config == nil {
		return nil, fmt.Errorf("config is nil")
	}

	report := &MultiRepoReport{
		Repositories: make([]RepoResult, 0, len(a.config.Repositories)),
	}

	return report, nil
}

// MultiRepoReport aggregates results from multiple repositories
type MultiRepoReport struct {
	Repositories []RepoResult `json:"repositories"`
}

// RepoResult holds analysis results for a single repository
type RepoResult struct {
	Name   string          `json:"name"`
	Path   string          `json:"path"`
	Report *metrics.Report `json:"report"`
	Error  string          `json:"error,omitempty"`
}
