# Project Overview

The `go-stats-generator` is a high-performance command-line tool that analyzes Go source code repositories to generate comprehensive statistical reports about code structure, complexity, and patterns. The project focuses on computing obscure and detailed metrics that standard linters don't typically capture, providing actionable insights for code quality assessment and refactoring decisions. The tool is designed for enterprise-scale codebases, supporting concurrent processing of 50,000+ files within 60 seconds while maintaining memory usage under 1GB.

The program is called `go-stats-generator` not gostats. Don't ever use the name `gostats` and if you discover a use of the name `gostats` correct it. `gostats` is a fucking stupid name.

The target audience includes software engineers, technical leads, and DevOps teams working with large Go codebases who need detailed code analysis beyond basic linting. The tool emphasizes advanced metrics like design pattern detection, struct member categorization, concurrency pattern analysis, and documentation quality assessment. It serves as a comprehensive code health assessment platform that helps teams identify technical debt, architectural issues, and optimization opportunities.

## Technical Stack
- **Primary Language**: Go 1.23.2+ (currently using latest stable version)
- **CLI Framework**: github.com/spf13/cobra v1.9.1 for command structure, github.com/spf13/viper v1.20.1 for configuration management
- **AST Processing**: Go standard library (go/parser, go/ast, go/token) for zero-dependency core functionality
- **Storage**: modernc.org/sqlite v1.31.1 for historical metrics storage and baseline management
- **Testing**: Go standard library testing package with github.com/stretchr/testify v1.10.0 for enhanced assertions and mocking
- **Build/Deploy**: Standard Go build tools, Makefile for automation, GitHub Actions CI/CD, GoReleaser for multi-platform releases

## Current Implementation Status (87% Complete)

### **Fully Implemented Core Features** ✅
- **CLI Framework**: Complete command structure with analyze, baseline, diff, trend, and version commands
- **Function Analysis**: Precise line counting (excluding comments/blank lines), cyclomatic complexity, signature analysis
- **Struct Analysis**: Detailed member categorization, method analysis, embedded type detection, tag analysis
- **Package Analysis**: Dependency tracking, circular detection, cohesion/coupling metrics
- **Interface Analysis**: Cross-file implementation tracking, embedding depth, signature complexity
- **Concurrency Analysis**: Goroutine patterns, channel analysis, sync primitives, worker pools, pipelines
- **Historical Storage**: SQLite backend for baseline management and trend analysis
- **Multi-format Output**: Console (rich tables), JSON, HTML, CSV, and Markdown reporting

### **Key Architectural Components**
