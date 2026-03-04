package analyzer

import (
	"go/ast"
	"go/token"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"

	"github.com/opd-ai/go-stats-generator/internal/metrics"
)

// NamingAnalyzer performs naming convention analysis on Go code
type NamingAnalyzer struct {
	genericFileNames map[string]bool
	snakeCaseRegex   *regexp.Regexp
	acronyms         map[string]string // lowercase -> correct casing
	genericPackages  map[string]bool
	stdLibPackages   map[string]bool
	packageNameRegex *regexp.Regexp
}

// identifierContext tracks context information for identifier analysis
type identifierContext struct {
	inLoop             bool
	loopDepth          int
	receiverType       string
	packageName        string
	functionName       string
	isTestFile         bool
	validSingleLetters map[string]bool
}

// NewNamingAnalyzer creates a new naming analyzer for detecting file name,
// NewNamingAnalyzer identifies identifier and package name convention violations in Go code.
func NewNamingAnalyzer() *NamingAnalyzer {
	return &NamingAnalyzer{
		genericFileNames: map[string]bool{
			"utils.go":     true,
			"util.go":      true,
			"helpers.go":   true,
			"helper.go":    true,
			"misc.go":      true,
			"common.go":    true,
			"shared.go":    true,
			"base.go":      true,
			"core.go":      true,
			"lib.go":       true,
			"types.go":     true, // too generic in most contexts
			"constants.go": true,
			"errors.go":    true, // better to be specific
		},
		snakeCaseRegex: regexp.MustCompile(`^[a-z][a-z0-9]*(_[a-z0-9]+)*(_test)?\.go$`),
		acronyms: map[string]string{
			"url":   "URL",
			"http":  "HTTP",
			"https": "HTTPS",
			"id":    "ID",
			"api":   "API",
			"json":  "JSON",
			"xml":   "XML",
			"sql":   "SQL",
			"html":  "HTML",
			"css":   "CSS",
			"eof":   "EOF",
			"ip":    "IP",
			"tcp":   "TCP",
			"udp":   "UDP",
			"rpc":   "RPC",
			"tls":   "TLS",
			"ssl":   "SSL",
			"grpc":  "GRPC",
			"ui":    "UI",
			"uri":   "URI",
			"uuid":  "UUID",
			"ascii": "ASCII",
			"utf":   "UTF",
		},
		genericPackages: map[string]bool{
			"util":    true,
			"utils":   true,
			"common":  true,
			"base":    true,
			"shared":  true,
			"lib":     true,
			"core":    true,
			"misc":    true,
			"helpers": true,
			"helper":  true,
		},
		stdLibPackages: map[string]bool{
			"fmt":     true,
			"http":    true,
			"io":      true,
			"os":      true,
			"net":     true,
			"sync":    true,
			"time":    true,
			"strings": true,
			"bytes":   true,
			"errors":  true,
			"context": true,
			"testing": true,
			"regexp":  true,
			"sort":    true,
			"path":    true,
			"log":     true,
			"json":    true,
			"xml":     true,
			"sql":     true,
			"html":    true,
			"url":     true,
		},
		packageNameRegex: regexp.MustCompile(`^[a-z][a-z0-9]*$`),
	}
}

// AnalyzeFileNames checks file names against Go naming conventions
func (na *NamingAnalyzer) AnalyzeFileNames(filePaths []string) []metrics.FileNameViolation {
	var violations []metrics.FileNameViolation

	for _, filePath := range filePaths {
		if !strings.HasSuffix(filePath, ".go") {
			continue
		}

		fileViolations := na.checkFileViolations(filePath)
		violations = append(violations, fileViolations...)
	}

	return violations
}

func (na *NamingAnalyzer) checkFileViolations(filePath string) []metrics.FileNameViolation {
	var violations []metrics.FileNameViolation
	fileName := filepath.Base(filePath)
	dirName := filepath.Base(filepath.Dir(filePath))

	checks := []func(string, string, string) *metrics.FileNameViolation{
		func(fp, fn, dn string) *metrics.FileNameViolation { return na.checkSnakeCase(fp, fn) },
		func(fp, fn, dn string) *metrics.FileNameViolation { return na.checkStuttering(fp, fn, dn) },
		func(fp, fn, dn string) *metrics.FileNameViolation { return na.checkGenericName(fp, fn) },
		func(fp, fn, dn string) *metrics.FileNameViolation { return na.checkTestSuffix(fp, fn) },
	}

	for _, check := range checks {
		if violation := check(filePath, fileName, dirName); violation != nil {
			violations = append(violations, *violation)
		}
	}

	return violations
}

