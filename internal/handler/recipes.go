package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/LilyEssence/cosmic-potions/internal/model"
	"github.com/LilyEssence/cosmic-potions/internal/store"
)

// ── Recipe Handlers ─────────────────────────────────────────────────

// RecipeHandler serves recipe-related API endpoints.
type RecipeHandler struct {
	recipes store.RecipeStore
}

// NewRecipeHandler creates a handler with the given recipe store.
func NewRecipeHandler(recipes store.RecipeStore) *RecipeHandler {
	return &RecipeHandler{recipes: recipes}
}

// List returns recipes, optionally filtered by difficulty.
// GET /api/recipes?difficulty=novice
func (h *RecipeHandler) List(w http.ResponseWriter, r *http.Request) {
	filter := model.RecipeFilter{}

	if v := r.URL.Query().Get("difficulty"); v != "" {
		diff := model.Difficulty(v)
		filter.Difficulty = &diff
	}

	recipes, err := h.recipes.ListRecipes(filter)
	if err != nil {
		mapError(w, err, "recipes")
		return
	}

	writeJSON(w, http.StatusOK, recipes)
}

// Get returns a single recipe with its required ingredients.
// GET /api/recipes/{id}
func (h *RecipeHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	recipe, err := h.recipes.GetRecipe(id)
	if err != nil {
		mapError(w, err, "recipe")
		return
	}

	writeJSON(w, http.StatusOK, recipe)
}
