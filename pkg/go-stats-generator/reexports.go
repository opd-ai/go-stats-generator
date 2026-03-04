package go_stats_generator

import "github.com/opd-ai/go-stats-generator/internal/metrics"

// Re-export commonly used types for public API

// Report represents a complete analysis report
type Report = metrics.Report

// FunctionMetrics contains detailed function analysis
type FunctionMetrics = metrics.FunctionMetrics

// StructMetrics contains detailed struct analysis
type StructMetrics = metrics.StructMetrics

// ComplexityScore represents various complexity measurements
type ComplexityScore = metrics.ComplexityScore

// OverviewMetrics provides high-level statistics
type OverviewMetrics = metrics.OverviewMetrics

// ReportMetadata contains information about the analysis run
type ReportMetadata = metrics.ReportMetadata
