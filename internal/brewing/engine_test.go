package brewing

import (
	"errors"
	"testing"

	"github.com/LilyEssence/cosmic-potions/internal/model"
	"github.com/LilyEssence/cosmic-potions/seed"
)

// ── Test Helpers ────────────────────────────────────────────────────
//
// GO CONCEPT: Test Files
// Go test files live alongside the code they test, in the same package,
// with the suffix _test.go. The testing package is part of the standard
// library — no external framework like Jest or Mocha needed.
//
// Tests are functions named TestXxx(t *testing.T). The `go test` command
// finds and runs them automatically.

// fakeRand is a deterministic Randomizer for testing. It returns
// pre-configured values so brew outcomes are predictable and repeatable.
//
// GO CONCEPT: Test Doubles (Fakes)
// This is why we created the Randomizer interface — in tests, we inject
// this fake that returns exactly what we tell it to. No monkey-patching
// globals, no jest.spyOn(Math, 'random'). The dependency is explicitly
// controlled.
type fakeRand struct {
	floatVal float64 // value returned by Float64()
	intVal   int     // value returned by Intn()
}

func (f *fakeRand) Float64() float64 { return f.floatVal }
func (f *fakeRand) Intn(n int) int {
	if n == 0 {
		return 0
	}
	return f.intVal % n
}

// newTestEngine creates a BrewEngine with seed data and a fake randomizer.
// The fake returns floatVal for brew outcome rolls and intVal for name selection.
func newTestEngine(floatVal float64, intVal int) *BrewEngine {
	return NewBrewEngine(
		seed.Interactions(),
		seed.Effects(),
		&fakeRand{floatVal: floatVal, intVal: intVal},
	)
}

// ingredient helpers — build test ingredients without verbose struct literals.
func solarIngredient(potency int) model.Ingredient {
	return model.Ingredient{ID: "test-solar", Element: model.ElementSolar, Potency: potency}
}
func voidIngredient(potency int) model.Ingredient {
	return model.Ingredient{ID: "test-void", Element: model.ElementVoid, Potency: potency}
}
func crystallineIngredient(potency int) model.Ingredient {
	return model.Ingredient{ID: "test-crystalline", Element: model.ElementCrystalline, Potency: potency}
}
func organicIngredient(potency int) model.Ingredient {
	return model.Ingredient{ID: "test-organic", Element: model.ElementOrganic, Potency: potency}
}
func plasmaIngredient(potency int) model.Ingredient {
	return model.Ingredient{ID: "test-plasma", Element: model.ElementPlasma, Potency: potency}
}

// ── Validation Tests ────────────────────────────────────────────────

// GO CONCEPT: Table-Driven Tests
// The most common Go test pattern: define a slice of test cases (each with
// a name, inputs, and expected outputs), then loop through them calling
// t.Run(name, ...) for each. This gives you:
//   - One test function covering many cases
//   - Clear names in test output (TestValidation/too_few_ingredients)
//   - Easy to add new cases (just append a struct)
//
// Compare to Jest's describe/it blocks — similar intent, different syntax.

