package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	// GO CONCEPT: Blank Import for Side-Effect Registration
	// This import doesn't give us a named variable — the underscore `_` discards
	// the package name. But importing it triggers the package's init() function,
	// which registers the "sqlite" driver with Go's database/sql package.
	//
	// After this import, sql.Open("sqlite", path) works. Without it, sql.Open
	// would return "unknown driver 'sqlite'". This is the standard Go pattern
	// for database drivers — the import IS the registration.
	//
	// We use modernc.org/sqlite (pure Go, no CGo) rather than mattn/go-sqlite3
	// (which requires a C compiler). Pure Go means `go build` works everywhere
	// without installing gcc or a C toolchain.
	_ "modernc.org/sqlite"

	"github.com/LilyEssence/cosmic-potions/internal/model"
)

// ── Compile-Time Interface Verification ─────────────────────────────
//
// Same pattern as MemoryStore — these lines prove at compile time that
// *SQLiteStore satisfies all four store interfaces. If we miss a method or
// get a signature wrong, the compiler catches it HERE, not at runtime.
var _ PlanetStore = (*SQLiteStore)(nil)
var _ IngredientStore = (*SQLiteStore)(nil)
var _ RecipeStore = (*SQLiteStore)(nil)
var _ BrewStore = (*SQLiteStore)(nil)

// ── SQLiteStore ─────────────────────────────────────────────────────
//
// GO CONCEPT: Same Interface, Different Implementation
// SQLiteStore and MemoryStore both satisfy the same four interfaces. main.go
// can swap between them based on an env var (STORE_TYPE). The handlers never
// know which one they're talking to — they only see PlanetStore, IngredientStore,
// etc. This is the Interface Segregation Principle in action.
//
// Key differences from MemoryStore:
//   - No sync.RWMutex — SQLite handles concurrency via its own locking
//   - No maps — data lives on disk, queried with SQL
//   - FlavorNotes stored as JSON text — needs marshal/unmarshal
//   - Relationships (recipe_effects, recipe_ingredients) use JOIN queries
//     instead of nested struct fields

// SQLiteStore implements all four store interfaces backed by a SQLite database.
type SQLiteStore struct {
	db *sql.DB
}

// NewSQLiteStore opens (or creates) a SQLite database at the given path,
// enables foreign key enforcement, and runs any pending migrations.
//
// GO CONCEPT: PRAGMA foreign_keys = ON
// SQLite has foreign key support but it's DISABLED by default for backwards
// compatibility. You must explicitly enable it on every connection. Without
// this PRAGMA, foreign key constraints in your schema are silently ignored —
// you could insert an ingredient referencing a planet that doesn't exist.
func NewSQLiteStore(dbPath string, migrationsDir string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("opening sqlite database: %w", err)
	}

	// Verify the connection actually works.
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("pinging sqlite database: %w", err)
	}

	// Enable foreign key enforcement.
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		db.Close()
		return nil, fmt.Errorf("enabling foreign keys: %w", err)
	}

	// Run pending migrations (creates tables + seeds data on first run).
	if err := RunMigrations(db, migrationsDir); err != nil {
		db.Close()
		return nil, fmt.Errorf("running migrations: %w", err)
	}

	return &SQLiteStore{db: db}, nil
}

// Close releases the database connection. Should be called when the server
// shuts down (typically via defer in main.go).
//
// GO CONCEPT: io.Closer Pattern
// Many Go types implement a Close() method. The convention is to defer the
// close right after creation:
//
//	store, err := NewSQLiteStore(path, dir)
//	if err != nil { log.Fatal(err) }
//	defer store.Close()
//
// This ensures the database connection is released when the program exits,
// even if it exits via panic or log.Fatal.
func (s *SQLiteStore) Close() error {
	return s.db.Close()
}

// ── PlanetStore implementation ──────────────────────────────────────

