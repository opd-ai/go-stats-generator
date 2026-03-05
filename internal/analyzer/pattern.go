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

// NewPatternAnalyzer creates analyzer for design pattern detection, identifying common Go idioms and architectural patterns.
// Detects patterns like singleton, factory, builder, strategy, observer, and middleware implementations through AST analysis.
// Pattern recognition helps assess code maintainability, architectural consistency, and adherence to established best practices.
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
			if found, line := pa.inspectVarDeclForSyncOnce(genDecl); found {
				hasSyncOnce = true
				onceLine = line
			}
		}
		return true
	})

	if hasSyncOnce {
		pa.addSingletonPattern(patterns, filePath, onceLine)
	}
}

// inspectVarDeclForSyncOnce checks if a variable declaration contains sync.Once
func (pa *PatternAnalyzer) inspectVarDeclForSyncOnce(genDecl *ast.GenDecl) (bool, int) {
	for _, spec := range genDecl.Specs {
		valueSpec, ok := spec.(*ast.ValueSpec)
		if !ok {
			continue
		}
		if found, line := pa.checkValueSpecForSyncOnce(valueSpec); found {
			return true, line
		}
	}
	return false, 0
}

// checkValueSpecForSyncOnce checks if a ValueSpec contains sync.Once
func (pa *PatternAnalyzer) checkValueSpecForSyncOnce(valueSpec *ast.ValueSpec) (bool, int) {
	for i, name := range valueSpec.Names {
		if pa.hasSyncOnceType(valueSpec.Type) {
			return true, pa.fset.Position(name.Pos()).Line
		}
		if pa.hasSyncOnceValue(valueSpec, i) {
			return true, pa.fset.Position(name.Pos()).Line
		}
	}
	return false, 0
}

// hasSyncOnceType checks if the type is sync.Once
func (pa *PatternAnalyzer) hasSyncOnceType(typeExpr ast.Expr) bool {
	return typeExpr != nil && pa.isSyncOnce(typeExpr)
}

// hasSyncOnceValue checks if the value at index i is a sync.Once composite literal
func (pa *PatternAnalyzer) hasSyncOnceValue(valueSpec *ast.ValueSpec, i int) bool {
	if i >= len(valueSpec.Values) {
		return false
	}
	compLit, ok := valueSpec.Values[i].(*ast.CompositeLit)
	return ok && pa.isSyncOnce(compLit.Type)
}

// addSingletonPattern appends a singleton pattern instance to the metrics
func (pa *PatternAnalyzer) addSingletonPattern(patterns *metrics.DesignPatternMetrics, filePath string, line int) {
	patterns.Singleton = append(patterns.Singleton, metrics.PatternInstance{
		Name:            "Singleton (sync.Once)",
		File:            filePath,
		Line:            line,
		ConfidenceScore: 0.95,
		Description:     "Thread-safe singleton using sync.Once",
		Example:         "sync.Once variable for singleton initialization",
	})
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
	typeBuilders := pa.collectBuilderCandidates(file)
	pa.appendBuilderPatterns(typeBuilders, filePath, patterns)
}

// collectBuilderCandidates scans methods to identify potential builder types
func (pa *PatternAnalyzer) collectBuilderCandidates(file *ast.File) map[string]*builderCandidate {
	typeBuilders := make(map[string]*builderCandidate)

	ast.Inspect(file, func(n ast.Node) bool {
		funcDecl, ok := n.(*ast.FuncDecl)
		if !ok {
			return true
		}

		if !pa.hasMethods(funcDecl) {
			return true
		}

		recvType := pa.getReceiverTypeName(funcDecl.Recv)
		if recvType == "" {
			return true
		}

		pa.ensureCandidateExists(typeBuilders, recvType, funcDecl)
		pa.updateCandidateFromMethod(typeBuilders[recvType], funcDecl, recvType)

		return true
	})

	return typeBuilders
}

// hasMethods checks if function declaration has receiver methods
func (pa *PatternAnalyzer) hasMethods(funcDecl *ast.FuncDecl) bool {
	return funcDecl.Recv != nil && len(funcDecl.Recv.List) > 0
}

