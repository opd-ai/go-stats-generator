package analyzer

import (
	"go/ast"
	"go/token"
	"reflect"
	"strings"

	"github.com/opd-ai/go-stats-generator/internal/metrics"
)

// StructAnalyzer analyzes struct declarations in Go source code
// It categorizes fields by type, analyzes embedded types, and calculates
// complexity metrics according to the project's requirements for detailed
// struct member categorization.
type StructAnalyzer struct {
	fset *token.FileSet
}

// NewStructAnalyzer creates a new struct analyzer
func NewStructAnalyzer(fset *token.FileSet) *StructAnalyzer {
	return &StructAnalyzer{
		fset: fset,
	}
}

// AnalyzeStructs analyzes all struct declarations in an AST file
func (sa *StructAnalyzer) AnalyzeStructs(file *ast.File, pkgName string) ([]metrics.StructMetrics, error) {
	var structs []metrics.StructMetrics

	// Find all struct declarations (both standalone and in type specs)
	for _, decl := range file.Decls {
		if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.TYPE {
			for _, spec := range genDecl.Specs {
				if typeSpec, ok := spec.(*ast.TypeSpec); ok {
					if structType, ok := typeSpec.Type.(*ast.StructType); ok {
						structMetric, err := sa.analyzeStruct(typeSpec, structType, file.Name.Name, pkgName, genDecl.Doc)
						if err != nil {
							continue // Log warning and continue
						}
						structs = append(structs, structMetric)
					}
				}
			}
		}
	}

	return structs, nil
}

// analyzeStruct analyzes a single struct declaration
func (sa *StructAnalyzer) analyzeStruct(typeSpec *ast.TypeSpec, structType *ast.StructType, fileName, pkgName string, doc *ast.CommentGroup) (metrics.StructMetrics, error) {
	pos := sa.fset.Position(typeSpec.Pos())

	structMetric := metrics.StructMetrics{
		Name:         typeSpec.Name.Name,
		Package:      pkgName,
		File:         fileName,
		Line:         pos.Line,
		IsExported:   ast.IsExported(typeSpec.Name.Name),
		FieldsByType: make(map[metrics.FieldType]int),
		Tags:         make(map[string]int),
	}

	// Analyze struct fields
	if structType.Fields != nil {
		structMetric.TotalFields = len(structType.Fields.List)

		for _, field := range structType.Fields.List {
			sa.analyzeField(field, &structMetric)
		}
	}

	// Calculate complexity score
	structMetric.Complexity = sa.calculateComplexity(structMetric)

	// Analyze documentation
	structMetric.Documentation = sa.analyzeDocumentation(doc)

	return structMetric, nil
}

// analyzeField analyzes a single struct field and updates metrics
func (sa *StructAnalyzer) analyzeField(field *ast.Field, structMetric *metrics.StructMetrics) {
	// Handle embedded types (fields without names)
	if len(field.Names) == 0 {
		embedded := sa.extractEmbeddedType(field.Type)
		if embedded.Name != "" {
			structMetric.EmbeddedTypes = append(structMetric.EmbeddedTypes, embedded)
			structMetric.FieldsByType[metrics.FieldTypeEmbedded]++
		}
		return
	}

	// Regular fields (with names)
	fieldType := sa.categorizeFieldType(field.Type)
	structMetric.FieldsByType[fieldType] += len(field.Names)

	// Analyze struct tags
	if field.Tag != nil {
		sa.analyzeTags(field.Tag.Value, structMetric)
	}
}

// categorizeFieldType determines the category of a field type
func (sa *StructAnalyzer) categorizeFieldType(expr ast.Expr) metrics.FieldType {
	switch t := expr.(type) {
	case *ast.Ident:
		// Built-in types or types in same package
		if sa.isPrimitiveType(t.Name) {
			return metrics.FieldTypePrimitive
		}
		return metrics.FieldTypeStruct // Custom type

	case *ast.SelectorExpr:
		// Types from other packages (pkg.Type)
		return metrics.FieldTypeStruct

	case *ast.StarExpr:
		// Pointer types
		return metrics.FieldTypePointer

	case *ast.ArrayType:
		// Arrays and slices
		return metrics.FieldTypeSlice

	case *ast.MapType:
		// Map types
		return metrics.FieldTypeMap

	case *ast.ChanType:
		// Channel types
		return metrics.FieldTypeChannel

	case *ast.InterfaceType:
		// Interface types
		return metrics.FieldTypeInterface

	case *ast.FuncType:
		// Function types
		return metrics.FieldTypeFunction

	default:
		// Default to struct for unknown types
		return metrics.FieldTypeStruct
	}
}

