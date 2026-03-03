// Package main provides the entry point for the go-stats-generator CLI application.
//
// go-stats-generator is a high-performance command-line tool that analyzes Go source
// code repositories to generate comprehensive statistical reports about code structure,
// complexity, and patterns.
package main

import "github.com/opd-ai/go-stats-generator/cmd"

// main is the entry point for the go-stats-generator CLI application.
func main() {
	cmd.Execute()
}
