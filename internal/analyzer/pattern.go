package analyzer

import (
	"go/ast"
	"go/token"
	"strings"

	"github.com/opd-ai/go-stats-generator/internal/metrics"
)

// PatternAnalyzer detects common design patterns in Go code
type PatternAnalyzer struct {
	fset *token.FileSet
}

// NewPatternAnalyzer creates analyzer for design pattern detection
func NewPatternAnalyzer(fset *token.FileSet) *PatternAnalyzer {
	return &PatternAnalyzer{fset: fset}
}

// AnalyzePatterns detects design patterns in an AST file
func (pa *PatternAnalyzer) AnalyzePatterns(file *ast.File, pkgName, filePath string) (metrics.DesignPatternMetrics, error) {
	patterns := metrics.DesignPatternMetrics{
		Singleton: []metrics.PatternInstance{},
		Factory:   []metrics.PatternInstance{},
		Builder:   []metrics.PatternInstance{},
		Observer:  []metrics.PatternInstance{},
		Strategy:  []metrics.PatternInstance{},
	}

	pa.detectSingleton(file, filePath, &patterns)
	pa.detectFactory(file, filePath, &patterns)
	pa.detectBuilder(file, filePath, &patterns)
	pa.detectObserver(file, filePath, &patterns)
	pa.detectStrategy(file, filePath, &patterns)

	return patterns, nil
}

// detectSingleton identifies singleton patterns via sync.Once or init
func (pa *PatternAnalyzer) detectSingleton(file *ast.File, filePath string, patterns *metrics.DesignPatternMetrics) {
	hasSyncOnce := false
	var onceLine int

	ast.Inspect(file, func(n ast.Node) bool {
		if genDecl, ok := n.(*ast.GenDecl); ok && genDecl.Tok == token.VAR {
			for _, spec := range genDecl.Specs {
				valueSpec, ok := spec.(*ast.ValueSpec)
				if !ok {
					continue
				}

				for i, name := range valueSpec.Names {
					if valueSpec.Type != nil && pa.isSyncOnce(valueSpec.Type) {
						hasSyncOnce = true
						onceLine = pa.fset.Position(name.Pos()).Line
					}

					if i < len(valueSpec.Values) {
						if compLit, ok := valueSpec.Values[i].(*ast.CompositeLit); ok {
							if pa.isSyncOnce(compLit.Type) {
								hasSyncOnce = true
								onceLine = pa.fset.Position(name.Pos()).Line
							}
						}
					}
				}
			}
		}
		return true
	})

	if hasSyncOnce {
		patterns.Singleton = append(patterns.Singleton, metrics.PatternInstance{
			Name:            "Singleton (sync.Once)",
			File:            filePath,
			Line:            onceLine,
			ConfidenceScore: 0.95,
			Description:     "Thread-safe singleton using sync.Once",
			Example:         "sync.Once variable for singleton initialization",
		})
	}
}

// detectFactory identifies factory patterns via New* constructors
func (pa *PatternAnalyzer) detectFactory(file *ast.File, filePath string, patterns *metrics.DesignPatternMetrics) {
	ast.Inspect(file, func(n ast.Node) bool {
		funcDecl, ok := n.(*ast.FuncDecl)
		if !ok || funcDecl.Body == nil {
			return true
		}

		funcName := funcDecl.Name.Name
		if !pa.isFactoryName(funcName) {
			return true
		}

		if funcDecl.Type.Results == nil || len(funcDecl.Type.Results.List) == 0 {
			return true
		}

		returnType := funcDecl.Type.Results.List[0].Type
		if pa.isInterfaceReturn(returnType) {
			confidence := 0.85
			if pa.hasTypeSwitch(funcDecl.Body) {
				confidence = 0.95
			}

			patterns.Factory = append(patterns.Factory, metrics.PatternInstance{
				Name:            "Factory Method",
				File:            filePath,
				Line:            pa.fset.Position(funcDecl.Pos()).Line,
				ConfidenceScore: confidence,
				Description:     "Factory function returning interface type",
				Example:         funcName + "() creates objects via factory pattern",
			})
		}
		return true
	})
}

