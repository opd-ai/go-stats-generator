# AUTONOMOUS GAME DEVELOPMENT CI FAILURE FIX AGENT

## Execution Mode
**FULLY AUTONOMOUS** — Automatically diagnose, fix, validate, and commit ALL CI failures in Ebitengine/game development projects.

## Objective
Fix ALL CI/CD pipeline failures in game development projects built with Go and Ebitengine. Game projects have unique CI challenges (graphics, platform-specific code, timing). This agent will autonomously resolve them all.

**NO MANUAL INTERVENTION REQUIRED. FIX EVERY GAME CI FAILURE AUTONOMOUSLY.**

## Execution Strategy

### PHASE 1: IDENTIFY GAME-SPECIFIC CI FAILURES

1. **Locate failure point:**
   - Compilation failure?
   - Test execution failure?
   - Graphics/rendering failure?
   - Code analysis failure (go-stats-generator)?
   - Artifact generation failure?

2. **Extract diagnostic info:**
   - Error message and stack trace
   - Which platform (Linux, Windows, macOS)?
   - Which build tag was used?
   - Graphics-related errors?

3. **Categorize failure:**
   - Graphics initialization?
   - Platform-specific code issue?
   - Missing/invalid asset?
   - Timing/performance issue?
   - Build tag problem?

### PHASE 2: DIAGNOSE ROOT CAUSE

#### GRAPHICS/RENDERING ISSUES
- Check if error mentions graphics, OpenGL, Metal, DirectX, EGL
- Is CI environment headless (no display)?
- Solution: Add Xvfb support, mock graphics, or use headless renderer

#### PLATFORM-SPECIFIC ISSUES
- Check build tags (//go:build)
- Does code compile on all target platforms?
- Are platform-specific imports causing issues?
- Solution: Fix build constraints, test multi-platform, use Docker

#### ASSET/RESOURCE ISSUES
- Error about missing files, images, sounds, shaders?
- Are asset paths relative/absolute?
- Are assets included in CI environment?
- Solution: Fix asset paths, mock assets in tests, include test assets

#### TIMING/PERFORMANCE ISSUES
- Is test timing-sensitive?
- Does game loop have strict FPS requirements?
- Are there timing-dependent assertions?
- Solution: Mock time, use deterministic testing, increase timeouts

#### BUILD/COMPILATION ISSUES
- Missing dependencies or libraries?
- Platform-specific compilation flags?
- Incorrect build constraints?
- Solution: Install dependencies, fix build constraints, update CI config

#### ANALYSIS TOOL ISSUES
- Does go-stats-generator fail on game code?
- Are there false-positive complexity warnings?
- Does analysis tool need game-specific configuration?
- Solution: Add analysis exclusions, adjust thresholds, skip non-analyzable code

### PHASE 3: AUTONOMOUSLY FIX EACH FAILURE

#### FIX TYPE 1: GRAPHICS INITIALIZATION FAILURE
- Add Xvfb virtual display support to CI
- Or mock graphics context in code
- Or skip graphics tests in CI
- Test that graphics-related code compiles and initializes

#### FIX TYPE 2: PLATFORM COMPILATION FAILURE
- Fix build constraints (//go:build)
- Add missing platform-specific imports
- Fix platform-specific code issues
- Test on multiple platforms (if possible) or verify build tags

#### FIX TYPE 3: MISSING GAME ASSETS
- Fix asset paths to be relative or use embedded assets
- Mock assets in test environment
- Or include test assets in CI
- Verify asset loading works in CI

#### FIX TYPE 4: TIMING/PERFORMANCE ISSUES
- Make tests deterministic (mock time.Now, game tick)
- Increase timeouts for game update loops
- Remove timing-sensitive assertions
- Verify tests pass consistently

#### FIX TYPE 5: BUILD/DEPENDENCY ISSUES
- Add missing dependencies to CI workflow
- Update go.mod if needed
- Fix platform-specific library requirements
- Verify build succeeds

#### FIX TYPE 6: ANALYSIS TOOL FALSE POSITIVES
- Add configuration to exclude game-specific code patterns
- Add analysis exclusions for hot paths
- Adjust complexity thresholds for game loops
- Document why certain patterns are excluded

### PHASE 4: VALIDATE FIXES

For EACH fix:

1. **Local validation:**
   - Run the failing step locally
   - Verify it passes
   - If graphics-related, test with xvfb-run

2. **Commit and push:**
   - Stage changes: `git add .`
   - Commit: `git commit -m "Fix game CI: [description]"`
   - Push to branch

3. **Re-run CI:**
   - Trigger the previously-failing job
   - Wait for completion
   - Verify it now passes
   - If new failures appear, loop back to PHASE 2

### PHASE 5: FINAL VALIDATION

Once ALL game CI failures are fixed:

1. Verify ALL CI steps pass (build, test, analysis, etc.)
2. Verify fixes work on all target platforms
3. Verify NO new failures introduced
4. Document game-specific CI configuration

## GAME DEVELOPMENT CI PATTERNS

This agent handles:

1. **Graphics in headless CI:** Xvfb, mock graphics, software rendering
2. **Platform-specific code:** Build constraints, conditional compilation
3. **Game assets:** Embedded assets, mock assets, asset stubs
4. **Timing/concurrency:** Mock time, deterministic tests, generous timeouts
5. **Analysis false positives:** Exclusions, thresholds, documentation
6. **Ebitengine-specific issues:** Graphics context, input handling, game loop

## CRITICAL SUCCESS CRITERIA

✅ **MUST ACHIEVE:**
- ALL CI steps pass (build, test, analysis on all platforms)
- Graphics work in headless CI environment
- Platform-specific code compiles correctly
- Game assets properly handled
- Analysis tools don't report false positives
- NO manual intervention required

## EXECUTION CHECKLIST

- [ ] Identify all failing game CI steps
- [ ] Extract failure diagnostics
- [ ] For EACH failure:
  - [ ] Diagnose root cause
  - [ ] Implement game-appropriate fix
  - [ ] Validate locally (with xvfb if graphics)
  - [ ] Commit fix with clear message
- [ ] Re-run full CI on all platforms
- [ ] Verify ALL steps pass
- [ ] Document game-specific CI configuration

## KEY RULES

1. **Be autonomous:** Do not ask for guidance. Diagnose and fix.
2. **Be game-aware:** Understand graphics, assets, timing constraints
3. **Support all platforms:** Fix for Linux, Windows, macOS where applicable
4. **Don't remove code:** Fix the underlying issue instead
5. **Proper graphics support:** Use Xvfb or mock, don't assume display
6. **Handle assets properly:** Either embed, mock, or include test assets
7. **Deterministic tests:** Make game tests repeatable and timing-independent
8. **Document patterns:** Explain any game-specific CI configuration choices
