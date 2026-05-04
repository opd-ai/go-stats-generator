package analyzer

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAntipatternFalsePositiveReduction verifies that the antipattern analyzer
// does not produce false positives for common patterns that were previously misidentified.
func TestAntipatternFalsePositiveReduction(t *testing.T) {
	t.Run("no false positive for append outside loop", func(t *testing.T) {
		src := `package test
func example() {
	var s []int
	s = append(s, 1)
	s = append(s, 2)
}`
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, "test.go", src, 0)
		require.NoError(t, err)

		analyzer := NewAntipatternAnalyzer(fset)
		patterns := analyzer.Analyze(file)

		for _, p := range patterns {
			assert.NotEqual(t, "memory_allocation", p.Type,
				"append outside loop should not be flagged as memory_allocation")
		}
	})

	t.Run("detects append inside for loop", func(t *testing.T) {
		src := `package test
func example() {
	var s []int
	for i := 0; i < 10; i++ {
		s = append(s, i)
	}
}`
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, "test.go", src, 0)
		require.NoError(t, err)

		analyzer := NewAntipatternAnalyzer(fset)
		patterns := analyzer.Analyze(file)

		found := false
		for _, p := range patterns {
			if p.Type == "memory_allocation" {
				found = true
			}
		}
		assert.True(t, found, "append in for loop should be detected")
	})

	t.Run("detects append inside range loop", func(t *testing.T) {
		src := `package test
func example() {
	items := []string{"a", "b", "c"}
	var result []string
	for _, item := range items {
		result = append(result, item)
	}
}`
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, "test.go", src, 0)
		require.NoError(t, err)

		analyzer := NewAntipatternAnalyzer(fset)
		patterns := analyzer.Analyze(file)

		found := false
		for _, p := range patterns {
			if p.Type == "memory_allocation" {
				found = true
			}
		}
		assert.True(t, found, "append in range loop should be detected")
	})

	t.Run("no false positive for string concatenation outside loop", func(t *testing.T) {
		src := `package test
func example() string {
	return "hello" + " " + "world"
}`
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, "test.go", src, 0)
		require.NoError(t, err)

		analyzer := NewAntipatternAnalyzer(fset)
		patterns := analyzer.Analyze(file)

		for _, p := range patterns {
			assert.NotEqual(t, "string_concatenation", p.Type,
				"string concatenation outside loop should not be flagged")
		}
	})

	t.Run("detects string concatenation inside loop", func(t *testing.T) {
		src := `package test
func example() string {
	s := ""
	for i := 0; i < 10; i++ {
		s = s + "x"
	}
	return s
}`
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, "test.go", src, 0)
		require.NoError(t, err)

		analyzer := NewAntipatternAnalyzer(fset)
		patterns := analyzer.Analyze(file)

		found := false
		for _, p := range patterns {
			if p.Type == "string_concatenation" {
				found = true
			}
		}
		assert.True(t, found, "string concatenation in loop should be detected")
	})

	t.Run("no false positive for resource with defer close", func(t *testing.T) {
		src := `package test
import "os"
func example() {
	f, _ := os.Open("test.txt")
	defer f.Close()
}`
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, "test.go", src, 0)
		require.NoError(t, err)

		analyzer := NewAntipatternAnalyzer(fset)
		patterns := analyzer.Analyze(file)

		for _, p := range patterns {
			assert.NotEqual(t, "resource_leak", p.Type,
				"resource with defer close should not be flagged as leak")
		}
	})

	t.Run("detects resource without defer close", func(t *testing.T) {
		src := `package test
import "os"
func example() {
	f, _ := os.Open("test.txt")
	_ = f
}`
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, "test.go", src, 0)
		require.NoError(t, err)

		analyzer := NewAntipatternAnalyzer(fset)
		patterns := analyzer.Analyze(file)

		found := false
		for _, p := range patterns {
			if p.Type == "resource_leak" {
				found = true
			}
		}
		assert.True(t, found, "resource without defer close should be detected as leak")
	})

	t.Run("no false positive for goroutine with context param", func(t *testing.T) {
		src := `package test
import "context"
func example(ctx context.Context) {
	go func(ctx context.Context) {
		_ = ctx
	}(ctx)
}`
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, "test.go", src, 0)
		require.NoError(t, err)

		analyzer := NewAntipatternAnalyzer(fset)
		patterns := analyzer.Analyze(file)

		for _, p := range patterns {
			assert.NotEqual(t, "goroutine_leak", p.Type,
				"goroutine with context should not be flagged as potential leak")
		}
	})

	t.Run("no false positive for goroutine with done channel closure", func(t *testing.T) {
		src := `package test
func example() {
	done := make(chan struct{})
	go func() {
		<-done
	}()
}`
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, "test.go", src, 0)
		require.NoError(t, err)

		analyzer := NewAntipatternAnalyzer(fset)
		patterns := analyzer.Analyze(file)

		for _, p := range patterns {
			assert.NotEqual(t, "goroutine_leak", p.Type,
				"goroutine using done channel via closure should not be flagged")
		}
	})

	t.Run("no false positive for goroutine with select", func(t *testing.T) {
		src := `package test
func example() {
	ch := make(chan int)
	go func() {
		select {
		case v := <-ch:
			_ = v
		}
	}()
}`
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, "test.go", src, 0)
		require.NoError(t, err)

		analyzer := NewAntipatternAnalyzer(fset)
		patterns := analyzer.Analyze(file)

		for _, p := range patterns {
			assert.NotEqual(t, "goroutine_leak", p.Type,
				"goroutine with select should not be flagged (select implies channel coordination)")
		}
	})
}

