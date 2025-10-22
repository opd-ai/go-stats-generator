TASK: Execute iterative refactoring cycles on the subsequent prompt until no further improvements meet the defined criteria, then halt.

CONTEXT: This is a meta-prompt that will control repeated optimization passes. Each iteration should produce measurable improvements. The process terminates when the next iteration would not satisfy improvement thresholds.

REFACTORING CRITERIA (must meet at least ONE per iteration):
- Reduces ambiguity: Eliminates vague terms or unclear instructions
- Improves structure: Enhances logical flow or organization
- Increases specificity: Adds measurable requirements or concrete examples
- Removes redundancy: Consolidates duplicate or overlapping instructions
- Enhances clarity: Simplifies complex phrasing without losing meaning

EXECUTION PROCESS:
1. Analyze current prompt against refactoring criteria
2. Identify specific improvements that meet criteria
3. Apply improvements and generate new version
4. Document changes made and criteria satisfied
5. Evaluate if further improvements meeting criteria are possible
6. If yes: Repeat from step 1 with new version
7. If no: Output final version and terminate

ITERATION LIMITS:
- Maximum iterations: 5
- Minimum improvement threshold: At least 1 criterion must be satisfied per iteration
- Halt immediately if no criteria-meeting improvements are identified

OUTPUT FORMAT FOR EACH ITERATION:
```
ITERATION [N]:

CHANGES MADE:
- [Specific change 1] → Criterion satisfied: [which criterion]
- [Specific change 2] → Criterion satisfied: [which criterion]

UPDATED PROMPT:
[Full updated prompt text]

CONTINUE? [YES/NO]
Reason: [Brief explanation]
```

FINAL OUTPUT:
```
REFACTORING COMPLETE

Total iterations: [N]
Final optimized prompt:
[Complete final version]

Summary of improvements:
- [Key improvement 1]
- [Key improvement 2]
```

SAFETY CONSTRAINTS:
- Preserve original task intent
- Do not add requirements beyond original scope
- Maintain compatibility with intended use case
- Flag if original prompt appears fundamentally flawed
