package cmd

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServeCommand(t *testing.T) {
	t.Run("flag parsing", func(t *testing.T) {
		cmd := serveCmd
		assert.NotNil(t, cmd)
		assert.Equal(t, "serve", cmd.Use)
		assert.Equal(t, "Start the REST API server", cmd.Short)

		portFlag := cmd.Flags().Lookup("port")
		assert.NotNil(t, portFlag)
		assert.Equal(t, "8080", portFlag.DefValue)
	})
}

func TestServeCommandHelp(t *testing.T) {
	rootCmd.SetArgs([]string{"serve", "--help"})
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)

	err := rootCmd.Execute()
	assert.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Start the go-stats-generator REST API server")
	assert.Contains(t, output, "--port")
}