// GetPlanet returns a planet by ID, or ErrNotFound.
//
// GO CONCEPT: sql.QueryRow + Scan
// QueryRow executes a query expected to return at most one row. You chain
// .Scan() to read column values into Go variables. The column order in Scan()
// must match the SELECT column order — there's no named mapping.
//
// If no row matches, Scan returns sql.ErrNoRows. We translate that to our
// own ErrNotFound sentinel so callers don't need to know about database/sql.
func (s *SQLiteStore) GetPlanet(id string) (*model.Planet, error) {
	row := s.db.QueryRow(`
		SELECT id, name, system, biome, description, discovered_at
		FROM planets WHERE id = ?
	`, id)

	p, err := scanPlanet(row)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("querying planet %s: %w", id, err)
	}
	return p, nil
}

// ListPlanets returns all planets sorted by name.
//
// GO CONCEPT: sql.Query + Rows Iteration
// Query (vs QueryRow) returns a *sql.Rows cursor for multiple rows. You
// iterate with rows.Next() + rows.Scan() in a loop. ALWAYS defer rows.Close()
// to release the database cursor — leaking rows = leaking connections.
//
// After the loop, check rows.Err() — it catches errors that occurred during
// iteration (like a connection dropping mid-read).
func (s *SQLiteStore) ListPlanets() ([]model.Planet, error) {
	rows, err := s.db.Query(`
		SELECT id, name, system, biome, description, discovered_at
		FROM planets ORDER BY name
	`)
	if err != nil {
		return nil, fmt.Errorf("listing planets: %w", err)
	}
	defer rows.Close()

	results := make([]model.Planet, 0)
	for rows.Next() {
		p, err := scanPlanetRow(rows)
		if err != nil {
			return nil, fmt.Errorf("scanning planet row: %w", err)
		}
		results = append(results, *p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating planet rows: %w", err)
	}

	return results, nil
}

// GetPlanetWithIngredients returns a planet and all ingredients found there.
//
// In the MemoryStore, this iterates the ingredients map filtering by PlanetID.
// Here, we use a single SQL query with WHERE planet_id = ? — the database
// does the filtering, which is more efficient as data grows.
func (s *SQLiteStore) GetPlanetWithIngredients(id string) (*model.PlanetWithIngredients, error) {
	// First, get the planet.
	planet, err := s.GetPlanet(id)
	if err != nil {
		return nil, err // ErrNotFound passes through
	}

	// Then, get its ingredients.
	rows, err := s.db.Query(`
		SELECT id, planet_id, name, element, potency, rarity, description, flavor_notes
		FROM ingredients WHERE planet_id = ? ORDER BY name
	`, id)
	if err != nil {
		return nil, fmt.Errorf("querying ingredients for planet %s: %w", id, err)
	}
	defer rows.Close()

	ingredients := make([]model.Ingredient, 0)
	for rows.Next() {
		ing, err := scanIngredientRow(rows)
		if err != nil {
			return nil, fmt.Errorf("scanning ingredient row: %w", err)
		}
		ingredients = append(ingredients, *ing)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating ingredient rows: %w", err)
	}

	return &model.PlanetWithIngredients{
		Planet:      *planet,
		Ingredients: ingredients,
	}, nil
}

// ── IngredientStore implementation ──────────────────────────────────

// GetIngredient returns an ingredient by ID, or ErrNotFound.
func (s *SQLiteStore) GetIngredient(id string) (*model.Ingredient, error) {
	row := s.db.QueryRow(`
		SELECT id, planet_id, name, element, potency, rarity, description, flavor_notes
		FROM ingredients WHERE id = ?
	`, id)

	ing, err := scanIngredient(row)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("querying ingredient %s: %w", id, err)
	}
	return ing, nil
}

