package handler

import (
	"net/http"

	"github.com/LilyEssence/cosmic-potions/internal/model"
)

// ── Effect Handlers ─────────────────────────────────────────────────
//
// The effects endpoint powers the Codex — it tells the frontend what "100%
// completion" looks like. The response includes every discoverable effect's
// ID, name, tier, and description. It does NOT include hints about which
// ingredients produce them — that's for the player to discover.

// EffectHandler serves the master effect list for the codex.
type EffectHandler struct {
	effects []model.Effect
}

// NewEffectHandler creates a handler with the full effect list.
// The effects slice is typically loaded from seed.Effects() at startup
// and shared read-only — no mutex needed since slices are safe for
// concurrent reads when no goroutine is writing.
func NewEffectHandler(effects []model.Effect) *EffectHandler {
	return &EffectHandler{effects: effects}
}

// List returns all discoverable effects for the Cosmic Alchemist's Codex.
// GET /api/effects
//
// The frontend uses this to render the codex grid. Each effect starts as
// "???" and is revealed when the player brews a potion that produces it.
// The tier field lets the frontend show difficulty hints even for undiscovered
// effects (e.g., a legendary silhouette looks different from a common one).
func (h *EffectHandler) List(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, h.effects)
}
