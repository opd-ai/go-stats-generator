package analyzer

import (
	"go/ast"
	"go/token"
	"regexp"
	"strings"

	"github.com/opd-ai/go-stats-generator/internal/metrics"
)

// DocumentationAnalyzer performs documentation quality analysis on Go code
type DocumentationAnalyzer struct {
	fset               *token.FileSet
	cfg                *DocumentationConfig
	annotationRegex    *regexp.Regexp
	severityClassifier map[string]string
}

// DocumentationConfig contains configuration for documentation analysis
type DocumentationConfig struct {
	RequireExportedDoc  bool
	RequirePackageDoc   bool
	StaleAnnotationDays int
	MinCommentWords     int
}

// NewDocumentationAnalyzer creates a new documentation analyzer with the given
// NewDocumentationAnalyzer uses sensible defaults if cfg is nil.
func NewDocumentationAnalyzer(fset *token.FileSet, cfg *DocumentationConfig) *DocumentationAnalyzer {
	if cfg == nil {
		cfg = &DocumentationConfig{
			RequireExportedDoc:  true,
			RequirePackageDoc:   true,
			StaleAnnotationDays: 180,
			MinCommentWords:     5,
		}
	}

	return &DocumentationAnalyzer{
		fset:            fset,
		cfg:             cfg,
		annotationRegex: regexp.MustCompile(`(?i)(TODO|FIXME|HACK|BUG|XXX|DEPRECATED|NOTE)[\s:]+(.*)`),
		severityClassifier: map[string]string{
			"FIXME":      "critical",
			"BUG":        "critical",
			"HACK":       "high",
			"TODO":       "medium",
			"XXX":        "medium",
			"DEPRECATED": "low",
			"NOTE":       "low",
		},
	}
}

// Analyze performs comprehensive documentation analysis including coverage,
// Analyze tracks quality and annotations for all provided AST files and packages.
func (d *DocumentationAnalyzer) Analyze(files []*ast.File, pkgs map[string]*ast.Package) *metrics.DocumentationMetrics {
	m := &metrics.DocumentationMetrics{
		Coverage:              metrics.DocumentationCoverage{},
		Quality:               metrics.DocumentationQuality{},
		TODOComments:          []metrics.TODOComment{},
		FIXMEComments:         []metrics.FIXMEComment{},
		HACKComments:          []metrics.HACKComment{},
		BUGComments:           []metrics.BUGComment{},
		XXXComments:           []metrics.XXXComment{},
		DEPRECATEDComments:    []metrics.DEPRECATEDComment{},
		NOTEComments:          []metrics.NOTEComment{},
		AnnotationsByCategory: make(map[string]int),
	}

	// Analyze exported symbols for documentation coverage
	d.analyzeExportedSymbols(files, m)

	// Analyze package-level documentation
	d.analyzePackageDocs(files, pkgs, m)

	// Analyze annotations (TODO, FIXME, HACK, etc.)
	d.analyzeAnnotations(files, m)

	// Analyze documentation quality
	d.analyzeQuality(files, m)

	return m
}

// analyzeExportedSymbols checks documentation coverage for exported symbols
func (d *DocumentationAnalyzer) analyzeExportedSymbols(files []*ast.File, m *metrics.DocumentationMetrics) {
	var totalFuncs, documentedFuncs int
	var totalTypes, documentedTypes int
	var totalMethods, documentedMethods int

	for _, file := range files {
		ast.Inspect(file, func(n ast.Node) bool {
			d.processNode(n, &totalFuncs, &documentedFuncs, &totalTypes, &documentedTypes, &totalMethods, &documentedMethods)
			return true
		})
	}

	d.calculateCoverageMetrics(m, totalFuncs, documentedFuncs, totalTypes, documentedTypes, totalMethods, documentedMethods)
}

// processNode processes a single AST node for documentation analysis
func (d *DocumentationAnalyzer) processNode(n ast.Node, totalFuncs, documentedFuncs, totalTypes, documentedTypes, totalMethods, documentedMethods *int) {
	switch decl := n.(type) {
	case *ast.FuncDecl:
		d.processFuncDecl(decl, totalFuncs, documentedFuncs, totalMethods, documentedMethods)
	case *ast.GenDecl:
		d.processGenDecl(decl, totalTypes, documentedTypes)
	}
}

// processFuncDecl processes function declarations
func (d *DocumentationAnalyzer) processFuncDecl(decl *ast.FuncDecl, totalFuncs, documentedFuncs, totalMethods, documentedMethods *int) {
	if !decl.Name.IsExported() {
		return
	}

	if decl.Recv == nil {
		*totalFuncs++
		if d.checkExportedDoc(decl.Name.Name, decl.Doc) {
			*documentedFuncs++
		}
	} else {
		*totalMethods++
		if d.checkExportedDoc(decl.Name.Name, decl.Doc) {
			*documentedMethods++
		}
	}
}

