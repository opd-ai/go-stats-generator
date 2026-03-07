# API Usage Examples

This directory contains compile-tested examples demonstrating how to use the `go-stats-generator` API.

## Basic API Usage

The `api_example.go` file demonstrates the fundamental API pattern:

```go
analyzer := generator.NewAnalyzer()
report, err := analyzer.AnalyzeDirectory(context.Background(), "./src")
if err != nil {
    // handle error
}

// Access analysis results
fmt.Printf("Found %d functions with average complexity %.1f\n",
    len(report.Functions), report.Complexity.AverageFunction)
```

## Running the Examples

From the repository root:

```bash
# Run the basic API example (will fail if ./src doesn't exist)
go run examples/api_example.go

# Or modify the path to analyze a real directory:
# Edit api_example.go and change "./src" to a valid directory
```

## Testing

All examples in this directory are compile-tested to ensure they stay in sync with the actual API.