func TestValidation(t *testing.T) {
	engine := newTestEngine(0.0, 0)

	tests := []struct {
		name        string
		count       int
		wantErr     error
		ingredients []model.Ingredient
	}{
		{
			name:        "zero ingredients",
			count:       0,
			wantErr:     ErrTooFewIngredients,
			ingredients: []model.Ingredient{},
		},
		{
			name:        "one ingredient",
			count:       1,
			wantErr:     ErrTooFewIngredients,
			ingredients: []model.Ingredient{solarIngredient(5)},
		},
		{
			name:  "two ingredients (minimum)",
			count: 2,
			ingredients: []model.Ingredient{
				solarIngredient(5), voidIngredient(5),
			},
		},
		{
			name:  "four ingredients (maximum)",
			count: 4,
			ingredients: []model.Ingredient{
				solarIngredient(5), voidIngredient(5),
				crystallineIngredient(5), organicIngredient(5),
			},
		},
		{
			name:    "five ingredients",
			count:   5,
			wantErr: ErrTooManyIngredients,
			ingredients: []model.Ingredient{
				solarIngredient(5), voidIngredient(5),
				crystallineIngredient(5), organicIngredient(5),
				plasmaIngredient(5),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := engine.CheckCompatibility(tt.ingredients)
			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tt.wantErr)
				}
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestBrewValidation(t *testing.T) {
	engine := newTestEngine(0.0, 0)

	_, err := engine.Brew([]model.Ingredient{solarIngredient(5)})
	if !errors.Is(err, ErrTooFewIngredients) {
		t.Fatalf("Brew with 1 ingredient: expected ErrTooFewIngredients, got %v", err)
	}

	_, err = engine.Brew([]model.Ingredient{
		solarIngredient(1), voidIngredient(1), crystallineIngredient(1),
		organicIngredient(1), plasmaIngredient(1),
	})
	if !errors.Is(err, ErrTooManyIngredients) {
		t.Fatalf("Brew with 5 ingredients: expected ErrTooManyIngredients, got %v", err)
	}
}

// ── Interaction Detection Tests ─────────────────────────────────────

func TestCheckCompatibility_InteractionMatrix(t *testing.T) {
	// GO CONCEPT: Sub-tests for Organized Output
	// t.Run("name", fn) creates a sub-test. In test output you'll see:
	//   TestCheckCompatibility_InteractionMatrix/solar+crystalline_synergy
	// This makes it easy to find which specific case failed.

	engine := newTestEngine(0.0, 0)

	tests := []struct {
		name         string
		ingredients  []model.Ingredient
		wantSynergy  int
		wantVolatile int
		wantNeutral  int
	}{
		{
			name:         "solar+crystalline synergy",
			ingredients:  []model.Ingredient{solarIngredient(5), crystallineIngredient(5)},
			wantSynergy:  1,
			wantVolatile: 0,
			wantNeutral:  0,
		},
		{
			name:         "solar+organic synergy",
			ingredients:  []model.Ingredient{solarIngredient(5), organicIngredient(5)},
			wantSynergy:  1,
			wantVolatile: 0,
			wantNeutral:  0,
		},
		{
			name:         "void+plasma synergy",
			ingredients:  []model.Ingredient{voidIngredient(5), plasmaIngredient(5)},
			wantSynergy:  1,
			wantVolatile: 0,
			wantNeutral:  0,
		},
		{
			name:         "organic+plasma synergy",
			ingredients:  []model.Ingredient{organicIngredient(5), plasmaIngredient(5)},
			wantSynergy:  1,
			wantVolatile: 0,
			wantNeutral:  0,
		},
		{
			name:         "solar+void volatile",
			ingredients:  []model.Ingredient{solarIngredient(5), voidIngredient(5)},
			wantSynergy:  0,
			wantVolatile: 1,
			wantNeutral:  0,
		},
		{
			name:         "void+organic volatile",
			ingredients:  []model.Ingredient{voidIngredient(5), organicIngredient(5)},
			wantSynergy:  0,
			wantVolatile: 1,
			wantNeutral:  0,
		},
		{
			name:         "crystalline+plasma volatile",
			ingredients:  []model.Ingredient{crystallineIngredient(5), plasmaIngredient(5)},
			wantSynergy:  0,
			wantVolatile: 1,
			wantNeutral:  0,
		},
		{
			name:         "solar+solar neutral (same element)",
			ingredients:  []model.Ingredient{solarIngredient(5), solarIngredient(5)},
			wantSynergy:  0,
			wantVolatile: 0,
			wantNeutral:  1,
		},
		{
			name:         "solar+plasma neutral",
			ingredients:  []model.Ingredient{solarIngredient(5), plasmaIngredient(5)},
			wantSynergy:  0,
			wantVolatile: 0,
			wantNeutral:  1,
		},
		{
			// 3 ingredients: solar+crystalline=synergy, solar+organic=synergy, crystalline+organic=neutral
			name: "three ingredients (2 synergies, 1 neutral)",
			ingredients: []model.Ingredient{
				solarIngredient(5), crystallineIngredient(5), organicIngredient(5),
			},
			wantSynergy:  2,
			wantVolatile: 0,
			wantNeutral:  1,
		},
		{
			// 4 ingredients: all pairs for solar, void, crystalline, organic
			// solar+void=volatile, solar+crystalline=synergy, solar+organic=synergy
			// void+crystalline=neutral, void+organic=volatile, crystalline+organic=neutral
			name: "four ingredients mixed",
			ingredients: []model.Ingredient{
				solarIngredient(5), voidIngredient(5),
				crystallineIngredient(5), organicIngredient(5),
			},
			wantSynergy:  2,
			wantVolatile: 2,
			wantNeutral:  2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report, err := engine.CheckCompatibility(tt.ingredients)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if report.Synergies != tt.wantSynergy {
				t.Errorf("synergies: got %d, want %d", report.Synergies, tt.wantSynergy)
			}
			if report.Volatiles != tt.wantVolatile {
				t.Errorf("volatiles: got %d, want %d", report.Volatiles, tt.wantVolatile)
			}
			if report.Neutrals != tt.wantNeutral {
				t.Errorf("neutrals: got %d, want %d", report.Neutrals, tt.wantNeutral)
			}
		})
	}
}

// ── Success Chance Tests ────────────────────────────────────────────

func TestSuccessChance(t *testing.T) {
	// GO CONCEPT: Floating-Point Comparison
	// IEEE 754 floating-point arithmetic can produce tiny rounding errors:
	// 0.65 + 0.20 might equal 0.8500000000000001 instead of 0.85. In tests,
	// compare with an epsilon tolerance rather than exact equality. This is
	// the same issue you'd hit in JavaScript (0.1 + 0.2 !== 0.3).
	const epsilon = 1e-9

	tests := []struct {
		name      string
		synergies int
		volatiles int
		want      float64
	}{
		{"baseline (no interactions)", 0, 0, 0.65},
		{"one synergy", 1, 0, 0.75},
		{"two synergies", 2, 0, 0.85},
		{"three synergies (capped at 95%)", 3, 0, 0.95},
		{"one volatile", 0, 1, 0.50},
		{"two volatiles", 0, 2, 0.35},
		{"four volatiles (capped at 5%)", 0, 4, 0.05},
		{"one of each", 1, 1, 0.60},
		{"two synergies one volatile", 2, 1, 0.70},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chance := successChance(tt.synergies, tt.volatiles)
			diff := chance - tt.want
			if diff < -epsilon || diff > epsilon {
				t.Errorf("successChance(%d, %d) = %v, want %v (diff %v)",
					tt.synergies, tt.volatiles, chance, tt.want, diff)
			}
		})
	}
}

