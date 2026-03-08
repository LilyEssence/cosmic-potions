package brewing

import (
	"errors"
	"fmt"
	"math/rand"
	"sort"
	"time"

	"github.com/LilyEssence/cosmic-potions/internal/model"
)

// ── Sentinel Errors ─────────────────────────────────────────────────

// ErrTooFewIngredients is returned when fewer than 2 ingredients are provided.
var ErrTooFewIngredients = errors.New("brewing requires at least 2 ingredients")

// ErrTooManyIngredients is returned when more than 4 ingredients are provided.
var ErrTooManyIngredients = errors.New("brewing accepts at most 4 ingredients")

// ── Randomizer Interface ────────────────────────────────────────────
//
// GO CONCEPT: Dependency Injection via Interface
// The Randomizer interface lets us control randomness in the brewing engine.
// In production, we use a real random source. In tests, we inject a fake that
// returns predetermined values — making brew outcomes deterministic and testable.
//
// This is Go's approach to dependency injection: no framework, no container,
// just an interface. The caller decides what implementation to provide.
//
// Compare to TypeScript: instead of `jest.spyOn(Math, 'random')`, you design
// randomness as a dependency from the start. It's more explicit but avoids the
// "monkey patching globals in tests" pattern.

// Randomizer abstracts random number generation so the brew engine can be
// tested with deterministic outcomes.
type Randomizer interface {
	// Float64 returns a random float in [0.0, 1.0).
	Float64() float64
	// Intn returns a random int in [0, n).
	Intn(n int) int
}

// DefaultRandomizer wraps Go's math/rand for production use.
//
// GO CONCEPT: Wrapping Standard Library Types
// Rather than using math/rand globally (which would be impossible to control
// in tests), we wrap it in a struct that satisfies our Randomizer interface.
// This is a common Go pattern: define a small interface, then wrap existing
// functionality to match it.
type DefaultRandomizer struct {
	rng *rand.Rand
}

