package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/LilyEssence/cosmic-potions/internal/model"
	"github.com/LilyEssence/cosmic-potions/internal/store"
)

// ── Ingredient Handlers ─────────────────────────────────────────────

// IngredientHandler serves ingredient-related API endpoints.
type IngredientHandler struct {
	ingredients store.IngredientStore
}

// NewIngredientHandler creates a handler with the given ingredient store.
func NewIngredientHandler(ingredients store.IngredientStore) *IngredientHandler {
	return &IngredientHandler{ingredients: ingredients}
}

// List returns ingredients, optionally filtered by query parameters.
// GET /api/ingredients?element=void&rarity=rare&planet_id=planet-verdanthia
//
// GO CONCEPT: Query Parameters
// r.URL.Query() returns a url.Values (which is a map[string][]string).
// The .Get(key) method returns the first value for a key, or "" if absent.
// This is how you read ?key=value from the URL — similar to Express's
// req.query.key but with explicit type handling.
//
// GO CONCEPT: Type Conversion
// model.Element(v) converts a plain string to our Element type. Go allows
// this because Element's underlying type is string. This is NOT a cast
// (which reinterprets bits) — it's a type conversion (which the compiler
// verifies is safe). The value is unchanged, only the type label changes.
//
// GO CONCEPT: Address-of Operator (&)
// &elem creates a pointer to elem. The filter's Element field is *Element
// (pointer), meaning "optional." Setting filter.Element = &elem says
// "I want to filter by this element." Leaving it nil says "don't filter."
// This is Go's pointer-for-optional pattern — no Option<T> or Maybe<T> needed.
func (h *IngredientHandler) List(w http.ResponseWriter, r *http.Request) {
	filter := model.IngredientFilter{}

	if v := r.URL.Query().Get("element"); v != "" {
		elem := model.Element(v)
		filter.Element = &elem
	}
	if v := r.URL.Query().Get("rarity"); v != "" {
		rar := model.Rarity(v)
		filter.Rarity = &rar
	}
	if v := r.URL.Query().Get("planet_id"); v != "" {
		filter.PlanetID = &v
	}

	ingredients, err := h.ingredients.ListIngredients(filter)
	if err != nil {
		mapError(w, err, "ingredients")
		return
	}

	writeJSON(w, http.StatusOK, ingredients)
}

// Get returns a single ingredient by ID.
// GET /api/ingredients/{id}
func (h *IngredientHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	ingredient, err := h.ingredients.GetIngredient(id)
	if err != nil {
		mapError(w, err, "ingredient")
		return
	}

	writeJSON(w, http.StatusOK, ingredient)
}