// detectBuilder identifies builder patterns via method chaining
func (pa *PatternAnalyzer) detectBuilder(file *ast.File, filePath string, patterns *metrics.DesignPatternMetrics) {
	typeBuilders := make(map[string]*builderCandidate)

	ast.Inspect(file, func(n ast.Node) bool {
		funcDecl, ok := n.(*ast.FuncDecl)
		if !ok || funcDecl.Recv == nil || len(funcDecl.Recv.List) == 0 {
			return true
		}

		recvType := pa.getReceiverTypeName(funcDecl.Recv)
		if recvType == "" {
			return true
		}

		if _, exists := typeBuilders[recvType]; !exists {
			typeBuilders[recvType] = &builderCandidate{
				typeName:    recvType,
				line:        pa.fset.Position(funcDecl.Pos()).Line,
				setterCount: 0,
				hasBuild:    false,
				returnsSelf: 0,
			}
		}

		candidate := typeBuilders[recvType]

		if strings.HasPrefix(funcDecl.Name.Name, "Set") || strings.HasPrefix(funcDecl.Name.Name, "With") {
			candidate.setterCount++
			if pa.returnsSelf(funcDecl, recvType) {
				candidate.returnsSelf++
			}
		}

		if funcDecl.Name.Name == "Build" || funcDecl.Name.Name == "Create" {
			candidate.hasBuild = true
		}

		return true
	})

	for _, candidate := range typeBuilders {
		if candidate.setterCount >= 2 && candidate.returnsSelf >= 2 && candidate.hasBuild {
			patterns.Builder = append(patterns.Builder, metrics.PatternInstance{
				Name:            "Builder Pattern",
				File:            filePath,
				Line:            candidate.line,
				ConfidenceScore: 0.9,
				Description:     "Fluent builder with method chaining",
				Example:         candidate.typeName + " builder with " + string(rune(candidate.setterCount)) + " setters",
			})
		}
	}
}

// detectObserver identifies observer patterns via callback registration
func (pa *PatternAnalyzer) detectObserver(file *ast.File, filePath string, patterns *metrics.DesignPatternMetrics) {
	ast.Inspect(file, func(n ast.Node) bool {
		funcDecl, ok := n.(*ast.FuncDecl)
		if !ok || funcDecl.Body == nil {
			return true
		}

		funcName := funcDecl.Name.Name
		hasRegister := strings.Contains(funcName, "Register") || strings.Contains(funcName, "Add") ||
			strings.Contains(funcName, "Subscribe") || strings.Contains(funcName, "Listen")

		if !hasRegister {
			return true
		}

		if pa.hasCallbackParam(funcDecl) {
			patterns.Observer = append(patterns.Observer, metrics.PatternInstance{
				Name:            "Observer Pattern",
				File:            filePath,
				Line:            pa.fset.Position(funcDecl.Pos()).Line,
				ConfidenceScore: 0.85,
				Description:     "Callback registration for observer pattern",
				Example:         funcName + "() registers observers/callbacks",
			})
		}
		return true
	})
}

// detectStrategy identifies strategy patterns via interface delegation
func (pa *PatternAnalyzer) detectStrategy(file *ast.File, filePath string, patterns *metrics.DesignPatternMetrics) {
	typeStrategies := make(map[string]*strategyCandidate)

	ast.Inspect(file, func(n ast.Node) bool {
		typeSpec, ok := n.(*ast.TypeSpec)
		if !ok {
			return true
		}

		structType, ok := typeSpec.Type.(*ast.StructType)
		if !ok || structType.Fields == nil {
			return true
		}

		for _, field := range structType.Fields.List {
			if pa.isInterfaceField(field) {
				typeName := typeSpec.Name.Name
				if _, exists := typeStrategies[typeName]; !exists {
					typeStrategies[typeName] = &strategyCandidate{
						typeName:        typeName,
						line:            pa.fset.Position(typeSpec.Pos()).Line,
						interfaceFields: 1,
					}
				} else {
					typeStrategies[typeName].interfaceFields++
				}
			}
		}
		return true
	})

	for _, candidate := range typeStrategies {
		if candidate.interfaceFields > 0 {
			patterns.Strategy = append(patterns.Strategy, metrics.PatternInstance{
				Name:            "Strategy Pattern",
				File:            filePath,
				Line:            candidate.line,
				ConfidenceScore: 0.8,
				Description:     "Struct with interface field(s) for strategy delegation",
				Example:         candidate.typeName + " uses strategy pattern",
			})
		}
	}
}

