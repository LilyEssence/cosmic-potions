package handler

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/LilyEssence/cosmic-potions/internal/brewing"
	"github.com/LilyEssence/cosmic-potions/internal/model"
	"github.com/LilyEssence/cosmic-potions/internal/store"
)

// ── Route Registration ──────────────────────────────────────────────
//
// GO CONCEPT: Centralized Route Table
// All routes are registered in one function so you can see the entire API
// surface at a glance. Each handler is created with only the interfaces it
// needs, then its methods are passed to the router as handler functions.
//
// chi.Router.Route("/api", fn) creates a subrouter where all paths are
// relative to /api. Inside, r.Get("/planets", ...) registers GET /api/planets.
// This grouping keeps the path hierarchy clean and makes it easy to add
// API versioning later (e.g., /api/v2 as a separate group).
//
// GO CONCEPT: Method Values as Handler Functions
// When we write `r.Get("/planets", planets.List)`, planets.List is a
// "method value" — it captures both the method AND the receiver (the
// PlanetHandler instance). The router stores it as a plain function:
//   func(http.ResponseWriter, *http.Request)
//
// This is equivalent to:
//   r.Get("/planets", func(w http.ResponseWriter, r *http.Request) {
//       planets.List(w, r)
//   })
//
// But method values are more concise and idiomatic.

// Deps holds all dependencies needed to create handlers and register routes.
// This struct makes it explicit what the handler layer requires — when
// reading main.go, you can see exactly what's being wired together.
type Deps struct {
	Planets      store.PlanetStore
	Ingredients  store.IngredientStore
	Recipes      store.RecipeStore
	Brews        store.BrewStore
	Engine       *brewing.BrewEngine
	Effects      []model.Effect
	Interactions []model.Interaction
}

// RegisterRoutes creates all handlers from the provided dependencies and
// mounts every API endpoint on the given router.
//
// This is the single source of truth for the API surface. Adding a new
// endpoint means: create a handler method, then add one line here.
func RegisterRoutes(r chi.Router, deps Deps) {
	// Create handlers — each receives only the interfaces it needs.
	planets := NewPlanetHandler(deps.Planets)
	ingredients := NewIngredientHandler(deps.Ingredients)
	recipes := NewRecipeHandler(deps.Recipes)
	brews := NewBrewHandler(deps.Brews, deps.Ingredients, deps.Engine)
	effects := NewEffectHandler(deps.Effects)
	interactions := NewInteractionHandler(deps.Interactions)

	// Mount all routes under /api.
	r.Route("/api", func(r chi.Router) {
		// ── Health Check ────────────────────────────────────────
		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, `{"status":"ok","service":"cosmic-potions"}`)
		})

		// ── Planets ─────────────────────────────────────────────
		r.Get("/planets", planets.List)
		r.Get("/planets/{id}", planets.Get)

		// ── Ingredients ─────────────────────────────────────────
		r.Get("/ingredients", ingredients.List)
		r.Get("/ingredients/{id}", ingredients.Get)

		// ── Recipes ─────────────────────────────────────────────
		r.Get("/recipes", recipes.List)
		r.Get("/recipes/{id}", recipes.Get)

		// ── Brewing ─────────────────────────────────────────────
		r.Post("/brew", brews.Brew)
		r.Post("/brew/check", brews.Check)
		r.Get("/brews", brews.List)

		// ── Codex Data ──────────────────────────────────────────
		r.Get("/effects", effects.List)
		r.Get("/interactions", interactions.List)
	})
}
