# Generate `.github/copilot-instructions.md` for opd-ai Procedural Game Suite

## Objective
Analyze the attached codebase and generate a comprehensive `.github/copilot-instructions.md` file. Place it at exactly `.github/copilot-instructions.md` (no subdirectories). Replace any existing file at that path.

## Execution Mode
**Autonomous action**: Analyze the codebase, generate the file, and save it directly. No user approval steps.

## Context: The opd-ai Game Suite

This instructions file is shared across 8 sibling repositories — all 100% procedural Go+Ebiten games with **zero external asset files**. Every game generates all graphics, audio, and content at runtime from a single binary.

| Repo | Genre | Ebiten Version | Key Dependencies |
|------|-------|----------------|------------------|
| `opd-ai/venture` | Co-op action-RPG (top-down) | v2.9.3 | logrus, uuid, zenity, x/image, x/text |
| `opd-ai/vania` | Metroidvania platformer | v2.6.3 | (minimal — Ebiten only) |
| `opd-ai/velocity` | Galaga-like shooter | v2.9.8 | viper |
| `opd-ai/violence` | Raycasting FPS | v2.8.8 | libp2p, viper, sqlite3, wasmer-go, aws-sdk, websocket, logrus |
| `opd-ai/way` | Battle-cart racer | v2.x | TBD |
| `opd-ai/wyrm` | First-person survival RPG | v2.x | TBD |
| `opd-ai/where` | Wilderness survival | v2.x | TBD |
| `opd-ai/whack` | Arena battle game | v2.x | TBD |

Tailor the output to the **attached codebase** but keep guidance fully compatible with all sibling repos.

---

## Analysis Steps

### 1. Project Foundation
- Extract purpose, target users, and setup requirements from README.md (first 3 paragraphs)
- Map ALL first and second-level directories with purpose descriptions
- Identify the `cmd/` entrypoint structure (client, server, mobile where applicable)
- Determine whether the project uses `pkg/` (public library packages) or `internal/` (private packages) or both
- Note the project layout variant:
  - **venture-style**: `cmd/client`, `cmd/server`, `cmd/mobile` + `pkg/` with 30+ sub-packages
  - **vania-style**: `cmd/vania` + `internal/` with domain-specific packages
  - **violence-style**: monolithic `main.go` (207KB+) + `cmd/` + `pkg/`
  - **velocity-style**: `cmd/` + `pkg/` with config file (yaml/toml)

### 2. Documentation & Dependencies
- Analyze ALL markdown docs in priority order: README.md, GAPS.md, AUDIT.md, ROADMAP.md, PLAN.md, REFACTORING_SUMMARY.md, UI_AUDIT.md, CONTRIBUTING.md, any docs/ subdirectory files
- Parse `go.mod` for exact Go version, framework versions, and dependency categories
- Extract coding standards, naming conventions, and architectural patterns
- Identify shared dependencies across the suite (Ebiten, logrus, viper, x/image, x/text)

### 3. Codebase Pattern Analysis
- **ECS Architecture**: Identify whether the project uses Entity-Component-System (components as pure data with `Type() string`, systems as logic with `Update(entities, deltaTime)`) or an alternative architecture
- **Procedural Generation**: Verify deterministic seed patterns — look for `rand.New(rand.NewSource(seed))` usage vs. global rand vs. `time.Now()` seeding
- **Logging**: Check for logrus structured logging with `logrus.WithFields(logrus.Fields{...})` or other logging patterns
- **Network Interface Usage**: Scan for concrete network types that should be interfaces — this is a hard constraint, see Networking section
- **Latency Tolerance**: Verify that networking code is designed for 200–5000ms latency environments
- **State Management**: Identify game state patterns (game loop, scene/state machines, ECS world)
- **Rendering Pipeline**: Note whether procedural rendering uses sprite caching, tiling, lighting, particle systems
- **Audio Synthesis**: Check for procedural audio generation patterns

### 4. CRITICAL: Feature Integration Completeness Audit
**This is the highest-priority section of the instructions file.** Dangling features in complex game codebases are a maintenance nightmare and a source of deep frustration. The instructions MUST include detailed guidance on this topic.

Analyze the codebase for signs of:
- **Instantiation Gaps**: Structs/systems defined but never instantiated outside `_test.go` files
- **Dead Integration Paths**: Integration packages that import dependencies but never call key methods
- **Orphaned Event Emitters/Listeners**: Events emitted with no handler, or handlers registered for events never fired
- **Stub Implementations**: Interface methods with empty bodies or hardcoded returns
- **Dangling Generators**: Procedural generators that satisfy an interface but are never invoked in runtime code
- **Partial Genre Coverage**: Generation paths that exist for some genres/themes but not others
- **Seed Propagation Breaks**: Seeds accepted but not forwarded to sub-generators (causing non-determinism)
- **Documentation-Code Drift**: Features described in docs with no corresponding code, or code with no documentation

