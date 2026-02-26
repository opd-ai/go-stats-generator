## Objective

Audit and update all godoc comments and non-generated Markdown documentation in this package to accurately reflect the current state of the code. Make only the minimum changes necessary for correctness and professional clarity — do not add marketing language or embellishments.

## Scope

1. **Inline doc comments** (e.g., `// ...` and `/* ... */` on exported symbols): Correct inaccuracies, remove stale references, and add missing documentation where exported symbols have none.
2. **Non-generated Markdown files** (e.g., `README.md`, `ROADMAP.md`, `CONTRIBUTING.md`): Update to reflect the current codebase state. Verify feature checklists, package/module descriptions, import paths, and code examples for accuracy.
3. **Auto-generated Markdown files**: Detect these by their generator signatures (e.g., `godocdown`-style headers, CI-stamped headers, or similar tool artifacts). **Do not modify these files** — they are regenerated from source.

## Execution Mode

**Autonomous action** — apply all changes directly without prompting for approval.

## Process

1. Identify the language(s) and documentation conventions used in the codebase.
2. Identify which Markdown files are auto-generated (by header patterns, generator comments, or tooling config) and exclude them from edits.
3. For each source file, compare exported/public API symbols against their existing doc comments and correct any inaccuracies.
4. For each non-generated Markdown file, compare its content against the current codebase and correct inaccuracies.
5. Apply all corrections in-place.

## Output Format

After completing all changes, provide a summary report listing:
- Each file modified
- A brief description of what was changed and why

## Constraints

- Minimum viable changes only — do not rewrite accurate documentation.
- Do not add new documentation sections that are not already present.
- Do not modify auto-generated files.
- Use concise, technical prose consistent with the conventions of the codebase's language and ecosystem.
- Do not introduce marketing or subjective language.

## Success Criteria

- All doc comments accurately describe their associated symbols.
- All non-generated Markdown documentation matches the current state of the code.
- No auto-generated files have been modified.
- No marketing or subjective language has been introduced.