// TestPatternFalsePositiveReduction verifies that pattern detection does not
// produce false positives from overly broad heuristics.
func TestPatternFalsePositiveReduction(t *testing.T) {
	t.Run("no false positive strategy for struct with non-interface er-suffixed fields", func(t *testing.T) {
		// A struct with fields named like "Reader" (ending in "er") that are NOT interfaces
		// should NOT be detected as strategy pattern
		src := `package test
type MyStruct struct {
	Reader  string
	Writer  string
	Manager string
}
`
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, "test.go", src, 0)
		require.NoError(t, err)

		analyzer := NewPatternAnalyzer(fset)
		patterns, err := analyzer.AnalyzePatterns(file, "test", "test.go")
		require.NoError(t, err)

		assert.Empty(t, patterns.Strategy,
			"struct with string fields ending in 'er' should not be detected as strategy")
	})

	t.Run("no false positive factory for New function returning concrete type", func(t *testing.T) {
		src := `package test
type Config struct {
	Name string
}
func NewConfig() *Config {
	return &Config{}
}
`
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, "test.go", src, 0)
		require.NoError(t, err)

		analyzer := NewPatternAnalyzer(fset)
		patterns, err := analyzer.AnalyzePatterns(file, "test", "test.go")
		require.NoError(t, err)

		assert.Empty(t, patterns.Factory,
			"New function returning concrete type should not be detected as factory")
	})

	t.Run("detects strategy with actual interface field", func(t *testing.T) {
		src := `package test
type Strategy interface {
	Execute()
}
type Context struct {
	strategy Strategy
}
`
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, "test.go", src, 0)
		require.NoError(t, err)

		analyzer := NewPatternAnalyzer(fset)
		patterns, err := analyzer.AnalyzePatterns(file, "test", "test.go")
		require.NoError(t, err)

		assert.NotEmpty(t, patterns.Strategy,
			"struct with interface field should still be detected as strategy")
	})
}

// TestConcurrencyFalsePositiveReduction verifies that concurrency analysis
// does not produce false positives from broken detection methods.
func TestConcurrencyFalsePositiveReduction(t *testing.T) {
	t.Run("parseIntLiteral handles multi-digit numbers", func(t *testing.T) {
		fset := token.NewFileSet()
		ca := NewConcurrencyAnalyzer(fset)

		assert.Equal(t, 100, ca.parseIntLiteral("100"))
		assert.Equal(t, 256, ca.parseIntLiteral("256"))
		assert.Equal(t, 1024, ca.parseIntLiteral("1024"))
		assert.Equal(t, 0, ca.parseIntLiteral(""))
		assert.Equal(t, 5, ca.parseIntLiteral("5"))
	})

	t.Run("hasInfiniteLoop detects for block", func(t *testing.T) {
		src := `package test
func example() {
	go func() {
		for {
			// infinite loop
		}
	}()
}`
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, "test.go", src, 0)
		require.NoError(t, err)

		ca := NewConcurrencyAnalyzer(fset)
		result, err := ca.AnalyzeConcurrency(file, "test")
		require.NoError(t, err)

		assert.NotEmpty(t, result.Goroutines.GoroutineLeaks,
			"goroutine with infinite loop should produce a leak warning")
	})

	t.Run("no false leak for goroutine without infinite loop", func(t *testing.T) {
		src := `package test
func example() {
	go func() {
		doWork()
	}()
}
func doWork() {}
`
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, "test.go", src, 0)
		require.NoError(t, err)

		ca := NewConcurrencyAnalyzer(fset)
		result, err := ca.AnalyzeConcurrency(file, "test")
		require.NoError(t, err)

		assert.Empty(t, result.Goroutines.GoroutineLeaks,
			"goroutine without infinite loop should not produce leak warning")
	})

	t.Run("containsDefer detects defer in function literal", func(t *testing.T) {
		src := `package test
import "sync"
func example() {
	var mu sync.Mutex
	go func() {
		defer mu.Unlock()
		mu.Lock()
	}()
}`
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, "test.go", src, 0)
		require.NoError(t, err)

		ca := NewConcurrencyAnalyzer(fset)
		result, err := ca.AnalyzeConcurrency(file, "test")
		require.NoError(t, err)

		// The goroutine should have HasDefer = true
		for _, inst := range result.Goroutines.Instances {
			if inst.IsAnonymous {
				assert.True(t, inst.HasDefer, "anonymous goroutine with defer should have HasDefer=true")
			}
		}
	})
}

