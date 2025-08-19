package analyzer

import (
	"fmt"
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
		ca.analyzeNode(n, &concurrency, pkgName)
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
	// Pattern: Multiple goroutines reading from the same channel + WaitGroup for synchronization

	// Group goroutines by file to analyze potential worker pools at file level
	fileGoroutines := make(map[string][]metrics.GoroutineInstance)
	for _, goroutine := range concurrency.Goroutines.Instances {
		fileGoroutines[goroutine.File] = append(fileGoroutines[goroutine.File], goroutine)
	}

	// Analyze each file for worker pool patterns
	for file, goroutines := range fileGoroutines {
		if len(goroutines) >= 1 { // Even 1 goroutine can be part of a worker pool in loops
			// Check if there are channels and WaitGroups in the same context
			hasChannels := ca.hasChannelsInContext(goroutines, concurrency.Channels.Instances)
			hasWaitGroup := ca.hasWaitGroupInContext(goroutines, concurrency.SyncPrims.WaitGroups)

			// Worker pools often have multiple anonymous goroutines
			anonymousCount := 0
			for _, g := range goroutines {
				if g.IsAnonymous {
					anonymousCount++
				}
			}

			// Worker pool pattern: channels + waitgroup + anonymous goroutines
			if hasChannels && hasWaitGroup && anonymousCount >= 1 {
				confidence := ca.calculateWorkerPoolConfidence(goroutines, hasChannels, hasWaitGroup)

				// This looks like a worker pool pattern
				pattern := metrics.PatternInstance{
					Name:            "Worker Pool",
					File:            file,
					Line:            goroutines[0].Line,
					ConfidenceScore: confidence,
					Description:     fmt.Sprintf("Worker pool pattern with %d goroutine(s), channels, and WaitGroup", len(goroutines)),
					Example:         ca.extractWorkerPoolExample(file, len(goroutines)),
				}
				concurrency.WorkerPools = append(concurrency.WorkerPools, pattern)
			}
		}
	}
}

func (ca *ConcurrencyAnalyzer) detectPipelines(concurrency *metrics.ConcurrencyPatternMetrics) {
	// Detect pipeline patterns based on channel chaining
	// Pattern: Sequential goroutines connected by channels (stage1 -> channel -> stage2 -> channel -> stage3)

	// Group channels by file to analyze potential pipelines
	fileChannels := make(map[string][]metrics.ChannelInstance)
	for _, channel := range concurrency.Channels.Instances {
		fileChannels[channel.File] = append(fileChannels[channel.File], channel)
	}

	// Group goroutines by file
	fileGoroutines := make(map[string][]metrics.GoroutineInstance)
	for _, goroutine := range concurrency.Goroutines.Instances {
		fileGoroutines[goroutine.File] = append(fileGoroutines[goroutine.File], goroutine)
	}

	// Look for files with multiple channels and goroutines (potential pipelines)
	for file, channels := range fileChannels {
		goroutines := fileGoroutines[file]

		if len(channels) >= 2 && len(goroutines) >= 2 {
			// This could be a pipeline pattern
			confidence := ca.calculatePipelineConfidence(channels, goroutines)

			if confidence > 0.6 { // Higher threshold for pipelines as they're more complex to detect
				pattern := metrics.PatternInstance{
					Name:            "Pipeline",
					File:            file,
					Line:            goroutines[0].Line, // Use first goroutine line
					ConfidenceScore: confidence,
					Description:     fmt.Sprintf("Pipeline with %d stages and %d channels", len(goroutines), len(channels)),
					Example:         fmt.Sprintf("File '%s' implements pipeline pattern", file),
				}
				concurrency.Pipelines = append(concurrency.Pipelines, pattern)
			}
		}
	}
}

