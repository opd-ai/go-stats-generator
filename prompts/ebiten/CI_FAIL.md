# TASK: Analyze CI/CD failures in the context of game development and Ebitengine projects

## Execution Mode
**Report generation only** — do NOT modify any source code unless specifically instructed.

## Objective
Analyze CI/CD pipeline failures that occur when testing or analyzing game development projects built with Go and Ebitengine. Game projects often have unique constraints and patterns that affect CI/CD pipeline behavior.

## Domain-Specific Considerations

Game development introduces special CI/CD challenges:

1. **Graphics/rendering issues** — Platform-specific graphics APIs (OpenGL, Metal, DirectX)
2. **Windowed/headless environments** — Games need display or virtual display
3. **Real-time performance** — Frame rate and latency requirements
4. **Platform-specific code** — Windows/Mac/Linux/Web require different handling
5. **External libraries** — Game engines often have heavy C/C++ dependencies
6. **Timing-sensitive tests** — Game loops have strict timing requirements

## Workflow

### Phase 1: Identify Game-Related CI Failures

1. **Locate the failure point:**
   - Does it fail during compilation?
   - Does it fail during test execution?
   - Does it fail during analysis (go-stats-generator)?
   - Does it fail during artifact generation?

2. **Determine failure type:**
   - Graphics initialization error?
   - Platform-specific build issue?
   - Performance regression?
   - Missing game assets?
   - Timing/frame rate issue?

3. **Check platform specificity:**
   - Does it fail on all platforms or specific ones?
   - Is it a graphics backend issue?
   - Is it a library availability issue?

### Phase 2: Game Development-Specific Failure Analysis

#### Graphics/Rendering Issues
- Ebitengine requires a graphics context (OpenGL/Metal/DirectX)
- Headless CI environments may not support graphics
- **Solution:** Use virtual display (xvfb), mock graphics, or use software rendering

#### Platform Compatibility
- Game code often has platform-specific imports/build tags
- CI may test on Linux but game targets Windows/Mac
- **Solution:** Test on all target platforms, use build constraints, test in Docker

#### Performance Analysis False Positives
- Game update/render loops may trigger complexity warnings
- Tight loops and state machines may appear over-complex
- **Solution:** Use domain-specific analysis rules, exclude hot paths, document patterns

#### Asset/Resource Issues
- Game resources (sprites, sounds, shaders) not available in CI
- Asset loading failures
- **Solution:** Mock resources in CI, use asset stubs, verify asset paths

#### Timing/Concurrency Issues
- Games run update loop at specific FPS
- Timing-sensitive tests may fail under load
- **Solution:** Mock time, use deterministic testing, increase timeouts

### Phase 3: Game-Specific Debugging Steps

1. **Check graphics environment:**
   ```bash
   glxinfo              # Check OpenGL support
   eglinfo              # Check EGL support
   xvfb-run -a ...     # Run with virtual display
   ```

2. **Verify build tags:**
   - Are platform-specific build tags correct?
   - Are game-specific tags being used?

3. **Check asset paths:**
   - Are relative paths working in CI?
   - Are assets being included in CI?

4. **Review game loop:**
   - Is timing too strict for CI?
   - Are frame rate assumptions breaking in CI?

5. **Test on target platforms:**
   - Does it fail on the actual target platform?
   - Is it a CI-specific environment issue?

### Phase 4: Create Resolution Plan

1. **For graphics issues:** Use xvfb or mock graphics
2. **For platform issues:** Add platform-specific handling or matrix testing
3. **For analysis issues:** Add domain-specific linting rules
4. **For asset issues:** Mock or stub assets in tests
5. **For timing issues:** Make tests deterministic, use mock time

## Output Format

```markdown
# Game Development CI Failure Analysis

## Summary
[Overview of failures specific to game development]

## Graphics Issues
[Issues related to rendering/graphics]

## Platform Issues
[Issues specific to Windows/Mac/Linux/Web]

## Performance Analysis Issues
[False positives from standard analysis tools]

## Asset/Resource Issues
[Missing or invalid game assets]

## Timing/Frame Rate Issues
[Issues from real-time game requirements]

## Resolution Plan
[Specific fixes for game development context]

## Validation
[How to test fixes in game development context]
```

## Common Game Development CI Patterns

- **Separate graphics testing:** Test game logic separately from graphics
- **Headless mode:** Support CI without graphics support
- **Platform matrix:** Test on all target platforms
- **Asset management:** Include test assets or mock them
- **Mock time:** Use deterministic time for testing
- **Performance budgets:** Set realistic performance targets for CI environments

## Tips

- Check if issue is CI-environment specific vs. actual code bug
- Use conditional compilation for game-specific code
- Mock expensive or platform-specific operations in tests
- Consider game development best practices from Ebitengine community
- Be aware that standard linters may not understand game development patterns
- Document any CI-specific workarounds in code comments
