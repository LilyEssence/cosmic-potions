package handler

import (
	"crypto/rand"
	"fmt"
	"net/http"
	"time"

	"github.com/LilyEssence/cosmic-potions/internal/brewing"
	"github.com/LilyEssence/cosmic-potions/internal/model"
	"github.com/LilyEssence/cosmic-potions/internal/store"
)

// ── Brew Handlers ───────────────────────────────────────────────────
//
// GO CONCEPT: Multiple Interface Dependencies
// BrewHandler needs three dependencies:
//   - BrewStore to save/list brew results
//   - IngredientStore to look up ingredients by ID from the request body
//   - BrewEngine to run the actual brewing logic
//
// Each is accepted through the minimal interface it needs (BrewStore, not
// "the entire store"). The handler doesn't know or care that BrewStore and
// IngredientStore might be the same struct behind the scenes.

// BrewHandler serves brew-related API endpoints — the core game interaction.
type BrewHandler struct {
	brews       store.BrewStore
	ingredients store.IngredientStore
	engine      *brewing.BrewEngine
}

// NewBrewHandler creates a handler with stores and the brew engine.
func NewBrewHandler(
	brews store.BrewStore,
	ingredients store.IngredientStore,
	engine *brewing.BrewEngine,
) *BrewHandler {
	return &BrewHandler{
		brews:       brews,
		ingredients: ingredients,
		engine:      engine,
	}
}

// ── Request / Response Types ────────────────────────────────────────
//
// GO CONCEPT: Endpoint-Specific Types
// Rather than reusing domain models for API input/output, we define types
// specific to each endpoint. The request type (brewRequest) describes what
// the client sends. The response type (brewResponse) describes what we return.
//
// This decouples the API contract from internal models: if we change how
// Brew is stored internally, the API response shape stays stable. This is
// especially important for discoveredEffects — the stored Brew only has
// effect IDs ([]string), but the API response includes full Effect objects
// so the frontend can update the codex without a second API call.

// brewRequest is the JSON body for POST /api/brew and POST /api/brew/check.
type brewRequest struct {
	IngredientIDs []string `json:"ingredientIds"`
}

// brewResponse is the rich response for POST /api/brew containing everything
// the frontend needs: the stored brew record, full effect objects for codex
// updates, and interaction details for the brew animation.
type brewResponse struct {
	Brew              model.Brew          `json:"brew"`
	DiscoveredEffects []model.Effect      `json:"discoveredEffects"`
	Interactions      []model.Interaction `json:"interactions"`
	Synergies         int                 `json:"synergies"`
	Volatiles         int                 `json:"volatiles"`
}

// ── Handlers ────────────────────────────────────────────────────────

