package cmd

import (
	"github.com/spf13/viper"

	"github.com/opd-ai/go-stats-generator/internal/analyzer"
	"github.com/opd-ai/go-stats-generator/internal/config"
)

// loadConfiguration creates and populates a Config object from viper settings
func loadConfiguration() *config.Config {
	cfg := config.DefaultConfig()
	loadOutputConfiguration(cfg)
	loadPerformanceConfiguration(cfg)
	loadFilterConfiguration(cfg)
	loadAnalysisConfiguration(cfg)
	return cfg
}

// loadOutputConfiguration loads output-related settings from viper
func loadOutputConfiguration(cfg *config.Config) {
	if viper.IsSet("output.format") {
		cfg.Output.Format = config.OutputFormat(viper.GetString("output.format"))
	}
	if viper.IsSet("output.destination") {
		cfg.Output.Destination = viper.GetString("output.destination")
	}
	if viper.IsSet("output.verbose") {
		cfg.Output.Verbose = viper.GetBool("output.verbose")
	}
	if viper.IsSet("output.show_progress") {
		cfg.Output.ShowProgress = viper.GetBool("output.show_progress")
	}
	applyVerboseDefaults(cfg)
	cfg.Output.Sections = mergeSectionFlags()
}

// applyVerboseDefaults enables progress display when verbose mode is active
func applyVerboseDefaults(cfg *config.Config) {
	if cfg.Output.Verbose {
		cfg.Output.ShowProgress = true
	}
}

// mergeSectionFlags combines --sections and --only flags with deduplication
func mergeSectionFlags() []string {
	seen := make(map[string]bool)
	var merged []string
	for _, src := range []string{"output.sections", "output.only"} {
		if viper.IsSet(src) {
			for _, s := range viper.GetStringSlice(src) {
				if !seen[s] {
					seen[s] = true
					merged = append(merged, s)
				}
			}
		}
	}
	return merged
}

// loadPerformanceConfiguration loads performance-related settings from viper
func loadPerformanceConfiguration(cfg *config.Config) {
	if viper.IsSet("performance.worker_count") {
		cfg.Performance.WorkerCount = viper.GetInt("performance.worker_count")
	}
	if viper.IsSet("performance.timeout") {
		cfg.Performance.Timeout = viper.GetDuration("performance.timeout")
	}
}

// loadFilterConfiguration loads file filtering settings from viper
func loadFilterConfiguration(cfg *config.Config) {
	if viper.IsSet("filters.skip_test_files") {
		cfg.Filters.SkipTestFiles = viper.GetBool("filters.skip_test_files")
	}
	if viper.IsSet("filters.skip_vendor") {
		cfg.Filters.SkipVendor = viper.GetBool("filters.skip_vendor")
	}
	if viper.IsSet("filters.skip_generated") {
		cfg.Filters.SkipGenerated = viper.GetBool("filters.skip_generated")
	}
	if viper.IsSet("filters.include_patterns") {
		cfg.Filters.IncludePatterns = viper.GetStringSlice("filters.include_patterns")
	}
	if viper.IsSet("filters.exclude_patterns") {
		cfg.Filters.ExcludePatterns = viper.GetStringSlice("filters.exclude_patterns")
	}
}

// loadAnalysisConfiguration loads all analysis-specific settings from viper
func loadAnalysisConfiguration(cfg *config.Config) {
	loadBasicAnalysisSettings(cfg)
	loadThresholdSettings(cfg)
	loadDuplicationSettings(cfg)
	loadPlacementSettings(cfg)
	loadOrganizationSettings(cfg)
	loadBurdenSettings(cfg)
	loadDocumentationSettings(cfg)
	loadScoringSettings(cfg)
}

// loadBasicAnalysisSettings loads core analysis toggles from viper
func loadBasicAnalysisSettings(cfg *config.Config) {
	loadBooleanAnalysisSettings(cfg)
	loadStringAnalysisSettings(cfg)
}

