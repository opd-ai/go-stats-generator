# Go Template Fixes - Detailed Analysis

This document provides a comprehensive breakdown of all template-to-struct mismatches found in the HTML templates and their exact fixes.

## 🚨 Critical Issues Summary

**Templates Analyzed:**
- `/internal/reporter/templates/html/report.html`
- `/internal/reporter/templates/html/diff.html`

**Data Structures:**
- `metrics.Report` (main report data)
- `metrics.ComplexityDiff` (diff report data)
- `config.OutputConfig` (template configuration)

**Total Mismatches Found:** 8 critical issues

---

## 📋 Individual Fix Details

### **Fix #1: Overview Total Lines Field**

**Location:** `report.html:35`
```html
<!-- BROKEN -->
<h3>{{.Report.Overview.TotalLines}}</h3>

<!-- FIXED -->
<h3>{{.Report.Overview.TotalLinesOfCode}}</h3>
```

**Root Cause:** Template uses `TotalLines` but struct field is `TotalLinesOfCode`

**Struct Definition:**
```go
type OverviewMetrics struct {
    TotalLinesOfCode int `json:"total_lines_of_code"`  // ✅ Correct field
    // No TotalLines field exists                      // ❌ Template assumption
}
```

**Impact:** Template renders empty/zero value instead of actual line count

**✅ COMPLETED**

---

### **Fix #2: Overview Average Complexity (Non-existent Field)**

**Location:** `report.html:41`
```html
<!-- BROKEN -->
<h3>{{formatFloat .Report.Overview.AverageComplexity}}</h3>

<!-- FIXED - Option A: Remove entirely -->
<!-- Remove this stat card completely -->

<!-- FIXED - Option B: Calculate from functions -->
<h3>{{if .Report.Functions}}{{formatFloat (calculateAverageComplexity .Report.Functions)}}{{else}}0.00{{end}}</h3>
```

**Root Cause:** `OverviewMetrics` struct has no `AverageComplexity` field

**Struct Definition:**
```go
type OverviewMetrics struct {
    TotalLinesOfCode int `json:"total_lines_of_code"`
    TotalFunctions   int `json:"total_functions"`
    TotalMethods     int `json:"total_methods"`
    // No AverageComplexity field exists
}
```

**Impact:** Template fails with "can't evaluate field AverageComplexity" error

**Recommended Fix:** Remove the stat card since this metric isn't calculated at overview level

**✅ COMPLETED**

---

### **Fix #3: Function Line Count Access**

**Location:** `report.html:57`
```html
<!-- BROKEN -->
<td>{{.LineCount}}</td>

<!-- FIXED -->
<td>{{.Lines.Code}}</td>
```

**Root Cause:** Function metrics store line data in nested structure

**Struct Definition:**
```go
type FunctionMetrics struct {
    Lines LineMetrics `json:"lines"`  // ✅ Nested structure
    // No LineCount field
}

type LineMetrics struct {
    Total    int `json:"total"`
    Code     int `json:"code"`     // ✅ Actual code lines (excluding comments/blanks)
    Comments int `json:"comments"`
    Blank    int `json:"blank"`
}
```

**Impact:** Template shows empty value instead of actual line count

**✅ COMPLETED**

---

### **Fix #4: Function Complexity Access**

**Location:** `report.html:58`
```html
<!-- BROKEN -->
<td>{{.CyclomaticComplexity}}</td>

<!-- FIXED -->
<td>{{.Complexity.Cyclomatic}}</td>
```

**Root Cause:** Complexity is stored in nested structure, not direct field

**Struct Definition:**
```go
type FunctionMetrics struct {
    Complexity ComplexityScore `json:"complexity"`  // ✅ Nested structure
    // No CyclomaticComplexity field
}

type ComplexityScore struct {
    Cyclomatic   int     `json:"cyclomatic"`    // ✅ Actual cyclomatic complexity
    Cognitive    int     `json:"cognitive"`
    NestingDepth int     `json:"nesting_depth"`
    Overall      float64 `json:"overall"`
}
```

**Impact:** Template shows empty value instead of cyclomatic complexity score

**✅ COMPLETED**

---

### **Fix #5: Function Parameters and Returns**

**Location:** `report.html:59-60`
```html
<!-- BROKEN -->
<td>{{.ParameterCount}}</td>
<td>{{.ReturnCount}}</td>

<!-- FIXED -->
<td>{{.Signature.ParameterCount}}</td>
<td>{{.Signature.ReturnCount}}</td>
```

**Root Cause:** Parameter/return counts are in signature structure, not direct fields

**Struct Definition:**
```go
type FunctionMetrics struct {
    Signature FunctionSignature `json:"signature"`  // ✅ Nested structure
    // No ParameterCount/ReturnCount fields
}

type FunctionSignature struct {
    ParameterCount  int `json:"parameter_count"`  // ✅ Actual parameter count
    ReturnCount     int `json:"return_count"`     // ✅ Actual return count
    VariadicUsage   bool
    ErrorReturn     bool
    // ... other signature fields
}
```

**Impact:** Template shows empty values for function signature information

**✅ COMPLETED**

---

### **Fix #6: Diff Summary Severity vs Trend**

