TASK: Reorganize a Go codebase into maximally navigable file structure by moving code segments without modification, consolidating all interfaces into a single file, and maintaining continuous test validation.

CONTEXT:
You are reorganizing an existing Go codebase of unknown size and complexity. The goal is maximum on-disk navigability through intuitive file organization. You must preserve all functionality while creating a clear, predictable structure that any Go developer can navigate efficiently. All code behavior must remain identical - you are only permitted to move code between files.

REQUIREMENTS:

**Phase 1: Initial Assessment and Preparation**
- Run `go test ./...` and save the output as baseline.txt
- Create a file inventory listing all .go files and their primary contents
- Identify all interfaces across the codebase
- Catalog all shared constants, types, and utility functions
- Document the current package structure

**Phase 2: Interface Consolidation**
- Create either `interfaces.go` or `types.go` (choose based on existing conventions)
- For each interface found:
  1. Copy the interface definition to the consolidated file
  2. Add comment: `// Originally from: [source_file.go]`
  3. Run `go build ./...` to check for compilation errors
  4. If successful, remove the interface from its original location
  5. Run `go test ./...` to verify no regressions
  6. If tests fail, revert the change and investigate dependencies

**Phase 3: Structural Reorganization**
Apply these patterns based on file content:

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
  4. Run `go build ./...` after each struct migration
  5. Remove successfully migrated code from original file
  6. Run `go test ./...` to confirm no regressions

- **For shared constants:**
  1. Create `constants.go` if it doesn't exist
  2. Group constants by logical category with section comments
  3. Move constants maintaining their original comments
  4. Add `// Originally defined in: [source_file.go]`
  5. Update imports in affected files
  6. Validate with `go build ./...` and `go test ./...`

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
- **For packages with 20+ files:**
  1. Identify logical subdomains
  2. Create subdirectories for each subdomain
  3. Move related files maintaining import paths
  4. Update internal imports
  5. Run full test suite after each subdirectory

- **For mixed responsibility files:**
  1. Separate by primary responsibility
  2. HTTP handlers → `handlers.go` or `[resource]_handler.go`
  3. Database operations → `db.go` or `[model]_db.go`
  4. External API clients → `[service]_client.go`
  5. Business logic → `[domain]_service.go`

**Phase 5: Documentation Enhancement**
For each file after reorganization:
1. Add file-level comment explaining the file's purpose
2. Ensure every exported function has a comment starting with its name
3. Document any non-obvious organizational decisions
4. Create README.md in each package explaining the structure

OUTPUT FORMAT:
After each reorganization step, output:
```
MOVED: [description of what was moved]
FROM: [source_file.go]
TO: [destination_file.go]
TESTS: [PASS/FAIL] - [number] tests, [number] passed
BUILD: [SUCCESS/FAIL]
```

Final summary:
```
REORGANIZATION COMPLETE
Files created: [number]
Files modified: [number]
Files deleted: [number]
Test status: [PASS/FAIL]
Build status: [SUCCESS/FAIL]
```

QUALITY CRITERIA:
- Zero test regressions: baseline.txt matches final test output
- Zero build errors throughout the process
- Every exported symbol has documentation
- File names clearly indicate contents
- Related code is co-located
- No code logic modifications
- All moves are traced with comments

EXAMPLE:
Input: Codebase with user.go containing User struct, UserRole interface, user constants, and database functions

Process:
```
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
```
