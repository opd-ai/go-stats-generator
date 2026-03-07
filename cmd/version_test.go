package cmd

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVersionCommand(t *testing.T) {
	t.Run("version command", func(t *testing.T) {
		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		rootCmd.SetArgs([]string{"version"})
		err := rootCmd.Execute()

		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		buf.ReadFrom(r)
		output := buf.String()

		assert.NoError(t, err)
		assert.Contains(t, output, "go-stats-generator v1.0.0")
		assert.Contains(t, output, "Go Source Code Statistics Generator")
		assert.Contains(t, output, "Copyright (c) 2025")
	})
}

func TestVersionCommandFlagParsing(t *testing.T) {
	cmd := versionCmd
	assert.NotNil(t, cmd)
	assert.Equal(t, "version", cmd.Use)
	assert.Equal(t, "Print the version number of go-stats-generator", cmd.Short)
}
