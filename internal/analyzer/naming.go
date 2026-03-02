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
}

// identifierContext tracks context information for identifier analysis
type identifierContext struct {
	inLoop        bool
	loopDepth     int
	receiverType  string
	packageName   string
	functionName  string
	isTestFile    bool
	validSingleLetters map[string]bool
}

// NewNamingAnalyzer creates a new naming analyzer
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
			"url":  "URL",
			"http": "HTTP",
			"https": "HTTPS",
			"id":   "ID",
			"api":  "API",
			"json": "JSON",
			"xml":  "XML",
			"sql":  "SQL",
			"html": "HTML",
			"css":  "CSS",
			"eof":  "EOF",
			"ip":   "IP",
			"tcp":  "TCP",
			"udp":  "UDP",
			"rpc":  "RPC",
			"tls":  "TLS",
			"ssl":  "SSL",
			"grpc": "GRPC",
			"ui":   "UI",
			"uri":  "URI",
			"uuid": "UUID",
			"ascii": "ASCII",
			"utf":  "UTF",
		},
	}
}

// AnalyzeFileNames checks file names against Go naming conventions
func (na *NamingAnalyzer) AnalyzeFileNames(filePaths []string) []metrics.FileNameViolation {
	var violations []metrics.FileNameViolation

	for _, filePath := range filePaths {
		// Skip non-Go files
		if !strings.HasSuffix(filePath, ".go") {
			continue
		}

		fileName := filepath.Base(filePath)
		dirName := filepath.Base(filepath.Dir(filePath))

		// Check snake_case
		if violation := na.checkSnakeCase(filePath, fileName); violation != nil {
			violations = append(violations, *violation)
		}

		// Check stuttering
		if violation := na.checkStuttering(filePath, fileName, dirName); violation != nil {
			violations = append(violations, *violation)
		}

		// Check generic names
		if violation := na.checkGenericName(filePath, fileName); violation != nil {
			violations = append(violations, *violation)
		}

		// Check test suffix
		if violation := na.checkTestSuffix(filePath, fileName); violation != nil {
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
	if totalFiles == 0 {
		return 1.0
	}

	// Weight violations by severity
	severityWeights := map[string]float64{
		"low":    0.1,
		"medium": 0.3,
		"high":   0.5,
	}

	totalPenalty := 0.0
	for _, v := range violations {
		weight, ok := severityWeights[v.Severity]
		if !ok {
			weight = 0.2 // default
		}
		totalPenalty += weight
	}

	// Normalize penalty (max penalty = 1.0 per file)
	normalizedPenalty := totalPenalty / float64(totalFiles)

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

// AnalyzeIdentifiers walks the AST to analyze all identifiers in a file
func (na *NamingAnalyzer) AnalyzeIdentifiers(file *ast.File, filePath string, fset *token.FileSet) []metrics.IdentifierViolation {
	var violations []metrics.IdentifierViolation
	
	ctx := &identifierContext{
		packageName:   file.Name.Name,
		isTestFile:    strings.HasSuffix(filePath, "_test.go"),
		validSingleLetters: make(map[string]bool),
	}
	
	// Walk the AST to analyze identifiers
	ast.Inspect(file, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.FuncDecl:
			// Analyze function/method name
			if node.Name != nil && node.Name.Name != "_" {
				idType := "function"
				if node.Recv != nil && len(node.Recv.List) > 0 {
					idType = "method"
					ctx.receiverType = na.getReceiverTypeName(node.Recv)
				} else {
					ctx.receiverType = ""
				}
				
				ctx.functionName = node.Name.Name
				pos := fset.Position(node.Pos())
				
				// Check all identifier rules
				if v := na.checkMixedCaps(node.Name.Name, ctx); v != nil {
					v.File = filePath
					v.Line = pos.Line
					v.Type = idType
					violations = append(violations, *v)
				}
				
				if v := na.checkAcronymCasing(node.Name.Name, ctx); v != nil {
					v.File = filePath
					v.Line = pos.Line
					v.Type = idType
					violations = append(violations, *v)
				}
				
				if v := na.checkIdentifierStuttering(node.Name.Name, ctx); v != nil {
					v.File = filePath
					v.Line = pos.Line
					v.Type = idType
					violations = append(violations, *v)
				}
			}
			
		case *ast.GenDecl:
			// Analyze type, const, and var declarations
			for _, spec := range node.Specs {
				switch s := spec.(type) {
				case *ast.TypeSpec:
					if s.Name != nil && s.Name.Name != "_" {
						pos := fset.Position(s.Pos())
						
						if v := na.checkMixedCaps(s.Name.Name, ctx); v != nil {
							v.File = filePath
							v.Line = pos.Line
							v.Type = "type"
							violations = append(violations, *v)
						}
						
						if v := na.checkAcronymCasing(s.Name.Name, ctx); v != nil {
							v.File = filePath
							v.Line = pos.Line
							v.Type = "type"
							violations = append(violations, *v)
						}
						
						if v := na.checkIdentifierStuttering(s.Name.Name, ctx); v != nil {
							v.File = filePath
							v.Line = pos.Line
							v.Type = "type"
							violations = append(violations, *v)
						}
					}
					
				case *ast.ValueSpec:
					// Analyze const and var names
					idType := "var"
					if node.Tok == token.CONST {
						idType = "const"
					}
					
					for _, name := range s.Names {
						if name != nil && name.Name != "_" {
							pos := fset.Position(name.Pos())
							
							if v := na.checkMixedCaps(name.Name, ctx); v != nil {
								v.File = filePath
								v.Line = pos.Line
								v.Type = idType
								violations = append(violations, *v)
							}
							
							if v := na.checkSingleLetterName(name.Name, idType, ctx); v != nil {
								v.File = filePath
								v.Line = pos.Line
								violations = append(violations, *v)
							}
							
							if v := na.checkAcronymCasing(name.Name, ctx); v != nil {
								v.File = filePath
								v.Line = pos.Line
								v.Type = idType
								violations = append(violations, *v)
							}
						}
					}
				}
			}
			
		case *ast.ForStmt, *ast.RangeStmt:
			// Track loop scope for valid single-letter names
			ctx.inLoop = true
			ctx.loopDepth++
			
		case *ast.AssignStmt:
			// Track loop variables
			if ctx.inLoop {
				for _, lhs := range node.Lhs {
					if ident, ok := lhs.(*ast.Ident); ok {
						if len(ident.Name) == 1 && (ident.Name == "i" || ident.Name == "j" || ident.Name == "k") {
							ctx.validSingleLetters[ident.Name] = true
						}
					}
				}
			}
		}
		
		return true
	})
	
	return violations
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
func (na *NamingAnalyzer) checkSingleLetterName(name string, idType string, ctx *identifierContext) *metrics.IdentifierViolation {
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

// checkAcronymCasing detects improper acronym casing
func (na *NamingAnalyzer) checkAcronymCasing(name string, ctx *identifierContext) *metrics.IdentifierViolation {
	// Check for common acronyms with wrong casing
	nameLower := strings.ToLower(name)
	
	for acronym, correctForm := range na.acronyms {
		acronymLen := len(acronym)
		
		// Look for the acronym in different positions
		// Beginning of name: "Url" -> "URL", "UrlParser" -> "URLParser"
		if strings.HasPrefix(nameLower, acronym) {
			actualPrefix := name[:acronymLen]
			
			// Check if it's incorrectly cased
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
		}
		
		// Middle/end of name: "GetUrl" -> "GetURL", "UserId" -> "UserID"
		// We need to find word boundaries in MixedCaps names
		for i := 1; i < len(name)-acronymLen+1; i++ {
			// Check if we're at a word boundary (uppercase letter before this position)
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
	}
	
	return nil
}

// checkIdentifierStuttering detects stuttering in identifiers
func (na *NamingAnalyzer) checkIdentifierStuttering(name string, ctx *identifierContext) *metrics.IdentifierViolation {
	nameLower := strings.ToLower(name)
	
	// Check method stuttering: User.GetUser, User.UserName
	if ctx.receiverType != "" {
		receiverLower := strings.ToLower(ctx.receiverType)
		
		// Method name starts with receiver type
		if strings.HasPrefix(nameLower, receiverLower) && len(name) > len(ctx.receiverType) {
			// GetUser from User is okay, but UserName from User stutters
			if !strings.HasPrefix(nameLower, "get"+receiverLower) &&
			   !strings.HasPrefix(nameLower, "set"+receiverLower) &&
			   !strings.HasPrefix(nameLower, "new"+receiverLower) {
				suggested := name[len(ctx.receiverType):]
				if len(suggested) > 0 {
					return &metrics.IdentifierViolation{
						Name:          name,
						ViolationType: "stuttering",
						Description:   "Method name repeats receiver type (e.g., User.UserName should be User.Name)",
						SuggestedName: suggested,
						Severity:      "low",
					}
				}
			}
		}
	}
	
	// Check package stuttering: package user, func NewUser (okay), type UserService (stutters)
	if ctx.packageName != "" && ctx.packageName != "main" {
		packageLower := strings.ToLower(ctx.packageName)
		
		// Exported name starts with package name
		if unicode.IsUpper(rune(name[0])) && strings.HasPrefix(nameLower, packageLower) && len(name) > len(ctx.packageName) {
			// NewUser, ParseUser are okay (common constructors/operations)
			// But UserService, UserHandler stutter
			if !strings.HasPrefix(ctx.functionName, "New") &&
			   !strings.HasPrefix(ctx.functionName, "Parse") &&
			   !strings.HasPrefix(ctx.functionName, "Make") {
				// Only flag types and exported vars/consts, not all functions
				// This is a softer check
				return &metrics.IdentifierViolation{
					Name:          name,
					ViolationType: "package_stuttering",
					Description:   "Exported name repeats package name (e.g., user.UserService should be user.Service)",
					SuggestedName: name[len(ctx.packageName):],
					Severity:      "low",
				}
			}
		}
	}
	
	return nil
}

// ComputeIdentifierQualityScore calculates an overall identifier naming quality score
func (na *NamingAnalyzer) ComputeIdentifierQualityScore(violations []metrics.IdentifierViolation, totalIdentifiers int) float64 {
	if totalIdentifiers == 0 {
		return 1.0
	}
	
	// Weight violations by severity
	severityWeights := map[string]float64{
		"low":    0.1,
		"medium": 0.3,
		"high":   0.5,
	}
	
	totalPenalty := 0.0
	for _, v := range violations {
		weight, ok := severityWeights[v.Severity]
		if !ok {
			weight = 0.2
		}
		totalPenalty += weight
	}
	
	// Normalize penalty
	normalizedPenalty := totalPenalty / float64(totalIdentifiers)
	
	score := 1.0 - normalizedPenalty
	if score < 0 {
		score = 0
	}
	if score > 1 {
		score = 1
	}
	
	return score
}

// Helper functions

func (na *NamingAnalyzer) getReceiverTypeName(recv *ast.FieldList) string {
	if recv == nil || len(recv.List) == 0 {
		return ""
	}
	
	field := recv.List[0]
	switch t := field.Type.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		if ident, ok := t.X.(*ast.Ident); ok {
			return ident.Name
		}
	}
	
	return ""
}

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
