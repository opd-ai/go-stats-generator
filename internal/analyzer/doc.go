// Package analyzer provides static analysis capabilities for Go source code.
//
// The analyzer package contains analyzers for various aspects of Go code including
// functions, structs, interfaces, complexity, documentation, naming conventions,
// code duplication, and file organization. Each analyzer operates on the AST
// representation of Go source files and produces structured metrics.
//
// The package supports concurrent analysis for large codebases and integrates
// with the metrics package for structured reporting.
package analyzer
