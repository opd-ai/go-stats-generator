package analyzer

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/opd-ai/go-stats-generator/internal/metrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDuplicationAnalyzer_ExtractBlocks(t *testing.T) {
	tests := []struct {
		name          string
		source        string
		minBlockLines int
		wantBlocks    int
	}{
		{
			name: "simple function with single block",
			source: `package main
func simple() {
	x := 1
	y := 2
	z := 3
	a := 4
	b := 5
	c := 6
}`,
			minBlockLines: 6,
			wantBlocks:    1, // One window of 6 statements
		},
		{
			name: "function with nested if",
			source: `package main
func nested() {
	if true {
		x := 1
		y := 2
		z := 3
		a := 4
		b := 5
		c := 6
	}
}`,
			minBlockLines: 6,
			wantBlocks:    1, // One block from inside if
		},
		{
			name: "function with for loop",
			source: `package main
func loop() {
	for i := 0; i < 10; i++ {
		x := 1
		y := 2
		z := 3
		a := 4
		b := 5
		c := 6
	}
}`,
			minBlockLines: 6,
			wantBlocks:    1, // One block from inside for
		},
		{
			name: "empty function",
			source: `package main
func empty() {
}`,
			minBlockLines: 6,
			wantBlocks:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "test.go", tt.source, parser.ParseComments)
			require.NoError(t, err)

			analyzer := NewDuplicationAnalyzer(fset)
			blocks := analyzer.ExtractBlocks(file, "test.go", tt.minBlockLines)

			assert.GreaterOrEqual(t, len(blocks), tt.wantBlocks,
				"Expected at least %d blocks, got %d", tt.wantBlocks, len(blocks))
		})
	}
}

func TestDuplicationAnalyzer_NormalizeBlock(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		expected bool // whether normalization should produce same structure for different names
	}{
		{
			name: "identical structure different names",
			source: `package main
func test1() {
	x := 1
	y := 2
}
func test2() {
	a := 1
	b := 2
}`,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "test.go", tt.source, parser.ParseComments)
			require.NoError(t, err)

			analyzer := NewDuplicationAnalyzer(fset)

			// Extract blocks from both functions
			var blocks []StatementBlock
			ast.Inspect(file, func(n ast.Node) bool {
				if funcDecl, ok := n.(*ast.FuncDecl); ok && funcDecl.Body != nil {
					block := StatementBlock{
						File:       "test.go",
						StartLine:  1,
						EndLine:    2,
						Statements: funcDecl.Body.List,
						NodeCount:  len(funcDecl.Body.List),
					}
					blocks = append(blocks, block)
				}
				return true
			})

			require.Len(t, blocks, 2, "Should have extracted 2 function blocks")

			// Normalize both blocks
			norm1 := analyzer.NormalizeBlock(blocks[0])
			norm2 := analyzer.NormalizeBlock(blocks[1])

			if tt.expected {
				assert.Equal(t, norm1.Structure, norm2.Structure,
					"Normalized structures should be identical for same code with different names")
			}
		})
	}
}