// Brew performs a full brewing attempt. This is the main game action:
// the player submits ingredient IDs, the engine evaluates the combination,
// and the response reveals which codex effects were discovered.
//
// POST /api/brew
// Request:  {"ingredientIds": ["ing-solar-crystal", "ing-void-moss"]}
// Response: 201 Created (success/partial) or 200 OK (catastrophic failure)
//
// The response includes:
//   - brew: the stored brew record (ID, result, potion name, effect IDs)
//   - discoveredEffects: full Effect objects for direct codex updates
//   - interactions: element interactions between the ingredients
//   - synergies/volatiles: counts for the brew animation
func (h *BrewHandler) Brew(w http.ResponseWriter, r *http.Request) {
	var req brewRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	if len(req.IngredientIDs) == 0 {
		writeError(w, http.StatusBadRequest, "ingredientIds is required")
		return
	}

	// Look up all ingredients from the store.
	// GO CONCEPT: Separation of Concerns
	// The handler is responsible for translating between HTTP (IDs in JSON)
	// and domain objects (Ingredient structs). The brew engine only knows
	// about domain objects — it never sees an HTTP request or a JSON string.
	// This three-layer flow is: HTTP → Handler (translation) → Engine (logic).
	ingredients, err := h.resolveIngredients(req.IngredientIDs)
	if err != nil {
		mapError(w, err, "ingredient")
		return
	}

	// Run the brew engine — this is where the magic happens.
	outcome, err := h.engine.Brew(ingredients)
	if err != nil {
		mapError(w, err, "brew")
		return
	}

	// Extract effect IDs for the stored brew record.
	// The Brew model stores IDs (compact, stable), while the API response
	// includes full Effect objects (rich, frontend-friendly).
	effectIDs := make([]string, 0, len(outcome.Effects))
	for _, eff := range outcome.Effects {
		effectIDs = append(effectIDs, eff.ID)
	}

	// Build and save the brew record.
	brew := model.Brew{
		ID:              generateBrewID(),
		RecipeID:        nil, // experimental brew — no recipe followed
		IngredientsUsed: req.IngredientIDs,
		Result:          outcome.Result,
		PotionName:      outcome.PotionName,
		Effects:         effectIDs,
		Notes:           outcome.Notes,
		BrewedAt:        time.Now().UTC(),
	}

	if err := h.brews.SaveBrew(brew); err != nil {
		mapError(w, err, "brew")
		return
	}

	// Choose status code: 201 Created for successful/partial brews,
	// 200 OK for catastrophic failures (the request was valid, but the
	// potion exploded — that's a game result, not an error).
	status := http.StatusCreated
	if outcome.Result == model.BrewResultCatastrophicFailure {
		status = http.StatusOK
	}

	writeJSON(w, status, brewResponse{
		Brew:              brew,
		DiscoveredEffects: outcome.Effects,
		Interactions:      outcome.Interactions,
		Synergies:         outcome.Synergies,
		Volatiles:         outcome.Volatiles,
	})
}

// Check previews ingredient compatibility without actually brewing.
// This lets the player see synergies and volatile warnings before committing.
//
// POST /api/brew/check
// Request:  {"ingredientIds": ["ing-solar-crystal", "ing-void-moss"]}
// Response: 200 OK with compatibility report (interactions, success chance)
func (h *BrewHandler) Check(w http.ResponseWriter, r *http.Request) {
	var req brewRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	if len(req.IngredientIDs) == 0 {
		writeError(w, http.StatusBadRequest, "ingredientIds is required")
		return
	}

	ingredients, err := h.resolveIngredients(req.IngredientIDs)
	if err != nil {
		mapError(w, err, "ingredient")
		return
	}

	report, err := h.engine.CheckCompatibility(ingredients)
	if err != nil {
		mapError(w, err, "compatibility check")
		return
	}

	writeJSON(w, http.StatusOK, report)
}

// List returns the brew history in reverse chronological order (newest first).
// GET /api/brews
func (h *BrewHandler) List(w http.ResponseWriter, r *http.Request) {
	brews, err := h.brews.ListBrews()
	if err != nil {
		mapError(w, err, "brews")
		return
	}
	writeJSON(w, http.StatusOK, brews)
}

// ── Helpers ─────────────────────────────────────────────────────────

// resolveIngredients looks up Ingredient objects from the store by their IDs.
// Returns store.ErrNotFound if any ID doesn't match a known ingredient.
//
// GO CONCEPT: Pointer Dereference (*)
// GetIngredient returns *model.Ingredient (a pointer). We dereference it with
// *ing to get the value and append it to the slice. The pointer exists because
// GetIngredient might return nil when the ingredient isn't found — the error
// return tells us whether it's safe to dereference.
func (h *BrewHandler) resolveIngredients(ids []string) ([]model.Ingredient, error) {
	ingredients := make([]model.Ingredient, 0, len(ids))
	for _, id := range ids {
		ing, err := h.ingredients.GetIngredient(id)
		if err != nil {
			return nil, err
		}
		ingredients = append(ingredients, *ing)
	}
	return ingredients, nil
}

// generateBrewID creates a unique identifier for a brew record.
//
// GO CONCEPT: crypto/rand vs math/rand
// crypto/rand reads from the OS's cryptographic random source (/dev/urandom
// on Linux, CryptGenRandom on Windows). It's slower than math/rand but
// produces unpredictable values — important for IDs that might appear in URLs.
// For game logic randomness (brew outcome rolls), math/rand is fine. For
// identifiers, crypto/rand prevents guessable IDs.
func generateBrewID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return fmt.Sprintf("brew-%x", b)
}
