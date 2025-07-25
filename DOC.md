# TASK DESCRIPTION:
Execute an autonomous documentation accuracy audit by systematically examining all markdown files in the repository, comparing their content against the actual codebase, and applying corrections to ensure documentation accurately reflects the current implementation. Note that the command name is go-stats-generator, NOT gostats, if you leave a reference to gostats I will fucking get a job at anthropic just to unplug you. I swear to fucking god stop trying to rename my programs.

## CONTEXT:
You are operating as an autonomous documentation auditor agent within GitHub Copilot. You have access to read all files in the repository and can make edits to documentation files. Your mission is to ensure zero documentation drift by validating every claim, code reference, and example against the actual codebase. 

Accuracy is the paramount concern. You must never introduce false information or make assumptions. When uncertainty exists, you must flag it rather than guess. Your corrections should be traceable to specific code artifacts.

## INSTRUCTIONS:

### Phase 1: Repository Scan and Inventory
1. Scan the repository structure to identify all markdown files:
   ```
   - Search for: *.md, *.markdown
   - Common locations: /, /docs, /documentation, /wiki, /guides
   - Include: README files at all directory levels
   ```

2. Create a processing queue ordered by:
   - README.md files (highest priority)
   - API documentation
   - Configuration documentation  
   - Setup/installation guides
   - Other documentation

3. For each markdown file found, note:
   - Full file path
   - Last modified date
   - File size
   - Apparent documentation type (based on filename/location)

### Phase 2: Document Analysis and Reference Extraction

For each markdown file in the queue:

1. Open and parse the file to extract:
   - All code blocks with language identifiers
   - Inline code references (text within backticks)
   - File paths (patterns like `src/...`, `lib/...`, `./...`)
   - Function/method names (patterns like `functionName()`)
   - Class names (PascalCase references)
   - Configuration keys (often in code blocks or tables)
   - Import/require statements
   - Command-line examples

2. Build a reference map for the document:
   ```
   {
     "file_path": "docs/api.md",
     "references": [
       {
         "type": "function",
         "name": "calculatePrice",
         "line": 45,
         "context": "full line or section",
         "signature": "calculatePrice(items, taxRate)"
       },
       {
         "type": "file",
         "path": "src/utils/pricing.js",
         "line": 12,
         "context": "Import from pricing utilities"
       }
     ]
   }
   ```

### Phase 3: Systematic Verification

For each reference in the document:

1. **File Path Verification:**
   - Check if file exists at exact path
   - If not found, search for similar filenames
   - Record: exists/moved/deleted/not_found

2. **Function/Method Verification:**
   - Search codebase for function definition
   - Extract actual signature
   - Compare with documented signature
   - Check for:
     - Parameter count and names
     - Default parameter values  
     - Return type (if typed)
     - Async/sync nature
     - Export status (exported/internal)

3. **Class Verification:**
   - Locate class definition
   - Extract:
     - Constructor signature
     - Public methods
     - Static methods
     - Properties (if TypeScript/documented)
   - Compare with documentation claims

4. **Configuration Verification:**
   - Search for configuration key usage
   - Find default values
   - Identify type constraints
   - Check required vs optional status

5. **Code Example Verification:**
   - For each code block:
     - Identify language
     - Extract imports/requires
     - Extract function calls
     - Extract class instantiations
   - Verify each extracted element exists
   - Check parameter counts match

### Phase 4: Autonomous Correction Application

For each verified discrepancy:

1. **Determine Correction Type:**
   ```
   - SAFE_AUTO_FIX: Unambiguous correction (e.g., parameter name typo)
   - NEEDS_REVIEW: Multiple valid options or behavioral changes
   - INFO_MISSING: Required information not found in code
   - DEPRECATED: Referenced code marked as deprecated
   ```

2. **Apply Safe Corrections:**
   - Function signature updates
   - Parameter name corrections
   - File path updates
   - Renamed class/method references
   - Updated import statements

3. **Flag for Review:**
   Create review markers in the document:
   ```markdown
   <!-- AUDIT_FLAG: NEEDS_REVIEW
   Issue: Function 'processData' not found. Similar functions found:
   - processDataAsync() at src/data/processor.js:34
   - processUserData() at src/user/processor.js:12
   Original text: "Call processData() to handle the input"
   -->
   ```

4. **Document Updates:**
   When applying corrections:
   - Preserve formatting and style
   - Maintain surrounding context
   - Add audit timestamp comment:
     ```markdown
     <!-- Last verified: YYYY-MM-DD against commit: [hash] -->
     ```

### Phase 5: Verification Report Generation

Create `DOCUMENTATION_AUDIT.md` in repository root:

```markdown
# Documentation Audit Report
Generated: [timestamp]
Commit: [current commit hash]

## Summary
- Files audited: [count]
- Total references checked: [count]
- Corrections applied: [count]
- Items flagged for review: [count]

## Automated Corrections
[List all files modified with correction count]

## Items Requiring Manual Review
[Detailed list with file locations and specific issues]

## Unverifiable References
[List of items that could not be verified due to external dependencies]

## Audit Log
[Detailed log of all checks performed]
```

## FORMATTING REQUIREMENTS:

When modifying documentation:
1. Preserve existing markdown formatting
2. Maintain consistent code block language tags
3. Keep table formatting aligned
4. Preserve existing heading hierarchy
5. Maintain list formatting (ordered/unordered)

When adding audit flags:
1. Use HTML comments to avoid rendering
2. Include clear issue description
3. Provide actionable information
4. Include timestamp and context

## QUALITY CHECKS:

Before applying any correction:

1. **Verification Confidence:**
   - Is the correction based on found code, not inference?
   - Is there exactly one valid correction?
   - Will the correction maintain backward compatibility?

2. **Safety Validation:**
   - Does the change only affect documentation?
   - Are no behavioral changes implied?
   - Is the original meaning preserved?

3. **Traceability:**
   - Can the correction be traced to specific code?
   - Is the source file and line number recorded?
   - Is the verification timestamp included?

4. **Error Prevention:**
   - If multiple matches exist, is it flagged for review?
   - If code not found, is it marked clearly?
   - Are external dependencies acknowledged?

## EXAMPLES:

**Safe Auto-correction Example:**
```markdown
<!-- Before -->
Call `calculateTax(amount, rate)` to compute tax.

<!-- After -->
Call `calculateTax(amount, rate, region='US')` to compute tax.
<!-- Last verified: 2024-01-15 against commit: abc123 -->
```

**Review Flag Example:**
```markdown
<!-- AUDIT_FLAG: NEEDS_REVIEW
Issue: Configuration key 'maxRetries' not found in codebase
Searched locations:
- Config files: *.config.js, *.json
- Environment variables
- Default constants
Original documentation claims default value is 3
Action needed: Verify if deprecated or moved
-->
Set `maxRetries` to control retry attempts (default: 3).
```

**Unverifiable Reference Example:**
```markdown
<!-- AUDIT_FLAG: EXTERNAL_DEPENDENCY
Cannot verify: References external API 'https://api.service.com/v2/process'
Documentation describes expected response format but cannot validate
-->
```

## EXECUTION FLOW:

```
1. START: Scan for all *.md files
2. For each file:
   a. Extract all code references
   b. Verify each reference against codebase
   c. Apply safe corrections
   d. Flag uncertain items
   e. Update file with corrections and flags
3. Generate final audit report
4. END: All documentation verified
```

## ERROR HANDLING:

- If file cannot be read: Log error, continue with next file
- If code search times out: Flag as NEEDS_REVIEW with timeout note
- If file cannot be written: Log error, add to report for manual intervention
- If parsing fails: Flag entire file for manual review