// ListIngredients returns ingredients matching the filter, sorted by name.
//
// GO CONCEPT: Dynamic WHERE Clause Building
// When filter fields are optional, we build the WHERE clause dynamically.
// Each non-nil filter adds a condition and a parameter. Using `?` placeholders
// (not string interpolation!) prevents SQL injection.
//
// The strings.Builder + args pattern is common in Go for building dynamic SQL.
// Some projects use query builders (like squirrel), but for simple filters,
// manual building is clear and avoids an extra dependency.
func (s *SQLiteStore) ListIngredients(filter model.IngredientFilter) ([]model.Ingredient, error) {
	// GO CONCEPT: strings.Builder
	// Builder provides an efficient way to build strings incrementally.
	// Unlike string concatenation (which copies on every +), Builder
	// appends to an internal buffer and allocates only when needed.
	var query strings.Builder
	var args []interface{}

	query.WriteString("SELECT id, planet_id, name, element, potency, rarity, description, flavor_notes FROM ingredients")

	// Build WHERE clause from non-nil filters.
	conditions := make([]string, 0)
	if filter.PlanetID != nil {
		conditions = append(conditions, "planet_id = ?")
		args = append(args, *filter.PlanetID)
	}
	if filter.Element != nil {
		conditions = append(conditions, "element = ?")
		args = append(args, string(*filter.Element))
	}
	if filter.Rarity != nil {
		conditions = append(conditions, "rarity = ?")
		args = append(args, string(*filter.Rarity))
	}

	if len(conditions) > 0 {
		query.WriteString(" WHERE ")
		query.WriteString(strings.Join(conditions, " AND "))
	}

	query.WriteString(" ORDER BY name")

	rows, err := s.db.Query(query.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("listing ingredients: %w", err)
	}
	defer rows.Close()

	results := make([]model.Ingredient, 0)
	for rows.Next() {
		ing, err := scanIngredientRow(rows)
		if err != nil {
			return nil, fmt.Errorf("scanning ingredient row: %w", err)
		}
		results = append(results, *ing)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating ingredient rows: %w", err)
	}

	return results, nil
}

// ── RecipeStore implementation ──────────────────────────────────────

// GetRecipe returns a recipe by ID, or ErrNotFound. This requires multiple
// queries: one for the recipe itself, one for its effects, one for its ingredients.
//
// GO CONCEPT: N+1 Query Problem (and why it's okay here)
// In ORM-heavy languages, loading a recipe + effects + ingredients is 3 queries
// (the "N+1 problem"). For a single recipe lookup, this is fine. If we were
// loading many recipes with their relationships, we'd use JOINs or batch queries.
// Since GetRecipe loads ONE recipe, separate queries are clearer than a complex JOIN.
func (s *SQLiteStore) GetRecipe(id string) (*model.Recipe, error) {
	// Load base recipe.
	row := s.db.QueryRow(`
		SELECT id, name, description, difficulty FROM recipes WHERE id = ?
	`, id)

	var r model.Recipe
	err := row.Scan(&r.ID, &r.Name, &r.Description, &r.Difficulty)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("querying recipe %s: %w", id, err)
	}

	// Load effects.
	r.Effects, err = s.loadRecipeEffects(id)
	if err != nil {
		return nil, err
	}

	// Load ingredients.
	r.Ingredients, err = s.loadRecipeIngredients(id)
	if err != nil {
		return nil, err
	}

	return &r, nil
}