**Location:** `diff.html:37-39`
```html
<!-- BROKEN -->
<div class="summary-card {{severityClass .Diff.Summary.OverallSeverity}}">
    <h3>{{.Diff.Summary.OverallSeverity}}</h3>
    <p>Overall Severity</p>
</div>

<!-- FIXED -->
<div class="summary-card {{trendClass .Diff.Summary.OverallTrend}}">
    <h3>{{.Diff.Summary.OverallTrend}}</h3>
    <p>Overall Trend</p>
</div>
```

**Root Cause:** Summary has trend direction, not severity level

**Struct Definition:**
```go
type DiffSummary struct {
    OverallTrend TrendDirection `json:"overall_trend"`  // ✅ Actual field
    // No OverallSeverity field
}

type TrendDirection string
const (
    TrendImproving TrendDirection = "improving"
    TrendStable    TrendDirection = "stable"
    TrendDegrading TrendDirection = "degrading"
    TrendVolatile  TrendDirection = "volatile"
)
```

**Impact:** Template shows empty value instead of trend direction

**Additional Fix Needed:** Add `trendClass` function to template helpers

**✅ COMPLETED**

---

### **Fix #7: Regression/Improvement Range Variables**

**Location:** `diff.html:61-67, 94-100`
```html
<!-- BROKEN -->
{{range .Diff.Regressions}}
    <td>{{.EntityType}}</td>
    <td>{{.EntityName}}</td>
    <td>{{.MetricName}}</td>
    <td>{{formatFloat .OldValue}}</td>
    <td>{{formatFloat .NewValue}}</td>
    <td class="{{changeClass .PercentChange}}">{{formatChange .PercentChange}}</td>
{{end}}

<!-- FIXED -->
{{range .Diff.Regressions}}
    <td>{{.Type}}</td>
    <td>{{.Location}}</td>
    <td>{{.Description}}</td>
    <td>{{formatFloat .OldValue}}</td>
    <td>{{formatFloat .NewValue}}</td>
    <td class="{{changeClass .Delta.Percentage}}">{{formatChange .Delta.Percentage}}</td>
{{end}}
```

**Root Cause:** Range variables use non-existent field names

**Struct Definition:**
```go
type Regression struct {
    Type        RegressionType `json:"type"`         // ✅ Not EntityType
    Location    string         `json:"location"`     // ✅ Not EntityName  
    Description string         `json:"description"`  // ✅ Not MetricName
    OldValue    interface{}    `json:"old_value"`    // ✅ Correct
    NewValue    interface{}    `json:"new_value"`    // ✅ Correct
    Delta       Delta          `json:"delta"`        // ✅ Nested structure
    // No EntityType, EntityName, MetricName, PercentChange fields
}

type Delta struct {
    Percentage float64 `json:"percentage"`  // ✅ Actual percentage field
}
```

**Impact:** All regression/improvement data shows as empty in tables

**✅ COMPLETED**

---

### **Fix #8: Changes Range Variables**

**Location:** `diff.html:115-122`
```html
<!-- BROKEN -->
{{range .Diff.Changes}}
    <td>{{.EntityType}}</td>
    <td>{{.EntityName}}</td>
    <td>{{.MetricName}}</td>
    <td>{{.ThresholdExceeded}}</td>
{{end}}

<!-- FIXED -->
{{range .Diff.Changes}}
    <td>{{.Category}}</td>
    <td>{{.Name}}</td>
    <td>{{.Description}}</td>
    <td>{{if gt .Delta.Percentage 10.0}}Exceeded{{else}}OK{{end}}</td>
{{end}}
```

**Root Cause:** MetricChange struct uses different field names + no threshold field

**Struct Definition:**
```go
type MetricChange struct {
    Category    string      `json:"category"`     // ✅ Not EntityType
    Name        string      `json:"name"`         // ✅ Not EntityName
    Description string      `json:"description"`  // ✅ Not MetricName
    Delta       Delta       `json:"delta"`
    // No EntityType, EntityName, MetricName, ThresholdExceeded fields
}
```

**Impact:** All change data shows as empty, threshold logic doesn't work

**✅ COMPLETED**

---

## 🛠️ Implementation Priority

1. **CRITICAL** (Breaks rendering): Fixes #1, #3, #4, #5, #7, #8
2. **HIGH** (Missing functionality): Fix #2, #6  
3. **MEDIUM** (Enhancement): Add missing template helper functions

## 🧪 Testing Verification

After applying fixes, test with:
```bash
# Generate HTML report to verify fixes
./gostats analyze ./testdata/simple --format=html --output=test_report.html

# Generate diff report to verify diff fixes  
./gostats baseline create test-baseline ./testdata/simple
./gostats diff test-baseline ./testdata/simple --format=html --output=test_diff.html
```

## 📝 Additional Template Helper Needed

Add `trendClass` function to HTML reporter:
```go
func trendClass(trend metrics.TrendDirection) string {
    switch trend {
    case metrics.TrendImproving:
        return "trend-improving"
    case metrics.TrendDegrading:
        return "trend-degrading"
    case metrics.TrendVolatile:
        return "trend-volatile"
    default:
        return "trend-stable"
    }
}
```

---

*Generated by go-stats-generator template analysis - August 19, 2025*
