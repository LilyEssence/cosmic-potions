package store

import (
	"errors"
	"testing"
	"time"

	"github.com/LilyEssence/cosmic-potions/internal/model"
)

// ── Store Tests ─────────────────────────────────────────────────────
//
// GO CONCEPT: Testing in the Same Package
// These tests are in `package store` (not `package store_test`), so they
// can access unexported fields and methods. For store tests this is fine —
// we're testing the implementation, not just the interface contract.

// ── PlanetStore Tests ───────────────────────────────────────────────

func TestGetPlanet_Found(t *testing.T) {
	s := NewMemoryStore()

	planet, err := s.GetPlanet("planet-verdanthia")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if planet.Name != "Verdanthia" {
		t.Errorf("expected Verdanthia, got %s", planet.Name)
	}
}

func TestGetPlanet_NotFound(t *testing.T) {
	s := NewMemoryStore()

	_, err := s.GetPlanet("planet-nonexistent")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestListPlanets(t *testing.T) {
	s := NewMemoryStore()

	planets, err := s.ListPlanets()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(planets) != 6 {
		t.Fatalf("expected 6 planets, got %d", len(planets))
	}

	// Verify sorted by name
	for i := 1; i < len(planets); i++ {
		if planets[i-1].Name > planets[i].Name {
			t.Errorf("planets not sorted: %s before %s", planets[i-1].Name, planets[i].Name)
		}
	}
}

func TestGetPlanetWithIngredients(t *testing.T) {
	s := NewMemoryStore()

	result, err := s.GetPlanetWithIngredients("planet-verdanthia")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Planet.Name != "Verdanthia" {
		t.Errorf("expected Verdanthia, got %s", result.Planet.Name)
	}
	if len(result.Ingredients) == 0 {
		t.Error("expected at least one ingredient for Verdanthia")
	}

	// Verify all returned ingredients belong to Verdanthia
	for _, ing := range result.Ingredients {
		if ing.PlanetID != "planet-verdanthia" {
			t.Errorf("ingredient %s has planetID %s, expected planet-verdanthia",
				ing.Name, ing.PlanetID)
		}
	}

	// Verify sorted by name
	for i := 1; i < len(result.Ingredients); i++ {
		if result.Ingredients[i-1].Name > result.Ingredients[i].Name {
			t.Errorf("ingredients not sorted: %s before %s",
				result.Ingredients[i-1].Name, result.Ingredients[i].Name)
		}
	}
}

func TestGetPlanetWithIngredients_NotFound(t *testing.T) {
	s := NewMemoryStore()

	_, err := s.GetPlanetWithIngredients("planet-nonexistent")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

// ── IngredientStore Tests ───────────────────────────────────────────

func TestGetIngredient_Found(t *testing.T) {
	s := NewMemoryStore()

	ing, err := s.GetIngredient("ing-solar-crystal")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ing.Element != model.ElementSolar {
		t.Errorf("expected solar element, got %v", ing.Element)
	}
}

func TestGetIngredient_NotFound(t *testing.T) {
	s := NewMemoryStore()

	_, err := s.GetIngredient("ing-nonexistent")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestListIngredients_NoFilter(t *testing.T) {
	s := NewMemoryStore()

	ingredients, err := s.ListIngredients(model.IngredientFilter{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ingredients) < 10 {
		t.Fatalf("expected at least 10 ingredients (seed has 18), got %d", len(ingredients))
	}

	// Verify sorted by name
	for i := 1; i < len(ingredients); i++ {
		if ingredients[i-1].Name > ingredients[i].Name {
			t.Errorf("ingredients not sorted: %s before %s",
				ingredients[i-1].Name, ingredients[i].Name)
		}
	}
}

func TestListIngredients_FilterByElement(t *testing.T) {
	s := NewMemoryStore()
	elem := model.ElementVoid
	filter := model.IngredientFilter{Element: &elem}

	ingredients, err := s.ListIngredients(filter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ingredients) == 0 {
		t.Fatal("expected at least one void ingredient")
	}
	for _, ing := range ingredients {
		if ing.Element != model.ElementVoid {
			t.Errorf("expected void element, got %v for %s", ing.Element, ing.Name)
		}
	}
}

func TestListIngredients_FilterByRarity(t *testing.T) {
	s := NewMemoryStore()
	rarity := model.RarityLegendary
	filter := model.IngredientFilter{Rarity: &rarity}

	ingredients, err := s.ListIngredients(filter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ingredients) == 0 {
		t.Fatal("expected at least one legendary ingredient")
	}
	for _, ing := range ingredients {
		if ing.Rarity != model.RarityLegendary {
			t.Errorf("expected legendary rarity, got %v for %s", ing.Rarity, ing.Name)
		}
	}
}

func TestListIngredients_FilterByPlanetID(t *testing.T) {
	s := NewMemoryStore()
	planetID := "planet-verdanthia"
	filter := model.IngredientFilter{PlanetID: &planetID}

	ingredients, err := s.ListIngredients(filter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ingredients) == 0 {
		t.Fatal("expected at least one ingredient from Verdanthia")
	}
	for _, ing := range ingredients {
		if ing.PlanetID != planetID {
			t.Errorf("expected planetID %s, got %s for %s", planetID, ing.PlanetID, ing.Name)
		}
	}
}

func TestListIngredients_CombinedFilters(t *testing.T) {
	s := NewMemoryStore()
	elem := model.ElementOrganic
	planetID := "planet-verdanthia"
	filter := model.IngredientFilter{Element: &elem, PlanetID: &planetID}

	ingredients, err := s.ListIngredients(filter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, ing := range ingredients {
		if ing.Element != model.ElementOrganic {
			t.Errorf("expected organic element, got %v for %s", ing.Element, ing.Name)
		}
		if ing.PlanetID != planetID {
			t.Errorf("expected planetID %s, got %s for %s", planetID, ing.PlanetID, ing.Name)
		}
	}
}

func TestListIngredients_NoMatchReturnsEmpty(t *testing.T) {
	s := NewMemoryStore()
	elem := model.Element("nonexistent")
	filter := model.IngredientFilter{Element: &elem}

	ingredients, err := s.ListIngredients(filter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ingredients) != 0 {
		t.Errorf("expected 0 ingredients for nonexistent element, got %d", len(ingredients))
	}
}

// ── RecipeStore Tests ───────────────────────────────────────────────

func TestGetRecipe_Found(t *testing.T) {
	s := NewMemoryStore()

	// Use a known seed recipe ID
	recipes, _ := s.ListRecipes(model.RecipeFilter{})
	if len(recipes) == 0 {
		t.Fatal("expected at least one seed recipe")
	}

	recipe, err := s.GetRecipe(recipes[0].ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if recipe.ID != recipes[0].ID {
		t.Errorf("expected ID %s, got %s", recipes[0].ID, recipe.ID)
	}
}

func TestGetRecipe_NotFound(t *testing.T) {
	s := NewMemoryStore()

	_, err := s.GetRecipe("recipe-nonexistent")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestListRecipes_NoFilter(t *testing.T) {
	s := NewMemoryStore()

	recipes, err := s.ListRecipes(model.RecipeFilter{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(recipes) < 5 {
		t.Fatalf("expected at least 5 seed recipes, got %d", len(recipes))
	}

	// Verify sorted by name
	for i := 1; i < len(recipes); i++ {
		if recipes[i-1].Name > recipes[i].Name {
			t.Errorf("recipes not sorted: %s before %s", recipes[i-1].Name, recipes[i].Name)
		}
	}
}

func TestListRecipes_FilterByDifficulty(t *testing.T) {
	s := NewMemoryStore()
	diff := model.DifficultyNovice
	filter := model.RecipeFilter{Difficulty: &diff}

	recipes, err := s.ListRecipes(filter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, r := range recipes {
		if r.Difficulty != model.DifficultyNovice {
			t.Errorf("expected novice difficulty, got %v for %s", r.Difficulty, r.Name)
		}
	}
}

// ── BrewStore Tests ─────────────────────────────────────────────────

func TestSaveAndGetBrew(t *testing.T) {
	s := NewMemoryStore()

	brew := model.Brew{
		ID:              "brew-test-001",
		IngredientsUsed: []string{"ing-solar-crystal", "ing-drift-dust"},
		Result:          model.BrewResultSuccess,
		PotionName:      "Test Elixir",
		Effects:         []string{"effect-night-vision"},
		Notes:           "A test brew.",
		BrewedAt:        time.Now().UTC(),
	}

	if err := s.SaveBrew(brew); err != nil {
		t.Fatalf("SaveBrew error: %v", err)
	}

	retrieved, err := s.GetBrew("brew-test-001")
	if err != nil {
		t.Fatalf("GetBrew error: %v", err)
	}
	if retrieved.PotionName != "Test Elixir" {
		t.Errorf("expected 'Test Elixir', got %s", retrieved.PotionName)
	}
	if retrieved.Result != model.BrewResultSuccess {
		t.Errorf("expected success, got %v", retrieved.Result)
	}
}

func TestGetBrew_NotFound(t *testing.T) {
	s := NewMemoryStore()

	_, err := s.GetBrew("brew-nonexistent")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestListBrews_ReverseChronological(t *testing.T) {
	s := NewMemoryStore()

	// Save 3 brews in order
	for i, id := range []string{"brew-1", "brew-2", "brew-3"} {
		brew := model.Brew{
			ID:              id,
			IngredientsUsed: []string{"ing-solar-crystal"},
			Result:          model.BrewResultSuccess,
			PotionName:      "Test " + id,
			Effects:         []string{},
			BrewedAt:        time.Now().Add(time.Duration(i) * time.Minute),
		}
		if err := s.SaveBrew(brew); err != nil {
			t.Fatalf("SaveBrew error: %v", err)
		}
	}

	brews, err := s.ListBrews()
	if err != nil {
		t.Fatalf("ListBrews error: %v", err)
	}
	if len(brews) != 3 {
		t.Fatalf("expected 3 brews, got %d", len(brews))
	}

	// Newest first
	if brews[0].ID != "brew-3" {
		t.Errorf("expected brew-3 first, got %s", brews[0].ID)
	}
	if brews[1].ID != "brew-2" {
		t.Errorf("expected brew-2 second, got %s", brews[1].ID)
	}
	if brews[2].ID != "brew-1" {
		t.Errorf("expected brew-1 third, got %s", brews[2].ID)
	}
}

func TestListBrews_EmptyStore(t *testing.T) {
	s := NewMemoryStore()

	brews, err := s.ListBrews()
	if err != nil {
		t.Fatalf("ListBrews error: %v", err)
	}
	if len(brews) != 0 {
		t.Errorf("expected 0 brews, got %d", len(brews))
	}
}
