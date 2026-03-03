// Package go_stats_generator provides a programmatic API for analyzing Go source code.
//
// The go_stats_generator package exposes the core analysis capabilities of the
// go-stats-generator tool as a library. It allows programmatic analysis of Go
// codebases, producing structured metrics reports about functions, structs,
// interfaces, complexity, and documentation coverage.
//
// Basic usage:
//
//	analyzer := go_stats_generator.NewAnalyzer()
//	report, err := analyzer.AnalyzeDirectory(ctx, "/path/to/code")
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Printf("Found %d functions\n", len(report.Functions))
//
// The package re-exports commonly used types from the internal metrics package
// for convenience.
package go_stats_generator