// processGenDecl processes general declarations (types)
func (d *DocumentationAnalyzer) processGenDecl(decl *ast.GenDecl, totalTypes, documentedTypes *int) {
	for _, spec := range decl.Specs {
		if ts, ok := spec.(*ast.TypeSpec); ok && ts.Name.IsExported() {
			*totalTypes++
			if d.checkExportedDoc(ts.Name.Name, decl.Doc) {
				*documentedTypes++
			}
		}
	}
}

// calculateCoverageMetrics computes and stores coverage percentages
func (d *DocumentationAnalyzer) calculateCoverageMetrics(m *metrics.DocumentationMetrics, totalFuncs, documentedFuncs, totalTypes, documentedTypes, totalMethods, documentedMethods int) {
	m.Coverage.Functions = calculatePercentage(documentedFuncs, totalFuncs)
	m.Coverage.Types = calculatePercentage(documentedTypes, totalTypes)
	m.Coverage.Methods = calculatePercentage(documentedMethods, totalMethods)

	total := totalFuncs + totalTypes + totalMethods
	documented := documentedFuncs + documentedTypes + documentedMethods
	m.Coverage.Overall = calculatePercentage(documented, total)
}

// calculatePercentage computes percentage with zero-division safety
func calculatePercentage(part, total int) float64 {
	if total == 0 {
		return 0.0
	}
	return float64(part) / float64(total) * 100.0
}

// checkExportedDoc validates GoDoc comment for an exported symbol
func (d *DocumentationAnalyzer) checkExportedDoc(name string, doc *ast.CommentGroup) bool {
	if doc == nil {
		return false
	}

	text := doc.Text()
	if text == "" {
		return false
	}

	if !strings.HasPrefix(strings.TrimSpace(text), name) {
		return false
	}

	words := strings.Fields(text)
	if len(words) <= d.cfg.MinCommentWords {
		return false
	}

	return true
}

// extractAnnotation parses annotation comments (TODO, FIXME, etc.)
func (d *DocumentationAnalyzer) extractAnnotation(comment string) (category, description string) {
	matches := d.annotationRegex.FindStringSubmatch(comment)
	if len(matches) < 3 {
		return "", ""
	}

	category = strings.ToUpper(matches[1])
	description = strings.TrimSpace(matches[2])
	return category, description
}

// getSeverity returns severity classification for an annotation
func (d *DocumentationAnalyzer) getSeverity(category string) string {
	if severity, ok := d.severityClassifier[strings.ToUpper(category)]; ok {
		return severity
	}
	return "low"
}

// analyzePackageDocs checks for package-level documentation
func (d *DocumentationAnalyzer) analyzePackageDocs(files []*ast.File, pkgs map[string]*ast.Package, m *metrics.DocumentationMetrics) {
	pkgDocs := make(map[string]bool)
	totalPkgs := 0

	for _, file := range files {
		pkgName := file.Name.Name
		if _, seen := pkgDocs[pkgName]; !seen {
			totalPkgs++
			pkgDocs[pkgName] = d.hasPackageDoc(file)
		}
	}

	documented := 0
	for _, hasDoc := range pkgDocs {
		if hasDoc {
			documented++
		}
	}

	m.Coverage.Packages = calculatePercentage(documented, totalPkgs)
}

// hasPackageDoc checks if a file has package-level documentation
func (d *DocumentationAnalyzer) hasPackageDoc(file *ast.File) bool {
	if file.Doc != nil && file.Doc.Text() != "" {
		return true
	}
	return false
}

// analyzeAnnotations scans all comments for annotations
func (d *DocumentationAnalyzer) analyzeAnnotations(files []*ast.File, m *metrics.DocumentationMetrics) {
	for _, file := range files {
		filePath := d.fset.Position(file.Pos()).Filename
		d.scanFileComments(file, filePath, m)
	}
}

// scanFileComments processes all comment groups in a file
func (d *DocumentationAnalyzer) scanFileComments(file *ast.File, filePath string, m *metrics.DocumentationMetrics) {
	for _, cg := range file.Comments {
		for _, comment := range cg.List {
			d.processComment(comment, filePath, m)
		}
	}
}

// processComment extracts and categorizes an annotation
func (d *DocumentationAnalyzer) processComment(comment *ast.Comment, filePath string, m *metrics.DocumentationMetrics) {
	category, description := d.extractAnnotation(comment.Text)
	if category == "" {
		return
	}

	line := d.fset.Position(comment.Pos()).Line
	m.AnnotationsByCategory[category]++
	d.addAnnotationToMetrics(category, filePath, line, description, m)
}

