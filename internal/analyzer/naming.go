package analyzer

import (
	"path/filepath"
	"regexp"
	"strings"
	"unicode"

	"github.com/opd-ai/go-stats-generator/internal/metrics"
)

// NamingAnalyzer performs naming convention analysis on Go code
type NamingAnalyzer struct {
	genericFileNames map[string]bool
	snakeCaseRegex   *regexp.Regexp
}

// NewNamingAnalyzer creates a new naming analyzer
func NewNamingAnalyzer() *NamingAnalyzer {
	return &NamingAnalyzer{
		genericFileNames: map[string]bool{
			"utils.go":    true,
			"util.go":     true,
			"helpers.go":  true,
			"helper.go":   true,
			"misc.go":     true,
			"common.go":   true,
			"shared.go":   true,
			"base.go":     true,
			"core.go":     true,
			"lib.go":      true,
			"types.go":    true, // too generic in most contexts
			"constants.go": true,
			"errors.go":   true, // better to be specific
		},
		snakeCaseRegex: regexp.MustCompile(`^[a-z][a-z0-9]*(_[a-z0-9]+)*(_test)?\.go$`),
	}
}

// AnalyzeFileNames checks file names against Go naming conventions
func (na *NamingAnalyzer) AnalyzeFileNames(filePaths []string) []metrics.FileNameViolation {
	var violations []metrics.FileNameViolation

	for _, filePath := range filePaths {
		// Skip non-Go files
		if !strings.HasSuffix(filePath, ".go") {
			continue
		}

		fileName := filepath.Base(filePath)
		dirName := filepath.Base(filepath.Dir(filePath))

		// Check snake_case
		if violation := na.checkSnakeCase(filePath, fileName); violation != nil {
			violations = append(violations, *violation)
		}

		// Check stuttering
		if violation := na.checkStuttering(filePath, fileName, dirName); violation != nil {
			violations = append(violations, *violation)
		}

		// Check generic names
		if violation := na.checkGenericName(filePath, fileName); violation != nil {
			violations = append(violations, *violation)
		}

		// Check test suffix
		if violation := na.checkTestSuffix(filePath, fileName); violation != nil {
			violations = append(violations, *violation)
		}
	}

	return violations
}

// checkSnakeCase verifies file names are in snake_case (lowercase with underscores)
func (na *NamingAnalyzer) checkSnakeCase(filePath, fileName string) *metrics.FileNameViolation {
	// Allow _test.go suffix
	if !na.snakeCaseRegex.MatchString(fileName) {
		// Try to suggest a snake_case version
		suggested := na.toSnakeCase(fileName)
		
		return &metrics.FileNameViolation{
			File:          filePath,
			ViolationType: "non_snake_case",
			Description:   "File name should be in snake_case (lowercase with underscores)",
			SuggestedName: suggested,
			Severity:      "medium",
		}
	}
	return nil
}

// checkStuttering detects when file name repeats directory name
func (na *NamingAnalyzer) checkStuttering(filePath, fileName, dirName string) *metrics.FileNameViolation {
	// Remove .go extension and _test suffix for comparison
	baseName := strings.TrimSuffix(fileName, ".go")
	baseName = strings.TrimSuffix(baseName, "_test")

	// Check if file name starts with directory name
	if dirName != "." && dirName != "/" && !strings.HasPrefix(dirName, ".") {
		dirNameLower := strings.ToLower(dirName)
		baseNameLower := strings.ToLower(baseName)

		// Exact match or prefix match indicates stuttering
		if baseNameLower == dirNameLower || strings.HasPrefix(baseNameLower, dirNameLower+"_") {
			suggested := strings.TrimPrefix(baseNameLower, dirNameLower+"_")
			if suggested == "" {
				// If the entire name was the directory, use a more descriptive name
				suggested = baseName + "_impl"
			}
			if strings.HasSuffix(fileName, "_test.go") {
				suggested += "_test.go"
			} else {
				suggested += ".go"
			}

			return &metrics.FileNameViolation{
				File:          filePath,
				ViolationType: "stuttering",
				Description:   "File name repeats package/directory name (e.g., http/http_client.go should be http/client.go)",
				SuggestedName: suggested,
				Severity:      "low",
			}
		}
	}
	return nil
}

