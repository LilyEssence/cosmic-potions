package store

import (
	"sort"
	"sync"

	"github.com/LilyEssence/cosmic-potions/internal/model"
	"github.com/LilyEssence/cosmic-potions/seed"
)

// ── Compile-Time Interface Check ────────────────────────────────────
//
// GO CONCEPT: Compile-Time Interface Verification
// These lines look strange but serve an important purpose. They verify AT
// COMPILE TIME that *MemoryStore satisfies all four interfaces.
//
// The syntax: var _ InterfaceName = (*StructType)(nil)
//   - `var _`       → declares a variable we'll never use (blank identifier)
//   - `= (*MemoryStore)(nil)` → a nil pointer of type *MemoryStore
//   - The assignment only works if *MemoryStore implements the interface
//
// If you forget to implement a method, or get the signature wrong, you'll get
// a compile error HERE (with a clear message) instead of wherever the struct
// is first used as that interface (which might be far away and confusing).
//
// This is a universal Go idiom. You'll see it in the standard library and
// virtually every well-structured Go project.
var _ PlanetStore = (*MemoryStore)(nil)
var _ IngredientStore = (*MemoryStore)(nil)
var _ RecipeStore = (*MemoryStore)(nil)
var _ BrewStore = (*MemoryStore)(nil)

// ── MemoryStore ─────────────────────────────────────────────────────
//
// GO CONCEPT: Single Struct, Multiple Interfaces
// MemoryStore is ONE struct that implements ALL four store interfaces. This
// is practical because in-memory data lives together (it's all maps in the
// same struct). But consumers never see "MemoryStore" — they see only the
// interface they need:
//
//   func NewPlanetHandler(planets store.PlanetStore) *PlanetHandler
//   func NewBrewHandler(brews store.BrewStore, ingredients store.IngredientStore) *BrewHandler
//
// The handler doesn't know (or care) that PlanetStore and IngredientStore
// are the same struct behind the scenes. It only has access to the methods
// declared in its interface parameter.
//
// GO CONCEPT: sync.RWMutex
// Go's HTTP server runs each request in its own goroutine (lightweight thread).
// Multiple goroutines reading/writing the same maps = data race = crashes or
// corruption. RWMutex provides two lock modes:
//
//   mu.RLock()  → shared read lock (many readers at once, blocks writers)
//   mu.Lock()   → exclusive write lock (one writer, blocks everyone)
//
// Since this game is read-heavy (browsing planets/ingredients) and write-light
// (occasional brews), RWMutex is ideal — concurrent GETs don't block each other.
// A plain sync.Mutex would serialize ALL access, even concurrent reads.

// MemoryStore holds all game data in memory, loaded from seed data at startup.
// Thread-safe for concurrent HTTP handler access.
type MemoryStore struct {
	mu          sync.RWMutex
	planets     map[string]model.Planet
	ingredients map[string]model.Ingredient
	recipes     map[string]model.Recipe
	brews       map[string]model.Brew
	brewOrder   []string // insertion order — maps don't preserve order in Go
}

// NewMemoryStore creates a MemoryStore pre-loaded with seed data.
//
// GO CONCEPT: Constructor Functions
// Go has no constructors or `new` keywords (well, `new()` exists but is rarely
// used). Instead, the convention is a `NewXxx()` function that returns a pointer
// to an initialized struct. The caller gets a ready-to-use value — no separate
// "init" step.
//
// Why return *MemoryStore instead of the interface? Convention: "Accept interfaces,
// return structs." Returning the concrete type gives callers maximum flexibility.
// The call site can assign it to any interface the struct satisfies:
//
//	store := store.NewMemoryStore()           // *MemoryStore
//	var planets store.PlanetStore = store     // ✅ works
//	var brews store.BrewStore = store         // ✅ also works
//
// GO CONCEPT: Pointer Receiver (*MemoryStore)
// The methods below are defined on *MemoryStore (pointer receiver), not MemoryStore
// (value receiver). This matters because:
//  1. The mutex must not be copied (copying a locked mutex = undefined behavior)
//  2. We want methods to see the same data (a value receiver gets a copy)
//
// Rule of thumb: if a struct has a mutex or mutable state, always use pointer receivers.
func NewMemoryStore() *MemoryStore {
	s := &MemoryStore{
		// GO CONCEPT: make()
		// Maps must be initialized before use. `make(map[K]V)` allocates an
		// empty, ready-to-use map. Writing to a nil map panics. Reading from
		// a nil map returns the zero value (safe but misleading).
		planets:     make(map[string]model.Planet),
		ingredients: make(map[string]model.Ingredient),
		recipes:     make(map[string]model.Recipe),
		brews:       make(map[string]model.Brew),
	}

	// Load seed data into maps for O(1) lookups by ID.
	for _, p := range seed.Planets() {
		s.planets[p.ID] = p
	}
	for _, i := range seed.Ingredients() {
		s.ingredients[i.ID] = i
	}
	for _, r := range seed.Recipes() {
		s.recipes[r.ID] = r
	}
	// No seed brews — brews are created at runtime by players.

	return s
}

