package cmd

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/opd-ai/go-stats-generator/internal/config"
	"github.com/opd-ai/go-stats-generator/pkg/generator"
)

// BenchmarkFullAnalysis_CurrentCodebase benchmarks the analyze command on the current codebase
func BenchmarkFullAnalysis_CurrentCodebase(b *testing.B) {
	// Get current directory (go-stats-generator root)
	cwd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	rootDir := filepath.Dir(cwd)

	// Count Go files to report
	var fileCount int
	filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() && filepath.Ext(path) == ".go" {
			fileCount++
		}
		return nil
	})

	b.Logf("Analyzing %d Go files in %s", fileCount, rootDir)

	cfg := config.DefaultConfig()
	cfg.Filters.SkipTestFiles = true

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		analyzer := generator.NewAnalyzerWithConfig(cfg)
		_, err := analyzer.AnalyzeDirectory(context.Background(), rootDir)
		if err != nil {
			b.Fatalf("Analysis failed: %v", err)
		}
	}
}

// BenchmarkFullAnalysis_WithMemoryTracking benchmarks with memory usage tracking
func BenchmarkFullAnalysis_WithMemoryTracking(b *testing.B) {
	// Suppress output during benchmark
	oldStdout := os.Stdout
	os.Stdout, _ = os.Open(os.DevNull)
	defer func() { os.Stdout = oldStdout }()

	cwd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	rootDir := filepath.Dir(cwd)

	var fileCount int
	filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() && filepath.Ext(path) == ".go" {
			fileCount++
		}
		return nil
	})

	cfg := config.DefaultConfig()
	cfg.Filters.SkipTestFiles = true

	var maxMemory uint64
	var totalDuration time.Duration

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runtime.GC()
		var m1 runtime.MemStats
		runtime.ReadMemStats(&m1)

		start := time.Now()

		analyzer := generator.NewAnalyzerWithConfig(cfg)
		_, err := analyzer.AnalyzeDirectory(context.Background(), rootDir)
		if err != nil {
			b.Fatalf("Analysis failed: %v", err)
		}

		duration := time.Since(start)
		totalDuration += duration

		var m2 runtime.MemStats
		runtime.ReadMemStats(&m2)
		memUsed := m2.Alloc - m1.Alloc
		if memUsed > maxMemory {
			maxMemory = memUsed
		}
	}

	avgDuration := totalDuration / time.Duration(b.N)
	b.ReportMetric(float64(fileCount), "files")
	b.ReportMetric(float64(maxMemory)/(1024*1024), "MB_peak")
	b.ReportMetric(avgDuration.Seconds(), "sec/op")

	filesPerSec := float64(fileCount) / avgDuration.Seconds()
	b.Logf("Performance: %.0f files/sec, Peak memory: %.1f MB", filesPerSec, float64(maxMemory)/(1024*1024))

	// Restore stdout for final report
	os.Stdout = oldStdout
	_, _ = io.WriteString(oldStdout, "\n")
}
