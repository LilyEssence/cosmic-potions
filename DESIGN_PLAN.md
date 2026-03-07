# Cosmic Potions Lab — Implementation Plan

A space-themed potion crafting game built in Go, with a React frontend on lixie.art. You're a space alchemist brewing potions from ingredients harvested across the galaxy. Combine ingredients to brew potions — some combinations synergize, some are volatile. Your goal: fill the **Cosmic Alchemist's Codex** by discovering every potion effect in the universe.

**Purpose:** Demonstrate Go proficiency, clean architecture, and fast language ramp-up

**Repos:**
- `cosmic-potions` — Standalone Go repo (new GitHub repo: `LilyEssence/cosmic-potions`)
- `lixiestudios` — Frontend pages + API client + assets (existing monorepo)

**Deploy:** Both on the same Railway project dashboard. Frontend already deployed; Go service added as a new service.

**Live at:** `lixie.art/cosmic-potions`

---

## Data Model

### Planet
Origin world for ingredients.
- ID, Name, System, Biome (volcanic, crystalline, oceanic, fungal, void), Description, DiscoveredAt

### Ingredient
Harvested from a planet.
- ID, PlanetID, Name, Element (solar, void, crystalline, organic, plasma), Potency (1-10), Rarity (common, uncommon, rare, legendary), Description, FlavorNotes (bitter, sweet, luminous, electric, etc.)

### Recipe
A known combination.
- ID, Name, Description, Difficulty (novice, adept, master), Effects (list of strings like "night vision", "gravity resistance", "temporary gills"), IngredientList (ingredient IDs + quantities)

### Brew
An attempt to make a potion.
- ID, RecipeID (nullable — could be experimental), IngredientsUsed, Result (success, partial, catastrophic_failure), PotionName, Effects, Notes, BrewedAt

### Interaction
Defines how two elements react.
- Element1, Element2, Result (synergy, neutral, volatile)
- Volatile combos have a chance of catastrophic failure. Synergies boost potency.

---

## Game Design: The Cosmic Alchemist's Codex

### Core Loop

Experiment with ingredient combinations → brew potions → discover new effects → fill the codex → become Grand Cosmic Alchemist.

The codex is a collection tracker. Every possible potion effect in the universe is listed, but they start hidden ("???"). When you successfully brew a potion, the effects it produces get revealed in your codex. The goal is to discover all of them.

### Effect Tiers

Effects are grouped by discovery difficulty:

| Tier | Count | How to Discover |
|------|-------|-----------------|
| Common | ~10 | Appear in many recipes — hard to miss |
| Uncommon | ~8 | Require specific element pairings or 3-ingredient combos |
| Rare | ~6 | Need rare ingredients + synergy-boosted brews |
| Legendary | ~4 | Demand mastery of volatile combos without catastrophic failure |

**Total: ~28 discoverable effects.** The exact number is defined in seed data.

### Discovery Mechanics

- **Known recipes** are a starting point — following them guarantees certain effects
- **Experimental brewing** (no recipe, just toss ingredients together) can discover effects that no known recipe produces
- **Synergy combos** sometimes produce bonus effects beyond what the individual ingredients would yield
- **Volatile combos** that succeed (risky but rewarding) can unlock legendary effects that safe combos never produce
- **Partial successes** still reveal some effects — you don't need a perfect brew every time
- **Catastrophic failures** reveal nothing (but are entertaining)

### Win Condition

**"Grand Cosmic Alchemist"** — Discover all effects in the codex. The codex page shows a progress bar (e.g., "23/28 Effects Discovered"). At 100%, a celebration screen + title.

No time pressure, no lives, no fail state. Pure exploration and experimentation at your own pace.

### Player State (Frontend Only)

All game state lives in **localStorage** — the Go API is stateless with respect to player progress.

- `discoveredEffects` — array of effect IDs + discovery timestamps
- `brewCount` — total brews attempted
- `successCount` / `failureCount` — brew outcome tallies
- `codexComplete` — boolean, flipped when all effects discovered