// checkSnakeCase verifies file names are in snake_case (lowercase with underscores)
func (na *NamingAnalyzer) checkSnakeCase(filePath, fileName string) *metrics.FileNameViolation {
	// Allow _test.go suffix
	if !na.snakeCaseRegex.MatchString(fileName) {
		// Try to suggest a snake_case version
		suggested := na.toSnakeCase(fileName)

		return &metrics.FileNameViolation{
			File:          filePath,
			ViolationType: "non_snake_case",
			Description:   "File name should be in snake_case (lowercase with underscores)",
			SuggestedName: suggested,
			Severity:      "medium",
		}
	}
	return nil
}

// checkStuttering detects when file name repeats directory name
func (na *NamingAnalyzer) checkStuttering(filePath, fileName, dirName string) *metrics.FileNameViolation {
	// Remove .go extension and _test suffix for comparison
	baseName := strings.TrimSuffix(fileName, ".go")
	baseName = strings.TrimSuffix(baseName, "_test")

	// Check if file name starts with directory name
	if dirName != "." && dirName != "/" && !strings.HasPrefix(dirName, ".") {
		dirNameLower := strings.ToLower(dirName)
		baseNameLower := strings.ToLower(baseName)

		// Exact match or prefix match indicates stuttering
		if baseNameLower == dirNameLower || strings.HasPrefix(baseNameLower, dirNameLower+"_") {
			suggested := strings.TrimPrefix(baseNameLower, dirNameLower+"_")
			if suggested == "" {
				// If the entire name was the directory, use a more descriptive name
				suggested = baseName + "_impl"
			}
			if strings.HasSuffix(fileName, "_test.go") {
				suggested += "_test.go"
			} else {
				suggested += ".go"
			}

			return &metrics.FileNameViolation{
				File:          filePath,
				ViolationType: "stuttering",
				Description:   "File name repeats package/directory name (e.g., http/http_client.go should be http/client.go)",
				SuggestedName: suggested,
				Severity:      "low",
			}
		}
	}
	return nil
}

// checkGenericName flags overly generic file names
func (na *NamingAnalyzer) checkGenericName(filePath, fileName string) *metrics.FileNameViolation {
	if na.genericFileNames[fileName] {
		return &metrics.FileNameViolation{
			File:          filePath,
			ViolationType: "generic_name",
			Description:   "File name is too generic; use a name that describes what the code does",
			SuggestedName: "", // Cannot suggest without understanding code
			Severity:      "low",
		}
	}
	return nil
}

// checkTestSuffix verifies _test.go suffix is only on test files
func (na *NamingAnalyzer) checkTestSuffix(filePath, fileName string) *metrics.FileNameViolation {
	hasTestSuffix := strings.HasSuffix(fileName, "_test.go")

	// If it has _test.go, it's presumably a test file (good)
	// We can't easily check if it's NOT a test file with _test.go without parsing
	// So we'll check for the opposite: non-test files trying to use test-like names

	// Check for improper test naming patterns
	if !hasTestSuffix && (strings.Contains(fileName, "test_") || strings.HasPrefix(fileName, "test")) {
		suggested := strings.Replace(fileName, "test_", "", 1)
		suggested = strings.TrimPrefix(suggested, "test")
		if suggested == ".go" {
			suggested = "impl.go"
		}

		return &metrics.FileNameViolation{
			File:          filePath,
			ViolationType: "improper_test_name",
			Description:   "Test-related files should use _test.go suffix, not test_ prefix or similar",
			SuggestedName: suggested,
			Severity:      "medium",
		}
	}

	return nil
}

