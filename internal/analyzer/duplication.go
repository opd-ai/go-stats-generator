package analyzer

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"hash/fnv"
	"sort"
	"strings"

	"github.com/opd-ai/go-stats-generator/internal/metrics"
)

// MaxDeepCopyNodes is the maximum number of AST nodes allowed for deep copy operations
// to prevent excessive memory usage on pathological code
const MaxDeepCopyNodes = 10000

// DuplicationAnalyzer performs code duplication detection using AST fingerprinting
type DuplicationAnalyzer struct {
	fset *token.FileSet
}

// NewDuplicationAnalyzer creates a new duplication analyzer for detecting
// NewDuplicationAnalyzer identifies code clones and near-duplicates using structural and hash-based comparison.
func NewDuplicationAnalyzer(fset *token.FileSet) *DuplicationAnalyzer {
	return &DuplicationAnalyzer{
		fset: fset,
	}
}

// FileSet returns the token.FileSet used by this analyzer for accurate source position mapping
// and line number tracking. Essential for reporting duplication locations in code clone detection,
// enabling developers to locate and refactor duplicated code blocks. Used throughout block extraction
// and fingerprinting operations.
func (da *DuplicationAnalyzer) FileSet() *token.FileSet {
	return da.fset
}

// StatementBlock represents a block of statements extracted from a function
type StatementBlock struct {
	File       string
	StartLine  int
	EndLine    int
	Statements []ast.Stmt
	NodeCount  int
}

// NormalizedBlock represents a structurally normalized block
type NormalizedBlock struct {
	Structure string
	NodeCount int
}

// BlockFingerprint represents a fingerprinted code block
type BlockFingerprint struct {
	Hash      string
	File      string
	StartLine int
	EndLine   int
	NodeCount int
	Original  StatementBlock
}

// ExtractBlocks walks function and method bodies to extract statement-level sub-trees for
// duplication analysis. It uses a sliding window approach to generate blocks of various sizes,
// respecting the minimum block size threshold. Extracted blocks are used for structural comparison
// to detect code clones (exact, renamed, and near-duplicates). Returns all candidate blocks.
func (da *DuplicationAnalyzer) ExtractBlocks(file *ast.File, filePath string, minBlockLines int) []StatementBlock {
	var blocks []StatementBlock

	// Visit all function declarations
	ast.Inspect(file, func(n ast.Node) bool {
		funcDecl, ok := n.(*ast.FuncDecl)
		if !ok || funcDecl.Body == nil {
			return true
		}

		// Extract blocks from the function body
		funcBlocks := da.extractBlocksFromStmtList(funcDecl.Body.List, filePath, minBlockLines)
		blocks = append(blocks, funcBlocks...)

		return true
	})

	return blocks
}

// extractBlocksFromStmtList extracts statement blocks from a list of statements
func (da *DuplicationAnalyzer) extractBlocksFromStmtList(stmts []ast.Stmt, filePath string, minBlockLines int) []StatementBlock {
	var blocks []StatementBlock

	// Use a sliding window to extract blocks of various sizes
	for windowSize := minBlockLines; windowSize <= len(stmts); windowSize++ {
		for start := 0; start <= len(stmts)-windowSize; start++ {
			end := start + windowSize
			blockStmts := stmts[start:end]

			// Calculate block metrics
			if len(blockStmts) == 0 {
				continue
			}

			startPos := da.fset.Position(blockStmts[0].Pos())
			endPos := da.fset.Position(blockStmts[len(blockStmts)-1].End())

			// Count nodes in the block
			nodeCount := CountNodes(blockStmts)

			block := StatementBlock{
				File:       filePath,
				StartLine:  startPos.Line,
				EndLine:    endPos.Line,
				Statements: blockStmts,
				NodeCount:  nodeCount,
			}

			blocks = append(blocks, block)
		}
	}

	// Also recursively extract blocks from nested statements
	for _, stmt := range stmts {
		blocks = append(blocks, da.extractNestedBlocks(stmt, filePath, minBlockLines)...)
	}

	return blocks
}