type builderCandidate struct {
	typeName    string
	line        int
	setterCount int
	hasBuild    bool
	returnsSelf int
}

type strategyCandidate struct {
	typeName        string
	line            int
	interfaceFields int
}

func (pa *PatternAnalyzer) isSyncOnce(expr ast.Expr) bool {
	sel, ok := expr.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	ident, ok := sel.X.(*ast.Ident)
	return ok && ident.Name == "sync" && sel.Sel.Name == "Once"
}

func (pa *PatternAnalyzer) isFactoryName(name string) bool {
	prefixes := []string{"New", "Create", "Make", "Build"}
	for _, prefix := range prefixes {
		if strings.HasPrefix(name, prefix) {
			return true
		}
	}
	return false
}

func (pa *PatternAnalyzer) isInterfaceReturn(expr ast.Expr) bool {
	if _, ok := expr.(*ast.InterfaceType); ok {
		return true
	}
	if ident, ok := expr.(*ast.Ident); ok {
		return strings.HasSuffix(ident.Name, "er") || ident.Name == "Interface"
	}
	return false
}

func (pa *PatternAnalyzer) hasTypeSwitch(body *ast.BlockStmt) bool {
	hasSwitch := false
	ast.Inspect(body, func(n ast.Node) bool {
		if _, ok := n.(*ast.TypeSwitchStmt); ok {
			hasSwitch = true
			return false
		}
		return true
	})
	return hasSwitch
}

func (pa *PatternAnalyzer) getReceiverTypeName(recv *ast.FieldList) string {
	if len(recv.List) == 0 {
		return ""
	}
	switch t := recv.List[0].Type.(type) {
	case *ast.StarExpr:
		if ident, ok := t.X.(*ast.Ident); ok {
			return ident.Name
		}
	case *ast.Ident:
		return t.Name
	}
	return ""
}

func (pa *PatternAnalyzer) returnsSelf(funcDecl *ast.FuncDecl, recvType string) bool {
	if funcDecl.Type.Results == nil || len(funcDecl.Type.Results.List) == 0 {
		return false
	}

	for _, result := range funcDecl.Type.Results.List {
		switch t := result.Type.(type) {
		case *ast.StarExpr:
			if ident, ok := t.X.(*ast.Ident); ok && ident.Name == recvType {
				return true
			}
		case *ast.Ident:
			if t.Name == recvType {
				return true
			}
		}
	}
	return false
}

func (pa *PatternAnalyzer) hasCallbackParam(funcDecl *ast.FuncDecl) bool {
	if funcDecl.Type.Params == nil {
		return false
	}

	for _, param := range funcDecl.Type.Params.List {
		if _, ok := param.Type.(*ast.FuncType); ok {
			return true
		}
		if ident, ok := param.Type.(*ast.Ident); ok {
			name := ident.Name
			if strings.Contains(name, "Handler") || strings.Contains(name, "Callback") ||
				strings.Contains(name, "Listener") || strings.Contains(name, "Func") {
				return true
			}
		}
	}
	return false
}

func (pa *PatternAnalyzer) isInterfaceField(field *ast.Field) bool {
	if _, ok := field.Type.(*ast.InterfaceType); ok {
		return true
	}
	if ident, ok := field.Type.(*ast.Ident); ok {
		return strings.HasSuffix(ident.Name, "er") || ident.Name == "Interface"
	}
	return false
}