// isPrimitiveType checks if a type name represents a Go primitive type
func (sa *StructAnalyzer) isPrimitiveType(typeName string) bool {
	primitives := map[string]bool{
		"bool":       true,
		"string":     true,
		"int":        true,
		"int8":       true,
		"int16":      true,
		"int32":      true,
		"int64":      true,
		"uint":       true,
		"uint8":      true,
		"uint16":     true,
		"uint32":     true,
		"uint64":     true,
		"uintptr":    true,
		"byte":       true,
		"rune":       true,
		"float32":    true,
		"float64":    true,
		"complex64":  true,
		"complex128": true,
	}
	return primitives[typeName]
}

// extractEmbeddedType extracts information about an embedded type
func (sa *StructAnalyzer) extractEmbeddedType(expr ast.Expr) metrics.EmbeddedType {
	embedded := metrics.EmbeddedType{}

	switch t := expr.(type) {
	case *ast.Ident:
		embedded.Name = t.Name
		embedded.Package = "" // Same package
		embedded.IsExported = ast.IsExported(t.Name)

	case *ast.SelectorExpr:
		// pkg.Type
		if pkgIdent, ok := t.X.(*ast.Ident); ok {
			embedded.Package = pkgIdent.Name
			embedded.Name = t.Sel.Name
			embedded.IsExported = ast.IsExported(t.Sel.Name)
		}

	case *ast.StarExpr:
		// *Type or *pkg.Type
		embedded.IsPointer = true
		if inner := sa.extractEmbeddedType(t.X); inner.Name != "" {
			embedded.Name = inner.Name
			embedded.Package = inner.Package
			embedded.IsExported = inner.IsExported
		}
	}

	return embedded
}

// analyzeTags parses struct tags and counts usage
func (sa *StructAnalyzer) analyzeTags(tagValue string, structMetric *metrics.StructMetrics) {
	// Remove quotes from tag value
	if len(tagValue) >= 2 && tagValue[0] == '`' && tagValue[len(tagValue)-1] == '`' {
		tagValue = tagValue[1 : len(tagValue)-1]
	}

	// Parse struct tag using reflect
	tag := reflect.StructTag(tagValue)

	// Count common tag types
	tagTypes := []string{"json", "xml", "yaml", "db", "form", "validate", "binding"}
	for _, tagType := range tagTypes {
		if value := tag.Get(tagType); value != "" {
			structMetric.Tags[tagType]++
		}
	}
}

// calculateComplexity calculates complexity score for a struct
func (sa *StructAnalyzer) calculateComplexity(structMetric metrics.StructMetrics) metrics.ComplexityScore {
	complexity := metrics.ComplexityScore{}

	// Base complexity from field count
	complexity.Cyclomatic = structMetric.TotalFields

	// Add complexity for different field types
	for fieldType, count := range structMetric.FieldsByType {
		switch fieldType {
		case metrics.FieldTypeMap, metrics.FieldTypeChannel, metrics.FieldTypeInterface:
			complexity.Cyclomatic += count * 2 // More complex types
		case metrics.FieldTypeFunction, metrics.FieldTypeEmbedded:
			complexity.Cyclomatic += count * 3 // Highest complexity
		default:
			complexity.Cyclomatic += count
		}
	}

	// Add complexity for embedded types
	complexity.NestingDepth = len(structMetric.EmbeddedTypes)

	// Calculate overall complexity score
	complexity.Overall = float64(complexity.Cyclomatic) + float64(complexity.NestingDepth)*0.5

	return complexity
}

// analyzeDocumentation analyzes struct documentation quality
func (sa *StructAnalyzer) analyzeDocumentation(doc *ast.CommentGroup) metrics.DocumentationInfo {
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
		strings.Contains(text, "//")

	// Calculate quality score
	docInfo.QualityScore = sa.calculateDocQualityScore(text)

	return docInfo
}

// calculateDocQualityScore calculates documentation quality score
func (sa *StructAnalyzer) calculateDocQualityScore(docText string) float64 {
	if len(docText) == 0 {
		return 0.0
	}

	score := 0.0

	// Base score for having documentation
	score += 0.3

	// Length-based scoring
	if len(docText) > 50 {
		score += 0.2
	}
	if len(docText) > 100 {
		score += 0.2
	}

	// Content quality indicators
	if strings.Contains(strings.ToLower(docText), "represents") ||
		strings.Contains(strings.ToLower(docText), "contains") ||
		strings.Contains(strings.ToLower(docText), "provides") {
		score += 0.2
	}

	// Example presence
	if strings.Contains(strings.ToLower(docText), "example") {
		score += 0.1
	}

	// Cap at 1.0
	if score > 1.0 {
		score = 1.0
	}

	return score
}