// ListRecipes returns recipes matching the filter, sorted by name. Each recipe
// includes its full effects and ingredients lists.
func (s *SQLiteStore) ListRecipes(filter model.RecipeFilter) ([]model.Recipe, error) {
	var query strings.Builder
	var args []interface{}

	query.WriteString("SELECT id, name, description, difficulty FROM recipes")

	if filter.Difficulty != nil {
		query.WriteString(" WHERE difficulty = ?")
		args = append(args, string(*filter.Difficulty))
	}

	query.WriteString(" ORDER BY name")

	rows, err := s.db.Query(query.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("listing recipes: %w", err)
	}
	defer rows.Close()

	results := make([]model.Recipe, 0)
	for rows.Next() {
		var r model.Recipe
		if err := rows.Scan(&r.ID, &r.Name, &r.Description, &r.Difficulty); err != nil {
			return nil, fmt.Errorf("scanning recipe row: %w", err)
		}

		// Load relationships for each recipe.
		r.Effects, err = s.loadRecipeEffects(r.ID)
		if err != nil {
			return nil, err
		}
		r.Ingredients, err = s.loadRecipeIngredients(r.ID)
		if err != nil {
			return nil, err
		}

		results = append(results, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating recipe rows: %w", err)
	}

	return results, nil
}

// loadRecipeEffects returns the effects for a recipe, sorted alphabetically.
func (s *SQLiteStore) loadRecipeEffects(recipeID string) ([]string, error) {
	rows, err := s.db.Query(`
		SELECT effect FROM recipe_effects WHERE recipe_id = ? ORDER BY effect
	`, recipeID)
	if err != nil {
		return nil, fmt.Errorf("loading effects for recipe %s: %w", recipeID, err)
	}
	defer rows.Close()

	effects := make([]string, 0)
	for rows.Next() {
		var e string
		if err := rows.Scan(&e); err != nil {
			return nil, fmt.Errorf("scanning effect: %w", err)
		}
		effects = append(effects, e)
	}
	return effects, rows.Err()
}

// loadRecipeIngredients returns the ingredients for a recipe, sorted by ingredient ID.
func (s *SQLiteStore) loadRecipeIngredients(recipeID string) ([]model.RecipeIngredient, error) {
	rows, err := s.db.Query(`
		SELECT ingredient_id, quantity FROM recipe_ingredients
		WHERE recipe_id = ? ORDER BY ingredient_id
	`, recipeID)
	if err != nil {
		return nil, fmt.Errorf("loading ingredients for recipe %s: %w", recipeID, err)
	}
	defer rows.Close()

	ingredients := make([]model.RecipeIngredient, 0)
	for rows.Next() {
		var ri model.RecipeIngredient
		if err := rows.Scan(&ri.IngredientID, &ri.Quantity); err != nil {
			return nil, fmt.Errorf("scanning recipe ingredient: %w", err)
		}
		ingredients = append(ingredients, ri)
	}
	return ingredients, rows.Err()
}

// ── BrewStore implementation ────────────────────────────────────────

// SaveBrew persists a brew result with its ingredients and effects.
//
// GO CONCEPT: SQL Transactions for Multi-Table Inserts
// A brew spans three tables: brews, brew_ingredients, brew_effects. If we
// insert into brews but fail on brew_ingredients, we'd have an orphaned brew.
// Wrapping everything in a transaction ensures atomicity — all or nothing.
//
// GO CONCEPT: Nullable Columns with Pointers
// The Brew struct has `RecipeID *string` (pointer = nullable). When inserting,
// a nil pointer becomes SQL NULL, and a non-nil pointer becomes the string value.
// database/sql handles this automatically for pointer types — no special code needed.
func (s *SQLiteStore) SaveBrew(brew model.Brew) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("starting brew transaction: %w", err)
	}
	defer tx.Rollback() // No-op after Commit.

	// Insert the base brew.
	_, err = tx.Exec(`
		INSERT INTO brews (id, recipe_id, result, potion_name, notes, brewed_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, brew.ID, brew.RecipeID, string(brew.Result), brew.PotionName, brew.Notes,
		brew.BrewedAt.Format(time.RFC3339))
	if err != nil {
		return fmt.Errorf("inserting brew: %w", err)
	}

	// Insert ingredient references.
	for _, ingID := range brew.IngredientsUsed {
		_, err = tx.Exec(`
			INSERT INTO brew_ingredients (brew_id, ingredient_id) VALUES (?, ?)
		`, brew.ID, ingID)
		if err != nil {
			return fmt.Errorf("inserting brew ingredient %s: %w", ingID, err)
		}
	}

	// Insert effect references.
	for _, effect := range brew.Effects {
		_, err = tx.Exec(`
			INSERT INTO brew_effects (brew_id, effect) VALUES (?, ?)
		`, brew.ID, effect)
		if err != nil {
			return fmt.Errorf("inserting brew effect %s: %w", effect, err)
		}
	}

	return tx.Commit()
}

// GetBrew returns a brew by ID, or ErrNotFound.
func (s *SQLiteStore) GetBrew(id string) (*model.Brew, error) {
	row := s.db.QueryRow(`
		SELECT id, recipe_id, result, potion_name, notes, brewed_at
		FROM brews WHERE id = ?
	`, id)

	brew, err := scanBrew(row)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("querying brew %s: %w", id, err)
	}

	// Load ingredients.
	brew.IngredientsUsed, err = s.loadBrewIngredients(id)
	if err != nil {
		return nil, err
	}

	// Load effects.
	brew.Effects, err = s.loadBrewEffects(id)
	if err != nil {
		return nil, err
	}

	return brew, nil
}

// ListBrews returns all brews in reverse chronological order (newest first).
func (s *SQLiteStore) ListBrews() ([]model.Brew, error) {
	rows, err := s.db.Query(`
		SELECT id, recipe_id, result, potion_name, notes, brewed_at
		FROM brews ORDER BY brewed_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("listing brews: %w", err)
	}
	defer rows.Close()

	results := make([]model.Brew, 0)
	for rows.Next() {
		brew, err := scanBrewRow(rows)
		if err != nil {
			return nil, fmt.Errorf("scanning brew row: %w", err)
		}

		brew.IngredientsUsed, err = s.loadBrewIngredients(brew.ID)
		if err != nil {
			return nil, err
		}
		brew.Effects, err = s.loadBrewEffects(brew.ID)
		if err != nil {
			return nil, err
		}

		results = append(results, *brew)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating brew rows: %w", err)
	}

	return results, nil
}