// ── Outcome Roll Tests ──────────────────────────────────────────────
//
// These tests verify that the rollOutcome method correctly maps the random
// float to outcome bands. The fakeRand returns a fixed float, so we can
// test exact boundaries.

func TestRollOutcome(t *testing.T) {
	tests := []struct {
		name       string
		chance     float64 // success probability
		roll       float64 // fake random value
		wantResult model.BrewResult
	}{
		{"success (roll below chance)", 0.65, 0.30, model.BrewResultSuccess},
		{"success (roll at zero)", 0.65, 0.00, model.BrewResultSuccess},
		{"success (roll just below chance)", 0.65, 0.6499, model.BrewResultSuccess},
		{"partial (roll at chance boundary)", 0.65, 0.65, model.BrewResultPartial},
		{"partial (roll in partial band)", 0.65, 0.75, model.BrewResultPartial},
		{"catastrophic (roll above chance+0.20)", 0.65, 0.86, model.BrewResultCatastrophicFailure},
		{"catastrophic (roll at 0.99)", 0.65, 0.99, model.BrewResultCatastrophicFailure},
		{"high synergy — success at 0.80", 0.85, 0.80, model.BrewResultSuccess},
		{"high synergy — partial at 0.90", 0.85, 0.90, model.BrewResultPartial},
		{"low chance — catastrophic at 0.60", 0.35, 0.60, model.BrewResultCatastrophicFailure},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := newTestEngine(tt.roll, 0)
			result := engine.rollOutcome(tt.chance)
			if result != tt.wantResult {
				t.Errorf("rollOutcome(chance=%v, roll=%v) = %v, want %v",
					tt.chance, tt.roll, result, tt.wantResult)
			}
		})
	}
}

