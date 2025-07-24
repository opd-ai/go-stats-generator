package analyzer

import (
	"go/ast"
	"go/token"

	"github.com/opd-ai/go-stats-generator/internal/metrics"
)

// ConcurrencyAnalyzer analyzes concurrency patterns in Go source code
type ConcurrencyAnalyzer struct {
	fset *token.FileSet
}

// NewConcurrencyAnalyzer creates a new concurrency analyzer
func NewConcurrencyAnalyzer(fset *token.FileSet) *ConcurrencyAnalyzer {
	return &ConcurrencyAnalyzer{
		fset: fset,
	}
}

// AnalyzeConcurrency analyzes concurrency patterns in an AST file
func (ca *ConcurrencyAnalyzer) AnalyzeConcurrency(file *ast.File, pkgName string) (metrics.ConcurrencyPatternMetrics, error) {
	concurrency := metrics.ConcurrencyPatternMetrics{
		WorkerPools: []metrics.PatternInstance{},
		Pipelines:   []metrics.PatternInstance{},
		FanOut:      []metrics.PatternInstance{},
		FanIn:       []metrics.PatternInstance{},
		Semaphores:  []metrics.PatternInstance{},
		Goroutines: metrics.GoroutineMetrics{
			Instances:      []metrics.GoroutineInstance{},
			GoroutineLeaks: []metrics.GoroutineLeakWarning{},
		},
		Channels: metrics.ChannelMetrics{
			Instances: []metrics.ChannelInstance{},
		},
		SyncPrims: metrics.SyncPrimitives{
			Mutexes:    []metrics.SyncPrimitiveInstance{},
			RWMutexes:  []metrics.SyncPrimitiveInstance{},
			WaitGroups: []metrics.SyncPrimitiveInstance{},
			Once:       []metrics.SyncPrimitiveInstance{},
			Cond:       []metrics.SyncPrimitiveInstance{},
			Atomic:     []metrics.SyncPrimitiveInstance{},
		},
	}

	// Walk through the AST to analyze concurrency patterns
	ast.Inspect(file, func(n ast.Node) bool {
		ca.analyzeNode(n, &concurrency, file.Name.Name)
		return true
	})

	// Calculate summary statistics
	ca.calculateSummaryStats(&concurrency)

	// Detect patterns
	ca.detectPatterns(&concurrency)

	return concurrency, nil
}

// analyzeNode analyzes a single AST node for concurrency patterns
func (ca *ConcurrencyAnalyzer) analyzeNode(n ast.Node, concurrency *metrics.ConcurrencyPatternMetrics, fileName string) {
	switch node := n.(type) {
	case *ast.GoStmt:
		ca.analyzeGoroutine(node, concurrency, fileName)
	case *ast.ChanType:
		ca.analyzeChannelType(node, concurrency, fileName)
	case *ast.CallExpr:
		ca.analyzeCallExpr(node, concurrency, fileName)
	case *ast.GenDecl:
		ca.analyzeGenDecl(node, concurrency, fileName)
	case *ast.FuncDecl:
		ca.analyzeFuncDecl(node, concurrency, fileName)
	}
}

// analyzeGoroutine analyzes goroutine statements
func (ca *ConcurrencyAnalyzer) analyzeGoroutine(goStmt *ast.GoStmt, concurrency *metrics.ConcurrencyPatternMetrics, fileName string) {
	pos := ca.fset.Position(goStmt.Pos())

	var functionName string
	var isAnonymous bool
	var context string

	// Determine if it's an anonymous function or named function call
	switch call := goStmt.Call.Fun.(type) {
	case *ast.FuncLit:
		isAnonymous = true
		functionName = "anonymous"
		context = ca.extractFunctionContext(call, 50) // Extract up to 50 characters
	case *ast.Ident:
		functionName = call.Name
		context = functionName
	case *ast.SelectorExpr:
		functionName = ca.extractSelectorName(call)
		context = functionName
	default:
		functionName = "unknown"
		context = "unknown"
	}

	// Check for defer statements in function (basic leak detection)
	hasDefer := ca.containsDefer(goStmt.Call)

	instance := metrics.GoroutineInstance{
		File:        fileName,
		Line:        pos.Line,
		Function:    functionName,
		IsAnonymous: isAnonymous,
		HasDefer:    hasDefer,
		Context:     context,
	}

	concurrency.Goroutines.Instances = append(concurrency.Goroutines.Instances, instance)

	// Check for potential goroutine leaks
	ca.checkGoroutineLeak(goStmt, concurrency, fileName, functionName)
}

