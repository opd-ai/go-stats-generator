package multirepo

import (
	"os"
	"path/filepath"
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

func TestAnalyze_ValidRepository(t *testing.T) {
	testDir := filepath.Join("..", "..", "testdata", "simple")
	if _, err := os.Stat(testDir); os.IsNotExist(err) {
		t.Skip("testdata/simple directory not found")
	}

	cfg := &Config{
		Repositories: []RepositoryConfig{
			{Name: "simple", Path: testDir},
		},
	}

	analyzer := NewAnalyzer(cfg)
	report, err := analyzer.Analyze()

	require.NoError(t, err)
	require.NotNil(t, report)
	require.Len(t, report.Repositories, 1)

	repoResult := report.Repositories[0]
	assert.Equal(t, "simple", repoResult.Name)
	assert.Equal(t, testDir, repoResult.Path)
	assert.Empty(t, repoResult.Error)
	assert.NotNil(t, repoResult.Report)
	assert.Greater(t, repoResult.Report.Overview.TotalFunctions, 0)
}

func TestAnalyze_MultipleRepositories(t *testing.T) {
	testDir1 := filepath.Join("..", "..", "testdata", "simple")
	testDir2 := filepath.Join("..", "..", "testdata", "naming")

	if _, err := os.Stat(testDir1); os.IsNotExist(err) {
		t.Skip("testdata/simple directory not found")
	}
	if _, err := os.Stat(testDir2); os.IsNotExist(err) {
		t.Skip("testdata/naming directory not found")
	}

	cfg := &Config{
		Repositories: []RepositoryConfig{
			{Name: "simple", Path: testDir1},
			{Name: "naming", Path: testDir2},
		},
	}

	analyzer := NewAnalyzer(cfg)
	report, err := analyzer.Analyze()

	require.NoError(t, err)
	require.NotNil(t, report)
	require.Len(t, report.Repositories, 2)

	assert.Equal(t, "simple", report.Repositories[0].Name)
	assert.Equal(t, "naming", report.Repositories[1].Name)
	assert.Empty(t, report.Repositories[0].Error)
	assert.Empty(t, report.Repositories[1].Error)
	assert.NotNil(t, report.Repositories[0].Report)
	assert.NotNil(t, report.Repositories[1].Report)
}

func TestAnalyze_NonExistentRepository(t *testing.T) {
	cfg := &Config{
		Repositories: []RepositoryConfig{
			{Name: "invalid", Path: "/nonexistent/path"},
		},
	}

	analyzer := NewAnalyzer(cfg)
	report, err := analyzer.Analyze()

	require.NoError(t, err)
	require.NotNil(t, report)
	require.Len(t, report.Repositories, 1)

	repoResult := report.Repositories[0]
	assert.Equal(t, "invalid", repoResult.Name)
	assert.NotEmpty(t, repoResult.Error)
	assert.Nil(t, repoResult.Report)
}