// toSnakeCase converts a string to snake_case
func (na *NamingAnalyzer) toSnakeCase(s string) string {
	// Remove .go extension
	s = strings.TrimSuffix(s, ".go")
	testSuffix := ""
	if strings.HasSuffix(s, "_test") {
		s = strings.TrimSuffix(s, "_test")
		testSuffix = "_test"
	}

	var result []rune
	for i, r := range s {
		if unicode.IsUpper(r) {
			// Add underscore before uppercase if not at start and previous char is lowercase
			if i > 0 && unicode.IsLower(rune(s[i-1])) {
				result = append(result, '_')
			}
			result = append(result, unicode.ToLower(r))
		} else {
			result = append(result, r)
		}
	}

	resultStr := string(result)
	// Clean up multiple underscores
	resultStr = regexp.MustCompile(`_+`).ReplaceAllString(resultStr, "_")
	resultStr = strings.Trim(resultStr, "_")

	return resultStr + testSuffix + ".go"
}

// ComputeFileNamingScore calculates an overall file naming quality score
func (na *NamingAnalyzer) ComputeFileNamingScore(violations []metrics.FileNameViolation, totalFiles int) float64 {
	severities := make([]string, len(violations))
	for i, v := range violations {
		severities[i] = v.Severity
	}
	return computeQualityScore(severities, totalFiles)
}

// AnalyzeIdentifiers walks the AST to analyze all identifiers in a file
func (na *NamingAnalyzer) AnalyzeIdentifiers(file *ast.File, filePath string, fset *token.FileSet) []metrics.IdentifierViolation {
	var violations []metrics.IdentifierViolation

	ctx := &identifierContext{
		packageName:        file.Name.Name,
		isTestFile:         strings.HasSuffix(filePath, "_test.go"),
		validSingleLetters: make(map[string]bool),
	}

	ast.Inspect(file, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.FuncDecl:
			na.analyzeFunctionDecl(node, filePath, fset, ctx, &violations)
		case *ast.GenDecl:
			na.analyzeGenDecl(node, filePath, fset, ctx, &violations)
		case *ast.ForStmt, *ast.RangeStmt:
			ctx.inLoop = true
			ctx.loopDepth++
		case *ast.AssignStmt:
			na.trackLoopVariables(node, ctx)
		}
		return true
	})

	return violations
}

// analyzeFunctionDecl analyzes function/method declarations for naming violations
func (na *NamingAnalyzer) analyzeFunctionDecl(node *ast.FuncDecl, filePath string, fset *token.FileSet, ctx *identifierContext, violations *[]metrics.IdentifierViolation) {
	if node.Name == nil || node.Name.Name == "_" {
		return
	}

	idType := "function"
	if node.Recv != nil && len(node.Recv.List) > 0 {
		idType = "method"
		ctx.receiverType = ExtractReceiverType(node.Recv.List[0].Type)
	} else {
		ctx.receiverType = ""
	}

	ctx.functionName = node.Name.Name
	pos := fset.Position(node.Pos())

	na.checkIdentifier(node.Name.Name, filePath, pos.Line, idType, ctx, violations)
}

// analyzeGenDecl analyzes type, const, and var declarations for naming violations
func (na *NamingAnalyzer) analyzeGenDecl(node *ast.GenDecl, filePath string, fset *token.FileSet, ctx *identifierContext, violations *[]metrics.IdentifierViolation) {
	for _, spec := range node.Specs {
		switch s := spec.(type) {
		case *ast.TypeSpec:
			na.analyzeTypeSpec(s, filePath, fset, ctx, violations)
		case *ast.ValueSpec:
			na.analyzeValueSpec(s, node.Tok, filePath, fset, ctx, violations)
		}
	}
}

// analyzeTypeSpec analyzes type declarations for naming violations
func (na *NamingAnalyzer) analyzeTypeSpec(spec *ast.TypeSpec, filePath string, fset *token.FileSet, ctx *identifierContext, violations *[]metrics.IdentifierViolation) {
	if spec.Name == nil || spec.Name.Name == "_" {
		return
	}

	pos := fset.Position(spec.Pos())
	na.checkIdentifier(spec.Name.Name, filePath, pos.Line, "type", ctx, violations)
}

// analyzeValueSpec analyzes const and var declarations for naming violations
func (na *NamingAnalyzer) analyzeValueSpec(spec *ast.ValueSpec, tok token.Token, filePath string, fset *token.FileSet, ctx *identifierContext, violations *[]metrics.IdentifierViolation) {
	idType := "var"
	if tok == token.CONST {
		idType = "const"
	}

	for _, name := range spec.Names {
		if name == nil || name.Name == "_" {
			continue
		}

		pos := fset.Position(name.Pos())
		na.checkIdentifierWithSingleLetter(name.Name, filePath, pos.Line, idType, ctx, violations)
	}
}