// analyzeChannelType analyzes channel type declarations
func (ca *ConcurrencyAnalyzer) analyzeChannelType(chanType *ast.ChanType, concurrency *metrics.ConcurrencyPatternMetrics, fileName string) {
	pos := ca.fset.Position(chanType.Pos())

	direction := "bidirectional"
	isDirectional := false

	if chanType.Dir == ast.SEND {
		direction = "send-only"
		isDirectional = true
	} else if chanType.Dir == ast.RECV {
		direction = "receive-only"
		isDirectional = true
	}

	instance := metrics.ChannelInstance{
		File:          fileName,
		Line:          pos.Line,
		Function:      ca.getCurrentFunction(chanType),
		Type:          ca.extractTypeString(chanType.Value),
		IsBuffered:    false, // Will be determined in make calls
		BufferSize:    0,
		IsDirectional: isDirectional,
		Direction:     direction,
	}

	concurrency.Channels.Instances = append(concurrency.Channels.Instances, instance)
}

// analyzeCallExpr analyzes function calls for concurrency-related functions
func (ca *ConcurrencyAnalyzer) analyzeCallExpr(call *ast.CallExpr, concurrency *metrics.ConcurrencyPatternMetrics, fileName string) {
	pos := ca.fset.Position(call.Pos())
	functionName := ca.getCurrentFunction(call)

	// Check for make(chan ...) calls
	if ca.isMakeCall(call, "chan") {
		ca.analyzeMakeChannel(call, concurrency, fileName, functionName, pos.Line)
		return
	}

	// Check for sync package usage
	if selector, ok := call.Fun.(*ast.SelectorExpr); ok {
		if ident, ok := selector.X.(*ast.Ident); ok {
			ca.analyzeSyncCall(ident.Name, selector.Sel.Name, call, concurrency, fileName, functionName, pos.Line)
		}
	}
}

// analyzeGenDecl analyzes general declarations for sync primitive variables
func (ca *ConcurrencyAnalyzer) analyzeGenDecl(genDecl *ast.GenDecl, concurrency *metrics.ConcurrencyPatternMetrics, fileName string) {
	if genDecl.Tok != token.VAR {
		return
	}

	for _, spec := range genDecl.Specs {
		if valueSpec, ok := spec.(*ast.ValueSpec); ok {
			ca.analyzeVarSpec(valueSpec, concurrency, fileName)
		}
	}
}

// analyzeFuncDecl analyzes function declarations for concurrency patterns
func (ca *ConcurrencyAnalyzer) analyzeFuncDecl(funcDecl *ast.FuncDecl, concurrency *metrics.ConcurrencyPatternMetrics, fileName string) {
	if funcDecl.Body == nil {
		return
	}

	// Look for worker pool patterns, pipelines, etc.
	ca.analyzeForPatterns(funcDecl, concurrency, fileName)
}

// analyzeMakeChannel analyzes make(chan) calls for buffer size and type
func (ca *ConcurrencyAnalyzer) analyzeMakeChannel(call *ast.CallExpr, concurrency *metrics.ConcurrencyPatternMetrics, fileName, functionName string, line int) {
	if len(call.Args) < 1 {
		return
	}

	chanType, ok := call.Args[0].(*ast.ChanType)
	if !ok {
		return
	}

	isBuffered := len(call.Args) > 1
	bufferSize := 0

	if isBuffered {
		if basicLit, ok := call.Args[1].(*ast.BasicLit); ok && basicLit.Kind == token.INT {
			// Parse buffer size (simplified)
			if size := ca.parseIntLiteral(basicLit.Value); size > 0 {
				bufferSize = size
			}
		}
	}

	direction := "bidirectional"
	isDirectional := false

	if chanType.Dir == ast.SEND {
		direction = "send-only"
		isDirectional = true
	} else if chanType.Dir == ast.RECV {
		direction = "receive-only"
		isDirectional = true
	}

	instance := metrics.ChannelInstance{
		File:          fileName,
		Line:          line,
		Function:      functionName,
		Type:          ca.extractTypeString(chanType.Value),
		IsBuffered:    isBuffered,
		BufferSize:    bufferSize,
		IsDirectional: isDirectional,
		Direction:     direction,
	}

	concurrency.Channels.Instances = append(concurrency.Channels.Instances, instance)
}

// analyzeSyncCall analyzes calls to sync package functions
func (ca *ConcurrencyAnalyzer) analyzeSyncCall(packageName, functionName string, call *ast.CallExpr, concurrency *metrics.ConcurrencyPatternMetrics, fileName, currentFunc string, line int) {
	if packageName != "sync" {
		return
	}

	instance := metrics.SyncPrimitiveInstance{
		File:     fileName,
		Line:     line,
		Function: currentFunc,
		Type:     functionName,
		Variable: ca.extractVariableName(call),
		Context:  ca.extractCallContext(call),
	}

	switch functionName {
	case "Mutex", "NewMutex":
		concurrency.SyncPrims.Mutexes = append(concurrency.SyncPrims.Mutexes, instance)
	case "RWMutex", "NewRWMutex":
		concurrency.SyncPrims.RWMutexes = append(concurrency.SyncPrims.RWMutexes, instance)
	case "WaitGroup", "NewWaitGroup":
		concurrency.SyncPrims.WaitGroups = append(concurrency.SyncPrims.WaitGroups, instance)
	case "Once", "NewOnce":
		concurrency.SyncPrims.Once = append(concurrency.SyncPrims.Once, instance)
	case "Cond", "NewCond":
		concurrency.SyncPrims.Cond = append(concurrency.SyncPrims.Cond, instance)
	}
}

