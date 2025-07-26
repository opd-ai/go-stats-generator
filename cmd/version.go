package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of go-stats-generator",
	Long:  `Print the version number and build information for go-stats-generator.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("go-stats-generator v1.0.0")
		fmt.Println("Go Source Code Statistics Generator")
		fmt.Println("Copyright (c) 2025")
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