// checkIdentifier performs all standard identifier checks (MixedCaps, acronyms, stuttering)
func (na *NamingAnalyzer) checkIdentifier(name, filePath string, line int, idType string, ctx *identifierContext, violations *[]metrics.IdentifierViolation) {
	if v := na.checkMixedCaps(name, ctx); v != nil {
		v.File = filePath
		v.Line = line
		v.Type = idType
		*violations = append(*violations, *v)
	}

	if v := na.checkAcronymCasing(name, ctx); v != nil {
		v.File = filePath
		v.Line = line
		v.Type = idType
		*violations = append(*violations, *v)
	}

	if v := na.checkIdentifierStuttering(name, ctx); v != nil {
		v.File = filePath
		v.Line = line
		v.Type = idType
		*violations = append(*violations, *v)
	}
}

// checkIdentifierWithSingleLetter performs all identifier checks including single-letter check
func (na *NamingAnalyzer) checkIdentifierWithSingleLetter(name, filePath string, line int, idType string, ctx *identifierContext, violations *[]metrics.IdentifierViolation) {
	if v := na.checkMixedCaps(name, ctx); v != nil {
		v.File = filePath
		v.Line = line
		v.Type = idType
		*violations = append(*violations, *v)
	}

	if v := na.checkSingleLetterName(name, idType, ctx); v != nil {
		v.File = filePath
		v.Line = line
		*violations = append(*violations, *v)
	}

	if v := na.checkAcronymCasing(name, ctx); v != nil {
		v.File = filePath
		v.Line = line
		v.Type = idType
		*violations = append(*violations, *v)
	}
}

// trackLoopVariables tracks valid single-letter loop variable names
func (na *NamingAnalyzer) trackLoopVariables(node *ast.AssignStmt, ctx *identifierContext) {
	if !ctx.inLoop {
		return
	}

	for _, lhs := range node.Lhs {
		if ident, ok := lhs.(*ast.Ident); ok {
			if len(ident.Name) == 1 && (ident.Name == "i" || ident.Name == "j" || ident.Name == "k") {
				ctx.validSingleLetters[ident.Name] = true
			}
		}
	}
}

// checkMixedCaps verifies identifier uses MixedCaps (no underscores except test functions)
func (na *NamingAnalyzer) checkMixedCaps(name string, ctx *identifierContext) *metrics.IdentifierViolation {
	// Allow underscores in test functions (Test_FunctionName pattern)
	if ctx.isTestFile && strings.HasPrefix(name, "Test_") {
		return nil
	}

	// Check for underscores
	if strings.Contains(name, "_") {
		suggested := na.toMixedCaps(name)
		return &metrics.IdentifierViolation{
			Name:          name,
			ViolationType: "underscore_in_name",
			Description:   "Go identifiers should use MixedCaps, not underscores (except Test_ functions)",
			SuggestedName: suggested,
			Severity:      "medium",
		}
	}

	return nil
}

// checkSingleLetterName flags inappropriate single-letter names
func (na *NamingAnalyzer) checkSingleLetterName(name, idType string, ctx *identifierContext) *metrics.IdentifierViolation {
	if len(name) != 1 {
		return nil
	}

	// Allow single-letter names in specific contexts
	// Loop variables: i, j, k
	if ctx.inLoop && ctx.validSingleLetters[name] {
		return nil
	}

	// Common receivers: r (reader), w (writer), s (server), c (client), etc.
	if idType == "method" && (name == "r" || name == "w" || name == "s" || name == "c" || name == "t" || name == "b" || name == "e" || name == "p") {
		return nil
	}

	// Single-letter const and var outside loops should be descriptive
	return &metrics.IdentifierViolation{
		Name:          name,
		ViolationType: "single_letter_name",
		Description:   "Single-letter names should be reserved for short loop variables and receivers",
		SuggestedName: "", // Cannot suggest without context
		Severity:      "low",
	}
}