Reference `GAPS.md`, `AUDIT.md`, and `ROADMAP.md` to understand known issues and prioritize instructions accordingly.

---

## Output Format

Generate a Markdown file following this structure. Keep total length between 4000–6000 words for sufficient detail without exceeding Copilot context limits.

```markdown
# Project Overview
[2-3 paragraphs: purpose, audience, key technologies, what makes it 100% procedural, single-binary philosophy]

## Sibling Repository Context
[Brief description of the 8-game suite and the expectation that repos share code patterns, conventions, and eventually shared library packages. List all 8 repos with one-line descriptions.]

## Technical Stack
- **Primary Language**: Go [exact version from go.mod]
- **Game Framework**: Ebiten [exact version from go.mod] — 2D game engine with cross-platform + WASM support
- **Key Dependencies**: [List each direct dependency with version and purpose]
- **Testing**: Go standard `testing` package, table-driven tests, benchmarks
- **Build/Deploy**: [Makefile targets, CI/CD, WASM builds, mobile builds if applicable]

## Project Structure
[Describe the actual directory layout with purpose annotations. Note which layout variant this repo uses.]

---

## ⚠️ CRITICAL: Complete Feature Integration (Zero Dangling Features)

**This is the single most important rule for this codebase.** Every feature, system, component, generator, and integration MUST be fully wired into the runtime. Dangling features are a maintenance burden, a source of frustration, and actively degrade code quality.

### The Dangling Feature Problem
In complex procedural game codebases, it is extremely common for features to be:
1. **Defined but never instantiated** — A system struct exists but is never created in `main()` or system registration
2. **Instantiated but never integrated** — A system runs but its output is never consumed by other systems
3. **Partially integrated** — A system works for one genre/theme but silently no-ops for others
4. **Tested in isolation but broken in context** — Unit tests pass but the system was never wired into the game loop

### Mandatory Checks Before Adding or Modifying Any Feature

**Before writing ANY new code, verify the full integration chain:**

1. **Definition → Instantiation**: Is the struct/system created at runtime? Trace from `main()` through system registration.
2. **Instantiation → Registration**: Is the system registered with the game world/engine/scene? Check system lists, init functions.
3. **Registration → Update Loop**: Does the system's `Update()` method actually get called each frame/tick?
4. **Update → Output**: Does the system produce outputs (components, events, state changes) that other systems consume?
5. **Output → Consumer**: Is there at least one other system that reads this system's output?
6. **Consumer → Player Effect**: Does the chain ultimately produce something visible, audible, or mechanically felt by the player?

If ANY link in this chain is missing, the feature is dangling. **Do not submit dangling features.**

### Specific Anti-Patterns to Reject

```go
// ❌ BAD: System defined but never added to the game world
type WeatherSystem struct { ... }
func (w *WeatherSystem) Update(entities []*Entity, dt float64) { ... }
// ...but NewWeatherSystem() is never called in main() or init

// ✅ GOOD: System defined, instantiated, registered, and consuming/producing
weather := NewWeatherSystem(seed)
world.AddSystem(weather)
// AND other systems react to weather state:
// render.ApplyWeatherEffects(weather.Current())
// audio.PlayWeatherAmbience(weather.Current())
```

```go
// ❌ BAD: Generator implements interface but is never called outside tests
type CyberpunkTerrainGen struct { ... }
func (g *CyberpunkTerrainGen) Generate(params GenParams) *Terrain { ... }
// Only called in cyberpunk_terrain_test.go, never in runtime

// ✅ GOOD: Generator registered in genre dispatch and called during world creation
genRegistry["cyberpunk"] = &CyberpunkTerrainGen{}
// AND terrain := genRegistry[currentGenre].Generate(params) is called in world init
```

```go
// ❌ BAD: Event emitted but no listener handles it
eventBus.Emit("player.levelup", playerID, newLevel)
// No eventBus.On("player.levelup", ...) anywhere in the codebase