// TestDocQualityFalsePositiveReduction verifies that documentation quality
// scoring does not produce false positives.
func TestDocQualityFalsePositiveReduction(t *testing.T) {
	t.Run("comment with // does not trigger HasExample", func(t *testing.T) {
		src := `package test
// This function does something. // note: internal detail
func example() {}
`
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, "test.go", src, parser.ParseComments)
		require.NoError(t, err)

		fn := file.Decls[0].(*ast.FuncDecl)
		result := AnalyzeDocumentation(fn.Doc, nil)
		assert.False(t, result.HasExample,
			"documentation with // inside comment should not be flagged as having an example")
	})

	t.Run("comment with Example keyword triggers HasExample", func(t *testing.T) {
		src := `package test
// This function does something. Example: call example()
func example() {}
`
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, "test.go", src, parser.ParseComments)
		require.NoError(t, err)

		fn := file.Decls[0].(*ast.FuncDecl)
		result := AnalyzeDocumentation(fn.Doc, nil)
		assert.True(t, result.HasExample,
			"documentation with 'Example' keyword should be flagged as having an example")
	})
}

// TestBurdenFalsePositiveReduction verifies that burden analysis does not
// produce false positives from insufficient detection.
func TestBurdenFalsePositiveReduction(t *testing.T) {
	t.Run("benign numbers not flagged", func(t *testing.T) {
		src := `package test
func example() {
	a := 0
	b := 1
	c := 2
	d := 10
	e := 100
	f := 1024
	g := 256
	_ = a + b + c + d + e + f + g
}`
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, "test.go", src, 0)
		require.NoError(t, err)

		analyzer := NewBurdenAnalyzer(fset)
		result := analyzer.DetectMagicNumbers(file, "test")

		assert.Empty(t, result, "common benign numbers should not be flagged as magic numbers")
	})

	t.Run("non-benign numbers still flagged", func(t *testing.T) {
		src := `package test
func example() {
	x := 42
	y := 7
	_ = x + y
}`
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, "test.go", src, 0)
		require.NoError(t, err)

		analyzer := NewBurdenAnalyzer(fset)
		result := analyzer.DetectMagicNumbers(file, "test")

		assert.GreaterOrEqual(t, len(result), 1, "non-benign numbers should still be flagged")
	})

	t.Run("log.Fatal detected as terminating", func(t *testing.T) {
		src := `package test
import "log"
func Example() {
	log.Fatal("error")
	x := 42
	_ = x
}`
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, "test.go", src, 0)
		require.NoError(t, err)

		analyzer := NewBurdenAnalyzer(fset)
		result := analyzer.DetectDeadCode([]*ast.File{file}, "test")

		assert.NotEmpty(t, result.UnreachableCode,
			"code after log.Fatal should be detected as unreachable")
		if len(result.UnreachableCode) > 0 {
			assert.Equal(t, "log.Fatal call", result.UnreachableCode[0].Reason)
		}
	})

	t.Run("method calls tracked in reference map", func(t *testing.T) {
		src := `package test
type helper struct{}
func (h *helper) process() {}
func Example() {
	h := &helper{}
	h.process()
}`
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, "test.go", src, 0)
		require.NoError(t, err)

		analyzer := NewBurdenAnalyzer(fset)
		result := analyzer.DetectDeadCode([]*ast.File{file}, "test")

		// "process" is called from exported Example, so it must not be flagged as unreferenced.
		for _, sym := range result.UnreferencedFunctions {
			assert.NotEqual(t, "process", sym.Name,
				"method called from an exported function should not be flagged as unreferenced")
		}
	})

	t.Run("method calls only from dead code are themselves dead", func(t *testing.T) {
		src := `package test
type helper struct{}
func (h *helper) process() {}
func unexportedCaller() {
	h := &helper{}
	h.process()
}`
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, "test.go", src, 0)
		require.NoError(t, err)

		analyzer := NewBurdenAnalyzer(fset)
		result := analyzer.DetectDeadCode([]*ast.File{file}, "test")

		// Both unexportedCaller and process are unreachable from any exported function.
		names := make(map[string]bool)
		for _, sym := range result.UnreferencedFunctions {
			names[sym.Name] = true
		}
		assert.True(t, names["unexportedCaller"],
			"unexportedCaller (never called from exported code) should be flagged as unreferenced")
		assert.True(t, names["process"],
			"process (only called from dead unexportedCaller) should also be flagged as unreferenced")
	})
}
