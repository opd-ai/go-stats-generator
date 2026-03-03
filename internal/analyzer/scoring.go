package analyzer

import (
	"github.com/opd-ai/go-stats-generator/internal/config"
	"github.com/opd-ai/go-stats-generator/internal/metrics"
)

// ScoringAnalyzer calculates composite maintenance burden scores
type ScoringAnalyzer struct {
	weights config.ScoringWeights
}

// NewScoringAnalyzer creates a new scoring analyzer with configured weights
// for computing Maintenance Burden Index across files and packages.
func NewScoringAnalyzer(weights config.ScoringWeights) *ScoringAnalyzer {
	return &ScoringAnalyzer{
		weights: weights,
	}
}

// CalculateFileMBI computes the Maintenance Burden Index for a file by combining
// weighted scores from duplication, naming, placement, documentation, and burden.
func (sa *ScoringAnalyzer) CalculateFileMBI(file string, report *metrics.Report) float64 {
	score := 0.0

	score += sa.calculateDuplicationScore(file, report) * sa.weights.Duplication
	score += sa.calculateNamingScore(file, report) * sa.weights.Naming
	score += sa.calculatePlacementScore(file, report) * sa.weights.Placement
	score += sa.calculateDocumentationScore(file, report) * sa.weights.Documentation
	score += sa.calculateOrganizationScore(file, report) * sa.weights.Organization
	score += sa.calculateBurdenScore(file, report) * sa.weights.Burden

	return normalizeToScore(score)
}

// calculateDuplicationScore computes duplication contribution (0-100)
func (sa *ScoringAnalyzer) calculateDuplicationScore(file string, report *metrics.Report) float64 {
	fileClones := countFileClones(file, report.Duplication.Clones)
	if fileClones == 0 {
		return 0
	}
	return min(float64(fileClones)*10.0, 100.0)
}

// calculateNamingScore computes naming violation contribution (0-100)
func (sa *ScoringAnalyzer) calculateNamingScore(file string, report *metrics.Report) float64 {
	violations := countFileNamingViolations(file, report.Naming)
	if violations == 0 {
		return 0
	}
	return min(float64(violations)*5.0, 100.0)
}

// calculatePlacementScore computes misplacement contribution (0-100)
func (sa *ScoringAnalyzer) calculatePlacementScore(file string, report *metrics.Report) float64 {
	misplaced := countFileMisplacements(file, report.Placement)
	if misplaced == 0 {
		return 0
	}
	return min(float64(misplaced)*7.0, 100.0)
}

// calculateDocumentationScore computes documentation gap contribution (0-100)
func (sa *ScoringAnalyzer) calculateDocumentationScore(file string, report *metrics.Report) float64 {
	coverage := getFileCoverage(file, report)
	if coverage >= 0.7 {
		return 0
	}
	gap := (0.7 - coverage) / 0.7
	return gap * 100.0
}

// calculateOrganizationScore computes organization issue contribution (0-100)
func (sa *ScoringAnalyzer) calculateOrganizationScore(file string, report *metrics.Report) float64 {
	if isOversized(file, report.Organization) {
		return 100.0
	}
	cohesion := getFileCohesion(file, report.Placement)
	if cohesion >= 0.3 {
		return 0
	}
	return (0.3 - cohesion) / 0.3 * 80.0
}

// calculateBurdenScore computes burden indicator contribution (0-100)
func (sa *ScoringAnalyzer) calculateBurdenScore(file string, report *metrics.Report) float64 {
	score := 0.0

	score += float64(countFileMagicNumbers(file, report.Burden.MagicNumbers)) * 0.5
	score += float64(countFileDeadCode(file, report.Burden.DeadCode)) * 5.0
	score += float64(countFileComplexSigs(file, report.Burden.ComplexSignatures)) * 8.0
	score += float64(countFileDeepNesting(file, report.Burden.DeeplyNestedFunctions)) * 10.0
	score += float64(countFileFeatureEnvy(file, report.Burden.FeatureEnvyMethods)) * 3.0

	return min(score, 100.0)
}

// normalizeScore ensures score is in 0-100 range (renamed to avoid conflict with function.go)
func normalizeToScore(score float64) float64 {
	if score < 0 {
		return 0
	}
	if score > 100 {
		return 100
	}
	return score
}