// ✅ GOOD: Event has both emitter and handler
eventBus.Emit("player.levelup", playerID, newLevel)
// In another system:
eventBus.On("player.levelup", func(id, level) {
    ui.ShowLevelUpAnimation(id, level)
    audio.PlayLevelUpSound()
    stats.RecalculateForLevel(id, level)
})
```

### Integration Verification Checklist (run before every PR)
- [ ] `grep -rn 'func New' --include='*.go' | grep -v _test.go` — Every constructor has at least one non-test caller
- [ ] `grep -rn 'TODO\|FIXME\|HACK\|XXX' --include='*.go'` — All TODOs are tracked in GAPS.md or ROADMAP.md
- [ ] No empty method bodies in non-test `.go` files
- [ ] Every interface in the project has at least one runtime (non-test) implementation
- [ ] Every procedural generator is reachable from the main game initialization path
- [ ] Seeds are propagated through the full generation chain (parent seed → child generators)

---

## Networking Best Practices (MANDATORY for all Go network code)

### Interface-Only Network Types (Hard Constraint)

When declaring network variables, ALWAYS use interface types. This is a **non-negotiable project rule** enforced by CI validation.

| ❌ Never Use (Concrete Type) | ✅ Always Use (Interface Type) |
|---|----|
| `*net.UDPAddr` | `net.Addr` |
| `*net.IPAddr` | `net.Addr` |
| `*net.TCPAddr` | `net.Addr` |
| `*net.UDPConn` | `net.PacketConn` |
| `*net.TCPConn` | `net.Conn` |
| `*net.UDPListener` | `net.Listener` |
| `*net.TCPListener` | `net.Listener` |

```go
// ✅ GOOD: Interface types everywhere
var addr net.Addr
var conn net.PacketConn
var tcpConn net.Conn
var listener net.Listener

// ❌ BAD: Concrete types — will fail CI
var addr *net.UDPAddr
var conn *net.UDPConn
var tcpConn *net.TCPConn
var listener *net.TCPListener
```

**Never use type switches or type assertions to convert from an interface type to a concrete type.** Use the interface methods instead.

```go
// ❌ BAD: Type assertion to access concrete methods
if udpConn, ok := conn.(*net.UDPConn); ok {
    udpConn.ReadFromUDP(buf)
}

// ✅ GOOD: Use the interface methods directly
n, addr, err := conn.ReadFrom(buf)  // PacketConn interface method
```

This enhances testability and flexibility when working with different network implementations or mocks.

### High-Latency Network Design (200–5000ms)

All multiplayer networking code MUST be designed to function correctly under **200–5000ms round-trip latency**. These games target diverse network conditions including mobile data, satellite internet, and intercontinental connections.

#### Mandatory Design Principles

1. **Client-Side Prediction**: The client must simulate game state locally and reconcile with server authoritative state when it arrives. Never block the game loop waiting for a server response.

2. **State Interpolation / Extrapolation**: Remote entity positions must be interpolated between known states. When packets are delayed beyond the interpolation window, extrapolate using last-known velocity.

3. **Jitter Buffers**: Incoming state updates must be buffered and played back at a consistent rate, absorbing latency variance (jitter). Design for ±500ms jitter tolerance minimum.

4. **Idempotent Messages**: Every network message must be safe to process multiple times. Retransmission at high latency is expected, not exceptional.

5. **No Synchronous RPC in Game Loops**: Never issue a blocking network call inside `Update()` or `Draw()`. All network I/O must be asynchronous with results consumed on the next available frame.

6. **Graceful Degradation**: At 5000ms latency the game must remain playable, not just connected. Reduce update frequency, increase prediction windows, and hide latency with animations.

7. **Timeout Tolerance**: Connection timeouts must be set to ≥10 seconds. Disconnect detection must use heartbeat absence over a sliding window (≥3 missed heartbeats at the expected interval), never a single missed packet.

```go
// ❌ BAD: Tight timeout that drops players on satellite connections
conn.SetReadDeadline(time.Now().Add(1 * time.Second))

// ✅ GOOD: Generous timeout for high-latency environments
conn.SetReadDeadline(time.Now().Add(10 * time.Second))

// ❌ BAD: Blocking RPC in game loop
func (g *Game) Update() error {
    state, err := g.server.GetWorldState()  // blocks until response
    g.world = state
    return nil
}

// ✅ GOOD: Async receive with interpolation
func (g *Game) Update() error {
    select {
    case state := <-g.stateChannel:
        g.interpolator.PushServerState(state)
    default:
        // No new state — continue with prediction
    }
    g.world = g.interpolator.GetInterpolatedState(time.Now())
    return nil
}
```

#### Latency Budget Allocation (per frame at 60 FPS = 16.6ms)
- **Input processing**: ≤1ms
- **Local simulation / prediction**: ≤4ms
- **State interpolation**: ≤1ms
- **Network send (non-blocking enqueue)**: ≤0.5ms
- **Rendering**: ≤10ms
- **Network I/O goroutines**: Run independently, never counted against frame budget

---

## Code Assistance Guidelines

### 1. Deterministic Procedural Generation
All content generation MUST be deterministic and seed-based. Given the same seed, the game MUST produce identical output across all platforms and runs.

```go
// ✅ GOOD: Explicit seed-based RNG, never global
rng := rand.New(rand.NewSource(seed))
value := rng.Intn(100)