// extractNestedBlocks extracts blocks from nested statement structures
func (da *DuplicationAnalyzer) extractNestedBlocks(stmt ast.Stmt, filePath string, minBlockLines int) []StatementBlock {
	var blocks []StatementBlock

	switch s := stmt.(type) {
	case *ast.BlockStmt:
		blocks = append(blocks, da.extractBlocksFromStmtList(s.List, filePath, minBlockLines)...)
	case *ast.IfStmt:
		blocks = append(blocks, da.extractFromIfStmt(s, filePath, minBlockLines)...)
	case *ast.ForStmt:
		blocks = append(blocks, da.extractFromLoopBody(s.Body, filePath, minBlockLines)...)
	case *ast.RangeStmt:
		blocks = append(blocks, da.extractFromLoopBody(s.Body, filePath, minBlockLines)...)
	case *ast.SwitchStmt:
		blocks = append(blocks, da.extractFromSwitchStmt(s, filePath, minBlockLines)...)
	case *ast.TypeSwitchStmt:
		blocks = append(blocks, da.extractFromTypeSwitchStmt(s, filePath, minBlockLines)...)
	case *ast.SelectStmt:
		blocks = append(blocks, da.extractFromSelectStmt(s, filePath, minBlockLines)...)
	}

	return blocks
}

// extractFromIfStmt extracts statement blocks from if statement body and else clause.
func (da *DuplicationAnalyzer) extractFromIfStmt(s *ast.IfStmt, filePath string, minBlockLines int) []StatementBlock {
	var blocks []StatementBlock
	if s.Body != nil {
		blocks = append(blocks, da.extractBlocksFromStmtList(s.Body.List, filePath, minBlockLines)...)
	}
	if s.Else != nil {
		blocks = append(blocks, da.extractNestedBlocks(s.Else, filePath, minBlockLines)...)
	}
	return blocks
}

// extractFromLoopBody extracts statement blocks from loop body statements.
func (da *DuplicationAnalyzer) extractFromLoopBody(body *ast.BlockStmt, filePath string, minBlockLines int) []StatementBlock {
	if body == nil {
		return nil
	}
	return da.extractBlocksFromStmtList(body.List, filePath, minBlockLines)
}

// extractFromSwitchStmt extracts statement blocks from switch statement case clauses.
func (da *DuplicationAnalyzer) extractFromSwitchStmt(s *ast.SwitchStmt, filePath string, minBlockLines int) []StatementBlock {
	var blocks []StatementBlock
	if s.Body == nil {
		return blocks
	}
	for _, clause := range s.Body.List {
		if cc, ok := clause.(*ast.CaseClause); ok {
			blocks = append(blocks, da.extractBlocksFromStmtList(cc.Body, filePath, minBlockLines)...)
		}
	}
	return blocks
}

// extractFromTypeSwitchStmt extracts statement blocks from type switch statement case clauses.
func (da *DuplicationAnalyzer) extractFromTypeSwitchStmt(s *ast.TypeSwitchStmt, filePath string, minBlockLines int) []StatementBlock {
	var blocks []StatementBlock
	if s.Body == nil {
		return blocks
	}
	for _, clause := range s.Body.List {
		if cc, ok := clause.(*ast.CaseClause); ok {
			blocks = append(blocks, da.extractBlocksFromStmtList(cc.Body, filePath, minBlockLines)...)
		}
	}
	return blocks
}

// extractFromSelectStmt extracts statement blocks from select statement communication clauses.
func (da *DuplicationAnalyzer) extractFromSelectStmt(s *ast.SelectStmt, filePath string, minBlockLines int) []StatementBlock {
	var blocks []StatementBlock
	if s.Body == nil {
		return blocks
	}
	for _, clause := range s.Body.List {
		if cc, ok := clause.(*ast.CommClause); ok {
			blocks = append(blocks, da.extractBlocksFromStmtList(cc.Body, filePath, minBlockLines)...)
		}
	}
	return blocks
}