// checkGenericName flags overly generic file names
func (na *NamingAnalyzer) checkGenericName(filePath, fileName string) *metrics.FileNameViolation {
	if na.genericFileNames[fileName] {
		return &metrics.FileNameViolation{
			File:          filePath,
			ViolationType: "generic_name",
			Description:   "File name is too generic; use a name that describes what the code does",
			SuggestedName: "", // Cannot suggest without understanding code
			Severity:      "low",
		}
	}
	return nil
}

// checkTestSuffix verifies _test.go suffix is only on test files
func (na *NamingAnalyzer) checkTestSuffix(filePath, fileName string) *metrics.FileNameViolation {
	hasTestSuffix := strings.HasSuffix(fileName, "_test.go")
	
	// If it has _test.go, it's presumably a test file (good)
	// We can't easily check if it's NOT a test file with _test.go without parsing
	// So we'll check for the opposite: non-test files trying to use test-like names
	
	// Check for improper test naming patterns
	if !hasTestSuffix && (strings.Contains(fileName, "test_") || strings.HasPrefix(fileName, "test")) {
		suggested := strings.Replace(fileName, "test_", "", 1)
		suggested = strings.TrimPrefix(suggested, "test")
		if suggested == ".go" {
			suggested = "impl.go"
		}

		return &metrics.FileNameViolation{
			File:          filePath,
			ViolationType: "improper_test_name",
			Description:   "Test-related files should use _test.go suffix, not test_ prefix or similar",
			SuggestedName: suggested,
			Severity:      "medium",
		}
	}

	return nil
}

// toSnakeCase converts a string to snake_case
func (na *NamingAnalyzer) toSnakeCase(s string) string {
	// Remove .go extension
	s = strings.TrimSuffix(s, ".go")
	testSuffix := ""
	if strings.HasSuffix(s, "_test") {
		s = strings.TrimSuffix(s, "_test")
		testSuffix = "_test"
	}

	var result []rune
	for i, r := range s {
		if unicode.IsUpper(r) {
			// Add underscore before uppercase if not at start and previous char is lowercase
			if i > 0 && unicode.IsLower(rune(s[i-1])) {
				result = append(result, '_')
			}
			result = append(result, unicode.ToLower(r))
		} else {
			result = append(result, r)
		}
	}

	resultStr := string(result)
	// Clean up multiple underscores
	resultStr = regexp.MustCompile(`_+`).ReplaceAllString(resultStr, "_")
	resultStr = strings.Trim(resultStr, "_")

	return resultStr + testSuffix + ".go"
}

// ComputeFileNamingScore calculates an overall file naming quality score
func (na *NamingAnalyzer) ComputeFileNamingScore(violations []metrics.FileNameViolation, totalFiles int) float64 {
	if totalFiles == 0 {
		return 1.0
	}

	// Weight violations by severity
	severityWeights := map[string]float64{
		"low":    0.1,
		"medium": 0.3,
		"high":   0.5,
	}

	totalPenalty := 0.0
	for _, v := range violations {
		weight, ok := severityWeights[v.Severity]
		if !ok {
			weight = 0.2 // default
		}
		totalPenalty += weight
	}

	// Normalize penalty (max penalty = 1.0 per file)
	normalizedPenalty := totalPenalty / float64(totalFiles)
	
	// Score is 1.0 - penalty, clamped to [0, 1]
	score := 1.0 - normalizedPenalty
	if score < 0 {
		score = 0
	}
	if score > 1 {
		score = 1
	}

	return score
}