// ❌ BAD: Global rand (non-deterministic, not thread-safe)
value := rand.Intn(100)

// ❌ BAD: Time-based seeding in generation code
rng := rand.New(rand.NewSource(time.Now().UnixNano()))

// ✅ GOOD: Derived seeds for sub-generators (deterministic hierarchy)
terrainSeed := seed ^ 0x54455252  // "TERR"
enemySeed := seed ^ 0x454E454D    // "ENEM"
terrainRNG := rand.New(rand.NewSource(terrainSeed))
enemyRNG := rand.New(rand.NewSource(enemySeed))
```

### 2. ECS Architecture Discipline
[Adapt based on what the codebase actually uses. If ECS:]
- Components are pure data structs with a `Type() string` method. NO logic in components.
- Systems contain ALL game logic. Systems operate on entity collections filtered by component type.
- Never store references to other entities directly; use entity IDs.
- Systems declare their dependencies explicitly (which component types they require).

[If not ECS, describe the actual architecture pattern found in the codebase.]

### 3. Structured Logging (logrus)
[Include only if logrus is a dependency in the attached codebase's go.mod]

```go
// ✅ GOOD: Structured logging with context
logrus.WithFields(logrus.Fields{
    "system":   "terrain",
    "seed":     seed,
    "biome":    biomeType,
    "duration": elapsed,
}).Info("Terrain generation complete")

// ❌ BAD: Unstructured fmt.Printf or log.Println
fmt.Printf("Generated terrain with seed %d\n", seed)
log.Println("terrain done")
```

Standard field names: `system`, `entity`, `player`, `seed`, `error`, `duration`, `count`.

### 4. Performance Requirements
- Target 60 FPS on mid-range hardware
- Client memory budget: <500MB
- Use spatial partitioning (quadtree/grid) for entity queries over collections >100
- Cache generated sprites — never regenerate the same sprite twice per session
- Use object pooling for frequently allocated/deallocated objects (bullets, particles, effects)
- Benchmark hot paths with `go test -bench=. -benchmem`

### 5. Zero External Assets
The single-binary philosophy means ALL content is generated at runtime:
- **Graphics**: Procedurally generated from code (pixel manipulation, shape primitives, noise functions)
- **Audio**: Synthesized from oscillators, envelopes, and effects
- **Levels/Maps**: Generated from algorithms (BSP, cellular automata, L-systems, Voronoi, wave function collapse)
- **Items/NPCs/Quests**: Generated from parameterized templates with seed-based variation
- **UI**: Built from code, not loaded from image files

Never add asset files (PNG, WAV, OGG, JSON level files) to the repository. If you need test fixtures, generate them in test setup code.

### 6. Error Handling
```go
// ✅ GOOD: Return errors, handle them at the call site
func GenerateTerrain(seed int64) (*Terrain, error) {
    if seed == 0 {
        return nil, fmt.Errorf("terrain generation requires non-zero seed")
    }
    // ...
}

// ❌ BAD: Panic in library/game code
func GenerateTerrain(seed int64) *Terrain {
    if seed == 0 {
        panic("zero seed")  // Never panic in game logic
    }
}