// NormalizeBlock strips identifiers, literals, and comments to produce a structural
// NormalizeBlock compares code structure rather than specific names for clone detection.
func (da *DuplicationAnalyzer) NormalizeBlock(block StatementBlock) NormalizedBlock {
	// Protect against excessive memory usage for pathological code
	if block.NodeCount > MaxDeepCopyNodes {
		// For extremely large blocks, skip deep copy and use a simplified fingerprint
		// This prevents memory exhaustion on blocks with >10,000 nodes
		return NormalizedBlock{
			Structure: fmt.Sprintf("LARGE_BLOCK_%d_nodes", block.NodeCount),
			NodeCount: block.NodeCount,
		}
	}

	var buf bytes.Buffer

	// Create a normalized version of each statement
	for _, stmt := range block.Statements {
		normalized := da.normalizeNode(stmt)
		// Print the normalized AST to a canonical string form
		if err := printer.Fprint(&buf, da.fset, normalized); err == nil {
			buf.WriteString("\n")
		}
	}

	return NormalizedBlock{
		Structure: buf.String(),
		NodeCount: block.NodeCount,
	}
}

// normalizeNode creates a normalized copy of an AST node
func (da *DuplicationAnalyzer) normalizeNode(node ast.Node) ast.Node {
	// Create a deep copy with normalization
	var normalized ast.Node

	switch n := node.(type) {
	case *ast.Ident:
		// Replace all identifiers with placeholder
		return &ast.Ident{Name: "_"}
	case *ast.BasicLit:
		// Replace literals with type-specific placeholders
		return da.normalizeLiteral(n)
	default:
		// For other nodes, recursively normalize children
		normalized = da.deepCopyAndNormalize(n)
	}

	return normalized
}

// normalizeLiteral replaces literals with type-specific placeholders
func (da *DuplicationAnalyzer) normalizeLiteral(lit *ast.BasicLit) *ast.BasicLit {
	var value string
	switch lit.Kind {
	case token.INT:
		value = "INT_"
	case token.FLOAT:
		value = "FLOAT_"
	case token.IMAG:
		value = "IMAG_"
	case token.CHAR:
		value = "CHAR_"
	case token.STRING:
		value = "STRING_"
	default:
		value = "LITERAL_"
	}
	return &ast.BasicLit{Kind: lit.Kind, Value: value}
}

// deepCopyAndNormalize performs a deep copy while normalizing identifiers and literals
func (da *DuplicationAnalyzer) deepCopyAndNormalize(node ast.Node) ast.Node {
	switch n := node.(type) {
	case *ast.BlockStmt:
		return da.normalizeBlockStmt(n)
	case *ast.ExprStmt:
		return da.normalizeExprStmt(n)
	case *ast.AssignStmt:
		return da.normalizeAssignStmt(n)
	case *ast.IfStmt:
		return da.normalizeIfStmt(n)
	case *ast.ForStmt:
		return da.normalizeForStmt(n)
	case *ast.RangeStmt:
		return da.normalizeRangeStmt(n)
	case *ast.ReturnStmt:
		return da.normalizeReturnStmt(n)
	case *ast.CallExpr:
		return da.normalizeCallExpr(n)
	case *ast.BinaryExpr:
		return da.normalizeBinaryExpr(n)
	case *ast.UnaryExpr:
		return da.normalizeUnaryExpr(n)
	case *ast.SelectorExpr:
		return da.normalizeSelectorExpr(n)
	case *ast.IndexExpr:
		return da.normalizeIndexExpr(n)
	case *ast.StarExpr:
		return &ast.StarExpr{X: da.normalizeNode(n.X).(ast.Expr)}
	case *ast.ParenExpr:
		return &ast.ParenExpr{X: da.normalizeNode(n.X).(ast.Expr)}
	default:
		return node
	}
}

