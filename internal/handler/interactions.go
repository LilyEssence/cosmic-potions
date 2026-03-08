package handler

import (
	"net/http"
	"sort"

	"github.com/LilyEssence/cosmic-potions/internal/model"
)

// ── Interaction Handlers ────────────────────────────────────────────
//
// The interactions endpoint returns the element compatibility matrix — the
// foundation of the brewing system. Players use this to plan ingredient
// combinations: synergies boost success, volatiles risk catastrophe.

// InteractionHandler serves the element interaction matrix.
type InteractionHandler struct {
	interactions []model.Interaction
}

// NewInteractionHandler creates a handler with the interaction data.
// The interactions are sorted once at construction time for deterministic
// API output — since the seed data comes from a map (random iteration order),
// we sort here to ensure the same response every time.
func NewInteractionHandler(interactions []model.Interaction) *InteractionHandler {
	// Sort for deterministic output: by Element1, then Element2.
	// GO CONCEPT: Sort at Construction Time
	// Rather than sorting on every request, we sort once when the handler is
	// created. The interactions slice is read-only after construction, so this
	// is both safe and efficient. This is a common Go pattern: do expensive
	// setup work once in the constructor, then serve cheap reads.
	sorted := make([]model.Interaction, len(interactions))
	copy(sorted, interactions)

	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Element1 != sorted[j].Element1 {
			return sorted[i].Element1 < sorted[j].Element1
		}
		return sorted[i].Element2 < sorted[j].Element2
	})

	return &InteractionHandler{interactions: sorted}
}

// List returns the full element interaction matrix.
// GET /api/interactions
//
// The response is a flat array of {element1, element2, result} objects
// covering all 25 pairs (5×5 matrix). The frontend can use this to render
// an interaction chart and to show compatibility hints in the brew station.
func (h *InteractionHandler) List(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, h.interactions)
}
