package analyzer

import (
	"testing"

	"github.com/opd-ai/go-stats-generator/internal/metrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNamingAnalyzer_CheckSnakeCase(t *testing.T) {
	tests := []struct {
		name          string
		fileName      string
		shouldViolate bool
		suggested     string
	}{
		{
			name:          "valid snake_case",
			fileName:      "user_service.go",
			shouldViolate: false,
		},
		{
			name:          "valid simple name",
			fileName:      "user.go",
			shouldViolate: false,
		},
		{
			name:          "valid test file",
			fileName:      "user_service_test.go",
			shouldViolate: false,
		},
		{
			name:          "invalid CamelCase",
			fileName:      "UserService.go",
			shouldViolate: true,
			suggested:     "user_service.go",
		},
		{
			name:          "invalid mixedCase",
			fileName:      "userService.go",
			shouldViolate: true,
			suggested:     "user_service.go",
		},
		{
			name:          "invalid with dash",
			fileName:      "user-service.go",
			shouldViolate: true,
		},
		{
			name:          "valid with numbers",
			fileName:      "http2_client.go",
			shouldViolate: false,
		},
	}

	na := NewNamingAnalyzer()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			violation := na.checkSnakeCase("/path/to/"+tt.fileName, tt.fileName)

			if tt.shouldViolate {
				require.NotNil(t, violation, "expected violation for %s", tt.fileName)
				assert.Equal(t, "non_snake_case", violation.ViolationType)
				if tt.suggested != "" {
					assert.Equal(t, tt.suggested, violation.SuggestedName)
				}
			} else {
				assert.Nil(t, violation, "expected no violation for %s", tt.fileName)
			}
		})
	}
}

func TestNamingAnalyzer_CheckStuttering(t *testing.T) {
	tests := []struct {
		name          string
		filePath      string
		dirName       string
		fileName      string
		shouldViolate bool
	}{
		{
			name:          "stuttering with prefix",
			filePath:      "http/http_client.go",
			dirName:       "http",
			fileName:      "http_client.go",
			shouldViolate: true,
		},
		{
			name:          "exact match stuttering",
			filePath:      "user/user.go",
			dirName:       "user",
			fileName:      "user.go",
			shouldViolate: true,
		},
		{
			name:          "no stuttering",
			filePath:      "http/client.go",
			dirName:       "http",
			fileName:      "client.go",
			shouldViolate: false,
		},
		{
			name:          "different name",
			filePath:      "http/server.go",
			dirName:       "http",
			fileName:      "server.go",
			shouldViolate: false,
		},
		{
			name:          "test file stuttering",
			filePath:      "user/user_test.go",
			dirName:       "user",
			fileName:      "user_test.go",
			shouldViolate: true,
		},
	}

	na := NewNamingAnalyzer()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			violation := na.checkStuttering(tt.filePath, tt.fileName, tt.dirName)

			if tt.shouldViolate {
				require.NotNil(t, violation, "expected stuttering violation")
				assert.Equal(t, "stuttering", violation.ViolationType)
				assert.Equal(t, "low", violation.Severity)
			} else {
				assert.Nil(t, violation, "expected no stuttering violation")
			}
		})
	}
}

func TestNamingAnalyzer_CheckGenericName(t *testing.T) {
	tests := []struct {
		name          string
		fileName      string
		shouldViolate bool
	}{
		{
			name:          "generic utils",
			fileName:      "utils.go",
			shouldViolate: true,
		},
		{
			name:          "generic helpers",
			fileName:      "helpers.go",
			shouldViolate: true,
		},
		{
			name:          "generic common",
			fileName:      "common.go",
			shouldViolate: true,
		},
		{
			name:          "generic misc",
			fileName:      "misc.go",
			shouldViolate: true,
		},
		{
			name:          "specific name",
			fileName:      "user_service.go",
			shouldViolate: false,
		},
		{
			name:          "specific utility",
			fileName:      "string_utils.go",
			shouldViolate: false,
		},
	}

	na := NewNamingAnalyzer()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			violation := na.checkGenericName("/path/to/"+tt.fileName, tt.fileName)

			if tt.shouldViolate {
				require.NotNil(t, violation, "expected generic name violation")
				assert.Equal(t, "generic_name", violation.ViolationType)
				assert.Equal(t, "low", violation.Severity)
			} else {
				assert.Nil(t, violation, "expected no generic name violation")
			}
		})
	}
}

