package analyzer

import (
	"go/parser"
	"go/token"
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

func TestNamingAnalyzer_CheckMixedCaps(t *testing.T) {
	tests := []struct {
		name          string
		identifier    string
		isTestFile    bool
		shouldViolate bool
	}{
		{
			name:          "valid MixedCaps",
			identifier:    "getUserID",
			isTestFile:    false,
			shouldViolate: false,
		},
		{
			name:          "valid with acronym",
			identifier:    "HTTPClient",
			isTestFile:    false,
			shouldViolate: false,
		},
		{
			name:          "invalid with underscore",
			identifier:    "get_user",
			isTestFile:    false,
			shouldViolate: true,
		},
		{
			name:          "invalid with multiple underscores",
			identifier:    "get_user_by_id",
			isTestFile:    false,
			shouldViolate: true,
		},
		{
			name:          "test function with underscore - allowed",
			identifier:    "Test_UserService",
			isTestFile:    true,
			shouldViolate: false,
		},
		{
			name:          "regular function in test file with underscore - not allowed",
			identifier:    "helper_function",
			isTestFile:    true,
			shouldViolate: true,
		},
	}

	na := NewNamingAnalyzer()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &identifierContext{
				isTestFile: tt.isTestFile,
			}
			violation := na.checkMixedCaps(tt.identifier, ctx)

			if tt.shouldViolate {
				require.NotNil(t, violation, "expected violation for %s", tt.identifier)
				assert.Equal(t, "underscore_in_name", violation.ViolationType)
				assert.NotEmpty(t, violation.SuggestedName)
			} else {
				assert.Nil(t, violation, "expected no violation for %s", tt.identifier)
			}
		})
	}
}

func TestNamingAnalyzer_CheckSingleLetterName(t *testing.T) {
	tests := []struct {
		name          string
		identifier    string
		idType        string
		inLoop        bool
		shouldViolate bool
	}{
		{
			name:          "loop variable i",
			identifier:    "i",
			idType:        "var",
			inLoop:        true,
			shouldViolate: false,
		},
		{
			name:          "single letter outside loop",
			identifier:    "x",
			idType:        "var",
			inLoop:        false,
			shouldViolate: true,
		},
		{
			name:          "receiver r for reader",
			identifier:    "r",
			idType:        "method",
			inLoop:        false,
			shouldViolate: false,
		},
		{
			name:          "receiver w for writer",
			identifier:    "w",
			idType:        "method",
			inLoop:        false,
			shouldViolate: false,
		},
		{
			name:          "descriptive name",
			identifier:    "user",
			idType:        "var",
			inLoop:        false,
			shouldViolate: false,
		},
	}

	na := NewNamingAnalyzer()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &identifierContext{
				inLoop:             tt.inLoop,
				validSingleLetters: make(map[string]bool),
			}
			if tt.inLoop && (tt.identifier == "i" || tt.identifier == "j" || tt.identifier == "k") {
				ctx.validSingleLetters[tt.identifier] = true
			}

			violation := na.checkSingleLetterName(tt.identifier, tt.idType, ctx)

			if tt.shouldViolate {
				require.NotNil(t, violation, "expected violation for %s", tt.identifier)
				assert.Equal(t, "single_letter_name", violation.ViolationType)
			} else {
				assert.Nil(t, violation, "expected no violation for %s", tt.identifier)
			}
		})
	}
}

func TestNamingAnalyzer_CheckAcronymCasing(t *testing.T) {
	tests := []struct {
		name          string
		identifier    string
		shouldViolate bool
		suggested     string
	}{
		{
			name:          "correct URL",
			identifier:    "URL",
			shouldViolate: false,
		},
		{
			name:          "correct URLParser",
			identifier:    "URLParser",
			shouldViolate: false,
		},
		{
			name:          "incorrect Url",
			identifier:    "Url",
			shouldViolate: true,
			suggested:     "URL",
		},
		{
			name:          "incorrect UrlParser",
			identifier:    "UrlParser",
			shouldViolate: true,
			suggested:     "URLParser",
		},
		{
			name:          "incorrect GetUrl",
			identifier:    "GetUrl",
			shouldViolate: true,
			suggested:     "GetURL",
		},
		{
			name:          "correct GetURL",
			identifier:    "GetURL",
			shouldViolate: false,
		},
		{
			name:          "incorrect UserId",
			identifier:    "UserId",
			shouldViolate: true,
			suggested:     "UserID",
		},
		{
			name:          "correct UserID",
			identifier:    "UserID",
			shouldViolate: false,
		},
		{
			name:          "incorrect HttpClient",
			identifier:    "HttpClient",
			shouldViolate: true,
			suggested:     "HTTPClient",
		},
		{
			name:          "correct HTTPClient",
			identifier:    "HTTPClient",
			shouldViolate: false,
		},
	}

	na := NewNamingAnalyzer()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &identifierContext{}
			violation := na.checkAcronymCasing(tt.identifier, ctx)

			if tt.shouldViolate {
				require.NotNil(t, violation, "expected violation for %s", tt.identifier)
				assert.Equal(t, "acronym_casing", violation.ViolationType)
				if tt.suggested != "" {
					assert.Equal(t, tt.suggested, violation.SuggestedName)
				}
			} else {
				assert.Nil(t, violation, "expected no violation for %s", tt.identifier)
			}
		})
	}
}