// ✅ GOOD: Log and recover gracefully in game systems
func (s *TerrainSystem) Update(entities []*Entity, dt float64) {
    terrain, err := GenerateTerrain(s.seed)
    if err != nil {
        logrus.WithError(err).Error("Terrain generation failed, using fallback")
        terrain = s.fallbackTerrain()
    }
}
```

Panics are acceptable ONLY in `main()` for unrecoverable startup failures. All game systems must handle errors gracefully with fallbacks.

---

## Cross-Repository Code Sharing Patterns

### Current State
The 8 opd-ai game repos share architectural patterns but currently duplicate code. As the suite matures, common functionality should be extracted into shared libraries.

### Shared Pattern Catalog
When implementing features, follow these patterns so code can be extracted into shared packages later:

| Pattern | Package Convention | Used By |
|---------|-------------------|---------|
| ECS core (World, Entity, Component, System) | `pkg/engine/` or `internal/engine/` | All repos |
| Procedural generation framework | `pkg/procgen/` or `internal/pcg/` | All repos |
| Seed management & derivation | `pkg/seed/` or inline | All repos |
| Sprite/tile generation | `pkg/rendering/` or `internal/graphics/` | All repos |
| Audio synthesis | `pkg/audio/` or `internal/audio/` | All repos |
| Input handling | `pkg/input/` or `internal/input/` | All repos |
| Camera systems | `pkg/camera/` or `internal/camera/` | All repos |
| Particle systems | `pkg/particles/` or `internal/particle/` | Most repos |
| Save/load | `pkg/saveload/` or `internal/save/` | All repos |
| Menu/UI framework | `pkg/rendering/ui/` or `internal/menu/` | All repos |
| Achievement system | `pkg/achievement/` or `internal/achievement/` | Most repos |
| Configuration (viper/flags) | `pkg/config/` | violence, velocity, venture |
| Networking (multiplayer) | `pkg/network/` | venture, violence |
| Physics | `pkg/engine/physics/` or `internal/physics/` | vania, venture, whack, way |

### Guidelines for Shareable Code
1. **Keep dependencies minimal**: Shared packages should depend only on stdlib + Ebiten. Game-specific logic stays in the game repo.
2. **Use interfaces at boundaries**: Define interfaces for game-specific behavior so shared code doesn't import game packages.
3. **Parameterize, don't specialize**: A terrain generator should accept parameters for any genre, not have hardcoded genre logic.
4. **Same naming conventions across repos**: If `venture` calls it `pkg/engine/World`, `vania` should use `internal/engine/World` with the same method signatures.
5. **Identical Component interface**: `Type() string` must be the universal component identifier across all repos.
6. **Identical System interface**: `Update(entities []*Entity, deltaTime float64)` must be the universal system signature.

### When Adding a Feature That Exists in a Sibling Repo
1. Check the sibling repo's implementation first (reference the repos listed above)
2. Use the same package structure and naming conventions
3. Match the interface signatures so future extraction is seamless
4. If the sibling implementation has known issues (check its GAPS.md), fix them in your implementation
5. Document divergences in your repo's ROADMAP.md with a note about future convergence

---

## Quality Standards

### Testing Requirements
- **Coverage**: ≥40% per package (≥30% for display/Ebiten-dependent packages that require xvfb)
- **Table-driven tests** for all business logic and generation functions
- **Benchmarks** for all hot-path code (rendering, physics, generation)
- **Race detection**: All tests must pass under `go test -race ./...`
- **Integration tests**: For cross-system interactions (e.g., "terrain generates → enemies spawn on terrain → combat resolves")

### Code Review Quality Gates
[Include only if the repo has validate-code-review.sh or similar]
- Build success (client + server)
- All tests pass
- Race-free (`go test -race`)
- Static analysis (`go vet`)
- Network type validation (`scripts/validate-network-types.sh`)
- No new TODO/FIXME without corresponding GAPS.md entry

### Documentation Requirements
- Every exported type and function has a godoc comment
- README.md stays in sync with CLI flags and features
- GAPS.md is updated when new gaps are discovered
- ROADMAP.md reflects current priorities

---

## Naming Conventions
- **Packages**: lowercase, single-word when possible (`engine`, `procgen`, `audio`, `render`)
- **Files**: snake_case (`terrain_generator.go`, `combat_system.go`)
- **Types**: PascalCase (`TerrainGenerator`, `CombatSystem`, `HealthComponent`)
- **Interfaces**: PascalCase, often ending in `-er` for single-method interfaces (`Generator`, `Renderer`)
- **Component types**: PascalCase + "Component" suffix (`HealthComponent`, `PositionComponent`)
- **System types**: PascalCase + "System" suffix (`CombatSystem`, `RenderSystem`)
- **Constants**: PascalCase for exported, camelCase for unexported
- **Seeds**: Always `int64`, always named `seed` in function parameters

## GAPS.md and AUDIT.md Protocol
These repos use GAPS.md and AUDIT.md files to track implementation gaps and audit findings. When Copilot identifies a potential gap:
1. Note it in your response
2. Suggest adding it to GAPS.md with severity (Critical/High/Medium/Low)
3. Include the file path and line number
4. Propose an actionable fix
```

---

## Success Criteria
1. Every dependency version matches `go.mod` exactly
2. Every guideline references patterns observable in the actual codebase
3. The "Complete Feature Integration" section is the longest and most detailed section
4. The "Networking Best Practices" section includes BOTH the interface-only type constraints AND the 200–5000ms latency design requirements
5. Cross-repository sharing patterns are documented with concrete package names from multiple sibling repos
6. A new contributor can understand the project purpose, build it, and know the rules from this file alone
7. Zero generic programming advice — all guidance is specific to this procedural game codebase
8. File is saved at exactly `.github/copilot-instructions.md`