// countFileClones returns the number of clone pairs involving the specified file
func countFileClones(file string, clones []metrics.ClonePair) int {
	count := 0
	for _, clone := range clones {
		for _, loc := range clone.Instances {
			if loc.File == file {
				count++
				break
			}
		}
	}
	return count
}

// countFileNamingViolations returns the total naming violations for a file
func countFileNamingViolations(file string, naming metrics.NamingMetrics) int {
	count := 0
	for _, v := range naming.IdentifierIssues {
		if v.File == file {
			count++
		}
	}
	for _, v := range naming.FileNameIssues {
		if v.File == file {
			count++
		}
	}
	return count
}

// countFileMisplacements returns the number of misplaced functions and methods in a file
func countFileMisplacements(file string, placement metrics.PlacementMetrics) int {
	count := 0
	for _, m := range placement.FunctionIssues {
		if m.CurrentFile == file {
			count++
		}
	}
	for _, m := range placement.MethodIssues {
		if m.CurrentFile == file {
			count++
		}
	}
	return count
}

// getFileCoverage calculates documentation coverage ratio for a file (0.0-1.0)
func getFileCoverage(file string, report *metrics.Report) float64 {
	fileFuncs := 0
	documentedFuncs := 0
	for _, fn := range report.Functions {
		if fn.File == file {
			fileFuncs++
			if fn.Documentation.HasComment {
				documentedFuncs++
			}
		}
	}
	if fileFuncs == 0 {
		return 1.0
	}
	return float64(documentedFuncs) / float64(fileFuncs)
}

// isOversized checks if a file exceeds organization size thresholds
func isOversized(file string, org metrics.OrganizationMetrics) bool {
	for _, f := range org.OversizedFiles {
		if f.File == file {
			return true
		}
	}
	return false
}

// getFileCohesion returns the cohesion score for a file, defaulting to 1.0 if not found
func getFileCohesion(file string, placement metrics.PlacementMetrics) float64 {
	for _, f := range placement.CohesionIssues {
		if f.File == file {
			return f.CohesionScore
		}
	}
	return 1.0
}

// countFileMagicNumbers returns the number of magic number occurrences in a file
func countFileMagicNumbers(file string, magics []metrics.MagicNumber) int {
	count := 0
	for _, m := range magics {
		if m.File == file {
			count++
		}
	}
	return count
}

// countFileDeadCode returns the number of dead code issues in a file
func countFileDeadCode(file string, deadCode metrics.DeadCodeMetrics) int {
	count := 0
	for _, f := range deadCode.UnreferencedFunctions {
		if f.File == file {
			count++
		}
	}
	for _, b := range deadCode.UnreachableCode {
		if b.File == file {
			count++
		}
	}
	return count
}

// countFileComplexSigs returns the number of complex function signatures in a file
func countFileComplexSigs(file string, sigs []metrics.SignatureIssue) int {
	count := 0
	for _, s := range sigs {
		if s.File == file {
			count++
		}
	}
	return count
}

// countFileDeepNesting returns the number of deeply nested functions in a file
func countFileDeepNesting(file string, nesting []metrics.NestingIssue) int {
	count := 0
	for _, n := range nesting {
		if n.File == file {
			count++
		}
	}
	return count
}

// countFileFeatureEnvy returns the number of feature envy methods in a file
func countFileFeatureEnvy(file string, envy []metrics.FeatureEnvyIssue) int {
	count := 0
	for _, e := range envy {
		if e.File == file {
			count++
		}
	}
	return count
}

// min returns the smaller of two float64 values
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// CalculatePackageMBI computes the Maintenance Burden Index for a package by
// averaging MBI scores across all files belonging to the package.
func (sa *ScoringAnalyzer) CalculatePackageMBI(pkg string, report *metrics.Report) float64 {
	pkgFiles := getPackageFiles(pkg, report)
	if len(pkgFiles) == 0 {
		return 0
	}

	totalScore := 0.0
	for _, file := range pkgFiles {
		totalScore += sa.CalculateFileMBI(file, report)
	}

	return totalScore / float64(len(pkgFiles))
}

