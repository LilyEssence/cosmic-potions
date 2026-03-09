# Cosmic Potions Lab

A space-themed potion crafting game API built in Go. You're a space alchemist brewing potions from ingredients harvested across the galaxy — combine elements, discover effects, and fill the **Cosmic Alchemist's Codex**.

**Live at:** [lixie.art/cosmic-potions](https://lixie.art/cosmic-potions)

## Quick Start

```bash
go run cmd/server/main.go       # Start server on :8080
```

The server uses in-memory storage by default, pre-loaded with seed data. Set `STORE_TYPE=sqlite` for persistent storage.

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/health` | Health check |
| GET | `/api/planets` | List all planets |
| GET | `/api/planets/{id}` | Planet detail + ingredients |
| GET | `/api/ingredients` | List with filters (`?element=`, `?rarity=`, `?planet_id=`) |
| GET | `/api/ingredients/{id}` | Single ingredient |
| GET | `/api/recipes` | Known recipes (`?difficulty=`) |
| GET | `/api/recipes/{id}` | Recipe detail with required ingredients |
| POST | `/api/brew` | Submit ingredients, get brew result |
| POST | `/api/brew/check` | Compatibility preview without brewing |
| GET | `/api/brews` | Brew history |
| GET | `/api/effects` | Master list of discoverable effects (for codex) |
| GET | `/api/interactions` | Element interaction matrix |

## Architecture

```
cmd/server/main.go          → Entry point, dependency wiring, graceful shutdown
internal/
  model/                    → Domain structs (Planet, Ingredient, Recipe, Brew, etc.)
  store/                    → Store interfaces + memory & SQLite implementations
  brewing/                  → Brew engine: pure business logic, no I/O
  handler/                  → HTTP handlers (chi router)
  middleware/               → CORS, request logging
seed/                       → Seed data: planets, ingredients, recipes, interactions, effects
migrations/                 → SQLite schema
```

**Key boundaries:**

- **Store layer** — four small interfaces (`PlanetStore`, `IngredientStore`, `RecipeStore`, `BrewStore`), not one monolithic store. Both the in-memory and SQLite implementations satisfy all four.
- **Brewing engine** — pure business logic with no I/O. Accepts a `Randomizer` interface for deterministic testing.
- **Handlers** — each handler receives only the store interface(s) it needs.

## Game Design

Players experiment with ingredient combinations to brew potions and discover effects. Every effect starts hidden ("???") in the codex. The goal: reveal them all and earn the title **Grand Cosmic Alchemist**.

- **Known recipes** guarantee certain effects
- **Experimental brewing** (freeform ingredient combos) can discover effects no recipe produces
- **Element synergies** boost potency and may unlock bonus effects
- **Volatile combos** are risky but can reveal legendary effects
- **Player state lives entirely in the frontend** (localStorage) — the API is stateless regarding player progress

## Build & Test

```bash
go build ./...              # Compile all packages
go vet ./...                # Static analysis
go test ./... -v -cover     # Run tests with coverage
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | Server port (Railway sets this automatically) |
| `STORE_TYPE` | `memory` | `memory` or `sqlite` |
| `DB_PATH` | `cosmic-potions.db` | SQLite database file path |
| `MIGRATIONS_DIR` | `migrations` | Path to SQL migration files |

## Deployment

Deployed on Railway via multi-stage Dockerfile. The frontend lives in the [lixiestudios](https://github.com/LilyEssence/lixiestudios) monorepo and calls this API.

## Learning Notes

This project is annotated with `// GO CONCEPT:` comments throughout the codebase, explaining Go patterns and idioms for developers coming from TypeScript/Node.js.

---

Built with Go — [GitHub](https://github.com/LilyEssence/cosmic-potions)
