# GitHub Copilot Instructions — Cosmic Potions

## Project Overview

**Cosmic Potions** — A space-themed potion crafting game API in Go, with a React frontend in the sibling `lixiestudios` monorepo. The owner is learning Go; this project demonstrates clean architecture and Go idioms.

**Repos:** `cosmic-potions` (Go API, standalone) + `lixiestudios` (React frontend, existing monorepo)
**Deploy:** Both on Railway. Frontend at `lixie.art/cosmic-potions`, Go API as a separate service.

## Architecture

```
cmd/server/main.go     → Entry point, wires deps, graceful shutdown
internal/
  model/               → Domain structs (Planet, Ingredient, Recipe, Brew, Interaction)
  store/               → Interface definitions + in-memory implementation (SQLite planned)
  brewing/             → Brew engine: pure business logic, no I/O (Step 5 — not yet built)
  handler/             → HTTP handlers (Step 6 — not yet built)
  middleware/          → CORS, logging, request ID (Step 7 — not yet built)
seed/                  → Seed data: 6 planets, 18 ingredients, 8 recipes, 5×5 interaction matrix
migrations/            → SQLite schema (Step 9 — placeholder only)
```

**Key boundary:** The Go API owns all domain logic and data rules. The frontend (in lixiestudios) owns everything visual. The API must be fully usable via `curl`.

**Game loop:** Players brew potions to discover effects and fill a "Codex" (collection tracker). Player state lives entirely in frontend localStorage — the Go API is stateless regarding player progress.

## Build & Run

```bash
go build ./...              # Compile all packages
go vet ./...                # Static analysis
go test ./... -v -cover     # Tests (target >80% coverage on brewing engine)
go run cmd/server/main.go   # Start server on :8080 (or PORT env var)
```

## Conventions & Patterns

### Learning-Oriented Comments
Every Go concept is annotated with `// GO CONCEPT: Name` comments explaining the "what" and "why" for a developer learning Go from TypeScript. **Preserve this pattern** when adding new code. Each new Go concept encountered should get a comment block.

### Store Layer (Interface Segregation)
Four small interfaces, not one big `Store`: `PlanetStore`, `IngredientStore`, `RecipeStore`, `BrewStore`. Consumers accept only the interface they need. `MemoryStore` is one struct satisfying all four — `var _ PlanetStore = (*MemoryStore)(nil)` lines verify this at compile time.

### Data Patterns
- **Filter structs with pointer fields** for optional criteria: `nil` = don't filter, non-nil = match. See `model.IngredientFilter`, `model.RecipeFilter`.
- **`make([]T, 0)` not `var x []T`** for slices returned as JSON — ensures `[]` instead of `null`.
- **Sorted output:** All list methods sort by name for deterministic API responses.
- **`sync.RWMutex`:** Every store method locks before touching maps. `RLock` for reads, `Lock` for writes. Always `defer` the unlock.

### Seed Data
Lives in `seed/seed.go` as functions returning fresh slices (not mutable package vars). IDs follow the pattern `planet-verdanthia`, `ing-solar-crystal`, `recipe-starlight-tonic`.

### Router
Using `chi/v5` — lightweight, builds on `net/http`. No heavy frameworks.

### Error Handling
Sentinel error `store.ErrNotFound` for missing entities. Handlers (when built) should map this to HTTP 404. All fallible functions return `(value, error)`.

## Implementation Status

Refer to `DESIGN_PLAN.md` for the full plan. Current state:

| Step | Status | Package |
|------|--------|---------|
| 1-2: Setup + structure | ✅ Done | `cmd/server`, project root |
| 3: Models + seed data | ✅ Done | `internal/model`, `seed` |
| 4: Store interfaces + memory impl | ✅ Done | `internal/store` |
| 5: Brewing engine | 🔲 Next | `internal/brewing` |
| 6: HTTP handlers | 🔲 | `internal/handler` |
| 7: Middleware | 🔲 | `internal/middleware` |
| 8: Tests | 🔲 | All packages |
| 9-12: SQLite, deploy, frontend | 🔲 | Various |