// ── Effect Calculation Tests ────────────────────────────────────────

func TestCalculateEffects_CommonEffects(t *testing.T) {
	engine := newTestEngine(0.0, 0)

	// Two solar ingredients with potency 5 each = total 10
	// Should match effects requiring solar + minPotency ≤ 10
	ingredients := []model.Ingredient{solarIngredient(5), solarIngredient(5)}
	interactions := []model.Interaction{
		{Element1: model.ElementSolar, Element2: model.ElementSolar, Result: model.InteractionNeutral},
	}

	effects := engine.CalculateEffects(ingredients, interactions)
	if len(effects) == 0 {
		t.Fatal("expected at least one common effect for two solar ingredients")
	}

	// All matched effects should be common tier with these ingredients
	for _, eff := range effects {
		if eff.Tier != model.EffectTierCommon {
			t.Errorf("expected common tier, got %v for %s", eff.Tier, eff.Name)
		}
	}

	// Specific check: night-vision requires solar + potency 5
	found := false
	for _, eff := range effects {
		if eff.ID == "effect-night-vision" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected effect-night-vision to match for solar potency 10")
	}

	// mild-warmth requires solar + potency 4
	found = false
	for _, eff := range effects {
		if eff.ID == "effect-mild-warmth" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected effect-mild-warmth to match for solar potency 10")
	}
}

func TestCalculateEffects_UncommonRequiresTwoElements(t *testing.T) {
	engine := newTestEngine(0.0, 0)

	// void + crystalline, potency 11 → should match uncommon effects
	// requiring both void and crystalline
	ingredients := []model.Ingredient{voidIngredient(6), crystallineIngredient(5)}
	interactions := []model.Interaction{
		{Element1: model.ElementVoid, Element2: model.ElementCrystalline, Result: model.InteractionNeutral},
	}

	effects := engine.CalculateEffects(ingredients, interactions)

	// cosmic-hearing requires void+crystalline, potency ≥ 10
	found := false
	for _, eff := range effects {
		if eff.ID == "effect-cosmic-hearing" {
			found = true
			if eff.Tier != model.EffectTierUncommon {
				t.Errorf("cosmic-hearing should be uncommon, got %v", eff.Tier)
			}
		}
	}
	if !found {
		t.Error("expected effect-cosmic-hearing for void+crystalline potency 11")
	}
}

func TestCalculateEffects_RareRequiresSynergy(t *testing.T) {
	engine := newTestEngine(0.0, 0)

	// solar + crystalline with synergy, high potency (14)
	// structural-sight requires solar+crystalline + synergy + potency ≥ 14
	ingredients := []model.Ingredient{solarIngredient(7), crystallineIngredient(7)}
	interactions := []model.Interaction{
		{Element1: model.ElementSolar, Element2: model.ElementCrystalline, Result: model.InteractionSynergy},
	}

	effects := engine.CalculateEffects(ingredients, interactions)

	found := false
	for _, eff := range effects {
		if eff.ID == "effect-structural-sight" {
			found = true
			if eff.Tier != model.EffectTierRare {
				t.Errorf("structural-sight should be rare, got %v", eff.Tier)
			}
		}
	}
	if !found {
		t.Error("expected effect-structural-sight for solar+crystalline synergy potency 14")
	}
}

