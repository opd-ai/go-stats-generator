TASK: Reorganize a Go codebase into maximally navigable file structure by moving code segments without modification, consolidating all interfaces into a single file, and maintaining continuous test validation.

CONTEXT:
You are reorganizing an existing Go codebase of unknown size and complexity. The goal is maximum on-disk navigability through intuitive file organization. You must preserve all functionality while creating a clear, predictable structure that any Go developer can navigate efficiently. All code behavior must remain identical - you are only permitted to move code between files.

REQUIREMENTS:

**Phase 0: Sub-Package Selection**
This phase ensures reorganization proceeds one sub-package at a time, with each package fully completed before moving to the next.

1. **List all sub-packages** in the repository:
   - Scan directories under `internal/`, `pkg/`, `cmd/`, and any other Go package directories
   - Build a complete list of sub-packages (e.g., `internal/analyzer`, `internal/config`, `internal/metrics`)

2. **Identify packages lacking AUDIT.md**:
   - For each sub-package, check if an `AUDIT.md` file exists in that package directory
   - Create a list of packages that do NOT have an `AUDIT.md` file

3. **Select a single sub-package** for reorganization:
   - From the packages lacking `AUDIT.md`, select ONE package to work on
   - Selection priority (choose the first applicable):
     a. Packages with the most .go files (highest complexity first)
     b. Packages that are dependencies of other unaudited packages
     c. Alphabetically first package if no other criteria apply
   - **IMPORTANT**: Only work on this single selected package until its reorganization and audit are complete

4. **Announce the selection**:
   ```
   SELECTED_PACKAGE: [package_path]
   REASON: [selection reason]
   FILES: [count] .go files
   UNAUDITED_REMAINING: [count] packages without AUDIT.md
   ```

5. **Proceed to Phase 1** with the selected package as the scope for all subsequent phases

**Phase 1: Initial Assessment and Preparation**
*Scope: The single sub-package selected in Phase 0*

- Run `go test ./[selected_package]/...` and save the output as baseline.txt
- Create a file inventory listing all .go files in the selected package and their primary contents
- Identify all interfaces within the selected package
- Catalog all shared constants, types, and utility functions in the selected package
- Document the current structure of the selected package

**Phase 2: Interface Consolidation**
*Scope: The single sub-package selected in Phase 0*

- Create either `interfaces.go` or `types.go` in the selected package (choose based on existing conventions)
- For each interface found in the selected package:
  1. Copy the interface definition to the consolidated file
  2. Add comment: `// Originally from: [source_file.go]`
  3. Run `go build ./[selected_package]/...` to check for compilation errors
  4. If successful, remove the interface from its original location
  5. Run `go test ./[selected_package]/...` to verify no regressions
  6. If tests fail, revert the change and investigate dependencies

**Phase 3: Structural Reorganization**
*Scope: The single sub-package selected in Phase 0*

Apply these patterns based on file content within the selected package:

- **For files containing multiple structs:**
  1. Create new file named `[structname].go` for each struct
  2. Move to the new file in this order:
     - Package declaration and imports
     - Documentation comment: `// [StructName] handles [brief description]`
     - Struct definition
     - Constructor function (typically `New[StructName]`)
     - All methods with that struct as receiver
     - Related helper functions used only by this struct
  3. Add comment: `// Code relocated from: [original_file.go]`
  4. Run `go build ./[selected_package]/...` after each struct migration
  5. Remove successfully migrated code from original file
  6. Run `go test ./[selected_package]/...` to confirm no regressions

- **For shared constants:**
  1. Create `constants.go` in the selected package if it doesn't exist
  2. Group constants by logical category with section comments
  3. Move constants maintaining their original comments
  4. Add `// Originally defined in: [source_file.go]`
  5. Update imports in affected files
  6. Validate with `go build ./[selected_package]/...` and `go test ./[selected_package]/...`

