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
// Execute is called by main.main() and only needs to happen once to the rootCmd.
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
	setupConfigFile()
	viper.AutomaticEnv()
	handleConfigLoad()
}

// setupConfigFile configures viper to use the correct config file.
func setupConfigFile() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
		return
	}

	setupDefaultConfigPaths()
}

// setupDefaultConfigPaths sets up default config search paths.
func setupDefaultConfigPaths() {
	home, err := os.UserHomeDir()
	cobra.CheckErr(err)

	viper.AddConfigPath(home)
	viper.AddConfigPath(".")
	viper.SetConfigType("yaml")
	viper.SetConfigName(".go-stats-generator")
}

// handleConfigLoad reads config and handles errors appropriately.
func handleConfigLoad() {
	if err := viper.ReadInConfig(); err == nil {
		if verbose {
			fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
		}
		return
	} else {
		handleConfigError(err)
	}
}

// handleConfigError processes config file load errors.
func handleConfigError(err error) {
	if cfgFile != "" {
		fmt.Fprintf(os.Stderr, "Warning: Failed to load config file '%s': %v\n", cfgFile, err)
	} else if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
		fmt.Fprintf(os.Stderr, "Warning: Config file found but failed to load: %v\n", err)
		fmt.Fprintln(os.Stderr, "Proceeding with default configuration and command-line flags.")
	}
}