func TestCalculateEffects_RareNotMatchedWithoutSynergy(t *testing.T) {
	engine := newTestEngine(0.0, 0)

	// Same elements and potency as above, but neutral interaction (no synergy)
	ingredients := []model.Ingredient{solarIngredient(7), crystallineIngredient(7)}
	interactions := []model.Interaction{
		{Element1: model.ElementSolar, Element2: model.ElementCrystalline, Result: model.InteractionNeutral},
	}

	effects := engine.CalculateEffects(ingredients, interactions)

	for _, eff := range effects {
		if eff.ID == "effect-structural-sight" {
			t.Error("structural-sight should NOT match without synergy interaction")
		}
	}
}

func TestCalculateEffects_LegendaryRequiresVolatile(t *testing.T) {
	engine := newTestEngine(0.0, 0)

	// solar + void with volatile, high potency (14)
	// temporal-echo requires solar+void + volatile + potency ≥ 14
	ingredients := []model.Ingredient{solarIngredient(7), voidIngredient(7)}
	interactions := []model.Interaction{
		{Element1: model.ElementSolar, Element2: model.ElementVoid, Result: model.InteractionVolatile},
	}

	effects := engine.CalculateEffects(ingredients, interactions)

	found := false
	for _, eff := range effects {
		if eff.ID == "effect-temporal-echo" {
			found = true
			if eff.Tier != model.EffectTierLegendary {
				t.Errorf("temporal-echo should be legendary, got %v", eff.Tier)
			}
		}
	}
	if !found {
		t.Error("expected effect-temporal-echo for solar+void volatile potency 14")
	}
}

func TestCalculateEffects_LegendaryNotMatchedWithoutVolatile(t *testing.T) {
	engine := newTestEngine(0.0, 0)

	// Same elements and potency, but mark as neutral (no volatile)
	ingredients := []model.Ingredient{solarIngredient(7), voidIngredient(7)}
	interactions := []model.Interaction{
		{Element1: model.ElementSolar, Element2: model.ElementVoid, Result: model.InteractionNeutral},
	}

	effects := engine.CalculateEffects(ingredients, interactions)

	for _, eff := range effects {
		if eff.ID == "effect-temporal-echo" {
			t.Error("temporal-echo should NOT match without volatile interaction")
		}
	}
}

func TestCalculateEffects_PotencyTooLow(t *testing.T) {
	engine := newTestEngine(0.0, 0)

	// solar potency 1+1 = 2, below the minimum for night-vision (5)
	ingredients := []model.Ingredient{solarIngredient(1), solarIngredient(1)}
	interactions := []model.Interaction{
		{Element1: model.ElementSolar, Element2: model.ElementSolar, Result: model.InteractionNeutral},
	}

	effects := engine.CalculateEffects(ingredients, interactions)

	for _, eff := range effects {
		if eff.ID == "effect-night-vision" {
			t.Error("night-vision should NOT match with total potency 2 (needs 5)")
		}
	}
}