// normalizeBlockStmt normalizes a block statement by recursively normalizing all statements within
func (da *DuplicationAnalyzer) normalizeBlockStmt(n *ast.BlockStmt) ast.Node {
	stmts := make([]ast.Stmt, len(n.List))
	for i, stmt := range n.List {
		stmts[i] = da.normalizeNode(stmt).(ast.Stmt)
	}
	return &ast.BlockStmt{List: stmts}
}

// normalizeExprStmt normalizes an expression statement by normalizing its expression
func (da *DuplicationAnalyzer) normalizeExprStmt(n *ast.ExprStmt) ast.Node {
	return &ast.ExprStmt{X: da.normalizeNode(n.X).(ast.Expr)}
}

// normalizeAssignStmt normalizes an assignment statement by normalizing both LHS and RHS expressions
func (da *DuplicationAnalyzer) normalizeAssignStmt(n *ast.AssignStmt) ast.Node {
	lhs := make([]ast.Expr, len(n.Lhs))
	for i, expr := range n.Lhs {
		lhs[i] = da.normalizeNode(expr).(ast.Expr)
	}
	rhs := make([]ast.Expr, len(n.Rhs))
	for i, expr := range n.Rhs {
		rhs[i] = da.normalizeNode(expr).(ast.Expr)
	}
	return &ast.AssignStmt{Lhs: lhs, Tok: n.Tok, Rhs: rhs}
}

// normalizeIfStmt normalizes an if statement by normalizing its condition, body, and else branches
func (da *DuplicationAnalyzer) normalizeIfStmt(n *ast.IfStmt) ast.Node {
	var init ast.Stmt
	if n.Init != nil {
		init = da.normalizeNode(n.Init).(ast.Stmt)
	}
	var elseBranch ast.Stmt
	if n.Else != nil {
		elseBranch = da.normalizeNode(n.Else).(ast.Stmt)
	}
	return &ast.IfStmt{
		Init: init,
		Cond: da.normalizeNode(n.Cond).(ast.Expr),
		Body: da.normalizeNode(n.Body).(*ast.BlockStmt),
		Else: elseBranch,
	}
}

// normalizeForStmt normalizes a for loop statement by normalizing init, condition, post, and body
func (da *DuplicationAnalyzer) normalizeForStmt(n *ast.ForStmt) ast.Node {
	var init, post ast.Stmt
	var cond ast.Expr
	if n.Init != nil {
		init = da.normalizeNode(n.Init).(ast.Stmt)
	}
	if n.Cond != nil {
		cond = da.normalizeNode(n.Cond).(ast.Expr)
	}
	if n.Post != nil {
		post = da.normalizeNode(n.Post).(ast.Stmt)
	}
	return &ast.ForStmt{
		Init: init,
		Cond: cond,
		Post: post,
		Body: da.normalizeNode(n.Body).(*ast.BlockStmt),
	}
}

// normalizeRangeStmt normalizes a range statement by normalizing key, value, and range expression
func (da *DuplicationAnalyzer) normalizeRangeStmt(n *ast.RangeStmt) ast.Node {
	var key, value ast.Expr
	if n.Key != nil {
		key = da.normalizeNode(n.Key).(ast.Expr)
	}
	if n.Value != nil {
		value = da.normalizeNode(n.Value).(ast.Expr)
	}
	return &ast.RangeStmt{
		Key:   key,
		Value: value,
		Tok:   n.Tok,
		X:     da.normalizeNode(n.X).(ast.Expr),
		Body:  da.normalizeNode(n.Body).(*ast.BlockStmt),
	}
}

