package go_stats_generator

import (
	"errors"
	"testing"
)

// TestErrorVariables_NotNil verifies that all error variables are properly initialized
func TestErrorVariables_NotNil(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{"ErrNoGoFiles", ErrNoGoFiles},
		{"ErrInvalidDirectory", ErrInvalidDirectory},
		{"ErrParsingFailed", ErrParsingFailed},
		{"ErrAnalysisFailed", ErrAnalysisFailed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err == nil {
				t.Errorf("%s should not be nil", tt.name)
			}
		})
	}
}

// TestErrorMessages_Content verifies that error messages are meaningful
func TestErrorMessages_Content(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		expectedString string
	}{
		{
			name:           "ErrNoGoFiles_Message",
			err:            ErrNoGoFiles,
			expectedString: "no Go files found",
		},
		{
			name:           "ErrInvalidDirectory_Message",
			err:            ErrInvalidDirectory,
			expectedString: "invalid directory",
		},
		{
			name:           "ErrParsingFailed_Message",
			err:            ErrParsingFailed,
			expectedString: "failed to parse Go file",
		},
		{
			name:           "ErrAnalysisFailed_Message",
			err:            ErrAnalysisFailed,
			expectedString: "analysis failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Error() != tt.expectedString {
				t.Errorf("Expected error message %q, got %q", tt.expectedString, tt.err.Error())
			}
		})
	}
}

// TestErrorVariables_Uniqueness verifies that each error variable is distinct
func TestErrorVariables_Uniqueness(t *testing.T) {
	errorVars := []error{
		ErrNoGoFiles,
		ErrInvalidDirectory,
		ErrParsingFailed,
		ErrAnalysisFailed,
	}

	// Check that no two errors are the same instance
	for i, err1 := range errorVars {
		for j, err2 := range errorVars {
			if i != j && err1 == err2 {
				t.Errorf("Error variables at index %d and %d are the same instance", i, j)
			}
		}
	}

	// Check that error messages are distinct
	messages := make(map[string]bool)
	for _, err := range errorVars {
		msg := err.Error()
		if messages[msg] {
			t.Errorf("Duplicate error message found: %q", msg)
		}
		messages[msg] = true
	}
}

// TestErrorVariables_TypeAssertion verifies errors implement error interface
func TestErrorVariables_TypeAssertion(t *testing.T) {
	tests := []struct {
		name string
		err  interface{}
	}{
		{"ErrNoGoFiles", ErrNoGoFiles},
		{"ErrInvalidDirectory", ErrInvalidDirectory},
		{"ErrParsingFailed", ErrParsingFailed},
		{"ErrAnalysisFailed", ErrAnalysisFailed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, ok := tt.err.(error); !ok {
				t.Errorf("%s does not implement error interface", tt.name)
			}
		})
	}
}

// TestErrorVariables_IsComparison verifies error comparison with errors.Is
func TestErrorVariables_IsComparison(t *testing.T) {
	tests := []struct {
		name     string
		target   error
		err      error
		expected bool
	}{
		{
			name:     "ErrNoGoFiles_SelfComparison",
			target:   ErrNoGoFiles,
			err:      ErrNoGoFiles,
			expected: true,
		},
		{
			name:     "ErrNoGoFiles_DifferentError",
			target:   ErrNoGoFiles,
			err:      ErrInvalidDirectory,
			expected: false,
		},
		{
			name:     "ErrInvalidDirectory_SelfComparison",
			target:   ErrInvalidDirectory,
			err:      ErrInvalidDirectory,
			expected: true,
		},
		{
			name:     "ErrParsingFailed_SelfComparison",
			target:   ErrParsingFailed,
			err:      ErrParsingFailed,
			expected: true,
		},
		{
			name:     "ErrAnalysisFailed_SelfComparison",
			target:   ErrAnalysisFailed,
			err:      ErrAnalysisFailed,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if errors.Is(tt.err, tt.target) != tt.expected {
				t.Errorf("errors.Is(%v, %v) = %v, want %v", tt.err, tt.target, !tt.expected, tt.expected)
			}
		})
	}
}

// TestErrorVariables_Wrapping verifies error wrapping behavior
func TestErrorVariables_Wrapping(t *testing.T) {
	tests := []struct {
		name      string
		baseError error
		wrapMsg   string
	}{
		{
			name:      "WrapErrNoGoFiles",
			baseError: ErrNoGoFiles,
			wrapMsg:   "failed to scan directory",
		},
		{
			name:      "WrapErrInvalidDirectory",
			baseError: ErrInvalidDirectory,
			wrapMsg:   "cannot access path",
		},
		{
			name:      "WrapErrParsingFailed",
			baseError: ErrParsingFailed,
			wrapMsg:   "syntax error detected",
		},
		{
			name:      "WrapErrAnalysisFailed",
			baseError: ErrAnalysisFailed,
			wrapMsg:   "analysis pipeline error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a wrapped error
			wrappedErr := errors.Join(errors.New(tt.wrapMsg), tt.baseError)

			// Verify the base error can be unwrapped
			if !errors.Is(wrappedErr, tt.baseError) {
				t.Errorf("Wrapped error should contain base error %v", tt.baseError)
			}

			// Verify error message contains both parts
			errMsg := wrappedErr.Error()
			if !containsSubstring(errMsg, tt.wrapMsg) {
				t.Errorf("Wrapped error message should contain %q, got %q", tt.wrapMsg, errMsg)
			}
		})
	}
}

// TestErrorVariables_StringRepresentation verifies string representation consistency
func TestErrorVariables_StringRepresentation(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		shouldNotBe string
	}{
		{
			name:        "ErrNoGoFiles_NotEmpty",
			err:         ErrNoGoFiles,
			shouldNotBe: "",
		},
		{
			name:        "ErrInvalidDirectory_NotEmpty",
			err:         ErrInvalidDirectory,
			shouldNotBe: "",
		},
		{
			name:        "ErrParsingFailed_NotEmpty",
			err:         ErrParsingFailed,
			shouldNotBe: "",
		},
		{
			name:        "ErrAnalysisFailed_NotEmpty",
			err:         ErrAnalysisFailed,
			shouldNotBe: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errStr := tt.err.Error()
			if errStr == tt.shouldNotBe {
				t.Errorf("Error string should not be empty for %s", tt.name)
			}

			// Verify string is consistent across multiple calls
			errStr2 := tt.err.Error()
			if errStr != errStr2 {
				t.Errorf("Error string should be consistent: got %q and %q", errStr, errStr2)
			}
		})
	}
}

// Helper function to check if a string contains a substring
func containsSubstring(s, substr string) bool {
	return len(substr) <= len(s) && (substr == "" || indexOfString(s, substr) >= 0)
}

// Helper function to find index of substring (simplified implementation)
func indexOfString(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