// loadBrewIngredients returns ingredient IDs for a brew, sorted alphabetically.
func (s *SQLiteStore) loadBrewIngredients(brewID string) ([]string, error) {
	rows, err := s.db.Query(`
		SELECT ingredient_id FROM brew_ingredients
		WHERE brew_id = ? ORDER BY ingredient_id
	`, brewID)
	if err != nil {
		return nil, fmt.Errorf("loading brew ingredients: %w", err)
	}
	defer rows.Close()

	ids := make([]string, 0)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scanning brew ingredient: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// loadBrewEffects returns effects for a brew, sorted alphabetically.
func (s *SQLiteStore) loadBrewEffects(brewID string) ([]string, error) {
	rows, err := s.db.Query(`
		SELECT effect FROM brew_effects WHERE brew_id = ? ORDER BY effect
	`, brewID)
	if err != nil {
		return nil, fmt.Errorf("loading brew effects: %w", err)
	}
	defer rows.Close()

	effects := make([]string, 0)
	for rows.Next() {
		var e string
		if err := rows.Scan(&e); err != nil {
			return nil, fmt.Errorf("scanning brew effect: %w", err)
		}
		effects = append(effects, e)
	}
	return effects, rows.Err()
}

// ── Row Scanners ────────────────────────────────────────────────────
//
// GO CONCEPT: Scanner Interface
// database/sql has two types that can Scan rows:
//   - *sql.Row    (from QueryRow, single row)
//   - *sql.Rows   (from Query, multiple rows — but can Scan one at a time)
//
// Both have a Scan() method, but they're different types with no shared
// interface in the standard library. We write separate helper functions
// for each. Some Go projects define their own Scanner interface to unify
// them, but for clarity as a learning project, we keep them explicit.

// scanPlanet scans a single planet from a QueryRow result.
func scanPlanet(row *sql.Row) (*model.Planet, error) {
	var p model.Planet
	var discoveredAt string
	err := row.Scan(&p.ID, &p.Name, &p.System, &p.Biome, &p.Description, &discoveredAt)
	if err != nil {
		return nil, err
	}
	p.DiscoveredAt, err = time.Parse(time.RFC3339, discoveredAt)
	if err != nil {
		return nil, fmt.Errorf("parsing discovered_at: %w", err)
	}
	return &p, nil
}

// scanPlanetRow scans a planet from a Rows iterator.
func scanPlanetRow(rows *sql.Rows) (*model.Planet, error) {
	var p model.Planet
	var discoveredAt string
	err := rows.Scan(&p.ID, &p.Name, &p.System, &p.Biome, &p.Description, &discoveredAt)
	if err != nil {
		return nil, err
	}
	p.DiscoveredAt, err = time.Parse(time.RFC3339, discoveredAt)
	if err != nil {
		return nil, fmt.Errorf("parsing discovered_at: %w", err)
	}
	return &p, nil
}

// scanIngredient scans a single ingredient from a QueryRow result.
//
// GO CONCEPT: JSON in SQLite
// FlavorNotes is a []string in Go but stored as a JSON array TEXT in SQLite
// (e.g., '["bitter","sweet"]'). We scan into a string, then json.Unmarshal
// into a []string. This is a common pattern for storing arrays in SQLite,
// which has no native array type.
func scanIngredient(row *sql.Row) (*model.Ingredient, error) {
	var ing model.Ingredient
	var flavorJSON string
	err := row.Scan(&ing.ID, &ing.PlanetID, &ing.Name, &ing.Element,
		&ing.Potency, &ing.Rarity, &ing.Description, &flavorJSON)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(flavorJSON), &ing.FlavorNotes); err != nil {
		return nil, fmt.Errorf("parsing flavor_notes JSON: %w", err)
	}
	return &ing, nil
}

