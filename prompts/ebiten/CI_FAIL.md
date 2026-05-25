# CI Failure Analysis - go-stats-generator (Ebitengine Analysis)

## Overview

This file documents CI failures specific to the Ebitengine-optimized prompt variants for go-stats-generator. For the main CI failures and linting issues, see the root `/prompts/CI_FAIL.md` file.

## Domain-Specific Considerations for Game Development

The Ebitengine prompts are tailored for analyzing game development projects with Go. When failures occur in this context, consider:

1. **Graphics API Issues** - Ebitengine uses different graphics backends (OpenGL, Metal, DirectX)
2. **Platform-Specific Problems** - Game libraries often behave differently across platforms
3. **Real-time Performance** - Game analysis may have stricter performance requirements
4. **Input/Event Handling** - Game-specific event loops and input systems

## Recent Failures

When Ebitengine-related analysis fails, follow this resolution process:

1. **Check the failure mode:** Is it a code quality issue (same as main CI) or Ebitengine-specific?
2. **Identify the component:** Determine which analysis system failed (complexity, duplication, performance, etc.)
3. **Review game-specific code patterns:** Game code often uses patterns that look unusual in standard Go analysis
4. **Implement specialized fix:** Apply fixes that account for game development conventions

## Common Ebitengine Analysis Failure Categories

### Performance Analysis Issues
- False positives in frame rate critical sections
- Misunderstanding of game loop patterns
- Incorrect complexity metrics for render loops

### Graphics API Abstraction
- Platform-specific conditional code confusing the analyzer
- Graphics buffer management patterns
- Shader compilation handling

### Game Development Patterns
- Event loop and game state patterns
- Resource loading and unloading cycles
- Input polling vs callback patterns

## Resolution Strategy

1. Check if it's a code quality issue (apply fixes from main CI_FAIL.md)
2. Identify if it's an Ebitengine-specific false positive
3. Review game-specific code patterns for legitimate deviations from standard Go
4. Document any analyzer limitations with game development code
5. Consider adding Ebitengine-specific linting rules if needed

## Validation Commands

```bash
# Run full test suite
make test

# Run linting
make lint

# Build the project
make build

# Analyze a game project with the Ebitengine-optimized prompts
go-stats-generator analyze ./examples --skip-tests \
    --prompt prompts/ebiten \
    --max-function-length 35 \
    --max-complexity 10
```

## Notes

- The Ebitengine community uses specific patterns for game development
- Some "issues" flagged by standard linting may be intentional in game code
- Performance is critical in games - ensure analysis doesn't miss real issues
- Consider maintaining separate linting rules for game vs. general Go code