func TestDuplicationAnalyzer_ComputeHash(t *testing.T) {
	tests := []struct {
		name     string
		source1  string
		source2  string
		sameHash bool
	}{
		{
			name: "identical code same hash",
			source1: `package main
func test() {
	x := 1
	y := 2
}`,
			source2: `package main
func test() {
	x := 1
	y := 2
}`,
			sameHash: true,
		},
		{
			name: "different variable names same hash",
			source1: `package main
func test() {
	x := 1
	y := 2
}`,
			source2: `package main
func test() {
	a := 1
	b := 2
}`,
			sameHash: true,
		},
		{
			name: "different literals same structure same hash",
			source1: `package main
func test() {
	x := 1
	y := 2
}`,
			source2: `package main
func test() {
	x := 100
	y := 200
}`,
			sameHash: true,
		},
		{
			name: "different structure different hash",
			source1: `package main
func test() {
	x := 1
	y := 2
}`,
			source2: `package main
func test() {
	x := 1
	y := 2
	z := 3
}`,
			sameHash: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset1 := token.NewFileSet()
			file1, err := parser.ParseFile(fset1, "test1.go", tt.source1, parser.ParseComments)
			require.NoError(t, err)

			fset2 := token.NewFileSet()
			file2, err := parser.ParseFile(fset2, "test2.go", tt.source2, parser.ParseComments)
			require.NoError(t, err)

			analyzer1 := NewDuplicationAnalyzer(fset1)
			analyzer2 := NewDuplicationAnalyzer(fset2)

			// Extract first function body from each file
			var block1, block2 StatementBlock
			ast.Inspect(file1, func(n ast.Node) bool {
				if funcDecl, ok := n.(*ast.FuncDecl); ok && funcDecl.Body != nil && len(block1.Statements) == 0 {
					block1 = StatementBlock{
						File:       "test1.go",
						Statements: funcDecl.Body.List,
						NodeCount:  len(funcDecl.Body.List),
					}
				}
				return true
			})

			ast.Inspect(file2, func(n ast.Node) bool {
				if funcDecl, ok := n.(*ast.FuncDecl); ok && funcDecl.Body != nil && len(block2.Statements) == 0 {
					block2 = StatementBlock{
						File:       "test2.go",
						Statements: funcDecl.Body.List,
						NodeCount:  len(funcDecl.Body.List),
					}
				}
				return true
			})

			norm1 := analyzer1.NormalizeBlock(block1)
			norm2 := analyzer2.NormalizeBlock(block2)

			hash1 := analyzer1.ComputeHash(norm1)
			hash2 := analyzer2.ComputeHash(norm2)

			if tt.sameHash {
				assert.Equal(t, hash1, hash2, "Hashes should match for structurally identical code")
			} else {
				assert.NotEqual(t, hash1, hash2, "Hashes should differ for structurally different code")
			}
		})
	}
}

func TestDuplicationAnalyzer_FingerprintBlocks(t *testing.T) {
	source := `package main
func test() {
	x := 1
	y := 2
	z := 3
	a := 4
	b := 5
	c := 6
}`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", source, parser.ParseComments)
	require.NoError(t, err)

	analyzer := NewDuplicationAnalyzer(fset)
	blocks := analyzer.ExtractBlocks(file, "test.go", 6)
	fingerprints := analyzer.FingerprintBlocks(blocks)

	assert.NotEmpty(t, fingerprints, "Should generate fingerprints")
	for _, fp := range fingerprints {
		assert.NotEmpty(t, fp.Hash, "Each fingerprint should have a hash")
		assert.Equal(t, "test.go", fp.File)
		assert.Greater(t, fp.NodeCount, 0)
		assert.Greater(t, fp.StartLine, 0)
		assert.GreaterOrEqual(t, fp.EndLine, fp.StartLine)
	}
}

func TestDuplicationAnalyzer_GroupFingerprintsByHash(t *testing.T) {
	fingerprints := []BlockFingerprint{
		{Hash: "abc123", File: "file1.go", StartLine: 1, EndLine: 5},
		{Hash: "abc123", File: "file2.go", StartLine: 10, EndLine: 15},
		{Hash: "def456", File: "file3.go", StartLine: 20, EndLine: 25},
		{Hash: "abc123", File: "file1.go", StartLine: 30, EndLine: 35},
	}

	fset := token.NewFileSet()
	analyzer := NewDuplicationAnalyzer(fset)
	groups := analyzer.GroupFingerprintsByHash(fingerprints)

	assert.Len(t, groups, 2, "Should have 2 distinct hash groups")
	assert.Len(t, groups["abc123"], 3, "Hash 'abc123' should have 3 instances")
	assert.Len(t, groups["def456"], 1, "Hash 'def456' should have 1 instance")
}

