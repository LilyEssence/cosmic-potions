package brewing

import "github.com/LilyEssence/cosmic-potions/internal/model"

// ── Effect Discovery Rules ──────────────────────────────────────────
//
// GO CONCEPT: Unexported Types
// The effectRule type is lowercase — unexported, visible only within the
// brewing package. External code (handlers, tests) interacts with the engine's
// public methods, not the rules directly. This is encapsulation via naming.
//
// Each rule defines the conditions under which a specific codex effect can
// be discovered. The engine evaluates all rules against the current brew's
// ingredients and interactions, then returns matching effects.
//
// Rule design philosophy for the Codex game:
//   - Common: single element + low potency → found in first few brews
//   - Uncommon: two elements or 3+ ingredients → rewards cross-element combos
//   - Rare: synergy interactions + high potency → rewards understanding the matrix
//   - Legendary: volatile interactions + high potency → rewards braving risk
//
// Element interaction reference (from seed data):
//   Synergy pairs:  solar+crystalline, solar+organic, void+plasma, organic+plasma
//   Volatile pairs: solar+void, void+organic, crystalline+plasma
//   All same-element pairs and remaining combos are neutral.

// effectRule defines when a specific codex effect can be discovered.
type effectRule struct {
	effectID          string          // links to model.Effect.ID
	requiredElements  []model.Element // ALL must be present in the brew
	minPotency        int             // minimum combined potency of all ingredients
	minIngredients    int             // minimum number of ingredients (0 = no minimum)
	minUniqueElements int             // minimum distinct elements (0 = no minimum)
	requiresSynergy   bool            // at least one synergy interaction must exist
	requiresVolatile  bool            // at least one volatile interaction must exist
}

// matches checks whether a brew meets this rule's conditions.
//
// GO CONCEPT: Method on Unexported Type
// Even though effectRule is unexported, it can have methods. These methods
// are only callable within the brewing package. This keeps the rule-matching
// logic close to the rule definition — good cohesion.
func (r *effectRule) matches(
	elements map[model.Element]bool,
	potency int,
	numIngredients int,
	hasSynergy bool,
	hasVolatile bool,
) bool {
	// Check all required elements are present
	for _, elem := range r.requiredElements {
		if !elements[elem] {
			return false
		}
	}

	// Check minimum potency
	if potency < r.minPotency {
		return false
	}

	// Check minimum ingredient count
	if r.minIngredients > 0 && numIngredients < r.minIngredients {
		return false
	}

	// Check minimum unique elements
	if r.minUniqueElements > 0 && len(elements) < r.minUniqueElements {
		return false
	}

	// Check synergy requirement
	if r.requiresSynergy && !hasSynergy {
		return false
	}

	// Check volatile requirement
	if r.requiresVolatile && !hasVolatile {
		return false
	}

	return true
}