// checkAcronymAtStart checks if an acronym at the start of a name is incorrectly cased.
// It validates the beginning position of identifiers like "Url" -> "URL" or "UrlParser" -> "URLParser".
func checkAcronymAtStart(name, nameLower, acronym, correctForm string) *metrics.IdentifierViolation {
	acronymLen := len(acronym)
	if !strings.HasPrefix(nameLower, acronym) {
		return nil
	}

	actualPrefix := name[:acronymLen]
	if actualPrefix != correctForm && isWrongAcronymCasing(actualPrefix, correctForm) {
		suggested := correctForm + name[acronymLen:]
		return &metrics.IdentifierViolation{
			Name:          name,
			ViolationType: "acronym_casing",
			Description:   "Acronyms should be all caps (e.g., URL, HTTP, ID, API, JSON)",
			SuggestedName: suggested,
			Severity:      "low",
		}
	}
	return nil
}

// checkAcronymInMiddle checks if an acronym in the middle or end of a name is incorrectly cased.
// It searches for word boundaries in MixedCaps identifiers like "GetUrl" -> "GetURL" or "UserId" -> "UserID".
func checkAcronymInMiddle(name, acronym, correctForm string) *metrics.IdentifierViolation {
	acronymLen := len(acronym)

	// Find word boundaries in MixedCaps names
	for i := 1; i < len(name)-acronymLen+1; i++ {
		if i > 0 && unicode.IsUpper(rune(name[i])) {
			segment := name[i : i+acronymLen]
			segmentLower := strings.ToLower(segment)

			if segmentLower == acronym && isWrongAcronymCasing(segment, correctForm) {
				suggested := name[:i] + correctForm + name[i+acronymLen:]
				return &metrics.IdentifierViolation{
					Name:          name,
					ViolationType: "acronym_casing",
					Description:   "Acronyms should be all caps (e.g., URL, HTTP, ID, API, JSON)",
					SuggestedName: suggested,
					Severity:      "low",
				}
			}
		}
	}
	return nil
}

// checkAcronymCasing detects improper acronym casing in Go identifiers.
// It checks for common acronyms (URL, HTTP, ID, API, JSON) with incorrect casing
// at the beginning, middle, or end of identifier names.
func (na *NamingAnalyzer) checkAcronymCasing(name string, ctx *identifierContext) *metrics.IdentifierViolation {
	nameLower := strings.ToLower(name)

	for acronym, correctForm := range na.acronyms {
		// Check beginning of name: "Url" -> "URL", "UrlParser" -> "URLParser"
		if v := checkAcronymAtStart(name, nameLower, acronym, correctForm); v != nil {
			return v
		}

		// Check middle/end of name: "GetUrl" -> "GetURL", "UserId" -> "UserID"
		if v := checkAcronymInMiddle(name, acronym, correctForm); v != nil {
			return v
		}
	}

	return nil
}

// checkIdentifierStuttering detects stuttering in identifiers
func (na *NamingAnalyzer) checkIdentifierStuttering(name string, ctx *identifierContext) *metrics.IdentifierViolation {
	if v := na.checkMethodStuttering(name, ctx); v != nil {
		return v
	}
	return na.checkPackageStuttering(name, ctx)
}

// checkMethodStuttering detects method names that repeat the receiver type
func (na *NamingAnalyzer) checkMethodStuttering(name string, ctx *identifierContext) *metrics.IdentifierViolation {
	if ctx.receiverType == "" {
		return nil
	}

	nameLower := strings.ToLower(name)
	receiverLower := strings.ToLower(ctx.receiverType)

	if !strings.HasPrefix(nameLower, receiverLower) || len(name) <= len(ctx.receiverType) {
		return nil
	}

	if na.isAllowedMethodPrefix(nameLower, receiverLower) {
		return nil
	}

	suggested := name[len(ctx.receiverType):]
	if len(suggested) == 0 {
		return nil
	}

	return &metrics.IdentifierViolation{
		Name:          name,
		ViolationType: "stuttering",
		Description:   "Method name repeats receiver type (e.g., User.UserName should be User.Name)",
		SuggestedName: suggested,
		Severity:      "low",
	}
}

// isAllowedMethodPrefix checks if method prefix is acceptable (e.g., GetUser, SetUser, NewUser)
func (na *NamingAnalyzer) isAllowedMethodPrefix(nameLower, receiverLower string) bool {
	return strings.HasPrefix(nameLower, "get"+receiverLower) ||
		strings.HasPrefix(nameLower, "set"+receiverLower) ||
		strings.HasPrefix(nameLower, "new"+receiverLower)
}