// normalizeReturnStmt normalizes a return statement by normalizing all return value expressions
func (da *DuplicationAnalyzer) normalizeReturnStmt(n *ast.ReturnStmt) ast.Node {
	results := make([]ast.Expr, len(n.Results))
	for i, expr := range n.Results {
		results[i] = da.normalizeNode(expr).(ast.Expr)
	}
	return &ast.ReturnStmt{Results: results}
}

// normalizeCallExpr normalizes a call expression by normalizing function and argument expressions
func (da *DuplicationAnalyzer) normalizeCallExpr(n *ast.CallExpr) ast.Node {
	args := make([]ast.Expr, len(n.Args))
	for i, arg := range n.Args {
		args[i] = da.normalizeNode(arg).(ast.Expr)
	}
	return &ast.CallExpr{
		Fun:      da.normalizeNode(n.Fun).(ast.Expr),
		Args:     args,
		Ellipsis: n.Ellipsis,
	}
}

// normalizeBinaryExpr normalizes a binary expression by normalizing operands while preserving operator
func (da *DuplicationAnalyzer) normalizeBinaryExpr(n *ast.BinaryExpr) ast.Node {
	return &ast.BinaryExpr{
		X:  da.normalizeNode(n.X).(ast.Expr),
		Op: n.Op,
		Y:  da.normalizeNode(n.Y).(ast.Expr),
	}
}

// normalizeUnaryExpr normalizes a unary expression by normalizing operand while preserving operator
func (da *DuplicationAnalyzer) normalizeUnaryExpr(n *ast.UnaryExpr) ast.Node {
	return &ast.UnaryExpr{
		Op: n.Op,
		X:  da.normalizeNode(n.X).(ast.Expr),
	}
}

// normalizeSelectorExpr normalizes a selector expression by replacing field names with placeholders
func (da *DuplicationAnalyzer) normalizeSelectorExpr(n *ast.SelectorExpr) ast.Node {
	return &ast.SelectorExpr{
		X:   da.normalizeNode(n.X).(ast.Expr),
		Sel: &ast.Ident{Name: "_"},
	}
}

// normalizeIndexExpr normalizes an index expression by normalizing both array and index expressions
func (da *DuplicationAnalyzer) normalizeIndexExpr(n *ast.IndexExpr) ast.Node {
	return &ast.IndexExpr{
		X:     da.normalizeNode(n.X).(ast.Expr),
		Index: da.normalizeNode(n.Index).(ast.Expr),
	}
}

// ComputeHash computes a structural hash using FNV-1a for fast duplicate lookup.
func (da *DuplicationAnalyzer) ComputeHash(normalized NormalizedBlock) string {
	h := fnv.New64a()
	h.Write([]byte(normalized.Structure))
	return fmt.Sprintf("%016x", h.Sum64())
}

// FingerprintBlocks creates fingerprints for all blocks by normalizing and hashing.
func (da *DuplicationAnalyzer) FingerprintBlocks(blocks []StatementBlock) []BlockFingerprint {
	fingerprints := make([]BlockFingerprint, 0, len(blocks))

	for _, block := range blocks {
		normalized := da.NormalizeBlock(block)
		hash := da.ComputeHash(normalized)

		fingerprint := BlockFingerprint{
			Hash:      hash,
			File:      block.File,
			StartLine: block.StartLine,
			EndLine:   block.EndLine,
			NodeCount: block.NodeCount,
			Original:  block,
		}

		fingerprints = append(fingerprints, fingerprint)
	}

	return fingerprints
}

// GroupFingerprintsByHash groups fingerprints by their hash value for clone detection.
func (da *DuplicationAnalyzer) GroupFingerprintsByHash(fingerprints []BlockFingerprint) map[string][]BlockFingerprint {
	groups := make(map[string][]BlockFingerprint)

	for _, fp := range fingerprints {
		groups[fp.Hash] = append(groups[fp.Hash], fp)
	}

	return groups
}

