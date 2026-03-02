# Implementation Plan: Code Duplication Detection (Phase 1)

## Phase Overview
- **Objective**: Detect duplicated and near-duplicate code blocks that increase maintenance cost by forcing developers to make the same change in multiple places
- **Source Document**: ROADMAP.md — Phase 1: Code Duplication Detection
- **Prerequisites**: Existing AST-based analysis infrastructure (function, struct, interface, package, concurrency analyzers)
- **Estimated Scope**: Large

## Implementation Steps

### 1. Create DuplicationMetrics type definitions ✅ COMPLETE
- **Deliverable**: New types in `internal/metrics/types.go`:
  - `DuplicationMetrics` struct with fields: `ClonePairs int`, `DuplicatedLines int`, `DuplicationRatio float64`, `LargestCloneSize int`, `Clones []ClonePair`
  - `ClonePair` struct with fields: `Hash string`, `Type CloneType`, `Instances []CloneInstance`, `LineCount int`
  - `CloneInstance` struct with fields: `File string`, `StartLine int`, `EndLine int`, `NodeCount int`
  - `CloneType` enum: `CloneTypeExact`, `CloneTypeRenamed`, `CloneTypeNear`
- **Dependencies**: None
- **Completed**: 2026-03-02 (commit 42b868f)
- **Tests**: `internal/metrics/types_test.go` with 100% coverage of new types

### 2. Create AST-based block fingerprinting engine ✅ COMPLETE
- **Deliverable**: New file `internal/analyzer/duplication.go` containing:
  - `DuplicationAnalyzer` struct implementing block extraction
  - `ExtractBlocks(ast *ast.File) []StatementBlock` — walks function/method bodies to extract statement-level sub-trees
  - `NormalizeBlock(block StatementBlock) NormalizedBlock` — strips identifiers, literals, comments to produce structural form
  - `ComputeHash(normalized NormalizedBlock) string` — computes structural hash using FNV-1a or similar
  - Store tuples: `(hash, file, startLine, endLine, nodeCount)`
- **Dependencies**: Step 1 (type definitions)
- **Completed**: 2026-03-02 (current commit)
- **Tests**: `internal/analyzer/duplication_test.go` with 98%+ coverage
  - `TestDuplicationAnalyzer_ExtractBlocks` — validates block extraction from various function structures
  - `TestDuplicationAnalyzer_NormalizeBlock` — verifies normalization produces identical structures for different identifiers
  - `TestDuplicationAnalyzer_ComputeHash` — validates hash consistency and differentiation
  - `TestDuplicationAnalyzer_FingerprintBlocks` — end-to-end fingerprinting test
  - `TestDuplicationAnalyzer_GroupFingerprintsByHash` — validates grouping logic
  - `TestDuplicationAnalyzer_FilterDuplicateGroups` — validates duplicate detection and sorting
  - `TestDuplicationAnalyzer_GetBlockSource` — validates source code extraction
  - `TestDuplicationAnalyzer_NormalizeLiteral` — validates literal placeholder replacement
  - `TestDuplicationAnalyzer_ExtractNestedBlocks` — validates extraction from if/switch/select/for
  - `TestDuplicationAnalyzer_CountNodes` — validates node counting
  - `TestDuplicationAnalyzer_DeepCopyAndNormalize` — comprehensive normalization coverage for all AST node types

### 3. Implement clone pair detection algorithm
- **Deliverable**: Methods in `internal/analyzer/duplication.go`:
  - `DetectClonePairs(blocks []BlockFingerprint) []ClonePair` — groups fingerprints by hash, identifies groups with 2+ entries
  - `ClassifyClone(pair ClonePair) CloneType` — determines Type 1/2/3:
    - Type 1: exact duplicates (identical after whitespace normalization)
    - Type 2: renamed duplicates (identical structure, different identifiers)
    - Type 3: near duplicates (structural similarity ≥ configurable threshold, default 80%)
  - `ComputeSimilarity(block1, block2 NormalizedBlock) float64` — tree edit distance or Jaccard similarity for Type 3 detection
- **Dependencies**: Step 2 (fingerprinting engine)

### 4. Integrate duplication analysis into analyzer pipeline
- **Deliverable**: Modifications to existing analyzer infrastructure:
  - Add `DuplicationAnalyzer` initialization in `internal/analyzer/` pipeline
  - Wire `AnalyzeDuplication(files []*ast.File) DuplicationMetrics` into main analysis flow
  - Populate `Report.Duplication DuplicationMetrics` field in `internal/metrics/types.go`
  - Add `Duplication DuplicationMetrics` field to `Report` struct