// calculatePipelineConfidence calculates confidence for pipeline pattern
func (ca *ConcurrencyAnalyzer) calculatePipelineConfidence(channels []metrics.ChannelInstance, goroutines []metrics.GoroutineInstance) float64 {
	confidence := 0.0

	// Multiple channels suggest pipeline stages
	if len(channels) >= 2 {
		confidence += 0.3
	}

	// Multiple goroutines suggest processing stages
	if len(goroutines) >= 2 {
		confidence += 0.3
	}

	// Sequential channels (mostly unbuffered) are typical for pipelines
	unbufferedCount := 0
	for _, channel := range channels {
		if !channel.IsBuffered {
			unbufferedCount++
		}
	}

	if float64(unbufferedCount)/float64(len(channels)) > 0.7 {
		confidence += 0.3
	}

	// More stages increase confidence
	if len(channels) >= 3 {
		confidence += 0.2
	}

	return confidence
}

// detectFanPatterns detects fan-out and fan-in concurrency patterns
func (ca *ConcurrencyAnalyzer) detectFanPatterns(concurrency *metrics.ConcurrencyPatternMetrics) {
	fileAnalysis := ca.groupConcurrencyByFile(concurrency)

	for file, analysis := range fileAnalysis {
		ca.detectFanOutPattern(file, analysis, concurrency)
		ca.detectFanInPattern(file, analysis, concurrency)
	}
}

// FileConcurrencyAnalysis holds concurrency data grouped by file
type FileConcurrencyAnalysis struct {
	Channels   []metrics.ChannelInstance
	Goroutines []metrics.GoroutineInstance
}

// groupConcurrencyByFile groups channels and goroutines by file for pattern analysis
func (ca *ConcurrencyAnalyzer) groupConcurrencyByFile(concurrency *metrics.ConcurrencyPatternMetrics) map[string]FileConcurrencyAnalysis {
	fileAnalysis := make(map[string]FileConcurrencyAnalysis)

	// Group channels by file
	for _, channel := range concurrency.Channels.Instances {
		analysis := fileAnalysis[channel.File]
		analysis.Channels = append(analysis.Channels, channel)
		fileAnalysis[channel.File] = analysis
	}

	// Group goroutines by file
	for _, goroutine := range concurrency.Goroutines.Instances {
		analysis := fileAnalysis[goroutine.File]
		analysis.Goroutines = append(analysis.Goroutines, goroutine)
		fileAnalysis[goroutine.File] = analysis
	}

	return fileAnalysis
}

// detectFanOutPattern detects fan-out patterns in a file
func (ca *ConcurrencyAnalyzer) detectFanOutPattern(file string, analysis FileConcurrencyAnalysis, concurrency *metrics.ConcurrencyPatternMetrics) {
	if !ca.hasSufficientConcurrencyForFanOut(analysis) {
		return
	}

	if len(analysis.Goroutines) > len(analysis.Channels)*2 {
		confidence := ca.calculateFanOutConfidence(analysis.Channels, analysis.Goroutines)
		if confidence > 0.6 {
			pattern := ca.createFanOutPattern(file, analysis, confidence)
			concurrency.FanOut = append(concurrency.FanOut, pattern)
		}
	}
}

// detectFanInPattern detects fan-in patterns in a file
func (ca *ConcurrencyAnalyzer) detectFanInPattern(file string, analysis FileConcurrencyAnalysis, concurrency *metrics.ConcurrencyPatternMetrics) {
	if !ca.hasSufficientConcurrencyForFanIn(analysis) {
		return
	}

	confidence := ca.calculateFanInConfidence(analysis.Channels, analysis.Goroutines)
	if confidence > 0.6 {
		pattern := ca.createFanInPattern(file, analysis, confidence)
		concurrency.FanIn = append(concurrency.FanIn, pattern)
	}
}

// hasSufficientConcurrencyForFanOut checks if there's enough concurrency for fan-out detection
func (ca *ConcurrencyAnalyzer) hasSufficientConcurrencyForFanOut(analysis FileConcurrencyAnalysis) bool {
	return len(analysis.Goroutines) >= 3 && len(analysis.Channels) >= 1
}

// hasSufficientConcurrencyForFanIn checks if there's enough concurrency for fan-in detection
func (ca *ConcurrencyAnalyzer) hasSufficientConcurrencyForFanIn(analysis FileConcurrencyAnalysis) bool {
	return len(analysis.Goroutines) >= 3 && len(analysis.Channels) <= 2
}