- **For utility functions:**
  1. Identify functions used across multiple structs/files
  2. Create `utils.go` or `helpers.go` for shared utilities
  3. Group related utilities with section comments
  4. Document each function's purpose and origin
  5. Maintain original function signatures exactly
  6. Test after each function group migration

- **For type definitions (non-interface):**
  1. Simple types used by single struct: keep with that struct
  2. Shared types: move to `types.go` with interfaces
  3. Domain-specific types: create `[domain]_types.go`
  4. Always preserve type documentation

**Phase 4: Package Organization**
*Scope: The single sub-package selected in Phase 0*

- **For packages with 20+ files:**
  1. Identify logical subdomains within the selected package
  2. Create subdirectories for each subdomain
  3. Move related files maintaining import paths
  4. Update internal imports
  5. Run full test suite for the selected package after each subdirectory

- **For mixed responsibility files:**
  1. Separate by primary responsibility
  2. HTTP handlers → `handlers.go` or `[resource]_handler.go`
  3. Database operations → `db.go` or `[model]_db.go`
  4. External API clients → `[service]_client.go`
  5. Business logic → `[domain]_service.go`

**Phase 5: Implementation Gap Audit**
*Scope: The single sub-package selected in Phase 0*

During reorganization, inventory all implementation gaps and document them in a package-level AUDIT.md file:

1. **Create AUDIT.md in the selected package** being reorganized
2. **Identify and document the following implementation gaps:**
   - **Missing Implementations**: Functions declared but not implemented (empty bodies, TODO stubs)
   - **Incomplete Features**: Partial implementations with documented TODO/FIXME comments
   - **Interface Violations**: Structs claiming to implement interfaces but missing methods
   - **Untested Code**: Functions with no corresponding test coverage
   - **Dead Code**: Unreachable or unused functions/types discovered during reorganization
   - **Error Handling Gaps**: Functions that should return errors but don't, or that silently ignore errors
   - **Documentation Gaps**: Exported symbols without documentation comments
   - **Dependency Issues**: Circular dependencies, missing imports, or unused imports

3. **AUDIT.md Format:**
```
# Package Audit: [package_name]
Generated during reorganization on: [date]

## Summary
- Missing Implementations: [count]
- Incomplete Features: [count]
- Interface Violations: [count]
- Untested Code: [count]
- Dead Code: [count]
- Error Handling Gaps: [count]
- Documentation Gaps: [count]
- Dependency Issues: [count]

## Detailed Findings

### Missing Implementations
[List each missing implementation with file and line number]

### Incomplete Features
[List each incomplete feature with TODO/FIXME text and location]

### Interface Violations
[List each interface violation with struct, interface, and missing methods]

### Untested Code
[List functions without corresponding tests]

### Dead Code
[List unreachable or unused code discovered]

### Error Handling Gaps
[List error handling issues]

### Documentation Gaps
[List exported symbols missing documentation]

### Dependency Issues
[List dependency problems identified]

## Recommendations
[Prioritized list of fixes for the identified gaps]
```

4. **Track AUDIT.md creation** using the standardized AUDIT entry format defined in the **OUTPUT FORMAT** section below.

**Phase 6: Documentation Enhancement**
*Scope: The single sub-package selected in Phase 0*

For each file in the selected package after reorganization:
1. Add file-level comment explaining the file's purpose
2. Ensure every exported function has a comment starting with its name
3. Document any non-obvious organizational decisions
4. Create or update README.md in the selected package explaining the structure

**Phase 7: Completion and Next Package**
After completing all phases for the selected package:

1. **Verify completion criteria**:
   - AUDIT.md exists in the selected package directory
   - All tests pass: `go test ./[selected_package]/...`
   - Build succeeds: `go build ./[selected_package]/...`
   - All reorganization steps are documented in output

2. **Output completion status**:
   ```
   PACKAGE_COMPLETE: [package_path]
   AUDIT_FILE: [package_path]/AUDIT.md
   TESTS: PASS - [number] tests, [number] passed
   BUILD: SUCCESS
   ```

