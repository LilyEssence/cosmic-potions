package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/LilyEssence/cosmic-potions/internal/store"
)

// ── Planet Handlers ─────────────────────────────────────────────────
//
// GO CONCEPT: Handler Structs (Method-Based Handlers)
// Instead of standalone functions, we attach handler methods to a struct that
// holds dependencies. This is Go's equivalent of a class with constructor
// injection:
//
//   TypeScript:  class PlanetController { constructor(private repo: PlanetRepo) {} }
//   Go:          type PlanetHandler struct { planets store.PlanetStore }
//
// Each method on PlanetHandler has the standard http.HandlerFunc signature:
//   func(w http.ResponseWriter, r *http.Request)
//
// chi's router accepts these as route handlers: r.Get("/planets", h.List)
// Here, h.List is a "method value" — a function bound to a specific receiver.
// chi doesn't know or care that it's a method; it just sees a function with
// the right signature.

// PlanetHandler serves planet-related API endpoints.
// It depends only on PlanetStore — it can't accidentally access brew or
// recipe data because the interface doesn't expose those methods.
type PlanetHandler struct {
	planets store.PlanetStore
}

// NewPlanetHandler creates a handler with the given planet store.
func NewPlanetHandler(planets store.PlanetStore) *PlanetHandler {
	return &PlanetHandler{planets: planets}
}

// List returns all planets in the galaxy.
// GET /api/planets
func (h *PlanetHandler) List(w http.ResponseWriter, r *http.Request) {
	planets, err := h.planets.ListPlanets()
	if err != nil {
		mapError(w, err, "planets")
		return
	}
	writeJSON(w, http.StatusOK, planets)
}

// Get returns a single planet along with all its harvestable ingredients.
// GET /api/planets/{id}
//
// GO CONCEPT: chi.URLParam
// chi extracts path parameters from the URL pattern. When the route is
// registered as "/planets/{id}", chi.URLParam(r, "id") retrieves the actual
// value from the request URL. This is similar to Express's req.params.id.
//
// The parameter value comes from the request's context (stored by chi's
// router middleware), not from parsing the URL yourself.
func (h *PlanetHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	result, err := h.planets.GetPlanetWithIngredients(id)
	if err != nil {
		mapError(w, err, "planet")
		return
	}

	writeJSON(w, http.StatusOK, result)
}