// FilterDuplicateGroups returns only groups with 2 or more instances
func (da *DuplicationAnalyzer) FilterDuplicateGroups(groups map[string][]BlockFingerprint) map[string][]BlockFingerprint {
	duplicates := make(map[string][]BlockFingerprint)

	for hash, group := range groups {
		if len(group) >= 2 {
			// Sort by file and line for consistent ordering
			sort.Slice(group, func(i, j int) bool {
				if group[i].File != group[j].File {
					return group[i].File < group[j].File
				}
				return group[i].StartLine < group[j].StartLine
			})
			duplicates[hash] = group
		}
	}

	return duplicates
}

// GetBlockSource returns the source code for a statement block
func (da *DuplicationAnalyzer) GetBlockSource(block StatementBlock) (string, error) {
	var buf bytes.Buffer
	for _, stmt := range block.Statements {
		if err := printer.Fprint(&buf, da.fset, stmt); err != nil {
			return "", err
		}
		buf.WriteString("\n")
	}
	return strings.TrimSpace(buf.String()), nil
}

// DetectClonePairs groups fingerprints by hash and identifies groups with 2+ entries.
// DetectClonePairs returns classified clone pairs sorted by line count.
func (da *DuplicationAnalyzer) DetectClonePairs(fingerprints []BlockFingerprint, similarityThreshold float64) []metrics.ClonePair {
	// Group fingerprints by hash
	groups := da.GroupFingerprintsByHash(fingerprints)
	duplicates := da.FilterDuplicateGroups(groups)

	// Convert to ClonePair format
	var clonePairs []metrics.ClonePair

	for hash, group := range duplicates {
		// Create instances from the group
		instances := make([]metrics.CloneInstance, len(group))
		lineCount := 0

		for i, fp := range group {
			instances[i] = metrics.CloneInstance{
				File:      fp.File,
				StartLine: fp.StartLine,
				EndLine:   fp.EndLine,
				NodeCount: fp.NodeCount,
			}
			// Calculate line count (use first instance as representative)
			if i == 0 {
				lineCount = fp.EndLine - fp.StartLine + 1
			}
		}

		// Create the clone pair
		pair := metrics.ClonePair{
			Hash:      hash,
			Type:      metrics.CloneTypeExact, // Default to exact, will be classified later
			Instances: instances,
			LineCount: lineCount,
		}

		// Classify the clone type
		pair.Type = da.ClassifyClone(pair, group, similarityThreshold)

		clonePairs = append(clonePairs, pair)
	}

	// Sort by line count (descending) for consistent ordering
	sort.Slice(clonePairs, func(i, j int) bool {
		return clonePairs[i].LineCount > clonePairs[j].LineCount
	})

	return clonePairs
}

// ClassifyClone determines the clone type (exact, renamed, or near-clone) based
// ClassifyClone compares source and similarity threshold.
func (da *DuplicationAnalyzer) ClassifyClone(pair metrics.ClonePair, group []BlockFingerprint, threshold float64) metrics.CloneType {
	if len(group) < 2 {
		return metrics.CloneTypeExact
	}

	// Get the original source for the first two instances
	source1, err1 := da.GetBlockSource(group[0].Original)
	source2, err2 := da.GetBlockSource(group[1].Original)

	if err1 != nil || err2 != nil {
		return metrics.CloneTypeExact
	}

	// Type 1: Exact duplicates (identical after whitespace normalization)
	if normalizeWhitespace(source1) == normalizeWhitespace(source2) {
		return metrics.CloneTypeExact
	}

	// Type 2: Renamed duplicates (identical structure, different identifiers)
	// Since we already have the same hash, and it's not exact, it's renamed
	// The hash is based on normalized structure, so same hash means same structure

	// Compute similarity to distinguish Type 2 from Type 3
	normalized1 := da.NormalizeBlock(group[0].Original)
	normalized2 := da.NormalizeBlock(group[1].Original)

	similarity := da.ComputeSimilarity(normalized1, normalized2)

	// Type 2: Very high similarity (>= 0.95), just different identifiers
	if similarity >= 0.95 {
		return metrics.CloneTypeRenamed
	}

	// Type 3: Near duplicates (similarity above threshold)
	if similarity >= threshold {
		return metrics.CloneTypeNear
	}

	// If below threshold, still consider it as near duplicate since we found it via hashing
	return metrics.CloneTypeNear
}

