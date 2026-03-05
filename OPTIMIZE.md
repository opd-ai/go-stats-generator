# TASK: Convert remaining ALL_CAPS prompt files to use `go-stats-generator` as their primary analysis and validation engine.

## Execution Mode
**Autonomous action** — rewrite each target file in-place. No user approval between files.

## Prerequisite
```bash
which go-stats-generator || go install github.com/opd-ai/go-stats-generator@latest
```

## Workflow

### Phase 1: Inventory
1. List all ALL_CAPS `.md` files in the repository root.
2. Classify each as: exemplar (do not modify), already optimized, non-prompt (skip), or conversion target.

### Phase 2: Convert
For each conversion target:
1. Read the existing prompt to understand its intent and execution mode.
2. Rewrite to follow this template structure:
   - **Header**: `# TASK:` one-sentence objective
   - **Execution mode**: stated explicitly
   - **Prerequisite**: single-line install check
   - **Workflow**: 3 phases (baseline, action, validate) using `go-stats-generator`
   - **Thresholds**: only the thresholds relevant to THIS prompt
   - **Output format**: 3–5 line structural outline
   - **Tiebreaker**: one sentence for ambiguous prioritization
3. Map each prompt to appropriate `go-stats-generator` features:
   - Complexity prompts → `--sections functions` with `--max-complexity`, `--max-function-length`
   - Duplication prompts → `--sections duplication` with `--min-block-lines`, `--similarity-threshold`
   - Documentation prompts → `--sections documentation,naming` with `--min-doc-coverage`
   - Architecture prompts → `--sections packages,structs,interfaces`
   - All prompts → `go-stats-generator diff baseline.json post.json` for validation
4. Remove bloat:
   - Cut redundant installation blocks (one line is enough)
   - Cut section exclusion lists (use `--sections <relevant>` instead)
   - Cut example workflows longer than 5 lines
   - Cut "you are absolutely forbidden" language

### Phase 3: Validate
- Verify each rewritten file is 80–150 lines.
- Verify no `gostats` references remain.
- Verify all `go-stats-generator` commands use valid flags.
- Verify execution mode is explicitly stated.

## Conversion Rules
- Preserve each prompt's original intent and tiebreaker rules.
- The tool is called `go-stats-generator` — never `gostats`.
- Use `--sections <relevant>` instead of listing excluded sections.
- One-line prerequisite: `which go-stats-generator || go install github.com/opd-ai/go-stats-generator@latest`

## Valid `--sections` Values
`functions`, `structs`, `interfaces`, `packages`, `patterns`, `complexity`, `documentation`, `generics`, `duplication`, `naming`, `placement`, `organization`, `burden`, `scores`, `suggestions`, `metadata`, `overview`

## Output Format
```
[file]: [old_lines] -> [new_lines] — [intent preserved: YES/NO]
```
## Prompt Structure Template
Each converted prompt should follow this exact structure:

```markdown
# TASK: [one-sentence objective]
## Execution Mode
[Autonomous action | Report generation only | Interactive]
## Prerequisite
[single-line install check]
## Workflow
### Phase 1: Baseline
[go-stats-generator analyze command]
### Phase 2: [Action]
[numbered steps specific to this prompt]
### Phase 3: Validate
[go-stats-generator diff command]
## Thresholds
[only relevant thresholds]
## Output Format
[3-5 line template]
## Tiebreaker
[one sentence]
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
