package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// flagBinding represents a single flag-to-viper binding configuration.
type flagBinding struct {
	flagName string
	viperKey string
}

// bindFlags binds multiple flags to viper keys using a command and binding specifications.
func bindFlags(cmd *cobra.Command, bindings []flagBinding) {
	for _, b := range bindings {
		viper.BindPFlag(b.viperKey, cmd.Flags().Lookup(b.flagName))
	}
}
