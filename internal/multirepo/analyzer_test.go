package multirepo

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAnalyzer(t *testing.T) {
	cfg := &Config{
		Repositories: []RepositoryConfig{
			{Name: "test", Path: "/test"},
		},
	}

	analyzer := NewAnalyzer(cfg)
	require.NotNil(t, analyzer)
	assert.Equal(t, cfg, analyzer.config)
}

func TestAnalyze_NilConfig(t *testing.T) {
	analyzer := &Analyzer{config: nil}
	_, err := analyzer.Analyze()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "nil")
}

func TestAnalyze_EmptyRepos(t *testing.T) {
	cfg := &Config{Repositories: []RepositoryConfig{}}
	analyzer := NewAnalyzer(cfg)

	report, err := analyzer.Analyze()
	require.NoError(t, err)
	assert.NotNil(t, report)
	assert.Empty(t, report.Repositories)
}
