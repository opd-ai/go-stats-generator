package analyzer

import (
	"go/ast"
	"strings"

	"github.com/opd-ai/go-stats-generator/internal/metrics"
)

// AnalyzeDocumentation analyzes documentation quality from AST comment group, extracting presence, length, and quality metrics.
// Evaluates whether comments exist, their character count, presence of code examples (heuristic-based detection),
// and overall quality score using the provided qualityScoreFunc. Returns structured DocumentationInfo for coverage tracking.
func AnalyzeDocumentation(doc *ast.CommentGroup, qualityScoreFunc func(string) float64) metrics.DocumentationInfo {
	docInfo := metrics.DocumentationInfo{}

	if doc == nil {
		return docInfo
	}

	docInfo.HasComment = true

	// Combine all comment lines
	var docText strings.Builder
	for _, comment := range doc.List {
		docText.WriteString(comment.Text)
		docText.WriteString(" ")
	}

	text := docText.String()
	docInfo.CommentLength = len(text)

	// Check for code examples (simple heuristic)
	docInfo.HasExample = strings.Contains(text, "Example") ||
		strings.Contains(text, "example") ||
		strings.Contains(text, "e.g.") ||
		strings.Contains(text, "E.g.")

	// Calculate quality score if function provided
	if qualityScoreFunc != nil {
		docInfo.QualityScore = qualityScoreFunc(text)
	}

	return docInfo
}

// CalculateDocQualityScore calculates documentation quality score using length, keywords, and structural heuristics.
// Base score (0.3) is awarded for non-empty documentation. Additional points for length (>50 chars), domain keyword usage,
// and presence of examples or explanatory phrases. Score ranges from 0.0 (no doc) to ~1.0 (excellent documentation).
func CalculateDocQualityScore(docText string, domainKeywords []string) float64 {
	if len(docText) == 0 {
		return 0.0
	}

	score := 0.3 // Base score for having documentation
	score += calculateDocLengthScore(docText)
	score += calculateDomainKeywordScore(docText, domainKeywords)
	score += calculateDocExampleScore(docText)

	return capDocScoreAtOne(score)
}

// calculateDocLengthScore returns score based on documentation length
func calculateDocLengthScore(docText string) float64 {
	score := 0.0
	if len(docText) > 50 {
		score += 0.2
	}
	if len(docText) > 100 {
		score += 0.2
	}
	return score
}

// calculateDomainKeywordScore returns score if domain-specific keywords are found
func calculateDomainKeywordScore(docText string, domainKeywords []string) float64 {
	if len(domainKeywords) == 0 {
		return 0.0
	}
	docLower := strings.ToLower(docText)
	for _, keyword := range domainKeywords {
		if strings.Contains(docLower, keyword) {
			return 0.2
		}
	}
	return 0.0
}

// calculateDocExampleScore returns score if documentation contains an example
func calculateDocExampleScore(docText string) float64 {
	if strings.Contains(strings.ToLower(docText), "example") {
		return 0.1
	}
	return 0.0
}

// capDocScoreAtOne ensures quality score does not exceed 1.0
func capDocScoreAtOne(score float64) float64 {
	if score > 1.0 {
		return 1.0
	}
	return score
}