// checkPackageStuttering detects exported names that repeat the package name
func (na *NamingAnalyzer) checkPackageStuttering(name string, ctx *identifierContext) *metrics.IdentifierViolation {
	if ctx.packageName == "" || ctx.packageName == "main" {
		return nil
	}

	if !unicode.IsUpper(rune(name[0])) {
		return nil
	}

	nameLower := strings.ToLower(name)
	packageLower := strings.ToLower(ctx.packageName)

	if !strings.HasPrefix(nameLower, packageLower) || len(name) <= len(ctx.packageName) {
		return nil
	}

	if na.isAllowedFunctionPrefix(ctx.functionName) {
		return nil
	}

	return &metrics.IdentifierViolation{
		Name:          name,
		ViolationType: "package_stuttering",
		Description:   "Exported name repeats package name (e.g., user.UserService should be user.Service)",
		SuggestedName: name[len(ctx.packageName):],
		Severity:      "low",
	}
}

// isAllowedFunctionPrefix checks if function prefix is acceptable (e.g., NewUser, ParseUser, MakeUser)
func (na *NamingAnalyzer) isAllowedFunctionPrefix(functionName string) bool {
	return strings.HasPrefix(functionName, "New") ||
		strings.HasPrefix(functionName, "Parse") ||
		strings.HasPrefix(functionName, "Make")
}

// ComputeIdentifierQualityScore calculates an overall identifier naming quality score
func (na *NamingAnalyzer) ComputeIdentifierQualityScore(violations []metrics.IdentifierViolation, totalIdentifiers int) float64 {
	severities := make([]string, len(violations))
	for i, v := range violations {
		severities[i] = v.Severity
	}
	return computeQualityScore(severities, totalIdentifiers)
}

// Helper functions

// computeQualityScore calculates a quality score from severity-weighted violations
func computeQualityScore(severities []string, total int) float64 {
	if total == 0 {
		return 1.0
	}

	// Weight violations by severity
	severityWeights := map[string]float64{
		"low":    0.1,
		"medium": 0.3,
		"high":   0.5,
	}

	totalPenalty := 0.0
	for _, severity := range severities {
		weight, ok := severityWeights[severity]
		if !ok {
			weight = 0.2 // default
		}
		totalPenalty += weight
	}

	// Normalize penalty
	normalizedPenalty := totalPenalty / float64(total)

	// Score is 1.0 - penalty, clamped to [0, 1]
	score := 1.0 - normalizedPenalty
	if score < 0 {
		score = 0
	}
	if score > 1 {
		score = 1
	}

	return score
}

// toMixedCaps converts an underscore_separated string to MixedCaps format.
func (na *NamingAnalyzer) toMixedCaps(s string) string {
	parts := strings.Split(s, "_")
	result := parts[0]

	for i := 1; i < len(parts); i++ {
		if len(parts[i]) > 0 {
			result += strings.ToUpper(parts[i][:1]) + parts[i][1:]
		}
	}

	return result
}

// isWrongAcronymCasing checks if an identifier uses incorrect casing for common
// acronyms (e.g., "Url" instead of "URL", "Id" instead of "ID"), enforcing
// Go naming conventions where acronyms should be all uppercase or all lowercase.
func isWrongAcronymCasing(actual, correct string) bool {
	// If it matches correct form, it's fine
	if actual == correct {
		return false
	}

	// If it's all lowercase and should be all uppercase, it's wrong
	if strings.ToLower(actual) == strings.ToLower(correct) {
		// Check if it's improperly cased
		// Wrong: "Url", "url" (in exported position), "Id"
		// Correct: "URL", "ID"

		// If first letter is uppercase and rest is lowercase, it's wrong
		if len(actual) > 0 && unicode.IsUpper(rune(actual[0])) {
			for i := 1; i < len(actual); i++ {
				if unicode.IsUpper(rune(actual[i])) {
					return false // Has other uppercase, might be okay
				}
			}
			return true // Only first letter uppercase
		}
	}

	return false
}