func TestDuplicationAnalyzer_FilterDuplicateGroups(t *testing.T) {
	groups := map[string][]BlockFingerprint{
		"abc123": {
			{Hash: "abc123", File: "file1.go", StartLine: 1, EndLine: 5},
			{Hash: "abc123", File: "file2.go", StartLine: 10, EndLine: 15},
			{Hash: "abc123", File: "file1.go", StartLine: 30, EndLine: 35},
		},
		"def456": {
			{Hash: "def456", File: "file3.go", StartLine: 20, EndLine: 25},
		},
		"ghi789": {
			{Hash: "ghi789", File: "file4.go", StartLine: 1, EndLine: 10},
			{Hash: "ghi789", File: "file5.go", StartLine: 1, EndLine: 10},
		},
	}

	fset := token.NewFileSet()
	analyzer := NewDuplicationAnalyzer(fset)
	duplicates := analyzer.FilterDuplicateGroups(groups)

	assert.Len(t, duplicates, 2, "Should filter to only groups with 2+ instances")
	assert.Contains(t, duplicates, "abc123")
	assert.Contains(t, duplicates, "ghi789")
	assert.NotContains(t, duplicates, "def456")

	// Verify sorting
	for _, group := range duplicates {
		for i := 1; i < len(group); i++ {
			prev := group[i-1]
			curr := group[i]
			if prev.File == curr.File {
				assert.LessOrEqual(t, prev.StartLine, curr.StartLine,
					"Group should be sorted by file then line")
			} else {
				assert.Less(t, prev.File, curr.File,
					"Group should be sorted by file then line")
			}
		}
	}
}

func TestDuplicationAnalyzer_GetBlockSource(t *testing.T) {
	source := `package main
func test() {
	x := 1
	y := 2
	z := x + y
}`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", source, parser.ParseComments)
	require.NoError(t, err)

	analyzer := NewDuplicationAnalyzer(fset)

	// Extract function body
	var block StatementBlock
	ast.Inspect(file, func(n ast.Node) bool {
		if funcDecl, ok := n.(*ast.FuncDecl); ok && funcDecl.Body != nil {
			block = StatementBlock{
				File:       "test.go",
				Statements: funcDecl.Body.List,
				NodeCount:  len(funcDecl.Body.List),
			}
			return false
		}
		return true
	})

	require.NotEmpty(t, block.Statements, "Should have extracted statements")

	blockSource, err := analyzer.GetBlockSource(block)
	require.NoError(t, err)
	assert.NotEmpty(t, blockSource)
	assert.Contains(t, blockSource, "x := 1")
	assert.Contains(t, blockSource, "y := 2")
	assert.Contains(t, blockSource, "z := x + y")
}

func TestDuplicationAnalyzer_NormalizeLiteral(t *testing.T) {
	tests := []struct {
		name     string
		kind     token.Token
		value    string
		expected string
	}{
		{"integer", token.INT, "42", "INT_"},
		{"float", token.FLOAT, "3.14", "FLOAT_"},
		{"string", token.STRING, `"hello"`, "STRING_"},
		{"char", token.CHAR, "'a'", "CHAR_"},
		{"imaginary", token.IMAG, "1i", "IMAG_"},
	}

	fset := token.NewFileSet()
	analyzer := NewDuplicationAnalyzer(fset)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lit := &ast.BasicLit{
				Kind:  tt.kind,
				Value: tt.value,
			}

			normalized := analyzer.normalizeLiteral(lit)
			assert.Equal(t, tt.expected, normalized.Value)
			assert.Equal(t, tt.kind, normalized.Kind)
		})
	}
}

func TestDuplicationAnalyzer_ExtractNestedBlocks(t *testing.T) {
	tests := []struct {
		name          string
		source        string
		minBlockLines int
		expectBlocks  bool
	}{
		{
			name: "nested if statement",
			source: `package main
func test() {
	if true {
		x := 1
		y := 2
		z := 3
		a := 4
		b := 5
		c := 6
	}
}`,
			minBlockLines: 6,
			expectBlocks:  true,
		},
		{
			name: "switch statement",
			source: `package main
func test() {
	switch x {
	case 1:
		a := 1
		b := 2
		c := 3
		d := 4
		e := 5
		f := 6
	}
}`,
			minBlockLines: 6,
			expectBlocks:  true,
		},
		{
			name: "select statement",
			source: `package main
func test() {
	select {
	case <-ch:
		a := 1
		b := 2
		c := 3
		d := 4
		e := 5
		f := 6
	}
}`,
			minBlockLines: 6,
			expectBlocks:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "test.go", tt.source, parser.ParseComments)
			require.NoError(t, err)

			analyzer := NewDuplicationAnalyzer(fset)
			blocks := analyzer.ExtractBlocks(file, "test.go", tt.minBlockLines)

			if tt.expectBlocks {
				assert.NotEmpty(t, blocks, "Should extract blocks from nested statements")
			}
		})
	}
}