// createFanOutPattern creates a fan-out pattern instance
func (ca *ConcurrencyAnalyzer) createFanOutPattern(file string, analysis FileConcurrencyAnalysis, confidence float64) metrics.PatternInstance {
	return metrics.PatternInstance{
		Name:            "Fan-Out",
		File:            file,
		Line:            analysis.Goroutines[0].Line,
		ConfidenceScore: confidence,
		Description:     fmt.Sprintf("Fan-out pattern: %d goroutines consuming from %d channels", len(analysis.Goroutines), len(analysis.Channels)),
		Example:         fmt.Sprintf("Multiple consumers reading from shared channels in '%s'", file),
	}
}

// createFanInPattern creates a fan-in pattern instance
func (ca *ConcurrencyAnalyzer) createFanInPattern(file string, analysis FileConcurrencyAnalysis, confidence float64) metrics.PatternInstance {
	return metrics.PatternInstance{
		Name:            "Fan-In",
		File:            file,
		Line:            analysis.Goroutines[0].Line,
		ConfidenceScore: confidence,
		Description:     fmt.Sprintf("Fan-in pattern: %d goroutines merging into %d channels", len(analysis.Goroutines), len(analysis.Channels)),
		Example:         fmt.Sprintf("Multiple producers writing to shared channels in '%s'", file),
	}
}

// calculateFanOutConfidence calculates confidence for fan-out pattern
func (ca *ConcurrencyAnalyzer) calculateFanOutConfidence(channels []metrics.ChannelInstance, goroutines []metrics.GoroutineInstance) float64 {
	confidence := 0.0

	confidence += ca.calculateGoroutineRatioScore(channels, goroutines)
	confidence += ca.calculateGoroutineCountScore(goroutines)
	confidence += ca.calculateUnbufferedChannelScore(channels)
	confidence += ca.calculateChannelGoroutineBalance(channels, goroutines)

	return confidence
}

// calculateGoroutineRatioScore calculates confidence based on goroutine to channel ratio
func (ca *ConcurrencyAnalyzer) calculateGoroutineRatioScore(channels []metrics.ChannelInstance, goroutines []metrics.GoroutineInstance) float64 {
	ratio := float64(len(goroutines)) / float64(len(channels))
	if ratio >= 3 {
		return 0.4
	} else if ratio >= 2 {
		return 0.2
	}
	return 0.0
}

// calculateGoroutineCountScore calculates confidence based on number of goroutines
func (ca *ConcurrencyAnalyzer) calculateGoroutineCountScore(goroutines []metrics.GoroutineInstance) float64 {
	if len(goroutines) >= 3 {
		return 0.3
	}
	return 0.0
}

// calculateUnbufferedChannelScore calculates confidence based on unbuffered channel ratio
func (ca *ConcurrencyAnalyzer) calculateUnbufferedChannelScore(channels []metrics.ChannelInstance) float64 {
	if len(channels) == 0 {
		return 0.0
	}

	unbufferedCount := 0
	for _, channel := range channels {
		if !channel.IsBuffered {
			unbufferedCount++
		}
	}

	if float64(unbufferedCount)/float64(len(channels)) > 0.5 {
		return 0.2
	}
	return 0.0
}

// calculateChannelGoroutineBalance calculates confidence based on channel/goroutine balance
func (ca *ConcurrencyAnalyzer) calculateChannelGoroutineBalance(channels []metrics.ChannelInstance, goroutines []metrics.GoroutineInstance) float64 {
	if len(channels) <= 2 && len(goroutines) >= 4 {
		return 0.3
	}
	return 0.0
}

// calculateFanInConfidence calculates confidence for fan-in pattern
func (ca *ConcurrencyAnalyzer) calculateFanInConfidence(channels []metrics.ChannelInstance, goroutines []metrics.GoroutineInstance) float64 {
	confidence := 0.0

	// Multiple goroutines writing to few channels
	if len(goroutines) >= 3 && len(channels) <= 2 {
		confidence += 0.4
	}

	// More goroutines increase confidence
	if len(goroutines) >= 4 {
		confidence += 0.3
	}

	// Single output channel is very characteristic of fan-in
	if len(channels) == 1 {
		confidence += 0.3
	}

	// Unbuffered channels are common in fan-in
	unbufferedCount := 0
	for _, channel := range channels {
		if !channel.IsBuffered {
			unbufferedCount++
		}
	}
	if float64(unbufferedCount)/float64(len(channels)) > 0.5 {
		confidence += 0.2
	}

	return confidence
}