This means every visitor starts fresh (or can reset). No accounts, no server saves. The Go API just serves the data and brewing logic.

### API Impact (Minimal)

One new endpoint: `GET /api/effects` — returns the master list of all discoverable effects (ID + name + tier). No hints about which ingredients produce them. This tells the frontend what "100%" looks like so it can render the codex grid.

The existing `POST /api/brew` response already returns effects — the frontend just needs to cross-reference against the codex.

---

## Architecture

### Repository Boundary

**Go repo owns:** Domain logic, data, and rules. "What ingredients exist, what happens when you brew them, is this combination valid." If it would exist in a text-only version of the game, it belongs here.

**Frontend repo owns:** Everything visual and interactive. "How do I show a planet, what animation plays when a brew fails, what does the ingredient card look like." If it only matters because humans have eyes, it belongs here.

Test: Could someone use the Go API from a terminal with `curl` and have the full experience minus visuals? Yes — the boundary is clean.

### Go Project Structure

cosmic-potions/
├── cmd/server/main.go # Entry point, wire deps, graceful shutdown
├── internal/
│ ├── model/ # Structs: Planet, Ingredient, Recipe, Brew, Interaction
│ ├── store/ # Interface + in-memory implementation (+ SQLite Day 2)
│ ├── brewing/ # Brew engine: compatibility, effect calc, outcome
│ ├── handler/ # HTTP handlers
│ └── middleware/ # Logging, CORS, request ID
├── seed/ # Fun seed data (planets, ingredients, known recipes)
├── migrations/ # SQLite schema (Day 2)
├── Dockerfile # Multi-stage build for Railway
├── railway.toml # Railway deployment config
└── README.md

### Frontend Integration (in lixiestudios)

- **API client:** `packages/frontend/src/lib/cosmicPotionsApi.ts`
  - Pattern: `VITE_COSMIC_POTIONS_API_URL || 'http://localhost:8080'`
- **Pages** (flat in `packages/frontend/src/pages/`, following existing convention):
  - `CosmicPotionsPage.tsx` — main page / planet explorer
  - `CosmicPotionsBrewPage.tsx` — brew station
  - `CosmicPotionsIngredientPage.tsx` — ingredient browser
  - `CosmicPotionsCodexPage.tsx` — effect codex (discovery tracker)
  - `CosmicPotionsBrewLogPage.tsx` — brew history
- **Routes:** Public (no `ProtectedRoute`) in `packages/frontend/src/App.tsx` under `/cosmic-potions/*`
- **Assets:** `packages/frontend/public/cosmic-potions/` for pixel art sprites
  - Start with colored placeholder squares (one per element: solar=gold, void=purple, crystalline=cyan, organic=green, plasma=orange)
  - Replace with real pixel art later

### Local Development

Use a VS Code Multi-Root Workspace (File → Add Folder to Workspace) with both repos. Two terminals:
1. Go server: `go run cmd/server/main.go` (port 8080)
2. Frontend: `npm run dev:frontend` (port 5173, calls localhost:8080 for data)

---

## Phase 1: Go Fundamentals + Core API

### Step 1 — Environment Setup (~1 hour)