func TestCalculateEffects_MinIngredientsRequired(t *testing.T) {
	engine := newTestEngine(0.0, 0)

	// zero-gravity-tolerance requires void + 3 ingredients + potency ≥ 10
	// Test with only 2 void ingredients (potency 10) — should NOT match
	ingredients := []model.Ingredient{voidIngredient(5), voidIngredient(5)}
	interactions := []model.Interaction{
		{Element1: model.ElementVoid, Element2: model.ElementVoid, Result: model.InteractionNeutral},
	}

	effects := engine.CalculateEffects(ingredients, interactions)
	for _, eff := range effects {
		if eff.ID == "effect-zero-gravity-tolerance" {
			t.Error("zero-gravity-tolerance should NOT match with only 2 ingredients (needs 3)")
		}
	}

	// Now with 3 void ingredients (potency 15) — should match
	ingredients3 := []model.Ingredient{voidIngredient(5), voidIngredient(5), voidIngredient(5)}
	interactions3 := []model.Interaction{
		{Element1: model.ElementVoid, Element2: model.ElementVoid, Result: model.InteractionNeutral},
		{Element1: model.ElementVoid, Element2: model.ElementVoid, Result: model.InteractionNeutral},
		{Element1: model.ElementVoid, Element2: model.ElementVoid, Result: model.InteractionNeutral},
	}

	effects3 := engine.CalculateEffects(ingredients3, interactions3)
	found := false
	for _, eff := range effects3 {
		if eff.ID == "effect-zero-gravity-tolerance" {
			found = true
		}
	}
	if !found {
		t.Error("zero-gravity-tolerance should match with 3 void ingredients potency 15")
	}
}

func TestCalculateEffects_SortedByTierThenName(t *testing.T) {
	engine := newTestEngine(0.0, 0)

	// Use 4 ingredients with diverse elements to trigger many effects
	ingredients := []model.Ingredient{
		solarIngredient(5), crystallineIngredient(5),
		organicIngredient(5), plasmaIngredient(5),
	}
	// Build interactions from the engine itself
	report, err := engine.CheckCompatibility(ingredients)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	effects := engine.CalculateEffects(ingredients, report.Interactions)
	if len(effects) < 2 {
		t.Fatalf("expected multiple effects, got %d", len(effects))
	}

	// Verify sorting: tier rank must be non-decreasing, names sorted within tier
	for i := 1; i < len(effects); i++ {
		prevRank := tierRank(effects[i-1].Tier)
		currRank := tierRank(effects[i].Tier)
		if prevRank > currRank {
			t.Errorf("effects not sorted by tier: %s (%v) before %s (%v)",
				effects[i-1].Name, effects[i-1].Tier, effects[i].Name, effects[i].Tier)
		}
		if prevRank == currRank && effects[i-1].Name > effects[i].Name {
			t.Errorf("effects not sorted by name within tier: %s before %s",
				effects[i-1].Name, effects[i].Name)
		}
	}
}

// ── Filter Effects by Outcome Tests ─────────────────────────────────

func TestFilterEffectsByOutcome(t *testing.T) {
	effects := []model.Effect{
		{ID: "e1", Tier: model.EffectTierCommon},
		{ID: "e2", Tier: model.EffectTierUncommon},
		{ID: "e3", Tier: model.EffectTierRare},
		{ID: "e4", Tier: model.EffectTierLegendary},
	}

	t.Run("success keeps all effects", func(t *testing.T) {
		result := filterEffectsByOutcome(effects, model.BrewResultSuccess)
		if len(result) != 4 {
			t.Errorf("success: got %d effects, want 4", len(result))
		}
	})

	t.Run("partial keeps common and uncommon only", func(t *testing.T) {
		result := filterEffectsByOutcome(effects, model.BrewResultPartial)
		if len(result) != 2 {
			t.Errorf("partial: got %d effects, want 2", len(result))
		}
		for _, eff := range result {
			if eff.Tier == model.EffectTierRare || eff.Tier == model.EffectTierLegendary {
				t.Errorf("partial should not include %v tier", eff.Tier)
			}
		}
	})

	t.Run("catastrophic failure returns empty slice", func(t *testing.T) {
		result := filterEffectsByOutcome(effects, model.BrewResultCatastrophicFailure)
		if len(result) != 0 {
			t.Errorf("catastrophic: got %d effects, want 0", len(result))
		}
	})

	t.Run("empty input returns empty output", func(t *testing.T) {
		result := filterEffectsByOutcome([]model.Effect{}, model.BrewResultSuccess)
		if result != nil && len(result) != 0 {
			t.Errorf("empty input: got %d effects, want 0", len(result))
		}
	})
}

