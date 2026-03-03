package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	verbose bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "go-stats-generator",
	Short: "Go Source Code Statistics Generator",
	Long: `go-stats-generator is a comprehensive command-line tool that analyzes Go source code repositories 
and generates detailed statistical reports about code structure, complexity, and patterns.

The tool provides actionable insights for code quality assessment and refactoring decisions,
with a focus on computing advanced metrics that standard linters don't typically capture.

Features:
  • Function and method length analysis with precise line counting
  • Object complexity metrics with detailed member categorization  
  • Advanced AST pattern detection (design patterns, anti-patterns)
  • Concurrent file processing with configurable worker pools
  • Multiple output formats (console, JSON, CSV, HTML, Markdown)
  • Historical comparison and trend analysis
  • Performance optimized for enterprise-scale codebases

Performance:
  • Process 50,000+ files within 60 seconds
  • Memory usage under 1GB for large repositories
  • Configurable concurrency (default: number of CPU cores)`,

	Version: "1.0.0",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// It is called by main.main() and only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

// init initializes the root command with global flags and configuration bindings.
func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.go-stats-generator.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")

	// Bind flags to viper
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".go-stats-generator" (without extension).
		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName(".go-stats-generator")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		if verbose {
			fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
		}
	} else {
		// Check if a config file was explicitly provided or exists but failed to load
		if cfgFile != "" {
			// Explicit config file provided but failed to load
			fmt.Fprintf(os.Stderr, "Warning: Failed to load config file '%s': %v\n", cfgFile, err)
		} else if _, configNotFoundErr := err.(viper.ConfigFileNotFoundError); !configNotFoundErr {
			// Config file exists but has errors (YAML syntax, permissions, etc.)
			fmt.Fprintf(os.Stderr, "Warning: Config file found but failed to load: %v\n", err)
			fmt.Fprintln(os.Stderr, "Proceeding with default configuration and command-line flags.")
		}
		// If config file simply doesn't exist, proceed silently with defaults
	}
}