// ensureCandidateExists creates builder candidate if not already tracked
func (pa *PatternAnalyzer) ensureCandidateExists(typeBuilders map[string]*builderCandidate, recvType string, funcDecl *ast.FuncDecl) {
	if _, exists := typeBuilders[recvType]; !exists {
		typeBuilders[recvType] = &builderCandidate{
			typeName:    recvType,
			line:        pa.fset.Position(funcDecl.Pos()).Line,
			setterCount: 0,
			hasBuild:    false,
			returnsSelf: 0,
		}
	}
}

// updateCandidateFromMethod updates candidate based on method characteristics
func (pa *PatternAnalyzer) updateCandidateFromMethod(candidate *builderCandidate, funcDecl *ast.FuncDecl, recvType string) {
	if pa.isSetterMethod(funcDecl.Name.Name) {
		candidate.setterCount++
		if pa.returnsSelf(funcDecl, recvType) {
			candidate.returnsSelf++
		}
	}

	if pa.isBuildMethod(funcDecl.Name.Name) {
		candidate.hasBuild = true
	}
}

// isSetterMethod checks if method name suggests a setter
func (pa *PatternAnalyzer) isSetterMethod(name string) bool {
	return strings.HasPrefix(name, "Set") || strings.HasPrefix(name, "With")
}

// isBuildMethod checks if method name suggests a build/create function
func (pa *PatternAnalyzer) isBuildMethod(name string) bool {
	return name == "Build" || name == "Create"
}

// appendBuilderPatterns adds qualified builder candidates to pattern list
func (pa *PatternAnalyzer) appendBuilderPatterns(typeBuilders map[string]*builderCandidate, filePath string, patterns *metrics.DesignPatternMetrics) {
	for _, candidate := range typeBuilders {
		if pa.isBuilderPattern(candidate) {
			patterns.Builder = append(patterns.Builder, pa.createBuilderPattern(candidate, filePath))
		}
	}
}

// isBuilderPattern checks if candidate meets builder pattern criteria
func (pa *PatternAnalyzer) isBuilderPattern(candidate *builderCandidate) bool {
	return candidate.setterCount >= 2 && candidate.returnsSelf >= 2 && candidate.hasBuild
}

// createBuilderPattern constructs pattern instance from candidate
func (pa *PatternAnalyzer) createBuilderPattern(candidate *builderCandidate, filePath string) metrics.PatternInstance {
	return metrics.PatternInstance{
		Name:            "Builder Pattern",
		File:            filePath,
		Line:            candidate.line,
		ConfidenceScore: 0.9,
		Description:     "Fluent builder with method chaining",
		Example:         candidate.typeName + " builder with " + string(rune(candidate.setterCount)) + " setters",
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
	typeStrategies := pa.collectStrategyCandidates(file)
	pa.appendStrategyPatterns(typeStrategies, filePath, patterns)
}

// collectStrategyCandidates scans AST for structs with interface fields
func (pa *PatternAnalyzer) collectStrategyCandidates(file *ast.File) map[string]*strategyCandidate {
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

		pa.processStructFieldsForStrategy(typeSpec, structType, typeStrategies)
		return true
	})

	return typeStrategies
}

// processStructFieldsForStrategy counts interface fields in a struct
func (pa *PatternAnalyzer) processStructFieldsForStrategy(typeSpec *ast.TypeSpec, structType *ast.StructType, typeStrategies map[string]*strategyCandidate) {
	for _, field := range structType.Fields.List {
		if pa.isInterfaceField(field) {
			pa.updateStrategyCandidate(typeSpec, typeStrategies)
		}
	}
}

