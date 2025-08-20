# Single File Analysis Feature

The `go-stats-generator analyze` command has been enhanced to support analyzing individual Go source files in addition to the existing directory analysis functionality.

## Usage

### Directory Analysis (existing functionality)
```bash
# Analyze entire directory recursively
go-stats-generator analyze ./src
go-stats-generator analyze .
```

### Single File Analysis (new functionality)
```bash
# Analyze a single Go file
go-stats-generator analyze ./main.go
go-stats-generator analyze ./internal/analyzer/function.go
go-stats-generator analyze ./pkg/mypackage/myfile.go
```

## Features

- **Automatic detection**: The tool automatically detects whether the target is a file or directory
- **File validation**: Only `.go` files are accepted for single file analysis
- **Full analysis**: Single files receive the same comprehensive analysis as directory mode:
  - Function and method analysis
  - Struct complexity metrics
  - Complexity calculations
  - Pattern detection
  - Documentation analysis
  - Package information

## Examples

### Basic single file analysis
```bash
./go-stats-generator analyze ./main.go
```

### Single file with verbose output
```bash
./go-stats-generator analyze ./internal/analyzer/function.go --verbose
```

### Single file with JSON output
```bash
./go-stats-generator analyze ./pkg/types.go --format json --output analysis.json
```

### Single file analysis with custom thresholds
```bash
./go-stats-generator analyze ./complex_file.go --max-function-length 100 --max-complexity 15
```

## Error Handling

The tool provides clear error messages for invalid inputs:

- Non-existent files: `Error: path does not exist: /path/to/file.go`
- Non-Go files: `Error: analysis failed: file /path/to/file.txt is not a Go source file`

## Output

Single file analysis generates the same comprehensive report structure as directory analysis, but with:
- `Files Processed: 1`
- Repository path set to the file's directory
- Package analysis focused on the single file's package
- All metrics calculated for the single file's contents

This feature is particularly useful for:
- Analyzing specific files during development
- CI/CD pipelines that want to check individual changed files
- Quick analysis of suspicious or complex files
- Educational purposes when studying specific code patterns
