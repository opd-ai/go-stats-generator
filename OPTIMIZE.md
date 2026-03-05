# TASK: Ensure all ALL_CAPS prompt files use `go-stats-generator` as their primary analysis and validation engine with consistent structure.

## Execution Mode
**Autonomous action** — rewrite each target file in-place. No user approval between files.

## Prerequisite
```bash
which go-stats-generator || go install github.com/opd-ai/go-stats-generator@latest
```

## Workflow

### Phase 0: Understand the Project
1. Read the project README to understand its purpose and build process.
2. List all ALL_CAPS `.md` files in the repository root.
3. Classify each as: exemplar (well-structured prompt), conversion target (needs rewrite), or non-prompt (skip).

### Phase 1: Inventory and Classify
For each ALL_CAPS `.md` file:
- Does it start with `# TASK:`? If not, it may be a report — skip.
- Does it include a `Phase 0: Understand the Codebase` step? If not, it needs one.
- Does it use `go-stats-generator` for baseline/validation? If not, it needs conversion.
- Does it hardcode project-specific paths, thresholds, or file references? If so, generalize.

### Phase 2: Convert
For each conversion target:
1. Read the existing prompt to understand its intent and execution mode.
2. Rewrite to follow consistent structure:
   - **Header**: `# TASK:` one-sentence objective
   - **Execution mode**: stated explicitly
   - **Phase 0**: Codebase understanding step tailored to the task
   - **Workflow**: baseline → action → validate using `go-stats-generator`
   - **Thresholds**: presented as tunable defaults
   - **Tiebreaker**: one sentence for ambiguous prioritization
3. Map each prompt to appropriate `go-stats-generator` sections and flags.
4. Remove bloat: redundant install blocks, excessive examples, overly prescriptive language.

### Phase 3: Validate
- Verify each rewritten file is ≤ its previous line count.
- Verify no hardcoded project-specific references remain.
- Verify all `go-stats-generator` commands use valid flags.

## Prompt Structure Template
```markdown
# TASK: [one-sentence objective]
## Execution Mode
## Prerequisite
## Phase 0: Understand the Codebase
## Workflow
### Phase 1: Baseline
### Phase 2: [Action]
### Phase 3: Validate
## Thresholds (tunable defaults)
## Output Format
## Tiebreaker
```

## Valid --sections Values
`functions`, `structs`, `interfaces`, `packages`, `patterns`, `complexity`, `documentation`, `generics`, `duplication`, `naming`, `placement`, `organization`, `burden`, `scores`, `suggestions`, `metadata`, `overview`

## Common Flag Combinations
| Domain | Sections | Extra Flags |
|--------|----------|-------------|
| Complexity | `functions` | `--max-complexity 10 --max-function-length 30` |
| Duplication | `duplication` | `--min-block-lines 6 --similarity-threshold 0.80` |
| Documentation | `documentation,naming` | `--min-doc-coverage 0.80` |
| Architecture | `packages,structs,interfaces` | — |
| Full audit | `functions,documentation,naming,packages,patterns` | — |

## Conversion Rules
- Preserve each prompt's original intent and tiebreaker rules.
- The tool is called `go-stats-generator` — never `gostats`.
- Present thresholds as tunable defaults, not absolutes.
- Every prompt must work against a codebase the agent has never seen before.

## Output Format
```
[file]: [old_lines] -> [new_lines] — [intent preserved: YES/NO]
```