- Install Go from golang.org
- Install VS Code Go extension (gopls, delve debugger)
- Quick pass through [Go by Example](https://gobyexample.com): variables, structs, interfaces, error handling, slices, maps, goroutines (skim)
- Create GitHub repo `LilyEssence/cosmic-potions`
- `go mod init github.com/LilyEssence/cosmic-potions`

### Step 2 — Project Structure (~30 min)

- Create the directory structure listed above
- `go get github.com/go-chi/chi/v5` (lightweight, idiomatic router)
- Set up `cmd/server/main.go` skeleton with graceful shutdown via `os.Signal`

### Step 3 — Models + Seed Data (~1 hour)

- Define all structs in `internal/model/` with proper JSON tags
- Create seed data file with:
  - ~6 planets (e.g., Verdanthia/fungal, The Obsidian Drift/void, Heliox Prime/solar, Crystara/crystalline, Thalassia/oceanic, Pyraxis/volcanic)
  - ~20 ingredients spread across planets and elements
  - ~8 known recipes of varying difficulty
  - Element interaction matrix (5×5: solar, void, crystalline, organic, plasma)
- Fun flavor text for everything

### Step 4 — Store Interfaces + In-Memory Implementation (~1.5 hours)

- Define interfaces: `PlanetStore`, `IngredientStore`, `RecipeStore`, `BrewStore`
- In-memory implementations using `sync.RWMutex` for thread safety
- Filter methods: `ListIngredients(filter IngredientFilter)` with optional planet, element, rarity
- Aggregate query: `GetPlanetWithIngredients(id)` — planet + its ingredients in one call

### Step 5 — Brewing Engine (~1.5 hours)

Separated from handlers and storage — pure business logic. This is the architectural centerpiece.

- `BrewEngine` struct with:
  - `CheckCompatibility(ingredients []Ingredient) → []Interaction` — reports synergies and volatility warnings
  - `CalculateEffects(ingredients []Ingredient, interactions []Interaction) → []Effect` — derives potion effects from ingredient properties
  - `Brew(ingredients []Ingredient) → BrewResult` — runs compatibility, rolls outcome, returns result
- Inject a `Randomizer` interface for deterministic testing
- Success odds affected by synergies (boost), volatility (reduce), and recipe difficulty

### Step 6 — HTTP Handlers + Router (~2 hours)

Using `chi` router. Endpoints:

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/planets` | List all planets |
| GET | `/api/planets/{id}` | Planet + its ingredients |
| GET | `/api/ingredients` | List with filters (`?element=void&rarity=rare&planet_id=X`) |
| GET | `/api/ingredients/{id}` | Single ingredient detail |
| GET | `/api/recipes` | List known recipes (optional `?difficulty=novice`) |
| GET | `/api/recipes/{id}` | Recipe detail with required ingredients |
| POST | `/api/brew` | Submit ingredients, get brew result |
| POST | `/api/brew/check` | Compatibility preview without brewing |
| GET | `/api/brews` | Brew history (successes and catastrophic failures) |
| GET | `/api/effects` | Master list of all discoverable effects (for codex) |
| GET | `/api/interactions` | The element interaction matrix |
| GET | `/api/health` | Health check |

- Input validation with meaningful error messages
- Proper HTTP status codes (201, 404, 422, etc.)
- Consistent JSON response wrapper

### Step 7 — Middleware (~30 min)

- Request logging (method, path, status, duration)
- CORS (for frontend at localhost:5173 and lixie.art)
- Request ID header injection
- Seed data auto-loader on first startup

### Step 8 — Table-Driven Tests (~2 hours)

- **Brewing engine tests** (bulk of tests): every element interaction combo, synergy boosts, volatile failure cases, effect calculation
- **Handler tests** with Go's built-in `net/http/httptest` — no external test framework
- **Store tests:** CRUD operations, filter accuracy
- Seeded random source for deterministic brew outcomes in tests
- Target: >80% coverage on the brewing engine

---

## Phase 2: Persistence, Deploy, Frontend Demo

### Step 9 — SQLite Persistence (~2 hours)

- Second store implementation behind the same interfaces (the interface investment pays off here)
- Use `modernc.org/sqlite` (pure Go, no CGO needed)
- Schema with foreign keys (ingredient → planet, brew → recipe)
- Migrations in `migrations/001_init.sql`
- Seed data runs via migration
- Toggle store type via `STORE_TYPE` env var (`memory` vs `sqlite`)

### Step 10 — Deploy to Railway (~1 hour)

**Dockerfile** (multi-stage):
- Stage 1: `golang:1.22-alpine` — build the binary
- Stage 2: `alpine:latest` — copy binary + migrations, minimal image

**railway.toml:**
- `builder = "DOCKERFILE"` (Go needs Docker, not Railpack)
- Health check at `/api/health`
- Env vars: `PORT` (Railway sets this), `STORE_TYPE=sqlite`, `DB_PATH=/data/ledger.db`
- Persistent volume at `/data` for SQLite file

**Railway dashboard:**
- Add new service pointing to `cosmic-potions` GitHub repo
- Add `VITE_COSMIC_POTIONS_API_URL` env var to the **frontend** service (set to the Go service's Railway domain)

### Step 11 — Frontend Demo Page (~2-3 hours)

In the `lixiestudios` monorepo:

**API client:** `packages/frontend/src/lib/cosmicPotionsApi.ts`
- Follow existing pattern: `VITE_COSMIC_POTIONS_API_URL || 'http://localhost:8080'`

**Pages:**
- **Planet Explorer** — cards for each planet with biome colors, click to see ingredients
- **Ingredient Browser** — filterable grid with element color coding and rarity badges
- **Brew Station** — select 2-4 ingredients, see compatibility preview (synergies glow green, volatile flashes red), hit "Brew" and get result (success: potion + effects, failure: fun catastrophe message). New discoveries highlighted with a "NEW!" badge.
- **Codex** — the game's core screen. Grid of all effects: discovered ones show name + tier + discovery date, undiscovered show "???" with tier hint. Progress bar at top. Celebration state at 100%.
- **Brew Log** — history of attempts, successes, and spectacular failures

**Routes in App.tsx:** Public routes under `/cosmic-potions`, `/cosmic-potions/brew`, `/cosmic-potions/ingredients`, `/cosmic-potions/codex`, `/cosmic-potions/log`

**Assets:** Colored square placeholders in `packages/frontend/public/cosmic-potions/`, upgrade to pixel art later. Potentially use pixelStickers to generate some.

**Footer on page:** "Built with Go — [GitHub](repo-link)"

### Step 12 — README + Polish (~30 min)

In the `cosmic-potions` repo:
- What it is, how to run locally, architecture decisions
- Link to live demo at lixie.art/cosmic-potions
- Test coverage badge (optional)

---

## Verification

- `go test ./... -v -cover` — all tests pass, >80% coverage on brewing engine
- `go vet ./...` — no warnings
- `curl` the deployed API endpoints manually
- Frontend brew station works end-to-end (select ingredients → preview → brew → see result)
- Codex updates on successful brew (new effects appear, progress bar advances)
- Codex persists across page reloads (localStorage)
- Deployed on Railway and accessible via lixie.art

---

## Cut List (if time-crunched)

Cut in this order if time gets tight:

1. Drop brew log page on frontend (keep the API endpoint)
2. Simplify codex to a plain list instead of a styled grid
3. Simplify frontend to just brew station + codex + ingredient list
4. Skip SQLite — in-memory with seed data is enough for the demo (interface still demonstrates the pattern)
5. Skip recipe book / known recipes — focus on freeform experimental brewing

---

## Seed Data Preview

### Planets
| Name | System | Biome | Vibe |
|------|--------|-------|------|
| Verdanthia | Emerald Reach | Fungal | Lush mushroom forests, bioluminescent spores |
| The Obsidian Drift | Null Expanse | Void | Floating dark-matter shards in empty space |
| Heliox Prime | Solari Belt | Volcanic | Molten rivers, solar flare geysers |
| Crystara | Prism Nebula | Crystalline | Living crystal formations that sing |
| Thalassia | Deep Ring | Oceanic | Submerged caverns, pressure-forged pearls |
| Pyraxis | Ember Chain | Volcanic | Plasma storms, glass deserts |

### Element Interactions (5×5 Matrix)
|  | Solar | Void | Crystalline | Organic | Plasma |
|--|-------|------|-------------|---------|--------|
| Solar | neutral | volatile | synergy | synergy | neutral |
| Void | volatile | neutral | neutral | volatile | synergy |
| Crystalline | synergy | neutral | neutral | neutral | volatile |
| Organic | synergy | volatile | neutral | neutral | synergy |
| Plasma | neutral | synergy | volatile | synergy | neutral |