// addAnnotationToMetrics appends the annotation to appropriate metrics list
func (d *DocumentationAnalyzer) addAnnotationToMetrics(category, filePath string, line int, description string, m *metrics.DocumentationMetrics) {
	switch category {
	case "TODO":
		m.TODOComments = append(m.TODOComments, metrics.TODOComment{
			File: filePath, Line: line, Description: description,
		})
	case "FIXME":
		m.FIXMEComments = append(m.FIXMEComments, metrics.FIXMEComment{
			File: filePath, Line: line, Description: description, Severity: d.getSeverity(category),
		})
	case "HACK":
		m.HACKComments = append(m.HACKComments, metrics.HACKComment{
			File: filePath, Line: line, Description: description, Reason: description,
		})
	case "BUG":
		m.BUGComments = append(m.BUGComments, metrics.BUGComment{
			File: filePath, Line: line, Description: description, Severity: d.getSeverity(category),
		})
	case "XXX":
		m.XXXComments = append(m.XXXComments, metrics.XXXComment{
			File: filePath, Line: line, Description: description,
		})
	case "DEPRECATED":
		m.DEPRECATEDComments = append(m.DEPRECATEDComments, metrics.DEPRECATEDComment{
			File: filePath, Line: line, Description: description,
		})
	case "NOTE":
		m.NOTEComments = append(m.NOTEComments, metrics.NOTEComment{
			File: filePath, Line: line, Description: description,
		})
	}
}

// analyzeQuality performs comprehensive quality analysis on documentation
func (d *DocumentationAnalyzer) analyzeQuality(files []*ast.File, m *metrics.DocumentationMetrics) {
	stats := d.collectQualityStats(files)
	d.populateQualityMetrics(stats, m)
}

// qualityStats holds intermediate quality analysis statistics
type qualityStats struct {
	totalLength    int
	commentCount   int
	inlineComments int
	blockComments  int
	codeExamples   int
}

// collectQualityStats gathers quality statistics from all files
func (d *DocumentationAnalyzer) collectQualityStats(files []*ast.File) *qualityStats {
	stats := &qualityStats{}
	for _, file := range files {
		d.processFileComments(file, stats)
	}
	return stats
}

// processFileComments analyzes all comment groups in a file
func (d *DocumentationAnalyzer) processFileComments(file *ast.File, stats *qualityStats) {
	for _, cg := range file.Comments {
		d.processCommentGroup(cg, stats)
	}
}

// processCommentGroup analyzes a single comment group
func (d *DocumentationAnalyzer) processCommentGroup(cg *ast.CommentGroup, stats *qualityStats) {
	commentText := cg.Text()
	stats.commentCount++
	stats.totalLength += len(commentText)

	d.classifyComments(cg, stats)

	if d.containsCodeExample(commentText) {
		stats.codeExamples++
	}
}

// classifyComments categorizes individual comments as inline or block
func (d *DocumentationAnalyzer) classifyComments(cg *ast.CommentGroup, stats *qualityStats) {
	for _, comment := range cg.List {
		if strings.HasPrefix(comment.Text, "//") {
			stats.inlineComments++
		} else if strings.HasPrefix(comment.Text, "/*") {
			stats.blockComments++
		}
	}
}

// populateQualityMetrics transfers collected stats to metrics structure
func (d *DocumentationAnalyzer) populateQualityMetrics(stats *qualityStats, m *metrics.DocumentationMetrics) {
	if stats.commentCount > 0 {
		m.Quality.AverageLength = float64(stats.totalLength) / float64(stats.commentCount)
	}
	m.Quality.InlineComments = stats.inlineComments
	m.Quality.BlockComments = stats.blockComments
	m.Quality.CodeExamples = stats.codeExamples
	m.Quality.QualityScore = d.calculateQualityScore(m)
}

// containsCodeExample detects code examples in comments
func (d *DocumentationAnalyzer) containsCodeExample(text string) bool {
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.Contains(trimmed, "Example:") ||
			strings.Contains(trimmed, "Usage:") ||
			(strings.HasPrefix(trimmed, "func ") || strings.HasPrefix(trimmed, "var ") || strings.HasPrefix(trimmed, "type ")) {
			return true
		}
	}
	return false
}

// calculateQualityScore computes overall quality score based on coverage and quality
func (d *DocumentationAnalyzer) calculateQualityScore(m *metrics.DocumentationMetrics) float64 {
	score := 0.0

	score += m.Coverage.Overall * 0.4

	if m.Quality.AverageLength > 50 {
		score += 20.0
	} else if m.Quality.AverageLength > 30 {
		score += 10.0
	}

	score += float64(m.Quality.CodeExamples) * 2.0
	if score > 100.0 {
		score = 100.0
	}

	return score
}