// ── PlanetStore implementation ──────────────────────────────────────

// GetPlanet returns a planet by ID, or ErrNotFound.
//
// GO CONCEPT: Map Lookup with Comma-Ok
// The two-value form `val, ok := myMap[key]` is how you check if a key exists.
//   - ok=true  → key exists, val is the value
//   - ok=false → key doesn't exist, val is the zero value
//
// This avoids the ambiguity of "is this a real empty value or a missing key?"
// that plagues JavaScript (where obj["missing"] returns undefined).
//
// GO CONCEPT: defer
// `defer s.mu.RUnlock()` schedules the unlock to run when this function returns,
// no matter how it returns (normal return, early return, or panic). It's Go's
// cleanup guarantee — like try/finally but less verbose. The deferred call runs
// AFTER the return value is set but BEFORE the caller receives it.
func (s *MemoryStore) GetPlanet(id string) (*model.Planet, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	planet, ok := s.planets[id]
	if !ok {
		return nil, ErrNotFound
	}
	return &planet, nil
}

// ListPlanets returns all planets, sorted by name for consistent ordering.
//
// GO CONCEPT: Nil Slices vs Empty Slices
// `var results []model.Planet` creates a nil slice (length 0, capacity 0).
// `append()` handles nil slices gracefully — it allocates as needed. A nil
// slice marshals to JSON `null`, while `[]model.Planet{}` marshals to `[]`.
// We initialize with make() here so the API always returns `[]`, not `null`.
func (s *MemoryStore) ListPlanets() ([]model.Planet, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	results := make([]model.Planet, 0, len(s.planets))
	for _, p := range s.planets {
		results = append(results, p)
	}

	// GO CONCEPT: sort.Slice
	// Go's sort package provides sort.Slice() which takes a slice and a "less"
	// function. The less function receives two indices and returns true if
	// element [i] should come before element [j]. This is Go's generic sorting
	// mechanism — it works on any slice type.
	sort.Slice(results, func(i, j int) bool {
		return results[i].Name < results[j].Name
	})

	return results, nil
}

// GetPlanetWithIngredients returns a planet and all ingredients harvested from it.
//
// In memory this is two lookups (planet map + iterate ingredients). The SQLite
// implementation (Step 9) will do this as a single JOIN query. Same interface,
// different performance profile — that's the payoff of the abstraction.
func (s *MemoryStore) GetPlanetWithIngredients(id string) (*model.PlanetWithIngredients, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	planet, ok := s.planets[id]
	if !ok {
		return nil, ErrNotFound
	}

	ingredients := make([]model.Ingredient, 0)
	for _, ing := range s.ingredients {
		if ing.PlanetID == id {
			ingredients = append(ingredients, ing)
		}
	}

	// Sort ingredients by name for consistent API output.
	sort.Slice(ingredients, func(i, j int) bool {
		return ingredients[i].Name < ingredients[j].Name
	})

	return &model.PlanetWithIngredients{
		Planet:      planet,
		Ingredients: ingredients,
	}, nil
}

// ── IngredientStore implementation ──────────────────────────────────

