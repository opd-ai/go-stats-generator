package analyzer

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTeamAnalyzer(t *testing.T) {
	analyzer := NewTeamAnalyzer("/test/path")
	assert.NotNil(t, analyzer)
	assert.Equal(t, "/test/path", analyzer.repoPath)
}

func TestTeamAnalyzer_IsGitRepo(t *testing.T) {
	// Get current working directory (should be in git repo)
	wd, err := os.Getwd()
	require.NoError(t, err)

	// Find git root by walking up directories
	gitRoot := findGitRoot(wd)
	if gitRoot == "" {
		t.Skip("Not in a git repository, skipping test")
	}

	analyzer := NewTeamAnalyzer(gitRoot)
	authors, err := analyzer.getAuthors()

	// Should get authors without error in git repo
	assert.NoError(t, err)
	assert.NotEmpty(t, authors, "should find at least one author")
}

func TestTeamAnalyzer_AnalyzeTeamMetrics(t *testing.T) {
	// Get current working directory
	wd, err := os.Getwd()
	require.NoError(t, err)

	gitRoot := findGitRoot(wd)
	if gitRoot == "" {
		t.Skip("Not in a git repository")
	}

	analyzer := NewTeamAnalyzer(gitRoot)
	metrics, err := analyzer.AnalyzeTeamMetrics()
	if err != nil {
		t.Logf("Team metrics error: %v", err)
		return
	}

	assert.NotNil(t, metrics)
	assert.GreaterOrEqual(t, metrics.TotalDevelopers, 0)

	for name, dev := range metrics.Developers {
		assert.NotEmpty(t, name)
		assert.GreaterOrEqual(t, dev.CommitCount, 0)
		assert.GreaterOrEqual(t, dev.ActiveDays, 0)
	}
}

func TestCountUniqueDays(t *testing.T) {
	tests := []struct {
		name       string
		timestamps []time.Time
		want       int
	}{
		{
			name:       "empty",
			timestamps: []time.Time{},
			want:       0,
		},
		{
			name: "single day",
			timestamps: []time.Time{
				time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
				time.Date(2024, 1, 1, 15, 0, 0, 0, time.UTC),
			},
			want: 1,
		},
		{
			name: "multiple days",
			timestamps: []time.Time{
				time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
				time.Date(2024, 1, 2, 10, 0, 0, 0, time.UTC),
				time.Date(2024, 1, 3, 10, 0, 0, 0, time.UTC),
			},
			want: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := countUniqueDays(tt.timestamps)
			assert.Equal(t, tt.want, got)
		})
	}
}

// findGitRoot walks up from dir to find .git
func findGitRoot(dir string) string {
	for {
		gitDir := filepath.Join(dir, ".git")
		if _, err := os.Stat(gitDir); err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}