// analyzeVarSpec analyzes variable specifications for sync primitives
func (ca *ConcurrencyAnalyzer) analyzeVarSpec(valueSpec *ast.ValueSpec, concurrency *metrics.ConcurrencyPatternMetrics, fileName string) {
	for i, name := range valueSpec.Names {
		if valueSpec.Type != nil {
			ca.checkSyncPrimitiveType(valueSpec.Type, name.Name, concurrency, fileName, ca.fset.Position(name.Pos()).Line)
		}
		if i < len(valueSpec.Values) && valueSpec.Values[i] != nil {
			ca.checkSyncPrimitiveValue(valueSpec.Values[i], name.Name, concurrency, fileName, ca.fset.Position(name.Pos()).Line)
		}
	}
}

// checkSyncPrimitiveType checks if a type is a sync primitive
func (ca *ConcurrencyAnalyzer) checkSyncPrimitiveType(typeExpr ast.Expr, varName string, concurrency *metrics.ConcurrencyPatternMetrics, fileName string, line int) {
	if selector, ok := typeExpr.(*ast.SelectorExpr); ok {
		if ident, ok := selector.X.(*ast.Ident); ok && ident.Name == "sync" {
			instance := metrics.SyncPrimitiveInstance{
				File:     fileName,
				Line:     line,
				Function: ca.getCurrentFunction(typeExpr),
				Type:     selector.Sel.Name,
				Variable: varName,
				Context:  "declaration",
			}

			switch selector.Sel.Name {
			case "Mutex":
				concurrency.SyncPrims.Mutexes = append(concurrency.SyncPrims.Mutexes, instance)
			case "RWMutex":
				concurrency.SyncPrims.RWMutexes = append(concurrency.SyncPrims.RWMutexes, instance)
			case "WaitGroup":
				concurrency.SyncPrims.WaitGroups = append(concurrency.SyncPrims.WaitGroups, instance)
			case "Once":
				concurrency.SyncPrims.Once = append(concurrency.SyncPrims.Once, instance)
			case "Cond":
				concurrency.SyncPrims.Cond = append(concurrency.SyncPrims.Cond, instance)
			}
		}
	}
}

// checkSyncPrimitiveValue checks if a value expression creates a sync primitive
func (ca *ConcurrencyAnalyzer) checkSyncPrimitiveValue(valueExpr ast.Expr, varName string, concurrency *metrics.ConcurrencyPatternMetrics, fileName string, line int) {
	// Check for sync.Mutex{}, &sync.Mutex{}, etc.
	if compLit, ok := valueExpr.(*ast.CompositeLit); ok {
		ca.checkSyncPrimitiveType(compLit.Type, varName, concurrency, fileName, line)
	}
	if unary, ok := valueExpr.(*ast.UnaryExpr); ok && unary.Op == token.AND {
		if compLit, ok := unary.X.(*ast.CompositeLit); ok {
			ca.checkSyncPrimitiveType(compLit.Type, varName, concurrency, fileName, line)
		}
	}
}

// calculateSummaryStats calculates summary statistics for concurrency metrics
func (ca *ConcurrencyAnalyzer) calculateSummaryStats(concurrency *metrics.ConcurrencyPatternMetrics) {
	// Calculate goroutine stats
	concurrency.Goroutines.TotalCount = len(concurrency.Goroutines.Instances)
	for _, instance := range concurrency.Goroutines.Instances {
		if instance.IsAnonymous {
			concurrency.Goroutines.AnonymousCount++
		} else {
			concurrency.Goroutines.NamedCount++
		}
	}

	// Calculate channel stats
	concurrency.Channels.TotalCount = len(concurrency.Channels.Instances)
	for _, instance := range concurrency.Channels.Instances {
		if instance.IsBuffered {
			concurrency.Channels.BufferedCount++
		} else {
			concurrency.Channels.UnbufferedCount++
		}
		if instance.IsDirectional {
			concurrency.Channels.DirectionalCount++
		}
	}
}

// detectPatterns detects higher-level concurrency patterns
func (ca *ConcurrencyAnalyzer) detectPatterns(concurrency *metrics.ConcurrencyPatternMetrics) {
	// Detect worker pool patterns
	ca.detectWorkerPools(concurrency)

	// Detect pipeline patterns
	ca.detectPipelines(concurrency)

	// Detect fan-out/fan-in patterns
	ca.detectFanPatterns(concurrency)

	// Detect semaphore patterns
	ca.detectSemaphores(concurrency)
}