// GetIngredient returns an ingredient by ID, or ErrNotFound.
func (s *MemoryStore) GetIngredient(id string) (*model.Ingredient, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ing, ok := s.ingredients[id]
	if !ok {
		return nil, ErrNotFound
	}
	return &ing, nil
}

// ListIngredients returns ingredients matching the filter criteria.
// Nil filter fields are ignored (act as "match all").
//
// GO CONCEPT: Pointer Dereferencing
// Filter fields are pointers (*Element, *Rarity, *string). To compare:
//
//  1. Check if the pointer is non-nil (field was provided)
//
//  2. Dereference with * to get the actual value
//
//     if filter.Element != nil && ing.Element != *filter.Element {
//     continue  // skip — doesn't match
//     }
//
// The `continue` keyword skips to the next loop iteration (like TypeScript's
// `continue` in a for loop). Combined with the nil checks, this implements
// "AND" filtering: every non-nil filter must match for the ingredient to be included.
func (s *MemoryStore) ListIngredients(filter model.IngredientFilter) ([]model.Ingredient, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	results := make([]model.Ingredient, 0)
	for _, ing := range s.ingredients {
		if filter.PlanetID != nil && ing.PlanetID != *filter.PlanetID {
			continue
		}
		if filter.Element != nil && ing.Element != *filter.Element {
			continue
		}
		if filter.Rarity != nil && ing.Rarity != *filter.Rarity {
			continue
		}
		results = append(results, ing)
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Name < results[j].Name
	})

	return results, nil
}

// ── RecipeStore implementation ──────────────────────────────────────

// GetRecipe returns a recipe by ID, or ErrNotFound.
func (s *MemoryStore) GetRecipe(id string) (*model.Recipe, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	recipe, ok := s.recipes[id]
	if !ok {
		return nil, ErrNotFound
	}
	return &recipe, nil
}

// ListRecipes returns recipes, optionally filtered by difficulty.
func (s *MemoryStore) ListRecipes(filter model.RecipeFilter) ([]model.Recipe, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	results := make([]model.Recipe, 0)
	for _, r := range s.recipes {
		if filter.Difficulty != nil && r.Difficulty != *filter.Difficulty {
			continue
		}
		results = append(results, r)
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Name < results[j].Name
	})

	return results, nil
}

// ── BrewStore implementation ────────────────────────────────────────

// SaveBrew persists a new brew to the in-memory log.
//
// GO CONCEPT: Write Lock vs Read Lock
// SaveBrew uses mu.Lock() (exclusive write lock) instead of mu.RLock() (shared
// read lock). During a write lock:
//   - No other goroutine can read OR write
//   - All RLock() calls block until the write lock is released
//   - All Lock() calls block until the write lock is released
//
// This guarantees that the map write and the brewOrder append happen atomically —
// no other goroutine can see a half-finished state.
//
// GO CONCEPT: brewOrder Slice
// Go maps have intentionally randomized iteration order. For the brew log,
// we want chronological order (newest first in API responses). The brewOrder
// slice records insertion order, so ListBrews can iterate it in reverse
// without needing to sort by timestamp every time.
func (s *MemoryStore) SaveBrew(brew model.Brew) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.brews[brew.ID] = brew
	s.brewOrder = append(s.brewOrder, brew.ID)
	return nil
}

// GetBrew returns a brew by ID, or ErrNotFound.
func (s *MemoryStore) GetBrew(id string) (*model.Brew, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	brew, ok := s.brews[id]
	if !ok {
		return nil, ErrNotFound
	}
	return &brew, nil
}

// ListBrews returns all brews in reverse chronological order (newest first).
//
// GO CONCEPT: Iterating in Reverse
// Go doesn't have a built-in reverse iterator. The idiomatic approach is a
// standard for loop counting backwards. This is more explicit than methods
// like JavaScript's .reverse() — you see exactly what's happening.
func (s *MemoryStore) ListBrews() ([]model.Brew, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	results := make([]model.Brew, 0, len(s.brewOrder))
	for i := len(s.brewOrder) - 1; i >= 0; i-- {
		id := s.brewOrder[i]
		if brew, ok := s.brews[id]; ok {
			results = append(results, brew)
		}
	}

	return results, nil
}