// NewDefaultRandomizer creates a Randomizer seeded with the current time.
func NewDefaultRandomizer() *DefaultRandomizer {
	return &DefaultRandomizer{
		rng: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (r *DefaultRandomizer) Float64() float64 { return r.rng.Float64() }
func (r *DefaultRandomizer) Intn(n int) int   { return r.rng.Intn(n) }

// ── Output Types ────────────────────────────────────────────────────

// CompatibilityReport is the result of checking ingredient compatibility
// without actually brewing. Used by the /api/brew/check preview endpoint.
type CompatibilityReport struct {
	Interactions  []model.Interaction `json:"interactions"`
	Synergies     int                 `json:"synergies"`
	Volatiles     int                 `json:"volatiles"`
	Neutrals      int                 `json:"neutrals"`
	SuccessChance float64             `json:"successChance"`
}

// BrewOutcome is the full result of a brewing attempt. Contains everything
// the handler needs to build an API response and the frontend needs to
// update the codex.
type BrewOutcome struct {
	Result       model.BrewResult    `json:"result"`
	PotionName   string              `json:"potionName"`
	Effects      []model.Effect      `json:"effects"`
	Interactions []model.Interaction `json:"interactions"`
	Synergies    int                 `json:"synergies"`
	Volatiles    int                 `json:"volatiles"`
	Notes        string              `json:"notes"`
}

// ── BrewEngine ──────────────────────────────────────────────────────
//
// GO CONCEPT: Pure Business Logic Package
// BrewEngine has no I/O dependencies — it doesn't read from the store, doesn't
// write HTTP responses, doesn't touch the filesystem. It receives data (ingredients,
// interaction matrix) and returns results. This makes it:
//   - Easy to test (no mocks for DB or HTTP)
//   - Easy to reason about (input → output, no side effects)
//   - Reusable (a CLI tool could use the same engine)
//
// The handler layer is responsible for fetching ingredients from the store and
// passing them to the engine. The engine doesn't know stores exist.

// BrewEngine evaluates ingredient combinations and produces brew outcomes.
// It holds the element interaction matrix, effect definitions, and discovery
// rules that determine the Codex game mechanics.
type BrewEngine struct {
	interactions map[model.Element]map[model.Element]model.InteractionResult
	effectMap    map[string]model.Effect
	rules        []effectRule
	rand         Randomizer
}

// NewBrewEngine creates a brew engine configured with game data.
//
// GO CONCEPT: Accept Interfaces, Return Structs
// The constructor accepts a Randomizer interface (flexible — tests can inject
// fakes) but returns a concrete *BrewEngine (callers get the real type and can
// choose which interface to view it through).
//
// The interactions map and effects slice typically come from the seed package:
//
//	engine := brewing.NewBrewEngine(seed.Interactions(), seed.Effects(), randomizer)
func NewBrewEngine(
	interactions map[model.Element]map[model.Element]model.InteractionResult,
	effects []model.Effect,
	rand Randomizer,
) *BrewEngine {
	em := make(map[string]model.Effect, len(effects))
	for _, e := range effects {
		em[e.ID] = e
	}
	return &BrewEngine{
		interactions: interactions,
		effectMap:    em,
		rules:        defaultRules,
		rand:         rand,
	}
}

// ── Public Methods ──────────────────────────────────────────────────

// CheckCompatibility analyzes ingredient elements and reports how they interact.
// This powers the /api/brew/check preview — the player can see synergies and
// volatile warnings before committing to a brew.
//
// Returns ErrTooFewIngredients or ErrTooManyIngredients for invalid counts.
func (e *BrewEngine) CheckCompatibility(ingredients []model.Ingredient) (*CompatibilityReport, error) {
	if err := validateIngredientCount(len(ingredients)); err != nil {
		return nil, err
	}

	interactions := e.findInteractions(ingredients)

	synergies, volatiles, neutrals := countInteractionTypes(interactions)
	chance := successChance(synergies, volatiles)

	return &CompatibilityReport{
		Interactions:  interactions,
		Synergies:     synergies,
		Volatiles:     volatiles,
		Neutrals:      neutrals,
		SuccessChance: chance,
	}, nil
}

// CalculateEffects determines which codex effects a set of ingredients could
// discover, given their interactions. This is deterministic — the same inputs
// always produce the same eligible effects. It does NOT consider brew outcome
// (success/partial/failure); the Brew method handles that filtering.
//
// The handler can call this alongside CheckCompatibility to show the player
// "if you brew this and succeed, you could discover these effects."
func (e *BrewEngine) CalculateEffects(ingredients []model.Ingredient, interactions []model.Interaction) []model.Effect {
	elements := uniqueElements(ingredients)
	potency := totalPotency(ingredients)
	numIngredients := len(ingredients)

	hasSynergy := false
	hasVolatile := false
	for _, inter := range interactions {
		if inter.Result == model.InteractionSynergy {
			hasSynergy = true
		}
		if inter.Result == model.InteractionVolatile {
			hasVolatile = true
		}
	}

	var matched []model.Effect
	for _, rule := range e.rules {
		if !rule.matches(elements, potency, numIngredients, hasSynergy, hasVolatile) {
			continue
		}
		if eff, ok := e.effectMap[rule.effectID]; ok {
			matched = append(matched, eff)
		}
	}

	// Sort by tier order (common → legendary), then by name within tier.
	// GO CONCEPT: Custom Sort with Multiple Keys
	// sort.Slice lets you define arbitrary ordering. Here we sort by tier
	// first (using a numeric rank), then alphabetically within the same tier.
	sort.Slice(matched, func(i, j int) bool {
		ri, rj := tierRank(matched[i].Tier), tierRank(matched[j].Tier)
		if ri != rj {
			return ri < rj
		}
		return matched[i].Name < matched[j].Name
	})

	return matched
}

// Brew performs a full brewing attempt: checks compatibility, rolls for outcome,
// determines which effects are discovered, and generates a potion name.
//
// This is the main entry point for the POST /api/brew endpoint. It orchestrates
// the full game loop in a single call:
//  1. Find element interactions between all ingredient pairs
//  2. Calculate success probability (synergies boost, volatiles reduce)
//  3. Roll for outcome: success, partial, or catastrophic_failure
//  4. Determine eligible effects based on ingredients
//  5. Filter effects by outcome tier (partials lose rare/legendary)
//  6. Generate a thematic potion name and flavor notes
//
// Returns ErrTooFewIngredients or ErrTooManyIngredients for invalid counts.
func (e *BrewEngine) Brew(ingredients []model.Ingredient) (*BrewOutcome, error) {
	if err := validateIngredientCount(len(ingredients)); err != nil {
		return nil, err
	}

	// Step 1: Compatibility analysis
	interactions := e.findInteractions(ingredients)
	synergies, volatiles, _ := countInteractionTypes(interactions)

	// Step 2-3: Calculate odds and roll
	chance := successChance(synergies, volatiles)
	result := e.rollOutcome(chance)

	// Step 4: Determine all eligible effects
	allEffects := e.CalculateEffects(ingredients, interactions)

	// Step 5: Filter by brew outcome
	//   - Success: all effects discovered
	//   - Partial: only common + uncommon (partial brews can't unlock rare/legendary)
	//   - Catastrophic failure: nothing discovered (but entertaining)
	discoveredEffects := filterEffectsByOutcome(allEffects, result)

	// Step 6: Generate name and notes
	dominant := dominantElement(ingredients)
	name := generatePotionName(dominant, result, e.rand)
	notes := generateNotes(result, synergies, volatiles, e.rand)

	return &BrewOutcome{
		Result:       result,
		PotionName:   name,
		Effects:      discoveredEffects,
		Interactions: interactions,
		Synergies:    synergies,
		Volatiles:    volatiles,
		Notes:        notes,
	}, nil
}

// ── Internal Helpers ────────────────────────────────────────────────

// findInteractions checks every pair of ingredients and looks up their element
// interaction in the matrix.
//
// GO CONCEPT: Nested Loops for Combinations
// To get all unique pairs from N items, use two loops where j starts at i+1.
// This produces N*(N-1)/2 pairs without duplicates or self-pairs:
//   - 2 ingredients → 1 pair
//   - 3 ingredients → 3 pairs
//   - 4 ingredients → 6 pairs
func (e *BrewEngine) findInteractions(ingredients []model.Ingredient) []model.Interaction {
	var interactions []model.Interaction
	for i := 0; i < len(ingredients); i++ {
		for j := i + 1; j < len(ingredients); j++ {
			elem1 := ingredients[i].Element
			elem2 := ingredients[j].Element

			result := model.InteractionNeutral // default if not in matrix
			if inner, ok := e.interactions[elem1]; ok {
				if r, ok := inner[elem2]; ok {
					result = r
				}
			}

			interactions = append(interactions, model.Interaction{
				Element1: elem1,
				Element2: elem2,
				Result:   result,
			})
		}
	}
	return interactions
}

// successChance calculates the probability of a successful brew based on
// the number of synergy and volatile interactions.
//
// Base success rate is 65%. Each synergy adds 10%, each volatile subtracts 15%.
// Clamped to [5%, 95%] — there's always a small chance of success or failure.
func successChance(synergies, volatiles int) float64 {
	chance := 0.65 + float64(synergies)*0.10 - float64(volatiles)*0.15
	if chance > 0.95 {
		chance = 0.95
	}
	if chance < 0.05 {
		chance = 0.05
	}
	return chance
}

// rollOutcome uses the randomizer to determine the brew result.
//
// The outcome bands are:
//
//	[0, successChance)                → Success
//	[successChance, successChance+20) → Partial
//	[successChance+20, 1.0)           → Catastrophic Failure
//
// With the default 65% success chance: 65% success, 20% partial, 15% catastrophic.
// With 2 synergies (85%): 85% success, 15% partial, 0% catastrophic.
// With 2 volatiles (35%): 35% success, 20% partial, 45% catastrophic.
func (e *BrewEngine) rollOutcome(chance float64) model.BrewResult {
	roll := e.rand.Float64()
	switch {
	case roll < chance:
		return model.BrewResultSuccess
	case roll < chance+0.20:
		return model.BrewResultPartial
	default:
		return model.BrewResultCatastrophicFailure
	}
}

// filterEffectsByOutcome removes effects the player doesn't get to discover
// based on how the brew turned out.
func filterEffectsByOutcome(effects []model.Effect, result model.BrewResult) []model.Effect {
	switch result {
	case model.BrewResultCatastrophicFailure:
		// Nothing discovered — the cauldron exploded
		return make([]model.Effect, 0)
	case model.BrewResultPartial:
		// Only common and uncommon effects survive a partial brew.
		// Rare and legendary require clean execution.
		filtered := make([]model.Effect, 0)
		for _, eff := range effects {
			if eff.Tier == model.EffectTierCommon || eff.Tier == model.EffectTierUncommon {
				filtered = append(filtered, eff)
			}
		}
		return filtered
	default:
		// Full success — all effects discovered
		return effects
	}
}

// validateIngredientCount checks that the ingredient count is within game bounds.
func validateIngredientCount(n int) error {
	if n < 2 {
		return fmt.Errorf("%w: got %d", ErrTooFewIngredients, n)
	}
	if n > 4 {
		return fmt.Errorf("%w: got %d", ErrTooManyIngredients, n)
	}
	return nil
}

// countInteractionTypes tallies synergy, volatile, and neutral interactions.
func countInteractionTypes(interactions []model.Interaction) (synergies, volatiles, neutrals int) {
	for _, inter := range interactions {
		switch inter.Result {
		case model.InteractionSynergy:
			synergies++
		case model.InteractionVolatile:
			volatiles++
		default:
			neutrals++
		}
	}
	return
}

// uniqueElements extracts the distinct elements present in a set of ingredients.
func uniqueElements(ingredients []model.Ingredient) map[model.Element]bool {
	elems := make(map[model.Element]bool)
	for _, ing := range ingredients {
		elems[ing.Element] = true
	}
	return elems
}

// totalPotency sums the potency values of all ingredients.
func totalPotency(ingredients []model.Ingredient) int {
	sum := 0
	for _, ing := range ingredients {
		sum += ing.Potency
	}
	return sum
}

// dominantElement finds the element with the most combined potency.
// Ties are broken by element name (deterministic ordering).
func dominantElement(ingredients []model.Ingredient) model.Element {
	potencyByElement := make(map[model.Element]int)
	for _, ing := range ingredients {
		potencyByElement[ing.Element] += ing.Potency
	}

	// Iterate in deterministic order to break ties consistently.
	var best model.Element
	bestPotency := 0
	for _, elem := range []model.Element{
		model.ElementCrystalline, model.ElementOrganic,
		model.ElementPlasma, model.ElementSolar, model.ElementVoid,
	} {
		if potencyByElement[elem] > bestPotency {
			bestPotency = potencyByElement[elem]
			best = elem
		}
	}
	return best
}

// tierRank converts an EffectTier to a numeric rank for sorting.
func tierRank(tier model.EffectTier) int {
	switch tier {
	case model.EffectTierCommon:
		return 0
	case model.EffectTierUncommon:
		return 1
	case model.EffectTierRare:
		return 2
	case model.EffectTierLegendary:
		return 3
	default:
		return 99
	}
}