// Helper methods

func (ca *ConcurrencyAnalyzer) extractFunctionContext(funcLit *ast.FuncLit, maxLen int) string {
	if funcLit.Body == nil || len(funcLit.Body.List) == 0 {
		return "empty function"
	}

	// Extract first statement as context (simplified)
	return "anonymous function"
}

func (ca *ConcurrencyAnalyzer) extractSelectorName(selector *ast.SelectorExpr) string {
	if ident, ok := selector.X.(*ast.Ident); ok {
		return ident.Name + "." + selector.Sel.Name
	}
	return selector.Sel.Name
}

func (ca *ConcurrencyAnalyzer) containsDefer(call *ast.CallExpr) bool {
	// Simple check - would need more sophisticated analysis for real detection
	return false
}

func (ca *ConcurrencyAnalyzer) checkGoroutineLeak(goStmt *ast.GoStmt, concurrency *metrics.ConcurrencyPatternMetrics, fileName, functionName string) {
	// Basic goroutine leak detection
	pos := ca.fset.Position(goStmt.Pos())

	// Check for infinite loops without proper exit conditions
	if ca.hasInfiniteLoop(goStmt.Call) {
		warning := metrics.GoroutineLeakWarning{
			File:           fileName,
			Line:           pos.Line,
			Function:       functionName,
			RiskLevel:      "medium",
			Description:    "Goroutine with potential infinite loop detected",
			Recommendation: "Ensure proper exit conditions using context, channels, or other signaling mechanisms",
		}
		concurrency.Goroutines.GoroutineLeaks = append(concurrency.Goroutines.GoroutineLeaks, warning)
	}
}

func (ca *ConcurrencyAnalyzer) getCurrentFunction(node ast.Node) string {
	// Find the enclosing function - simplified implementation
	return "unknown"
}

func (ca *ConcurrencyAnalyzer) extractTypeString(expr ast.Expr) string {
	if ident, ok := expr.(*ast.Ident); ok {
		return ident.Name
	}
	return "unknown"
}

func (ca *ConcurrencyAnalyzer) isMakeCall(call *ast.CallExpr, targetType string) bool {
	if ident, ok := call.Fun.(*ast.Ident); ok && ident.Name == "make" {
		if len(call.Args) > 0 {
			if chanType, ok := call.Args[0].(*ast.ChanType); ok && targetType == "chan" {
				return chanType != nil
			}
		}
	}
	return false
}

func (ca *ConcurrencyAnalyzer) parseIntLiteral(value string) int {
	// Simplified integer parsing
	if len(value) > 0 && value[0] >= '0' && value[0] <= '9' {
		return int(value[0] - '0') // Very basic - just first digit
	}
	return 0
}

func (ca *ConcurrencyAnalyzer) extractVariableName(call *ast.CallExpr) string {
	// Extract variable name from call context - simplified
	return "unknown"
}

func (ca *ConcurrencyAnalyzer) extractCallContext(call *ast.CallExpr) string {
	// Extract call context - simplified
	return "method call"
}

func (ca *ConcurrencyAnalyzer) analyzeForPatterns(funcDecl *ast.FuncDecl, concurrency *metrics.ConcurrencyPatternMetrics, fileName string) {
	// Analyze function body for concurrency patterns - placeholder for complex pattern detection
}

func (ca *ConcurrencyAnalyzer) hasInfiniteLoop(call *ast.CallExpr) bool {
	// Detect infinite loops - simplified implementation
	return false
}

func (ca *ConcurrencyAnalyzer) detectWorkerPools(concurrency *metrics.ConcurrencyPatternMetrics) {
	// Detect worker pool patterns based on goroutine and channel usage
	// This is a complex analysis that would look for:
	// - Multiple goroutines reading from the same channel
	// - Channel used for work distribution
	// - WaitGroup for synchronization
}

func (ca *ConcurrencyAnalyzer) detectPipelines(concurrency *metrics.ConcurrencyPatternMetrics) {
	// Detect pipeline patterns based on channel chaining
	// Look for channels that are connected in sequence
}

func (ca *ConcurrencyAnalyzer) detectFanPatterns(concurrency *metrics.ConcurrencyPatternMetrics) {
	// Detect fan-out and fan-in patterns
	// Fan-out: one source, multiple consumers
	// Fan-in: multiple sources, one consumer
}

func (ca *ConcurrencyAnalyzer) detectSemaphores(concurrency *metrics.ConcurrencyPatternMetrics) {
	// Detect semaphore patterns using buffered channels
	// Look for buffered channels used for limiting concurrency
}