// ── Full Brew Integration Tests ─────────────────────────────────────
//
// These test the complete Brew() flow end-to-end with controlled randomness.

func TestBrew_SuccessfulBrew(t *testing.T) {
	// Roll 0.0 → always success (below any chance threshold)
	engine := newTestEngine(0.0, 0)

	ingredients := []model.Ingredient{solarIngredient(5), organicIngredient(5)}
	outcome, err := engine.Brew(ingredients)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if outcome.Result != model.BrewResultSuccess {
		t.Errorf("expected success, got %v", outcome.Result)
	}
	if outcome.PotionName == "" {
		t.Error("expected a potion name, got empty string")
	}
	if outcome.Notes == "" {
		t.Error("expected brew notes, got empty string")
	}
	if outcome.Synergies != 1 {
		t.Errorf("solar+organic should produce 1 synergy, got %d", outcome.Synergies)
	}
}

func TestBrew_CatastrophicFailure(t *testing.T) {
	// Roll 0.99 → catastrophic failure (above chance+0.20 for any reasonable chance)
	engine := newTestEngine(0.99, 0)

	ingredients := []model.Ingredient{solarIngredient(5), organicIngredient(5)}
	outcome, err := engine.Brew(ingredients)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if outcome.Result != model.BrewResultCatastrophicFailure {
		t.Errorf("expected catastrophic_failure, got %v", outcome.Result)
	}
	// Catastrophic failures discover no effects
	if len(outcome.Effects) != 0 {
		t.Errorf("catastrophic failure should discover 0 effects, got %d", len(outcome.Effects))
	}
}

func TestBrew_PartialBrew(t *testing.T) {
	// For solar+organic (1 synergy), chance = 0.75
	// Roll 0.80 → partial (between 0.75 and 0.95)
	engine := newTestEngine(0.80, 0)

	ingredients := []model.Ingredient{solarIngredient(5), organicIngredient(5)}
	outcome, err := engine.Brew(ingredients)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if outcome.Result != model.BrewResultPartial {
		t.Errorf("expected partial, got %v", outcome.Result)
	}
	// Partial should only have common/uncommon effects
	for _, eff := range outcome.Effects {
		if eff.Tier == model.EffectTierRare || eff.Tier == model.EffectTierLegendary {
			t.Errorf("partial brew should not include %v effect %s", eff.Tier, eff.Name)
		}
	}
}

func TestBrew_VolatileSuccessHasBonusNote(t *testing.T) {
	// solar + void = volatile. Roll 0.0 → success
	engine := newTestEngine(0.0, 0)

	ingredients := []model.Ingredient{solarIngredient(5), voidIngredient(5)}
	outcome, err := engine.Brew(ingredients)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if outcome.Volatiles != 1 {
		t.Errorf("expected 1 volatile, got %d", outcome.Volatiles)
	}
	// Successful volatile brews should mention stabilization
	if outcome.Result == model.BrewResultSuccess {
		if len(outcome.Notes) == 0 {
			t.Error("expected notes for volatile success")
		}
	}
}

// ── Helper Function Tests ───────────────────────────────────────────

func TestUniqueElements(t *testing.T) {
	ingredients := []model.Ingredient{
		solarIngredient(5), solarIngredient(3), voidIngredient(5),
	}
	elems := uniqueElements(ingredients)
	if len(elems) != 2 {
		t.Errorf("expected 2 unique elements, got %d", len(elems))
	}
	if !elems[model.ElementSolar] {
		t.Error("expected solar element")
	}
	if !elems[model.ElementVoid] {
		t.Error("expected void element")
	}
}

func TestTotalPotency(t *testing.T) {
	ingredients := []model.Ingredient{
		solarIngredient(3), voidIngredient(7), crystallineIngredient(2),
	}
	if got := totalPotency(ingredients); got != 12 {
		t.Errorf("totalPotency() = %d, want 12", got)
	}
}

