package analyzer

import (
	"go/parser"
	"go/token"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPatternAnalyzer_DetectSingleton(t *testing.T) {
	src := `package test
import "sync"

var (
	instance *Database
	once sync.Once
)

func GetInstance() *Database {
	once.Do(func() {
		instance = &Database{}
	})
	return instance
}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, parser.ParseComments)
	require.NoError(t, err)

	analyzer := NewPatternAnalyzer(fset)
	patterns, err := analyzer.AnalyzePatterns(file, "test", "test.go")
	require.NoError(t, err)

	assert.Len(t, patterns.Singleton, 1)
	assert.Equal(t, "Singleton (sync.Once)", patterns.Singleton[0].Name)
	assert.Greater(t, patterns.Singleton[0].ConfidenceScore, 0.9)
}

func TestPatternAnalyzer_DetectFactory(t *testing.T) {
	src := `package test

type Reader interface {
	Read() string
}

func NewReader(typ string) Reader {
	switch typ {
	case "file":
		return &FileReader{}
	case "network":
		return &NetworkReader{}
	default:
		return nil
	}
}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, parser.ParseComments)
	require.NoError(t, err)

	analyzer := NewPatternAnalyzer(fset)
	patterns, err := analyzer.AnalyzePatterns(file, "test", "test.go")
	require.NoError(t, err)

	assert.Len(t, patterns.Factory, 1)
	assert.Equal(t, "Factory Method", patterns.Factory[0].Name)
	assert.Greater(t, patterns.Factory[0].ConfidenceScore, 0.8)
}

func TestPatternAnalyzer_DetectBuilder(t *testing.T) {
	src := `package test

type RequestBuilder struct {
	url     string
	method  string
	headers map[string]string
}

func (b *RequestBuilder) SetURL(url string) *RequestBuilder {
	b.url = url
	return b
}

func (b *RequestBuilder) SetMethod(method string) *RequestBuilder {
	b.method = method
	return b
}

func (b *RequestBuilder) WithHeader(key, val string) *RequestBuilder {
	if b.headers == nil {
		b.headers = make(map[string]string)
	}
	b.headers[key] = val
	return b
}

func (b *RequestBuilder) Build() *Request {
	return &Request{
		URL:     b.url,
		Method:  b.method,
		Headers: b.headers,
	}
}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, parser.ParseComments)
	require.NoError(t, err)

	analyzer := NewPatternAnalyzer(fset)
	patterns, err := analyzer.AnalyzePatterns(file, "test", "test.go")
	require.NoError(t, err)

	assert.Len(t, patterns.Builder, 1)
	assert.Equal(t, "Builder Pattern", patterns.Builder[0].Name)
	assert.Greater(t, patterns.Builder[0].ConfidenceScore, 0.85)
}

func TestPatternAnalyzer_DetectObserver(t *testing.T) {
	src := `package test

type EventHandler func(event string)

type EventManager struct {
	handlers []EventHandler
}

func (em *EventManager) RegisterHandler(handler EventHandler) {
	em.handlers = append(em.handlers, handler)
}

func (em *EventManager) AddListener(handler EventHandler) {
	em.handlers = append(em.handlers, handler)
}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, parser.ParseComments)
	require.NoError(t, err)

	analyzer := NewPatternAnalyzer(fset)
	patterns, err := analyzer.AnalyzePatterns(file, "test", "test.go")
	require.NoError(t, err)

	assert.GreaterOrEqual(t, len(patterns.Observer), 2)
	if len(patterns.Observer) > 0 {
		assert.Equal(t, "Observer Pattern", patterns.Observer[0].Name)
	}
}

func TestPatternAnalyzer_DetectStrategy(t *testing.T) {
	src := `package test

type Sorter interface {
	Sort(data []int) []int
}

type DataProcessor struct {
	strategy Sorter
	logger   Logger
}

func (dp *DataProcessor) Process(data []int) []int {
	return dp.strategy.Sort(data)
}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, parser.ParseComments)
	require.NoError(t, err)

	analyzer := NewPatternAnalyzer(fset)
	patterns, err := analyzer.AnalyzePatterns(file, "test", "test.go")
	require.NoError(t, err)

	assert.Len(t, patterns.Strategy, 1)
	assert.Equal(t, "Strategy Pattern", patterns.Strategy[0].Name)
	assert.Greater(t, patterns.Strategy[0].ConfidenceScore, 0.75)
}

func TestPatternAnalyzer_NoFalsePositives(t *testing.T) {
	src := `package test

type SimpleStruct struct {
	value string
}

func (s *SimpleStruct) GetValue() string {
	return s.value
}

func helperFunction(x int) int {
	return x * 2
}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, parser.ParseComments)
	require.NoError(t, err)

	analyzer := NewPatternAnalyzer(fset)
	patterns, err := analyzer.AnalyzePatterns(file, "test", "test.go")
	require.NoError(t, err)

	assert.Empty(t, patterns.Singleton)
	assert.Empty(t, patterns.Factory)
	assert.Empty(t, patterns.Builder)
	assert.Empty(t, patterns.Observer)
	assert.Empty(t, patterns.Strategy)
}
