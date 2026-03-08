package store

import (
	"errors"

	"github.com/LilyEssence/cosmic-potions/internal/model"
)

// ── Sentinel Errors ─────────────────────────────────────────────────
//
// GO CONCEPT: Sentinel Errors
// A sentinel error is a predefined, package-level error value that callers can
// check against. Instead of comparing error strings (fragile), callers use:
//
//	if errors.Is(err, store.ErrNotFound) { ... }
//
// This is Go's equivalent of typed exceptions — but errors are just values,
// not special language constructs. errors.New() creates an error with a message;
// the variable name (ErrNotFound) is the stable identifier code checks against.
//
// Why export these? Handlers need to distinguish "not found" (→ 404) from
// unexpected errors (→ 500). Without sentinel errors, every handler would be
// guessing based on error message strings.

// ErrNotFound indicates that the requested entity does not exist.
var ErrNotFound = errors.New("not found")

// ── Store Interfaces ────────────────────────────────────────────────
//
// GO CONCEPT: Interface Design Philosophy
// Go interfaces are typically small — one to three methods is ideal. The standard
// library's most powerful interface, io.Reader, has exactly one method. Small
// interfaces are easier to implement, easier to mock, and easier to compose.
//
// We define four separate interfaces rather than one large Store interface.
// Each consumer (handler, brewing engine) declares which interface it needs:
//
//   func NewIngredientHandler(s store.IngredientStore) *IngredientHandler
//   func NewBrewHandler(s store.BrewStore, i store.IngredientStore) *BrewHandler
//
// The in-memory implementation is ONE struct that satisfies ALL four interfaces.
// But consumers only see their slice. This is the Interface Segregation Principle
// emerging naturally from Go's structural typing — no explicit "implements."
//
// GO CONCEPT: Return Patterns
// Every method that can fail returns (value, error). This is the universal Go
// pattern. Callers MUST check the error before using the value:
//
//   planet, err := store.GetPlanet(id)
//   if err != nil {
//       return err  // or handle it
//   }
//   // safe to use planet here
//
// For "get one" methods, we return a pointer (*model.Planet) so that the zero
// value is nil — making it obvious if you accidentally use it without checking
// the error. For "list" methods, we return a slice, which is nil-safe (a nil
// slice behaves like an empty slice for most operations).

// PlanetStore defines read operations for planet data.
type PlanetStore interface {
	// GetPlanet returns a single planet by ID.
	// Returns ErrNotFound if the planet doesn't exist.
	GetPlanet(id string) (*model.Planet, error)

	// ListPlanets returns all planets.
	ListPlanets() ([]model.Planet, error)

	// GetPlanetWithIngredients returns a planet bundled with all ingredients
	// found on that world. This exists so the handler can serve a rich detail
	// page in one API call instead of requiring the client to make two requests.
	//
	// The in-memory version iterates the ingredients map. The SQLite version
	// (Step 9) will do this as a JOIN — same interface, different performance
	// characteristics.
	GetPlanetWithIngredients(id string) (*model.PlanetWithIngredients, error)
}

// IngredientStore defines read operations for ingredient data.
type IngredientStore interface {
	// GetIngredient returns a single ingredient by ID.
	// Returns ErrNotFound if the ingredient doesn't exist.
	GetIngredient(id string) (*model.Ingredient, error)

	// ListIngredients returns ingredients matching the given filter criteria.
	// A zero-value filter (all nil fields) returns all ingredients.
	ListIngredients(filter model.IngredientFilter) ([]model.Ingredient, error)
}

// RecipeStore defines read operations for recipe data.
type RecipeStore interface {
	// GetRecipe returns a single recipe by ID.
	// Returns ErrNotFound if the recipe doesn't exist.
	GetRecipe(id string) (*model.Recipe, error)

	// ListRecipes returns recipes matching the given filter criteria.
	// A zero-value filter (all nil fields) returns all recipes.
	ListRecipes(filter model.RecipeFilter) ([]model.Recipe, error)
}

// BrewStore defines read/write operations for the brew log.
//
// This is the only store with a write method (SaveBrew). Planets, ingredients,
// and recipes are read-only seed data. Brews are created at runtime when players
// use the brew endpoint.
type BrewStore interface {
	// SaveBrew persists a brew result.
	SaveBrew(brew model.Brew) error

	// GetBrew returns a single brew by ID.
	// Returns ErrNotFound if the brew doesn't exist.
	GetBrew(id string) (*model.Brew, error)

	// ListBrews returns all brews in reverse chronological order (newest first).
	ListBrews() ([]model.Brew, error)
}