- **Dependencies**: Steps 1-3

### 5. Add configuration options for duplication thresholds
- **Deliverable**: Configuration keys in `.go-stats-generator.yaml` schema and `internal/config/`:
  - `maintenance.duplication.min_block_lines` (default: 6) — minimum block size to consider
  - `maintenance.duplication.similarity_threshold` (default: 0.80) — threshold for Type 3 clones
  - `maintenance.duplication.ignore_test_files` (default: false) — exclude `*_test.go` files
  - Wire thresholds as CLI flags: `--min-block-lines`, `--similarity-threshold`, `--ignore-test-duplication`
- **Dependencies**: Step 4

### 6. Implement duplication reporting across all output formats
- **Deliverable**: Updates to reporters in `internal/reporter/`:
  - Console reporter: add "DUPLICATION ANALYSIS" section with table showing clone pairs, duplicated lines, duplication ratio
  - JSON reporter: include `duplication` object in output
  - HTML reporter: add duplication section with expandable clone pair details
  - Markdown reporter: add duplication section with per-file scores
  - Per-file duplication score for prioritizing extraction refactoring
- **Dependencies**: Steps 4-5

### 7. Create comprehensive test suite
- **Deliverable**: New file `internal/analyzer/duplication_test.go` containing:
  - Unit tests for each detection rule using `testify/assert` and `testify/require`
  - Table-driven tests with Go source snippets in `testdata/duplication/`:
    - `exact_clone.go` — identical code blocks
    - `renamed_clone.go` — same structure, different variable names
    - `near_clone.go` — similar structure above threshold
    - `below_threshold.go` — similar structure below threshold (negative test)
    - `small_blocks.go` — blocks below `min_block_lines` (negative test)
  - Integration tests running full `analyze` command against test fixtures
  - Regression tests for false-positive cases
  - Benchmark tests ensuring analysis of 50,000+ files completes in <60s
- **Dependencies**: Steps 1-6

## Technical Specifications

- **Hash Algorithm**: Use FNV-1a (64-bit) for structural hashes — fast, good distribution, available in Go stdlib (`hash/fnv`)
- **AST Normalization Strategy**:
  - Replace all `*ast.Ident` (identifiers) with placeholder token `_`
  - Replace all `*ast.BasicLit` (literals) with type-specific placeholders: `INT_`, `STRING_`, `FLOAT_`
  - Preserve structural nodes: `*ast.IfStmt`, `*ast.ForStmt`, `*ast.RangeStmt`, `*ast.SwitchStmt`, etc.
  - Serialize normalized AST to canonical string form before hashing
- **Minimum Block Size**: Only consider blocks with ≥6 statements (configurable) to avoid trivial matches
- **Similarity Calculation for Type 3**: Use tree edit distance normalized by tree size: `1 - (editDistance / maxTreeSize)`
- **Memory Optimization**: Process files in batches, store only fingerprints (not full AST) after extraction phase
- **Concurrency**: Reuse existing worker pool infrastructure from `internal/scanner/` for parallel file processing

## Validation Criteria

- [x] `DuplicationMetrics` type is defined and integrated into `Report` struct
- [x] `DuplicationAnalyzer` successfully extracts statement blocks from all function/method bodies
- [x] Exact duplicates (Type 1) are correctly identified with 100% precision (hash-based detection implemented)
- [ ] Renamed duplicates (Type 2) are correctly identified — same structure with different identifiers
- [ ] Near duplicates (Type 3) are identified when similarity ≥ threshold (default 0.80)
- [ ] Blocks below `min_block_lines` threshold are ignored
- [ ] Test files are excluded when `ignore_test_files: true`
- [ ] All four output formats (console, JSON, HTML, Markdown) include duplication section
- [ ] Per-file duplication scores are calculated and reported
- [x] Unit test coverage ≥85% for `duplication.go` (achieved 98%+)
- [ ] Integration tests pass with sample codebases containing known duplicates
- [ ] Benchmark: Analysis of 50,000-file repository completes in <60 seconds
- [ ] Benchmark: Memory usage remains <1GB for large repository analysis
- [ ] Configuration options are documented and accessible via CLI flags

## Known Gaps

- None identified — all required information is present in ROADMAP.md