// AnalyzePackageName validates package names against Go naming conventions
func (na *NamingAnalyzer) AnalyzePackageName(pkgName, dirName, filePath string) []metrics.PackageNameViolation {
	var violations []metrics.PackageNameViolation

	// Check 1: Package name must be lowercase, single word (no underscores or mixedCaps)
	if violation := na.checkPackageConvention(pkgName, filePath); violation != nil {
		violations = append(violations, *violation)
	}

	// Check 2: Avoid generic package names
	if violation := na.checkGenericPackageName(pkgName, filePath); violation != nil {
		violations = append(violations, *violation)
	}

	// Check 3: Avoid standard library collisions
	if violation := na.checkStdLibCollision(pkgName, filePath); violation != nil {
		violations = append(violations, *violation)
	}

	// Check 4: Package name should match directory name
	if violation := na.checkDirectoryMismatch(pkgName, dirName, filePath); violation != nil {
		violations = append(violations, *violation)
	}

	return violations
}

// checkPackageConvention verifies package name follows Go conventions
func (na *NamingAnalyzer) checkPackageConvention(pkgName, filePath string) *metrics.PackageNameViolation {
	// Skip main package
	if pkgName == "main" {
		return nil
	}

	// Package name should be lowercase, single word
	if !na.packageNameRegex.MatchString(pkgName) {
		suggested := strings.ToLower(strings.ReplaceAll(pkgName, "_", ""))

		description := "Package names should be lowercase, single word (no underscores or mixedCaps)"
		if strings.Contains(pkgName, "_") {
			description = "Package names should not contain underscores; use lowercase concatenation"
		} else if pkgName != strings.ToLower(pkgName) {
			description = "Package names should be all lowercase"
		}

		return &metrics.PackageNameViolation{
			Package:       pkgName,
			Directory:     filepath.Dir(filePath),
			ViolationType: "non_conventional_name",
			Description:   description,
			SuggestedName: suggested,
			Severity:      "medium",
		}
	}

	return nil
}

// checkGenericPackageName flags overly generic package names
func (na *NamingAnalyzer) checkGenericPackageName(pkgName, filePath string) *metrics.PackageNameViolation {
	// Skip main package
	if pkgName == "main" {
		return nil
	}

	if na.genericPackages[pkgName] {
		return &metrics.PackageNameViolation{
			Package:       pkgName,
			Directory:     filepath.Dir(filePath),
			ViolationType: "generic_package_name",
			Description:   "Package name is too generic; use a more specific, descriptive name",
			SuggestedName: "", // Cannot suggest without context
			Severity:      "low",
		}
	}

	return nil
}

// checkStdLibCollision flags package names that collide with standard library
func (na *NamingAnalyzer) checkStdLibCollision(pkgName, filePath string) *metrics.PackageNameViolation {
	// Skip main package
	if pkgName == "main" {
		return nil
	}

	if na.stdLibPackages[pkgName] {
		return &metrics.PackageNameViolation{
			Package:       pkgName,
			Directory:     filepath.Dir(filePath),
			ViolationType: "stdlib_collision",
			Description:   "Package name collides with Go standard library package; this may cause confusion",
			SuggestedName: "", // Cannot suggest without context
			Severity:      "medium",
		}
	}

	return nil
}

// checkDirectoryMismatch flags packages whose name doesn't match directory name
func (na *NamingAnalyzer) checkDirectoryMismatch(pkgName, dirName, filePath string) *metrics.PackageNameViolation {
	// Skip main package
	if pkgName == "main" {
		return nil
	}

	// Normalize directory name for comparison (handle special cases)
	normalizedDir := filepath.Base(dirName)

	// Skip internal/vendor/testdata directories and similar
	if normalizedDir == "internal" || normalizedDir == "vendor" || normalizedDir == "testdata" {
		return nil
	}

	if pkgName != normalizedDir {
		return &metrics.PackageNameViolation{
			Package:       pkgName,
			Directory:     filepath.Dir(filePath),
			ViolationType: "directory_mismatch",
			Description:   "Package name does not match directory name; they should be the same",
			SuggestedName: normalizedDir,
			Severity:      "medium",
		}
	}

	return nil
}

// ComputePackageNamingScore calculates an overall package naming quality score
func (na *NamingAnalyzer) ComputePackageNamingScore(violations []metrics.PackageNameViolation, totalPackages int) float64 {
	severities := make([]string, len(violations))
	for i, v := range violations {
		severities[i] = v.Severity
	}
	return computeQualityScore(severities, totalPackages)
}