// CalculateAllScores computes MBI for all files and packages in the report.
func (sa *ScoringAnalyzer) CalculateAllScores(report *metrics.Report) *metrics.ScoringMetrics {
	fileScores := sa.calculateFileScores(report)
	packageScores := sa.calculatePackageScores(report)

	return &metrics.ScoringMetrics{
		FileScores:    fileScores,
		PackageScores: packageScores,
	}
}

// calculateFileScores computes MBI scores for all files in the report
func (sa *ScoringAnalyzer) calculateFileScores(report *metrics.Report) []metrics.FileScore {
	fileScores := make([]metrics.FileScore, 0)
	seenFiles := make(map[string]bool)

	for _, fn := range report.Functions {
		if !seenFiles[fn.File] {
			seenFiles[fn.File] = true
			score := sa.CalculateFileMBI(fn.File, report)
			fileScores = append(fileScores, metrics.FileScore{
				File:      fn.File,
				Score:     score,
				Risk:      getRiskLevel(score),
				Breakdown: sa.getFileBreakdown(fn.File, report),
			})
		}
	}

	return fileScores
}

// calculatePackageScores computes MBI scores for all packages in the report
func (sa *ScoringAnalyzer) calculatePackageScores(report *metrics.Report) []metrics.PackageScore {
	packageScores := make([]metrics.PackageScore, 0)
	seenPackages := make(map[string]bool)

	for _, pkg := range report.Packages {
		if !seenPackages[pkg.Name] {
			seenPackages[pkg.Name] = true
			score := sa.CalculatePackageMBI(pkg.Name, report)
			packageScores = append(packageScores, metrics.PackageScore{
				Package:   pkg.Name,
				Score:     score,
				Risk:      getRiskLevel(score),
				Breakdown: sa.getPackageBreakdown(pkg.Name, report),
			})
		}
	}

	return packageScores
}

// getFileBreakdown calculates weighted score contributions for a file
func (sa *ScoringAnalyzer) getFileBreakdown(file string, report *metrics.Report) metrics.ScoreBreakdown {
	return metrics.ScoreBreakdown{
		Duplication:   sa.calculateDuplicationScore(file, report) * sa.weights.Duplication,
		Naming:        sa.calculateNamingScore(file, report) * sa.weights.Naming,
		Placement:     sa.calculatePlacementScore(file, report) * sa.weights.Placement,
		Documentation: sa.calculateDocumentationScore(file, report) * sa.weights.Documentation,
		Organization:  sa.calculateOrganizationScore(file, report) * sa.weights.Organization,
		Burden:        sa.calculateBurdenScore(file, report) * sa.weights.Burden,
	}
}

// getPackageBreakdown calculates averaged score breakdown for all files in a package
func (sa *ScoringAnalyzer) getPackageBreakdown(pkg string, report *metrics.Report) metrics.ScoreBreakdown {
	pkgFiles := getPackageFiles(pkg, report)
	if len(pkgFiles) == 0 {
		return metrics.ScoreBreakdown{}
	}

	total := metrics.ScoreBreakdown{}
	for _, file := range pkgFiles {
		breakdown := sa.getFileBreakdown(file, report)
		total.Duplication += breakdown.Duplication
		total.Naming += breakdown.Naming
		total.Placement += breakdown.Placement
		total.Documentation += breakdown.Documentation
		total.Organization += breakdown.Organization
		total.Burden += breakdown.Burden
	}

	count := float64(len(pkgFiles))
	return metrics.ScoreBreakdown{
		Duplication:   total.Duplication / count,
		Naming:        total.Naming / count,
		Placement:     total.Placement / count,
		Documentation: total.Documentation / count,
		Organization:  total.Organization / count,
		Burden:        total.Burden / count,
	}
}

// getRiskLevel maps a numeric score to a risk level category
func getRiskLevel(score float64) string {
	if score >= 70 {
		return "critical"
	} else if score >= 50 {
		return "high"
	} else if score >= 30 {
		return "medium"
	} else if score >= 10 {
		return "low"
	}
	return "minimal"
}

// getPackageFiles returns the unique list of files belonging to a package
func getPackageFiles(pkg string, report *metrics.Report) []string {
	files := make([]string, 0)
	seenFiles := make(map[string]bool)

	for _, fn := range report.Functions {
		if fn.Package == pkg && !seenFiles[fn.File] {
			seenFiles[fn.File] = true
			files = append(files, fn.File)
		}
	}

	return files
}
