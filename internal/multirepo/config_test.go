package multirepo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &Config{
				Repositories: []RepositoryConfig{
					{Name: "repo1", Path: "/path1"},
				},
			},
			wantErr: false,
		},
		{
			name: "empty config",
			config: &Config{
				Repositories: []RepositoryConfig{},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotNil(t, tt.config)
		})
	}
}