// loadBooleanAnalysisSettings loads boolean analysis toggles.
func loadBooleanAnalysisSettings(cfg *config.Config) {
	boolSettings := map[string]*bool{
		"analysis.include_patterns":      &cfg.Analysis.IncludePatterns,
		"analysis.include_complexity":    &cfg.Analysis.IncludeComplexity,
		"analysis.include_documentation": &cfg.Analysis.IncludeDocumentation,
		"analysis.include_generics":      &cfg.Analysis.IncludeGenerics,
		"analysis.enable_team_metrics":   &cfg.Analysis.EnableTeamMetrics,
	}

	for key, target := range boolSettings {
		if viper.IsSet(key) {
			*target = viper.GetBool(key)
		}
	}
}

// loadStringAnalysisSettings loads string analysis settings.
func loadStringAnalysisSettings(cfg *config.Config) {
	if viper.IsSet("analysis.coverage_profile") {
		cfg.Analysis.CoverageProfile = viper.GetString("analysis.coverage_profile")
	}
}

// loadThresholdSettings loads quality threshold settings from viper
func loadThresholdSettings(cfg *config.Config) {
	if viper.IsSet("analysis.max_function_length") {
		cfg.Analysis.MaxFunctionLength = viper.GetInt("analysis.max_function_length")
	}
	if viper.IsSet("analysis.max_cyclomatic_complexity") {
		cfg.Analysis.MaxCyclomaticComplexity = viper.GetInt("analysis.max_cyclomatic_complexity")
	}
	if viper.IsSet("analysis.min_documentation_coverage") {
		cfg.Analysis.MinDocumentationCoverage = viper.GetFloat64("analysis.min_documentation_coverage")
	}
	if viper.IsSet("analysis.min_package_doc_coverage") {
		cfg.Analysis.MinPackageDocCoverage = viper.GetFloat64("analysis.min_package_doc_coverage")
	}
	if viper.IsSet("analysis.enforce_thresholds") {
		cfg.Analysis.EnforceThresholds = viper.GetBool("analysis.enforce_thresholds")
	}
}

// loadDuplicationSettings loads code duplication detection settings from viper
func loadDuplicationSettings(cfg *config.Config) {
	if viper.IsSet("analysis.duplication.min_block_lines") {
		cfg.Analysis.Duplication.MinBlockLines = viper.GetInt("analysis.duplication.min_block_lines")
	}
	if viper.IsSet("analysis.duplication.similarity_threshold") {
		cfg.Analysis.Duplication.SimilarityThreshold = viper.GetFloat64("analysis.duplication.similarity_threshold")
	}
	if viper.IsSet("analysis.duplication.ignore_test_files") {
		cfg.Analysis.Duplication.IgnoreTestFiles = viper.GetBool("analysis.duplication.ignore_test_files")
	}
}

// loadPlacementSettings loads function placement analysis settings from viper
func loadPlacementSettings(cfg *config.Config) {
	if viper.IsSet("analysis.placement.affinity_margin") {
		cfg.Analysis.Placement.AffinityMargin = viper.GetFloat64("analysis.placement.affinity_margin")
	}
	if viper.IsSet("analysis.placement.min_cohesion") {
		cfg.Analysis.Placement.MinCohesion = viper.GetFloat64("analysis.placement.min_cohesion")
	}
}

// loadOrganizationSettings loads code organization settings from viper
func loadOrganizationSettings(cfg *config.Config) {
	loadMaxFileLines(cfg)
	loadMaxFileFunctions(cfg)
	loadMaxFileTypes(cfg)
	loadMaxPackageFiles(cfg)
	loadMaxExportedSymbols(cfg)
	loadMaxDirectoryDepth(cfg)
	loadMaxFileImports(cfg)
}

// loadMaxFileLines loads max_file_lines setting from viper
func loadMaxFileLines(cfg *config.Config) {
	if viper.IsSet("analysis.organization.max_file_lines") {
		cfg.Analysis.Organization.MaxFileLines = viper.GetInt("analysis.organization.max_file_lines")
	}
}

// loadMaxFileFunctions loads max_file_functions setting from viper
func loadMaxFileFunctions(cfg *config.Config) {
	if viper.IsSet("analysis.organization.max_file_functions") {
		cfg.Analysis.Organization.MaxFileFunctions = viper.GetInt("analysis.organization.max_file_functions")
	}
}

// loadMaxFileTypes loads max_file_types setting from viper
func loadMaxFileTypes(cfg *config.Config) {
	if viper.IsSet("analysis.organization.max_file_types") {
		cfg.Analysis.Organization.MaxFileTypes = viper.GetInt("analysis.organization.max_file_types")
	}
}