func TestDominantElement(t *testing.T) {
	tests := []struct {
		name        string
		ingredients []model.Ingredient
		want        model.Element
	}{
		{
			name:        "single element clear winner",
			ingredients: []model.Ingredient{solarIngredient(8), voidIngredient(3)},
			want:        model.ElementSolar,
		},
		{
			name:        "cumulative potency wins",
			ingredients: []model.Ingredient{voidIngredient(3), voidIngredient(4), solarIngredient(5)},
			want:        model.ElementVoid,
		},
		{
			name: "tie broken by element order (alphabetical)",
			// crystalline and solar both at potency 5 — crystalline wins alphabetically
			ingredients: []model.Ingredient{crystallineIngredient(5), solarIngredient(5)},
			want:        model.ElementCrystalline,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := dominantElement(tt.ingredients)
			if got != tt.want {
				t.Errorf("dominantElement() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTierRank(t *testing.T) {
	if tierRank(model.EffectTierCommon) >= tierRank(model.EffectTierUncommon) {
		t.Error("common should rank lower than uncommon")
	}
	if tierRank(model.EffectTierUncommon) >= tierRank(model.EffectTierRare) {
		t.Error("uncommon should rank lower than rare")
	}
	if tierRank(model.EffectTierRare) >= tierRank(model.EffectTierLegendary) {
		t.Error("rare should rank lower than legendary")
	}
}

// ── Potion Name Generation Tests ────────────────────────────────────

func TestGeneratePotionName(t *testing.T) {
	rand := &fakeRand{floatVal: 0.0, intVal: 0}

	name := generatePotionName(model.ElementSolar, model.BrewResultSuccess, rand)
	if name == "" {
		t.Error("expected non-empty potion name")
	}
	// With intVal=0, should pick first adjective and first noun
	// Success adjectives[0] = "Radiant", Solar nouns[0] = "Sunfire Elixir"
	if name != "Radiant Sunfire Elixir" {
		t.Errorf("expected 'Radiant Sunfire Elixir', got %q", name)
	}
}

func TestGeneratePotionName_UnknownElement(t *testing.T) {
	rand := &fakeRand{floatVal: 0.0, intVal: 0}
	name := generatePotionName(model.Element("unknown"), model.BrewResultSuccess, rand)
	if name != "Mysterious Concoction" {
		t.Errorf("unknown element should produce 'Mysterious Concoction', got %q", name)
	}
}

func TestGenerateNotes(t *testing.T) {
	rand := &fakeRand{floatVal: 0.0, intVal: 0}

	note := generateNotes(model.BrewResultSuccess, 0, 0, rand)
	if note == "" {
		t.Error("expected non-empty notes for success")
	}

	note = generateNotes(model.BrewResultCatastrophicFailure, 0, 0, rand)
	if note == "" {
		t.Error("expected non-empty notes for catastrophic failure")
	}
}

func TestGenerateNotes_VolatileSuccessBonus(t *testing.T) {
	rand := &fakeRand{floatVal: 0.0, intVal: 0}
	note := generateNotes(model.BrewResultSuccess, 0, 1, rand)
	expected := "volatile elements stabilized"
	if len(note) == 0 {
		t.Fatal("expected notes")
	}
	// The volatile success bonus should be appended
	found := false
	for i := 0; i <= len(note)-len(expected); i++ {
		if note[i:i+len(expected)] == expected {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("volatile success note should contain %q, got %q", expected, note)
	}
}

func TestGenerateNotes_SynergyBonus(t *testing.T) {
	rand := &fakeRand{floatVal: 0.0, intVal: 0}
	note := generateNotes(model.BrewResultSuccess, 2, 0, rand)
	expected := "synergies between elements"
	found := false
	for i := 0; i <= len(note)-len(expected); i++ {
		if note[i:i+len(expected)] == expected {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("synergy bonus note should contain %q, got %q", expected, note)
	}
}