// ComputeSimilarity calculates structural similarity between two normalized blocks.
// ComputeSimilarity returns a value between 0.0 and 1.0.
func (da *DuplicationAnalyzer) ComputeSimilarity(block1, block2 NormalizedBlock) float64 {
	// Use Jaccard similarity on token sets
	tokens1 := tokenize(block1.Structure)
	tokens2 := tokenize(block2.Structure)

	// Create sets
	set1 := make(map[string]bool)
	set2 := make(map[string]bool)

	for _, t := range tokens1 {
		set1[t] = true
	}
	for _, t := range tokens2 {
		set2[t] = true
	}

	// Calculate intersection and union
	intersection := 0
	for token := range set1 {
		if set2[token] {
			intersection++
		}
	}

	union := len(set1) + len(set2) - intersection

	if union == 0 {
		return 1.0
	}

	return float64(intersection) / float64(union)
}

// normalizeWhitespace removes all whitespace for comparison
func normalizeWhitespace(s string) string {
	return strings.Join(strings.Fields(s), "")
}

// tokenize splits a string into tokens for similarity comparison
func tokenize(s string) []string {
	// Split on whitespace and common delimiters
	replacer := strings.NewReplacer(
		"(", " ( ",
		")", " ) ",
		"{", " { ",
		"}", " } ",
		"[", " [ ",
		"]", " ] ",
		";", " ; ",
		",", " , ",
		".", " . ",
	)
	s = replacer.Replace(s)
	tokens := strings.Fields(s)
	return tokens
}

// AnalyzeDuplication performs duplication analysis on a collection of files.
// AnalyzeDuplication returns DuplicationMetrics summarizing all detected code clones.
func (da *DuplicationAnalyzer) AnalyzeDuplication(files map[string]*ast.File, minBlockLines int, similarityThreshold float64) metrics.DuplicationMetrics {
	// Extract blocks from all files
	var allBlocks []StatementBlock
	totalLines := 0

	for filePath, file := range files {
		blocks := da.ExtractBlocks(file, filePath, minBlockLines)
		allBlocks = append(allBlocks, blocks...)

		// Calculate total lines in this file
		if file != nil && file.End().IsValid() {
			endPos := da.fset.Position(file.End())
			totalLines += endPos.Line
		}
	}

	// Fingerprint all blocks
	fingerprints := da.FingerprintBlocks(allBlocks)

	// Detect clone pairs
	clonePairs := da.DetectClonePairs(fingerprints, similarityThreshold)

	// Calculate duplicated lines
	// Count unique duplicated lines (each set of duplicates counts once, not per instance)
	duplicatedLines := 0
	largestCloneSize := 0

	for _, pair := range clonePairs {
		// Each clone pair represents duplicated code
		// Count the size once (it represents the duplicated code block size)
		// Then multiply by (instances - 1) to get wasted lines
		// Total duplicated = original + all copies
		duplicatedLines += pair.LineCount * (len(pair.Instances) - 1)

		if pair.LineCount > largestCloneSize {
			largestCloneSize = pair.LineCount
		}
	}

	// Calculate duplication ratio
	duplicationRatio := 0.0
	if totalLines > 0 {
		duplicationRatio = float64(duplicatedLines) / float64(totalLines)
	}

	return metrics.DuplicationMetrics{
		ClonePairs:       len(clonePairs),
		DuplicatedLines:  duplicatedLines,
		DuplicationRatio: duplicationRatio,
		LargestCloneSize: largestCloneSize,
		Clones:           clonePairs,
	}
}