// loadMaxPackageFiles loads max_package_files setting from viper
func loadMaxPackageFiles(cfg *config.Config) {
	if viper.IsSet("analysis.organization.max_package_files") {
		cfg.Analysis.Organization.MaxPackageFiles = viper.GetInt("analysis.organization.max_package_files")
	}
}

// loadMaxExportedSymbols loads max_exported_symbols setting from viper
func loadMaxExportedSymbols(cfg *config.Config) {
	if viper.IsSet("analysis.organization.max_exported_symbols") {
		cfg.Analysis.Organization.MaxExportedSymbols = viper.GetInt("analysis.organization.max_exported_symbols")
	}
}

// loadMaxDirectoryDepth loads max_directory_depth setting from viper
func loadMaxDirectoryDepth(cfg *config.Config) {
	if viper.IsSet("analysis.organization.max_directory_depth") {
		cfg.Analysis.Organization.MaxDirectoryDepth = viper.GetInt("analysis.organization.max_directory_depth")
	}
}

// loadMaxFileImports loads max_file_imports setting from viper
func loadMaxFileImports(cfg *config.Config) {
	if viper.IsSet("analysis.organization.max_file_imports") {
		cfg.Analysis.Organization.MaxFileImports = viper.GetInt("analysis.organization.max_file_imports")
	}
}

// loadBurdenSettings loads maintenance burden analysis settings from viper
func loadBurdenSettings(cfg *config.Config) {
	if viper.IsSet("analysis.burden.max_params") {
		cfg.Analysis.Burden.MaxParams = viper.GetInt("analysis.burden.max_params")
	}
	if viper.IsSet("analysis.burden.max_returns") {
		cfg.Analysis.Burden.MaxReturns = viper.GetInt("analysis.burden.max_returns")
	}
	if viper.IsSet("analysis.burden.max_nesting") {
		cfg.Analysis.Burden.MaxNesting = viper.GetInt("analysis.burden.max_nesting")
	}
	if viper.IsSet("analysis.burden.feature_envy_ratio") {
		cfg.Analysis.Burden.FeatureEnvyRatio = viper.GetFloat64("analysis.burden.feature_envy_ratio")
	}
}

// loadDocumentationSettings loads documentation analysis settings from viper
func loadDocumentationSettings(cfg *config.Config) {
	if viper.IsSet("analysis.documentation.require_exported_doc") {
		cfg.Analysis.Documentation.RequireExportedDoc = viper.GetBool("analysis.documentation.require_exported_doc")
	}
	if viper.IsSet("analysis.documentation.require_package_doc") {
		cfg.Analysis.Documentation.RequirePackageDoc = viper.GetBool("analysis.documentation.require_package_doc")
	}
	if viper.IsSet("analysis.documentation.stale_annotation_days") {
		cfg.Analysis.Documentation.StaleAnnotationDays = viper.GetInt("analysis.documentation.stale_annotation_days")
	}
	if viper.IsSet("analysis.documentation.min_comment_words") {
		cfg.Analysis.Documentation.MinCommentWords = viper.GetInt("analysis.documentation.min_comment_words")
	}
}

// loadScoringSettings loads MBI scoring settings from viper
func loadScoringSettings(cfg *config.Config) {
	if viper.IsSet("analysis.scoring.max_burden_score") {
		cfg.Analysis.Scoring.MaxBurdenScore = viper.GetFloat64("analysis.scoring.max_burden_score")
	}
}

// getOrganizationConfig extracts organization analysis configuration from the main config
func getOrganizationConfig(cfg *config.Config) analyzer.OrganizationConfig {
	return analyzer.OrganizationConfig{
		MaxFileLines:       cfg.Analysis.Organization.MaxFileLines,
		MaxFileFunctions:   cfg.Analysis.Organization.MaxFileFunctions,
		MaxFileTypes:       cfg.Analysis.Organization.MaxFileTypes,
		MaxPackageFiles:    cfg.Analysis.Organization.MaxPackageFiles,
		MaxExportedSymbols: cfg.Analysis.Organization.MaxExportedSymbols,
		MaxDirectoryDepth:  cfg.Analysis.Organization.MaxDirectoryDepth,
		MaxFileImports:     cfg.Analysis.Organization.MaxFileImports,
	}
}