// defaultRules defines discovery conditions for all 28 codex effects.
//
// GO CONCEPT: Package-Level Variables
// This is a package-level var (not const, because slices can't be const in Go).
// It's initialized once and never modified. The BrewEngine references it but
// doesn't mutate it, so no mutex is needed.
//
// Using a var here (instead of a function returning a fresh slice like seed data)
// is fine because the engine never writes to it. Seed data uses functions to
// prevent mutation because multiple callers might use it; here, only the engine
// reads it, and it's truly immutable by convention.
var defaultRules = []effectRule{
	// ── Common Effects ──────────────────────────────────────────
	// Low barrier: single element, modest potency. Players discover
	// several of these in their very first brew.
	{
		effectID:         "effect-night-vision",
		requiredElements: []model.Element{model.ElementSolar},
		minPotency:       5,
	},
	{
		effectID:         "effect-mild-euphoria",
		requiredElements: []model.Element{model.ElementOrganic},
		minPotency:       3,
	},
	{
		effectID:         "effect-bioluminescence",
		requiredElements: []model.Element{model.ElementOrganic},
		minPotency:       6,
	},
	{
		effectID:         "effect-enhanced-perception",
		requiredElements: []model.Element{model.ElementCrystalline},
		minPotency:       6,
	},
	{
		effectID:         "effect-inner-calm",
		requiredElements: []model.Element{model.ElementOrganic},
		minPotency:       8,
	},
	{
		effectID:         "effect-mild-warmth",
		requiredElements: []model.Element{model.ElementSolar},
		minPotency:       4,
	},
	{
		effectID:         "effect-skin-regeneration",
		requiredElements: []model.Element{model.ElementOrganic},
		minPotency:       7,
	},
	{
		effectID:         "effect-enhanced-focus",
		requiredElements: []model.Element{model.ElementCrystalline},
		minPotency:       5,
	},
	{
		effectID:         "effect-energy-boost",
		requiredElements: []model.Element{model.ElementPlasma},
		minPotency:       5,
	},
	{
		effectID:         "effect-basic-shielding",
		requiredElements: []model.Element{model.ElementCrystalline},
		minPotency:       7,
	},

	// ── Uncommon Effects ────────────────────────────────────────
	// Require two specific elements or 3+ ingredients. This nudges
	// players toward cross-planet ingredient mixing.
	{
		effectID:         "effect-cosmic-hearing",
		requiredElements: []model.Element{model.ElementVoid, model.ElementCrystalline},
		minPotency:       10,
	},
	{
		effectID:         "effect-zero-gravity-tolerance",
		requiredElements: []model.Element{model.ElementVoid},
		minIngredients:   3,
		minPotency:       10,
	},
	{
		effectID:         "effect-pressure-resistance",
		requiredElements: []model.Element{model.ElementOrganic, model.ElementPlasma},
		minPotency:       10,
	},
	{
		effectID:         "effect-gravity-resistance",
		requiredElements: []model.Element{model.ElementOrganic, model.ElementCrystalline},
		minPotency:       9,
	},
	{
		effectID:         "effect-bone-density-boost",
		requiredElements: []model.Element{model.ElementCrystalline, model.ElementPlasma},
		minPotency:       10,
	},
	{
		effectID:         "effect-harmonic-sensitivity",
		requiredElements: []model.Element{model.ElementCrystalline},
		minIngredients:   3,
		minPotency:       9,
	},
	{
		effectID:         "effect-time-dilation",
		requiredElements: []model.Element{model.ElementVoid, model.ElementOrganic},
		minPotency:       10,
	},
	{
		effectID:         "effect-temporary-translucency",
		requiredElements: []model.Element{model.ElementVoid, model.ElementCrystalline},
		minPotency:       11,
	},

	// ── Rare Effects ────────────────────────────────────────────
	// Require synergy interactions + high combined potency. The
	// required element pairs all naturally produce synergy:
	//   solar+crystalline, void+plasma, organic+plasma, solar+organic
	{
		effectID:         "effect-structural-sight",
		requiredElements: []model.Element{model.ElementSolar, model.ElementCrystalline},
		requiresSynergy:  true,
		minPotency:       14,
	},
	{
		effectID:         "effect-electrical-manipulation",
		requiredElements: []model.Element{model.ElementPlasma, model.ElementVoid},
		requiresSynergy:  true,
		minIngredients:   3,
		minPotency:       16,
	},
	{
		effectID:         "effect-plasma-generation",
		requiredElements: []model.Element{model.ElementPlasma, model.ElementOrganic},
		requiresSynergy:  true,
		minPotency:       13,
	},
	{
		effectID:         "effect-heat-invulnerability",
		requiredElements: []model.Element{model.ElementSolar, model.ElementOrganic},
		requiresSynergy:  true,
		minPotency:       18,
	},
	{
		effectID:         "effect-entropy-reversal",
		requiredElements: []model.Element{model.ElementVoid, model.ElementPlasma},
		requiresSynergy:  true,
		minIngredients:   3,
		minPotency:       18,
	},
	{
		effectID:         "effect-telepathic-link",
		requiredElements: []model.Element{model.ElementOrganic, model.ElementSolar},
		requiresSynergy:  true,
		minIngredients:   3,
		minPotency:       16,
	},

	// ── Legendary Effects ───────────────────────────────────────
	// Require volatile interactions + high potency. The required
	// element pairs naturally produce volatile interactions:
	//   solar+void, crystalline+plasma, void+organic
	//
	// This creates a risk/reward dynamic: volatile pairs reduce
	// success chance, but successful volatile brews unlock effects
	// that safe brews never can.
	{
		effectID:         "effect-temporal-echo",
		requiredElements: []model.Element{model.ElementSolar, model.ElementVoid},
		requiresVolatile: true,
		minPotency:       14,
	},
	{
		effectID:          "effect-cosmic-omniscience",
		minUniqueElements: 4,
		requiresVolatile:  true,
		minPotency:        20,
	},
	{
		effectID:         "effect-stellar-rebirth",
		requiredElements: []model.Element{model.ElementPlasma, model.ElementCrystalline},
		requiresVolatile: true,
		minPotency:       18,
	},
	{
		effectID:         "effect-void-walking",
		requiredElements: []model.Element{model.ElementVoid, model.ElementOrganic},
		requiresVolatile: true,
		minPotency:       16,
	},
}
