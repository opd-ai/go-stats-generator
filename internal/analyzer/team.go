package analyzer

import (
	"bufio"
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/opd-ai/go-stats-generator/internal/metrics"
)

// TeamAnalyzer analyzes Git history for team metrics
type TeamAnalyzer struct {
	repoPath string
}

// NewTeamAnalyzer creates a team analyzer instance
func NewTeamAnalyzer(repoPath string) *TeamAnalyzer {
	return &TeamAnalyzer{
		repoPath: repoPath,
	}
}

// AnalyzeTeamMetrics gathers developer productivity
func (a *TeamAnalyzer) AnalyzeTeamMetrics() (*metrics.TeamMetrics, error) {
	authors, err := a.getAuthors()
	if err != nil {
		return nil, fmt.Errorf("get authors: %w", err)
	}

	devMetrics := make(map[string]*metrics.DeveloperMetrics)
	for _, author := range authors {
		m, err := a.analyzeAuthor(author)
		if err != nil {
			continue
		}
		devMetrics[author] = m
	}

	return &metrics.TeamMetrics{
		Developers:      devMetrics,
		TotalDevelopers: len(devMetrics),
	}, nil
}

// getAuthors returns unique commit authors
func (a *TeamAnalyzer) getAuthors() ([]string, error) {
	cmd := exec.Command("git", "-C", a.repoPath,
		"log", "--format=%aN", "--all")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	seen := make(map[string]bool)
	var authors []string
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		author := strings.TrimSpace(scanner.Text())
		if author != "" && !seen[author] {
			seen[author] = true
			authors = append(authors, author)
		}
	}
	return authors, scanner.Err()
}

// analyzeAuthor gathers metrics for one developer
func (a *TeamAnalyzer) analyzeAuthor(author string) (*metrics.DeveloperMetrics, error) {
	stats, err := a.getAuthorStats(author)
	if err != nil {
		return nil, err
	}

	ownership, err := a.getFileOwnership(author)
	if err != nil {
		return nil, err
	}

	return &metrics.DeveloperMetrics{
		Name:            author,
		CommitCount:     stats.commits,
		LinesAdded:      stats.additions,
		LinesRemoved:    stats.deletions,
		FilesModified:   len(ownership),
		FirstCommitDate: stats.firstCommit,
		LastCommitDate:  stats.lastCommit,
		ActiveDays:      stats.activeDays,
	}, nil
}

type authorStats struct {
	commits     int
	additions   int
	deletions   int
	firstCommit time.Time
	lastCommit  time.Time
	activeDays  int
}

// getAuthorStats returns commit statistics
func (a *TeamAnalyzer) getAuthorStats(author string) (*authorStats, error) {
	out, err := a.fetchGitLogOutput(author)
	if err != nil {
		return nil, err
	}

	stats := &authorStats{}
	timestamps := a.parseGitLogOutput(out, stats)
	a.finalizeAuthorStats(stats, timestamps)

	return stats, nil
}

func (a *TeamAnalyzer) fetchGitLogOutput(author string) ([]byte, error) {
	cmd := exec.Command("git", "-C", a.repoPath, "log",
		"--author="+author, "--numstat", "--format=%at",
		"--all")
	return cmd.Output()
}

func (a *TeamAnalyzer) parseGitLogOutput(out []byte, stats *authorStats) []time.Time {
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	var timestamps []time.Time

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		if ts := a.tryParseTimestamp(line); ts != nil {
			timestamps = append(timestamps, *ts)
			stats.commits++
			continue
		}

		a.parseNumstatLine(line, stats)
	}

	return timestamps
}

func (a *TeamAnalyzer) tryParseTimestamp(line string) *time.Time {
	if ts, err := strconv.ParseInt(line, 10, 64); err == nil {
		t := time.Unix(ts, 0)
		return &t
	}
	return nil
}

func (a *TeamAnalyzer) parseNumstatLine(line string, stats *authorStats) {
	parts := strings.Fields(line)
	if len(parts) < 2 {
		return
	}

	if add, err := strconv.Atoi(parts[0]); err == nil {
		stats.additions += add
	}
	if del, err := strconv.Atoi(parts[1]); err == nil {
		stats.deletions += del
	}
}

func (a *TeamAnalyzer) finalizeAuthorStats(stats *authorStats, timestamps []time.Time) {
	if len(timestamps) == 0 {
		return
	}
	stats.lastCommit = timestamps[0]
	stats.firstCommit = timestamps[len(timestamps)-1]
	stats.activeDays = countUniqueDays(timestamps)
}

// getFileOwnership returns files primarily owned
func (a *TeamAnalyzer) getFileOwnership(author string) ([]string, error) {
	cmd := exec.Command("git", "-C", a.repoPath, "ls-files", "*.go")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var owned []string
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		file := strings.TrimSpace(scanner.Text())
		if file == "" {
			continue
		}

		filePath := filepath.Join(a.repoPath, file)
		if a.isPrimaryOwner(filePath, author) {
			owned = append(owned, file)
		}
	}

	return owned, scanner.Err()
}

// isPrimaryOwner checks if author owns most lines
func (a *TeamAnalyzer) isPrimaryOwner(file, author string) bool {
	cmd := exec.Command("git", "blame", "--line-porcelain", file)
	out, err := cmd.Output()
	if err != nil {
		return false
	}

	authorLines := 0
	totalLines := 0

	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "author ") {
			totalLines++
			if strings.TrimPrefix(line, "author ") == author {
				authorLines++
			}
		}
	}

	if totalLines == 0 {
		return false
	}
	return float64(authorLines)/float64(totalLines) > 0.5
}

// countUniqueDays returns unique calendar days
func countUniqueDays(timestamps []time.Time) int {
	days := make(map[string]bool)
	for _, t := range timestamps {
		day := t.Format("2006-01-02")
		days[day] = true
	}
	return len(days)
}
