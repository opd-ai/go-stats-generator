package storage

import (
	"testing"

	"github.com/opd-ai/go-stats-generator/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestNew_ReturnsMemoryBackend(t *testing.T) {
	cfg := config.DefaultConfig()

	store := New(cfg)

	assert.NotNil(t, store)
	_, ok := store.(*Memory)
	assert.True(t, ok, "expected Memory implementation")
}

func TestNew_ImplementsInterface(t *testing.T) {
	cfg := config.DefaultConfig()

	var _ ResultStore = New(cfg)
}