// updateStrategyCandidate creates or updates strategy candidate for a type
func (pa *PatternAnalyzer) updateStrategyCandidate(typeSpec *ast.TypeSpec, typeStrategies map[string]*strategyCandidate) {
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

// appendStrategyPatterns adds qualified strategy candidates to pattern list
func (pa *PatternAnalyzer) appendStrategyPatterns(typeStrategies map[string]*strategyCandidate, filePath string, patterns *metrics.DesignPatternMetrics) {
	for _, candidate := range typeStrategies {
		if candidate.interfaceFields > 0 {
			patterns.Strategy = append(patterns.Strategy, pa.createStrategyPattern(candidate, filePath))
		}
	}
}

// createStrategyPattern constructs pattern instance from candidate
func (pa *PatternAnalyzer) createStrategyPattern(candidate *strategyCandidate, filePath string) metrics.PatternInstance {
	return metrics.PatternInstance{
		Name:            "Strategy Pattern",
		File:            filePath,
		Line:            candidate.line,
		ConfidenceScore: 0.8,
		Description:     "Struct with interface field(s) for strategy delegation",
		Example:         candidate.typeName + " uses strategy pattern",
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

// isFactoryName checks if a function name follows factory method naming conventions.
func (pa *PatternAnalyzer) isFactoryName(name string) bool {
	prefixes := []string{"New", "Create", "Make", "Build"}
	for _, prefix := range prefixes {
		if strings.HasPrefix(name, prefix) {
			return true
		}
	}
	return false
}

// isInterfaceReturn determines if an expression represents an interface type.
func (pa *PatternAnalyzer) isInterfaceReturn(expr ast.Expr) bool {
	if _, ok := expr.(*ast.InterfaceType); ok {
		return true
	}
	if ident, ok := expr.(*ast.Ident); ok {
		if ident.Name == "Interface" {
			return true
		}
		if ident.Obj != nil {
			if typeSpec, ok := ident.Obj.Decl.(*ast.TypeSpec); ok {
				if _, ok := typeSpec.Type.(*ast.InterfaceType); ok {
					return true
				}
			}
		}
	}
	if sel, ok := expr.(*ast.SelectorExpr); ok {
		name := sel.Sel.Name
		if name == "Interface" {
			return true
		}
	}
	return false
}

// hasTypeSwitch checks if a function body contains a type switch statement.
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

// getReceiverTypeName extracts the type name from a method receiver field list.
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

// returnsSelf checks if a method returns its receiver type for builder pattern detection.
func (pa *PatternAnalyzer) returnsSelf(funcDecl *ast.FuncDecl, recvType string) bool {
	if funcDecl.Type.Results == nil || len(funcDecl.Type.Results.List) == 0 {
		return false
	}
	for _, result := range funcDecl.Type.Results.List {
		if pa.resultMatchesReceiverType(result.Type, recvType) {
			return true
		}
	}
	return false
}

// resultMatchesReceiverType checks if a result type matches the receiver type
func (pa *PatternAnalyzer) resultMatchesReceiverType(expr ast.Expr, recvType string) bool {
	switch t := expr.(type) {
	case *ast.StarExpr:
		if ident, ok := t.X.(*ast.Ident); ok {
			return ident.Name == recvType
		}
	case *ast.Ident:
		return t.Name == recvType
	}
	return false
}

// hasCallbackParam checks if a function accepts a function parameter for strategy/callback patterns.
func (pa *PatternAnalyzer) hasCallbackParam(funcDecl *ast.FuncDecl) bool {
	if funcDecl.Type.Params == nil {
		return false
	}
	for _, param := range funcDecl.Type.Params.List {
		if pa.isCallbackType(param.Type) {
			return true
		}
	}
	return false
}

// isCallbackType checks if a type is a callback type
func (pa *PatternAnalyzer) isCallbackType(expr ast.Expr) bool {
	if _, ok := expr.(*ast.FuncType); ok {
		return true
	}
	if ident, ok := expr.(*ast.Ident); ok {
		name := ident.Name
		return strings.Contains(name, "Handler") || strings.Contains(name, "Callback") ||
			strings.Contains(name, "Listener") || strings.Contains(name, "Func")
	}
	return false
}

// isInterfaceField determines if a struct field is an interface type.
func (pa *PatternAnalyzer) isInterfaceField(field *ast.Field) bool {
	if _, ok := field.Type.(*ast.InterfaceType); ok {
		return true
	}
	if ident, ok := field.Type.(*ast.Ident); ok {
		if ident.Name == "Interface" {
			return true
		}
		if ident.Obj != nil {
			if typeSpec, ok := ident.Obj.Decl.(*ast.TypeSpec); ok {
				if _, ok := typeSpec.Type.(*ast.InterfaceType); ok {
					return true
				}
			}
		}
	}
	return false
}