func (ca *ConcurrencyAnalyzer) detectSemaphores(concurrency *metrics.ConcurrencyPatternMetrics) {
	// Detect semaphore patterns using buffered channels
	// Pattern: Buffered channel used for limiting concurrency (typically with struct{} type)

	for _, channel := range concurrency.Channels.Instances {
		if channel.IsBuffered && channel.BufferSize > 1 {
			// This could be a semaphore pattern
			confidence := ca.calculateSemaphoreConfidence(channel)

			if confidence > 0.5 { // Only report if reasonably confident
				pattern := metrics.PatternInstance{
					Name:            "Semaphore",
					File:            channel.File,
					Line:            channel.Line,
					ConfidenceScore: confidence,
					Description:     fmt.Sprintf("Buffered channel with size %d used as semaphore", channel.BufferSize),
					Example:         fmt.Sprintf("Channel of type '%s' with buffer size %d", channel.Type, channel.BufferSize),
				}
				concurrency.Semaphores = append(concurrency.Semaphores, pattern)
			}
		}
	}
}

// calculateSemaphoreConfidence calculates confidence for semaphore pattern
func (ca *ConcurrencyAnalyzer) calculateSemaphoreConfidence(channel metrics.ChannelInstance) float64 {
	confidence := 0.0

	// Buffered channels are potential semaphores
	if channel.IsBuffered {
		confidence += 0.4
	}

	// struct{} type is very common for semaphores
	if channel.Type == "unknown" || channel.Type == "struct{}" {
		confidence += 0.4
	}

	// Small buffer sizes (2-10) are typical for semaphores
	if channel.BufferSize >= 2 && channel.BufferSize <= 10 {
		confidence += 0.3
	} else if channel.BufferSize > 10 {
		confidence += 0.1 // Still possible but less likely
	}

	return confidence
}

// Helper methods for pattern detection

// hasChannelsInContext checks if there are channels in the same context as goroutines
func (ca *ConcurrencyAnalyzer) hasChannelsInContext(goroutines []metrics.GoroutineInstance, channels []metrics.ChannelInstance) bool {
	if len(channels) == 0 {
		return false
	}

	// Check if any channels are in the same file as the goroutines
	for _, goroutine := range goroutines {
		for _, channel := range channels {
			if channel.File == goroutine.File {
				return true
			}
		}
	}
	return false
}

// hasWaitGroupInContext checks if there are WaitGroups in the same context as goroutines
func (ca *ConcurrencyAnalyzer) hasWaitGroupInContext(goroutines []metrics.GoroutineInstance, waitGroups []metrics.SyncPrimitiveInstance) bool {
	if len(waitGroups) == 0 {
		return false
	}

	// Check if any WaitGroups are in the same file as the goroutines
	for _, goroutine := range goroutines {
		for _, wg := range waitGroups {
			if wg.File == goroutine.File {
				return true
			}
		}
	}
	return false
}

// calculateWorkerPoolConfidence calculates confidence score for worker pool detection
func (ca *ConcurrencyAnalyzer) calculateWorkerPoolConfidence(goroutines []metrics.GoroutineInstance, hasChannels, hasWaitGroup bool) float64 {
	confidence := 0.0

	// Base confidence for multiple goroutines
	if len(goroutines) >= 2 {
		confidence += 0.3
	}

	// Higher confidence for more workers
	if len(goroutines) >= 3 {
		confidence += 0.2
	}

	// Channels increase confidence significantly
	if hasChannels {
		confidence += 0.3
	}

	// WaitGroup increases confidence significantly
	if hasWaitGroup {
		confidence += 0.3
	}

	// Cap at 1.0
	if confidence > 1.0 {
		confidence = 1.0
	}

	return confidence
}

// extractWorkerPoolExample creates an example description for the worker pool
func (ca *ConcurrencyAnalyzer) extractWorkerPoolExample(file string, workerCount int) string {
	return fmt.Sprintf("File '%s' launches %d goroutines in worker pool pattern", file, workerCount)
}