func TestNamingAnalyzer_CheckTestSuffix(t *testing.T) {
	tests := []struct {
		name          string
		fileName      string
		shouldViolate bool
	}{
		{
			name:          "proper test file",
			fileName:      "user_test.go",
			shouldViolate: false,
		},
		{
			name:          "improper test prefix",
			fileName:      "test_user.go",
			shouldViolate: true,
		},
		{
			name:          "improper test name",
			fileName:      "testuser.go",
			shouldViolate: true,
		},
		{
			name:          "normal file",
			fileName:      "user.go",
			shouldViolate: false,
		},
	}

	na := NewNamingAnalyzer()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			violation := na.checkTestSuffix("/path/to/"+tt.fileName, tt.fileName)

			if tt.shouldViolate {
				require.NotNil(t, violation, "expected test suffix violation")
				assert.Equal(t, "improper_test_name", violation.ViolationType)
			} else {
				assert.Nil(t, violation, "expected no test suffix violation")
			}
		})
	}
}

func TestNamingAnalyzer_ToSnakeCase(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "CamelCase to snake_case",
			input:    "UserService.go",
			expected: "user_service.go",
		},
		{
			name:     "mixedCase to snake_case",
			input:    "userService.go",
			expected: "user_service.go",
		},
		{
			name:     "already snake_case",
			input:    "user_service.go",
			expected: "user_service.go",
		},
		{
			name:     "with test suffix",
			input:    "UserService_test.go",
			expected: "user_service_test.go",
		},
		{
			name:     "single word",
			input:    "User.go",
			expected: "user.go",
		},
		{
			name:     "acronym",
			input:    "HTTPClient.go",
			expected: "httpclient.go",
		},
	}

	na := NewNamingAnalyzer()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := na.toSnakeCase(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNamingAnalyzer_AnalyzeFileNames(t *testing.T) {
	tests := []struct {
		name               string
		filePaths          []string
		expectedViolations int
	}{
		{
			name: "all valid files",
			filePaths: []string{
				"pkg/user/service.go",
				"pkg/user/repository.go",
				"pkg/user/service_test.go",
			},
			expectedViolations: 0,
		},
		{
			name: "mixed violations",
			filePaths: []string{
				"pkg/user/UserService.go", // CamelCase
				"pkg/http/http_client.go", // stuttering
				"pkg/common/utils.go",     // generic
				"pkg/auth/test_auth.go",   // improper test
			},
			expectedViolations: 4,
		},
		{
			name: "ignores non-go files",
			filePaths: []string{
				"README.md",
				"Makefile",
				"pkg/user/service.go",
			},
			expectedViolations: 0,
		},
	}

	na := NewNamingAnalyzer()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			violations := na.AnalyzeFileNames(tt.filePaths)
			assert.Len(t, violations, tt.expectedViolations)
		})
	}
}

func TestNamingAnalyzer_ComputeFileNamingScore(t *testing.T) {
	na := NewNamingAnalyzer()

	tests := []struct {
		name       string
		violations int
		totalFiles int
		minScore   float64
		maxScore   float64
	}{
		{
			name:       "perfect score",
			violations: 0,
			totalFiles: 10,
			minScore:   1.0,
			maxScore:   1.0,
		},
		{
			name:       "some violations",
			violations: 2,
			totalFiles: 10,
			minScore:   0.8,
			maxScore:   1.0,
		},
		{
			name:       "many violations",
			violations: 8,
			totalFiles: 10,
			minScore:   0.0,
			maxScore:   0.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create dummy violations using the metrics package
			var violations []metrics.FileNameViolation
			for i := 0; i < tt.violations; i++ {
				violations = append(violations, metrics.FileNameViolation{
					File:     "file.go",
					Severity: "low",
				})
			}

			score := na.ComputeFileNamingScore(violations, tt.totalFiles)

			assert.GreaterOrEqual(t, score, 0.0)
			assert.LessOrEqual(t, score, 1.0)
		})
	}
}