func TestDuplicationAnalyzer_CountNodes(t *testing.T) {
	source := `package main
func test() {
	x := 1 + 2
	y := x * 3
}`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", source, parser.ParseComments)
	require.NoError(t, err)

	analyzer := NewDuplicationAnalyzer(fset)

	var stmts []ast.Stmt
	ast.Inspect(file, func(n ast.Node) bool {
		if funcDecl, ok := n.(*ast.FuncDecl); ok && funcDecl.Body != nil {
			stmts = funcDecl.Body.List
			return false
		}
		return true
	})

	require.NotEmpty(t, stmts)

	count := analyzer.countNodes(stmts)
	assert.Greater(t, count, len(stmts), "Node count should be greater than statement count due to sub-nodes")
}

func TestDuplicationAnalyzer_DeepCopyAndNormalize(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name: "function calls",
			source: `package main
func test() {
	result := doSomething(x, y, z)
}`,
		},
		{
			name: "binary expressions",
			source: `package main
func test() {
	z := x + y
	w := a * b
}`,
		},
		{
			name: "unary expressions",
			source: `package main
func test() {
	x := -y
	z := !w
}`,
		},
		{
			name: "selector expressions",
			source: `package main
func test() {
	x := obj.field
	y := pkg.Func()
}`,
		},
		{
			name: "index expressions",
			source: `package main
func test() {
	x := arr[i]
	y := m[key]
}`,
		},
		{
			name: "star and paren expressions",
			source: `package main
func test() {
	x := *ptr
	y := (a + b)
}`,
		},
		{
			name: "range statement",
			source: `package main
func test() {
	for k, v := range items {
		process(k, v)
	}
}`,
		},
		{
			name: "return statement",
			source: `package main
func test() int {
	return x + y
}`,
		},
		{
			name: "if with else",
			source: `package main
func test() {
	if x > 0 {
		doThis()
	} else {
		doThat()
	}
}`,
		},
		{
			name: "for with all parts",
			source: `package main
func test() {
	for i := 0; i < 10; i++ {
		process(i)
	}
}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "test.go", tt.source, parser.ParseComments)
			require.NoError(t, err)

			analyzer := NewDuplicationAnalyzer(fset)

			// Extract function body
			var block StatementBlock
			ast.Inspect(file, func(n ast.Node) bool {
				if funcDecl, ok := n.(*ast.FuncDecl); ok && funcDecl.Body != nil {
					block = StatementBlock{
						File:       "test.go",
						Statements: funcDecl.Body.List,
						NodeCount:  len(funcDecl.Body.List),
					}
					return false
				}
				return true
			})

			require.NotEmpty(t, block.Statements, "Should have extracted statements")

			// Test normalization doesn't panic
			normalized := analyzer.NormalizeBlock(block)
			assert.NotEmpty(t, normalized.Structure, "Should produce normalized structure")

			// Test hash computation
			hash := analyzer.ComputeHash(normalized)
			assert.NotEmpty(t, hash, "Should produce hash")
		})
	}
}

func TestDuplicationAnalyzer_DetectClonePairs(t *testing.T) {
	fset := token.NewFileSet()
	analyzer := NewDuplicationAnalyzer(fset)

	tests := []struct {
		name                string
		sources             []string
		minBlockLines       int
		similarityThreshold float64
		wantClonePairs      int
		wantMinInstances    int
	}{
		{
			name: "exact duplicates",
			sources: []string{
				`package main
func foo() {
	x := 1
	y := 2
	z := 3
}`,
				`package main
func bar() {
	x := 1
	y := 2
	z := 3
}`,
			},
			minBlockLines:       3,
			similarityThreshold: 0.80,
			wantClonePairs:      1,
			wantMinInstances:    2,
		},
		{
			name: "renamed duplicates",
			sources: []string{
				`package main
func foo() {
	count := 1
	sum := 2
	avg := 3
}`,
				`package main
func bar() {
	total := 1
	result := 2
	output := 3
}`,
			},
			minBlockLines:       3,
			similarityThreshold: 0.80,
			wantClonePairs:      1,
			wantMinInstances:    2,
		},
		{
			name: "no duplicates",
			sources: []string{
				`package main
func foo() {
	x := 1
}`,
				`package main
func bar() {
	y := 2
	z := 3
}`,
			},
			minBlockLines:       3,
			similarityThreshold: 0.80,
			wantClonePairs:      0,
			wantMinInstances:    0,
		},
		{
			name: "three instances of same block",
			sources: []string{
				`package main
func foo() {
	a := 1
	b := 2
	c := 3
}`,
				`package main
func bar() {
	a := 1
	b := 2
	c := 3
}`,
				`package main
func baz() {
	a := 1
	b := 2
	c := 3
}`,
			},
			minBlockLines:       3,
			similarityThreshold: 0.80,
			wantClonePairs:      1,
			wantMinInstances:    3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var allFingerprints []BlockFingerprint

			for i, source := range tt.sources {
				file, err := parser.ParseFile(fset, fmt.Sprintf("test%d.go", i), source, 0)
				require.NoError(t, err)

				blocks := analyzer.ExtractBlocks(file, fmt.Sprintf("test%d.go", i), tt.minBlockLines)
				fingerprints := analyzer.FingerprintBlocks(blocks)
				allFingerprints = append(allFingerprints, fingerprints...)
			}

			clonePairs := analyzer.DetectClonePairs(allFingerprints, tt.similarityThreshold)

			assert.Equal(t, tt.wantClonePairs, len(clonePairs), "Should detect correct number of clone pairs")

			if tt.wantClonePairs > 0 {
				// Verify each clone pair has at least the minimum instances
				for _, pair := range clonePairs {
					assert.GreaterOrEqual(t, len(pair.Instances), tt.wantMinInstances,
						"Clone pair should have at least %d instances", tt.wantMinInstances)
					assert.NotEmpty(t, pair.Hash, "Clone pair should have a hash")
					assert.Greater(t, pair.LineCount, 0, "Clone pair should have positive line count")
				}
			}
		})
	}
}

func TestDuplicationAnalyzer_ClassifyClone(t *testing.T) {
	fset := token.NewFileSet()
	analyzer := NewDuplicationAnalyzer(fset)

	tests := []struct {
		name      string
		sources   []string
		wantType  string
		threshold float64
	}{
		{
			name: "exact clone - identical code",
			sources: []string{
				`package main
func foo() {
	x := 1
	y := 2
	return x + y
}`,
				`package main
func bar() {
	x := 1
	y := 2
	return x + y
}`,
			},
			wantType:  "exact",
			threshold: 0.80,
		},
		{
			name: "renamed clone - different identifiers",
			sources: []string{
				`package main
func foo() {
	count := 10
	sum := 20
	return count + sum
}`,
				`package main
func bar() {
	total := 10
	result := 20
	return total + result
}`,
			},
			wantType:  "renamed", // Different identifiers, same structure
			threshold: 0.80,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var fingerprints []BlockFingerprint

			for i, source := range tt.sources {
				file, err := parser.ParseFile(fset, fmt.Sprintf("test%d.go", i), source, 0)
				require.NoError(t, err)

				blocks := analyzer.ExtractBlocks(file, fmt.Sprintf("test%d.go", i), 3)
				fps := analyzer.FingerprintBlocks(blocks)
				fingerprints = append(fingerprints, fps...)
			}

			if len(fingerprints) >= 2 {
				// Group by hash to get duplicates
				groups := analyzer.GroupFingerprintsByHash(fingerprints)
				for _, group := range groups {
					if len(group) >= 2 {
						// Create a sample ClonePair for classification
						instances := make([]metrics.CloneInstance, len(group))
						for i, fp := range group {
							instances[i] = metrics.CloneInstance{
								File:      fp.File,
								StartLine: fp.StartLine,
								EndLine:   fp.EndLine,
								NodeCount: fp.NodeCount,
							}
						}

						pair := metrics.ClonePair{
							Hash:      group[0].Hash,
							Instances: instances,
							LineCount: group[0].EndLine - group[0].StartLine + 1,
						}

						cloneType := analyzer.ClassifyClone(pair, group, tt.threshold)
						assert.Equal(t, tt.wantType, string(cloneType),
							"Should classify as %s clone type", tt.wantType)
					}
				}
			}
		})
	}
}

func TestDuplicationAnalyzer_ComputeSimilarity(t *testing.T) {
	fset := token.NewFileSet()
	analyzer := NewDuplicationAnalyzer(fset)

	tests := []struct {
		name          string
		source1       string
		source2       string
		minSimilarity float64
		maxSimilarity float64
	}{
		{
			name: "identical structures",
			source1: `package main
func foo() {
	x := 1
	y := 2
	z := x + y
}`,
			source2: `package main
func bar() {
	a := 1
	b := 2
	c := a + b
}`,
			minSimilarity: 0.90,
			maxSimilarity: 1.00,
		},
		{
			name: "completely different structures",
			source1: `package main
func foo() {
	x := 1
	return x
}`,
			source2: `package main
func bar() {
	for i := 0; i < 10; i++ {
		fmt.Println(i)
	}
}`,
			minSimilarity: 0.0,
			maxSimilarity: 0.5,
		},
		{
			name: "similar but not identical",
			source1: `package main
func foo() {
	x := 1
	y := 2
	z := 3
	return x + y + z
}`,
			source2: `package main
func bar() {
	a := 1
	b := 2
	return a + b
}`,
			minSimilarity: 0.40,
			maxSimilarity: 1.00, // After normalization, these can be very similar
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file1, err := parser.ParseFile(fset, "test1.go", tt.source1, 0)
			require.NoError(t, err)

			file2, err := parser.ParseFile(fset, "test2.go", tt.source2, 0)
			require.NoError(t, err)

			blocks1 := analyzer.ExtractBlocks(file1, "test1.go", 1)
			blocks2 := analyzer.ExtractBlocks(file2, "test2.go", 1)

			if len(blocks1) > 0 && len(blocks2) > 0 {
				norm1 := analyzer.NormalizeBlock(blocks1[0])
				norm2 := analyzer.NormalizeBlock(blocks2[0])

				similarity := analyzer.ComputeSimilarity(norm1, norm2)

				assert.GreaterOrEqual(t, similarity, tt.minSimilarity,
					"Similarity should be at least %.2f", tt.minSimilarity)
				assert.LessOrEqual(t, similarity, tt.maxSimilarity,
					"Similarity should be at most %.2f", tt.maxSimilarity)
			}
		})
	}
}

func TestNormalizeWhitespace(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "removes all whitespace",
			input: "x := 1\n\ty := 2",
			want:  "x:=1y:=2",
		},
		{
			name:  "handles multiple spaces",
			input: "x   :=   1",
			want:  "x:=1",
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeWhitespace(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestTokenize(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantTokens []string
	}{
		{
			name:       "simple expression",
			input:      "x := 1",
			wantTokens: []string{"x", ":=", "1"},
		},
		{
			name:       "function call",
			input:      "foo(a, b)",
			wantTokens: []string{"foo", "(", "a", ",", "b", ")"},
		},
		{
			name:       "complex statement",
			input:      "result := obj.Method(x, y)",
			wantTokens: []string{"result", ":=", "obj", ".", "Method", "(", "x", ",", "y", ")"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tokenize(tt.input)
			assert.Equal(t, tt.wantTokens, got)
		})
	}
}
