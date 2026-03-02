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

// DuplicationAnalyzer performs code duplication detection using AST fingerprinting
type DuplicationAnalyzer struct {
	fset *token.FileSet
}

// NewDuplicationAnalyzer creates a new duplication analyzer
func NewDuplicationAnalyzer(fset *token.FileSet) *DuplicationAnalyzer {
	return &DuplicationAnalyzer{
		fset: fset,
	}
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

// ExtractBlocks walks function and method bodies to extract statement-level sub-trees
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
			nodeCount := da.countNodes(blockStmts)

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
		if s.Body != nil {
			blocks = append(blocks, da.extractBlocksFromStmtList(s.Body.List, filePath, minBlockLines)...)
		}
		if s.Else != nil {
			blocks = append(blocks, da.extractNestedBlocks(s.Else, filePath, minBlockLines)...)
		}
	case *ast.ForStmt:
		if s.Body != nil {
			blocks = append(blocks, da.extractBlocksFromStmtList(s.Body.List, filePath, minBlockLines)...)
		}
	case *ast.RangeStmt:
		if s.Body != nil {
			blocks = append(blocks, da.extractBlocksFromStmtList(s.Body.List, filePath, minBlockLines)...)
		}
	case *ast.SwitchStmt:
		if s.Body != nil {
			for _, clause := range s.Body.List {
				if cc, ok := clause.(*ast.CaseClause); ok {
					blocks = append(blocks, da.extractBlocksFromStmtList(cc.Body, filePath, minBlockLines)...)
				}
			}
		}
	case *ast.TypeSwitchStmt:
		if s.Body != nil {
			for _, clause := range s.Body.List {
				if cc, ok := clause.(*ast.CaseClause); ok {
					blocks = append(blocks, da.extractBlocksFromStmtList(cc.Body, filePath, minBlockLines)...)
				}
			}
		}
	case *ast.SelectStmt:
		if s.Body != nil {
			for _, clause := range s.Body.List {
				if cc, ok := clause.(*ast.CommClause); ok {
					blocks = append(blocks, da.extractBlocksFromStmtList(cc.Body, filePath, minBlockLines)...)
				}
			}
		}
	}

	return blocks
}

// countNodes counts the number of AST nodes in a set of statements
func (da *DuplicationAnalyzer) countNodes(stmts []ast.Stmt) int {
	count := 0
	for _, stmt := range stmts {
		ast.Inspect(stmt, func(n ast.Node) bool {
			if n != nil {
				count++
			}
			return true
		})
	}
	return count
}

// NormalizeBlock strips identifiers, literals, and comments to produce structural form
func (da *DuplicationAnalyzer) NormalizeBlock(block StatementBlock) NormalizedBlock {
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
	// Use ast.Inspect to walk and copy the tree
	var result ast.Node

	switch n := node.(type) {
	case *ast.BlockStmt:
		stmts := make([]ast.Stmt, len(n.List))
		for i, stmt := range n.List {
			stmts[i] = da.normalizeNode(stmt).(ast.Stmt)
		}
		result = &ast.BlockStmt{List: stmts}

	case *ast.ExprStmt:
		result = &ast.ExprStmt{X: da.normalizeNode(n.X).(ast.Expr)}

	case *ast.AssignStmt:
		lhs := make([]ast.Expr, len(n.Lhs))
		for i, expr := range n.Lhs {
			lhs[i] = da.normalizeNode(expr).(ast.Expr)
		}
		rhs := make([]ast.Expr, len(n.Rhs))
		for i, expr := range n.Rhs {
			rhs[i] = da.normalizeNode(expr).(ast.Expr)
		}
		result = &ast.AssignStmt{Lhs: lhs, Tok: n.Tok, Rhs: rhs}

	case *ast.IfStmt:
		var init ast.Stmt
		if n.Init != nil {
			init = da.normalizeNode(n.Init).(ast.Stmt)
		}
		var elseBranch ast.Stmt
		if n.Else != nil {
			elseBranch = da.normalizeNode(n.Else).(ast.Stmt)
		}
		result = &ast.IfStmt{
			Init: init,
			Cond: da.normalizeNode(n.Cond).(ast.Expr),
			Body: da.normalizeNode(n.Body).(*ast.BlockStmt),
			Else: elseBranch,
		}

	case *ast.ForStmt:
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
		result = &ast.ForStmt{
			Init: init,
			Cond: cond,
			Post: post,
			Body: da.normalizeNode(n.Body).(*ast.BlockStmt),
		}

	case *ast.RangeStmt:
		var key, value ast.Expr
		if n.Key != nil {
			key = da.normalizeNode(n.Key).(ast.Expr)
		}
		if n.Value != nil {
			value = da.normalizeNode(n.Value).(ast.Expr)
		}
		result = &ast.RangeStmt{
			Key:   key,
			Value: value,
			Tok:   n.Tok,
			X:     da.normalizeNode(n.X).(ast.Expr),
			Body:  da.normalizeNode(n.Body).(*ast.BlockStmt),
		}

	case *ast.ReturnStmt:
		results := make([]ast.Expr, len(n.Results))
		for i, expr := range n.Results {
			results[i] = da.normalizeNode(expr).(ast.Expr)
		}
		result = &ast.ReturnStmt{Results: results}

	case *ast.CallExpr:
		args := make([]ast.Expr, len(n.Args))
		for i, arg := range n.Args {
			args[i] = da.normalizeNode(arg).(ast.Expr)
		}
		result = &ast.CallExpr{
			Fun:      da.normalizeNode(n.Fun).(ast.Expr),
			Args:     args,
			Ellipsis: n.Ellipsis,
		}

	case *ast.BinaryExpr:
		result = &ast.BinaryExpr{
			X:  da.normalizeNode(n.X).(ast.Expr),
			Op: n.Op,
			Y:  da.normalizeNode(n.Y).(ast.Expr),
		}

	case *ast.UnaryExpr:
		result = &ast.UnaryExpr{
			Op: n.Op,
			X:  da.normalizeNode(n.X).(ast.Expr),
		}

	case *ast.SelectorExpr:
		result = &ast.SelectorExpr{
			X:   da.normalizeNode(n.X).(ast.Expr),
			Sel: &ast.Ident{Name: "_"},
		}

	case *ast.IndexExpr:
		result = &ast.IndexExpr{
			X:     da.normalizeNode(n.X).(ast.Expr),
			Index: da.normalizeNode(n.Index).(ast.Expr),
		}

	case *ast.StarExpr:
		result = &ast.StarExpr{X: da.normalizeNode(n.X).(ast.Expr)}

	case *ast.ParenExpr:
		result = &ast.ParenExpr{X: da.normalizeNode(n.X).(ast.Expr)}

	default:
		// For unhandled types, return the node as-is
		result = node
	}

	return result
}

// ComputeHash computes a structural hash using FNV-1a
func (da *DuplicationAnalyzer) ComputeHash(normalized NormalizedBlock) string {
	h := fnv.New64a()
	h.Write([]byte(normalized.Structure))
	return fmt.Sprintf("%016x", h.Sum64())
}

// FingerprintBlocks creates fingerprints for all blocks
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

// GroupFingerprintsByHash groups fingerprints by their hash value
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

// DetectClonePairs groups fingerprints by hash and identifies groups with 2+ entries
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

// ClassifyClone determines the clone type (Type 1/2/3)
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

// ComputeSimilarity calculates structural similarity between two normalized blocks
// Returns a value between 0.0 and 1.0
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