// scanIngredientRow scans an ingredient from a Rows iterator.
func scanIngredientRow(rows *sql.Rows) (*model.Ingredient, error) {
	var ing model.Ingredient
	var flavorJSON string
	err := rows.Scan(&ing.ID, &ing.PlanetID, &ing.Name, &ing.Element,
		&ing.Potency, &ing.Rarity, &ing.Description, &flavorJSON)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(flavorJSON), &ing.FlavorNotes); err != nil {
		return nil, fmt.Errorf("parsing flavor_notes JSON: %w", err)
	}
	return &ing, nil
}

// scanBrew scans a single brew from a QueryRow result.
func scanBrew(row *sql.Row) (*model.Brew, error) {
	var b model.Brew
	var brewedAt string
	err := row.Scan(&b.ID, &b.RecipeID, &b.Result, &b.PotionName, &b.Notes, &brewedAt)
	if err != nil {
		return nil, err
	}
	b.BrewedAt, err = time.Parse(time.RFC3339, brewedAt)
	if err != nil {
		return nil, fmt.Errorf("parsing brewed_at: %w", err)
	}
	// Initialize empty slices so JSON encodes [] not null.
	b.IngredientsUsed = make([]string, 0)
	b.Effects = make([]string, 0)
	return &b, nil
}

// scanBrewRow scans a brew from a Rows iterator.
func scanBrewRow(rows *sql.Rows) (*model.Brew, error) {
	var b model.Brew
	var brewedAt string
	err := rows.Scan(&b.ID, &b.RecipeID, &b.Result, &b.PotionName, &b.Notes, &brewedAt)
	if err != nil {
		return nil, err
	}
	b.BrewedAt, err = time.Parse(time.RFC3339, brewedAt)
	if err != nil {
		return nil, fmt.Errorf("parsing brewed_at: %w", err)
	}
	b.IngredientsUsed = make([]string, 0)
	b.Effects = make([]string, 0)
	return &b, nil
}

// ── Interaction Helper ──────────────────────────────────────────────
//
// Interactions are read-only data that the server needs for the brew engine
// and the /api/interactions endpoint. We load them all at startup rather
// than querying per-request.

// LoadInteractions reads all element interactions from the database.
// This is used by main.go to pass to the brew engine, similar to how
// seed.Interactions() works for the in-memory store.
func (s *SQLiteStore) LoadInteractions() ([]model.Interaction, error) {
	rows, err := s.db.Query(`SELECT element1, element2, result FROM interactions`)
	if err != nil {
		return nil, fmt.Errorf("loading interactions: %w", err)
	}
	defer rows.Close()

	results := make([]model.Interaction, 0)
	for rows.Next() {
		var ix model.Interaction
		if err := rows.Scan(&ix.Element1, &ix.Element2, &ix.Result); err != nil {
			return nil, fmt.Errorf("scanning interaction: %w", err)
		}
		results = append(results, ix)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating interactions: %w", err)
	}

	// Sort for deterministic output (same as seed package).
	sort.Slice(results, func(i, j int) bool {
		if results[i].Element1 != results[j].Element1 {
			return results[i].Element1 < results[j].Element1
		}
		return results[i].Element2 < results[j].Element2
	})

	return results, nil
}

// LoadInteractionMap builds the Element1→Element2→Result map used by the
// brew engine, loaded from the database instead of the seed package.
func (s *SQLiteStore) LoadInteractionMap() (map[model.Element]map[model.Element]model.InteractionResult, error) {
	interactions, err := s.LoadInteractions()
	if err != nil {
		return nil, err
	}

	matrix := make(map[model.Element]map[model.Element]model.InteractionResult)
	for _, ix := range interactions {
		if matrix[ix.Element1] == nil {
			matrix[ix.Element1] = make(map[model.Element]model.InteractionResult)
		}
		matrix[ix.Element1][ix.Element2] = ix.Result
	}
	return matrix, nil
}
