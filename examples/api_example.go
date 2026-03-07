package main

import (
	"context"
	"fmt"
	"os"

	"github.com/opd-ai/go-stats-generator/pkg/generator"
)

func main() {
	analyzer := generator.NewAnalyzer()

	report, err := analyzer.AnalyzeDirectory(context.Background(), "./src")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Analysis failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Found %d functions with average complexity %.1f\n",
		len(report.Functions), report.Complexity.AverageFunction)
}