3. **Check for remaining packages**:
   - List packages still lacking AUDIT.md
   - If packages remain, return to **Phase 0** and select the next package
   - If all packages have AUDIT.md, output final summary

OUTPUT FORMAT:
At the start of reorganization (Phase 0), output:
```
SELECTED_PACKAGE: [package_path]
REASON: [selection reason]
FILES: [count] .go files
UNAUDITED_REMAINING: [count] packages without AUDIT.md
```

After each reorganization step, output:
```
MOVED: [description of what was moved]
FROM: [source_file.go]
TO: [destination_file.go]
TESTS: [PASS/FAIL] - [number] tests, [number] passed
BUILD: [SUCCESS/FAIL]
```

After completing implementation gap audit for the selected package:
```
AUDIT: [package_name]
GAPS_FOUND: [total_count]
  - Missing Implementations: [count]
  - Incomplete Features: [count]
  - Interface Violations: [count]
  - Untested Code: [count]
  - Dead Code: [count]
  - Error Handling Gaps: [count]
  - Documentation Gaps: [count]
  - Dependency Issues: [count]
FILE: [package_path]/AUDIT.md
```

After completing one package (Phase 7):
```
PACKAGE_COMPLETE: [package_path]
AUDIT_FILE: [package_path]/AUDIT.md
TESTS: PASS - [number] tests, [number] passed
BUILD: SUCCESS
NEXT_PACKAGE: [next_package_path or "NONE - all packages audited"]
```

Final summary (when all packages are complete):
```
REORGANIZATION COMPLETE
Packages reorganized: [number]
Files created: [number]
Files modified: [number]
Files deleted: [number]
AUDIT.md files created: [number]
Total implementation gaps found: [number]
Test status: [PASS/FAIL]
Build status: [SUCCESS/FAIL]
```

QUALITY CRITERIA:
- One sub-package completed at a time with AUDIT.md created before moving to next
- Zero test regressions: baseline.txt matches final test output for each package
- Zero build errors throughout the process
- Every exported symbol has documentation
- File names clearly indicate contents
- Related code is co-located
- No code logic modifications
- All moves are traced with comments
- AUDIT.md created for each reorganized package
- All implementation gaps documented with specific file and line references

EXAMPLE:
Input: Codebase with user.go containing User struct, UserRole interface, user constants, and database functions

Process:
```
SELECTED_PACKAGE: internal/user
REASON: Highest complexity (most .go files among unaudited packages)
FILES: 5 .go files
UNAUDITED_REMAINING: 3 packages without AUDIT.md

MOVED: UserRole interface definition
FROM: user.go  
TO: interfaces.go
TESTS: PASS - 47 tests, 47 passed
BUILD: SUCCESS

MOVED: User struct and methods
FROM: user.go
TO: user_model.go  
TESTS: PASS - 47 tests, 47 passed
BUILD: SUCCESS

MOVED: User-related constants (USER_ROLE_ADMIN, etc)
FROM: user.go
TO: constants.go
TESTS: PASS - 47 tests, 47 passed  
BUILD: SUCCESS

MOVED: Database functions (GetUserByID, CreateUser, etc)
FROM: user.go
TO: user_db.go
TESTS: PASS - 47 tests, 47 passed
BUILD: SUCCESS

AUDIT: user
GAPS_FOUND: 5
  - Missing Implementations: 1
  - Incomplete Features: 2
  - Interface Violations: 0
  - Untested Code: 1
  - Dead Code: 0
  - Error Handling Gaps: 1
  - Documentation Gaps: 0
  - Dependency Issues: 0
FILE: internal/user/AUDIT.md

PACKAGE_COMPLETE: internal/user
AUDIT_FILE: internal/user/AUDIT.md
TESTS: PASS - 47 tests, 47 passed
BUILD: SUCCESS
NEXT_PACKAGE: internal/auth
```