func TestNamingAnalyzer_CheckIdentifierStuttering(t *testing.T) {
	tests := []struct {
		name          string
		identifier    string
		receiverType  string
		packageName   string
		functionName  string
		shouldViolate bool
		suggested     string
	}{
		{
			name:          "method stuttering - UserName",
			identifier:    "UserName",
			receiverType:  "User",
			shouldViolate: true,
			suggested:     "Name",
		},
		{
			name:          "no stuttering - GetName",
			identifier:    "GetName",
			receiverType:  "User",
			shouldViolate: false,
		},
		{
			name:          "package stuttering - UserService (flagged)",
			identifier:    "UserService",
			packageName:   "user",
			shouldViolate: true,
			suggested:     "Service",
		},
		{
			name:          "no package stuttering - Service",
			identifier:    "Service",
			packageName:   "user",
			shouldViolate: false,
		},
		{
			name:          "receiver Get prefix - okay",
			identifier:    "GetUser",
			receiverType:  "User",
			shouldViolate: false,
		},
	}

	na := NewNamingAnalyzer()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &identifierContext{
				receiverType: tt.receiverType,
				packageName:  tt.packageName,
				functionName: tt.functionName,
			}
			violation := na.checkIdentifierStuttering(tt.identifier, ctx)

			if tt.shouldViolate {
				require.NotNil(t, violation, "expected violation for %s", tt.identifier)
				assert.Contains(t, violation.ViolationType, "stuttering")
				if tt.suggested != "" {
					assert.Equal(t, tt.suggested, violation.SuggestedName)
				}
			} else {
				assert.Nil(t, violation, "expected no violation for %s", tt.identifier)
			}
		})
	}
}

func TestNamingAnalyzer_ComputeIdentifierQualityScore(t *testing.T) {
	na := NewNamingAnalyzer()

	tests := []struct {
		name             string
		violations       int
		totalIdentifiers int
		expectedMinScore float64
		expectedMaxScore float64
	}{
		{
			name:             "perfect score",
			violations:       0,
			totalIdentifiers: 100,
			expectedMinScore: 1.0,
			expectedMaxScore: 1.0,
		},
		{
			name:             "some violations",
			violations:       5,
			totalIdentifiers: 100,
			expectedMinScore: 0.9,
			expectedMaxScore: 1.0,
		},
		{
			name:             "many violations",
			violations:       50,
			totalIdentifiers: 100,
			expectedMinScore: 0.5,
			expectedMaxScore: 1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var violations []metrics.IdentifierViolation
			for i := 0; i < tt.violations; i++ {
				violations = append(violations, metrics.IdentifierViolation{
					Name:     "test",
					Severity: "low",
				})
			}

			score := na.ComputeIdentifierQualityScore(violations, tt.totalIdentifiers)

			assert.GreaterOrEqual(t, score, 0.0)
			assert.LessOrEqual(t, score, 1.0)
			assert.GreaterOrEqual(t, score, tt.expectedMinScore)
			assert.LessOrEqual(t, score, tt.expectedMaxScore)
		})
	}
}

func TestNamingAnalyzer_AnalyzeIdentifiers(t *testing.T) {
	sourceCode := `package user

type UserId int  // Should flag: "Id" -> "ID"

type User struct {
UserName string  // Should flag: stuttering
Age      int
}

func (u User) GetUrl() string {  // Should flag: "Url" -> "URL"
return ""
}

func get_user() {  // Should flag: underscore
}

var x = 5  // Should flag: single letter outside loop
`

	// Parse the source code
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", sourceCode, parser.ParseComments)
	require.NoError(t, err)

	na := NewNamingAnalyzer()
	violations := na.AnalyzeIdentifiers(file, "test.go", fset)

	// Should find violations for:
	// 1. UserId (acronym)
	// 2. UserName (stuttering)
	// 3. GetUrl (acronym)
	// 4. get_user (underscore)
	// 5. x (single letter)
	assert.GreaterOrEqual(t, len(violations), 5, "should find at least 5 violations")

	// Check specific violations
	violationTypes := make(map[string]int)
	for _, v := range violations {
		violationTypes[v.ViolationType]++
	}

	assert.Greater(t, violationTypes["acronym_casing"], 0, "should find acronym violations")
	assert.Greater(t, violationTypes["underscore_in_name"], 0, "should find underscore violations")
	assert.Greater(t, violationTypes["single_letter_name"], 0, "should find single letter violations")
}
