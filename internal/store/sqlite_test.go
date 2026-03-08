package store

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/LilyEssence/cosmic-potions/internal/model"
)

// ── SQLite Store Tests ──────────────────────────────────────────────
//
// GO CONCEPT: Test Helpers
// newTestSQLiteStore creates a temporary SQLite database with migrations
// applied. Each test gets its own database file (via t.TempDir()), so tests
// are fully isolated. The cleanup is automatic — t.TempDir() is removed
// when the test finishes.
//
// GO CONCEPT: t.TempDir()
// Creates a temporary directory unique to this test. Go automatically cleans
// it up when the test completes. This is perfect for SQLite test databases —
// no manual cleanup needed, and no interference between parallel tests.
//
// We could use ":memory:" (SQLite's in-memory mode), but using a file more
// closely matches production behavior and tests file I/O paths too.

// testMigrationsDir returns the absolute path to the migrations directory.
// Since tests run from the package directory (internal/store/), we need to
// navigate up to the project root to find the migrations folder.
func testMigrationsDir(t *testing.T) string {
	t.Helper()

	// Walk up from internal/store/ to project root.
	// GO CONCEPT: Testing Working Directory
	// Go tests run with the working directory set to the package directory.
	// For internal/store/, that's <project>/internal/store/. To find project
	// root files, we must navigate relative to that or use absolute paths.
	dir, err := filepath.Abs(filepath.Join("..", "..", "migrations"))
	if err != nil {
		t.Fatalf("resolving migrations directory: %v", err)
	}

	// Verify the directory exists.
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Fatalf("migrations directory not found at %s", dir)
	}

	return dir
}

// newTestSQLiteStore creates a fresh SQLite store backed by a temp file
// with all migrations applied (including seed data).
func newTestSQLiteStore(t *testing.T) *SQLiteStore {
	t.Helper()

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	migrationsDir := testMigrationsDir(t)

	s, err := NewSQLiteStore(dbPath, migrationsDir)
	if err != nil {
		t.Fatalf("failed to create SQLite store: %v", err)
	}

	// Register cleanup to close the database when the test finishes.
	// GO CONCEPT: t.Cleanup
	// Registers a function to run when the test (and its subtests) complete.
	// Like defer, but scoped to the test lifecycle rather than the function.
	// Useful when you return the resource to the caller (defer would close
	// it immediately when this helper function returns).
	t.Cleanup(func() {
		s.Close()
	})

	return s
}

// ── PlanetStore Tests ───────────────────────────────────────────────

func TestSQLiteGetPlanet_Found(t *testing.T) {
	s := newTestSQLiteStore(t)

	planet, err := s.GetPlanet("planet-verdanthia")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if planet.Name != "Verdanthia" {
		t.Errorf("expected Verdanthia, got %s", planet.Name)
	}
	if planet.Biome != model.BiomeFungal {
		t.Errorf("expected fungal biome, got %v", planet.Biome)
	}
}

func TestSQLiteGetPlanet_NotFound(t *testing.T) {
	s := newTestSQLiteStore(t)

	_, err := s.GetPlanet("planet-nonexistent")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestSQLiteListPlanets(t *testing.T) {
	s := newTestSQLiteStore(t)

	planets, err := s.ListPlanets()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(planets) != 6 {
		t.Fatalf("expected 6 planets, got %d", len(planets))
	}

	// Verify sorted by name.
	for i := 1; i < len(planets); i++ {
		if planets[i-1].Name > planets[i].Name {
			t.Errorf("planets not sorted: %s before %s", planets[i-1].Name, planets[i].Name)
		}
	}
}

func TestSQLiteGetPlanetWithIngredients(t *testing.T) {
	s := newTestSQLiteStore(t)

	result, err := s.GetPlanetWithIngredients("planet-verdanthia")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Planet.Name != "Verdanthia" {
		t.Errorf("expected Verdanthia, got %s", result.Planet.Name)
	}
	// Verdanthia has 3 ingredients in seed data.
	if len(result.Ingredients) != 3 {
		t.Fatalf("expected 3 ingredients for Verdanthia, got %d", len(result.Ingredients))
	}

	// Verify all returned ingredients belong to Verdanthia.
	for _, ing := range result.Ingredients {
		if ing.PlanetID != "planet-verdanthia" {
			t.Errorf("ingredient %s has planetID %s, expected planet-verdanthia",
				ing.Name, ing.PlanetID)
		}
	}

	// Verify sorted by name.
	for i := 1; i < len(result.Ingredients); i++ {
		if result.Ingredients[i-1].Name > result.Ingredients[i].Name {
			t.Errorf("ingredients not sorted: %s before %s",
				result.Ingredients[i-1].Name, result.Ingredients[i].Name)
		}
	}
}

func TestSQLiteGetPlanetWithIngredients_NotFound(t *testing.T) {
	s := newTestSQLiteStore(t)

	_, err := s.GetPlanetWithIngredients("planet-nonexistent")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

// ── IngredientStore Tests ───────────────────────────────────────────

func TestSQLiteGetIngredient_Found(t *testing.T) {
	s := newTestSQLiteStore(t)

	ing, err := s.GetIngredient("ing-solar-crystal")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ing.Element != model.ElementSolar {
		t.Errorf("expected solar element, got %v", ing.Element)
	}
	if ing.Potency != 6 {
		t.Errorf("expected potency 6, got %d", ing.Potency)
	}
	// Verify flavor notes were deserialized from JSON.
	if len(ing.FlavorNotes) == 0 {
		t.Error("expected non-empty flavor notes")
	}
}

func TestSQLiteGetIngredient_NotFound(t *testing.T) {
	s := newTestSQLiteStore(t)

	_, err := s.GetIngredient("ing-nonexistent")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestSQLiteListIngredients_NoFilter(t *testing.T) {
	s := newTestSQLiteStore(t)

	ingredients, err := s.ListIngredients(model.IngredientFilter{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ingredients) != 18 {
		t.Fatalf("expected 18 ingredients (seed data), got %d", len(ingredients))
	}

	// Verify sorted by name.
	for i := 1; i < len(ingredients); i++ {
		if ingredients[i-1].Name > ingredients[i].Name {
			t.Errorf("ingredients not sorted: %s before %s",
				ingredients[i-1].Name, ingredients[i].Name)
		}
	}
}

func TestSQLiteListIngredients_FilterByElement(t *testing.T) {
	s := newTestSQLiteStore(t)
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

func TestSQLiteListIngredients_FilterByRarity(t *testing.T) {
	s := newTestSQLiteStore(t)
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

func TestSQLiteListIngredients_FilterByPlanetID(t *testing.T) {
	s := newTestSQLiteStore(t)
	planetID := "planet-verdanthia"
	filter := model.IngredientFilter{PlanetID: &planetID}

	ingredients, err := s.ListIngredients(filter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ingredients) != 3 {
		t.Fatalf("expected 3 Verdanthia ingredients, got %d", len(ingredients))
	}
	for _, ing := range ingredients {
		if ing.PlanetID != planetID {
			t.Errorf("expected planetID %s, got %s for %s", planetID, ing.PlanetID, ing.Name)
		}
	}
}

func TestSQLiteListIngredients_CombinedFilters(t *testing.T) {
	s := newTestSQLiteStore(t)
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

func TestSQLiteListIngredients_NoMatchReturnsEmpty(t *testing.T) {
	s := newTestSQLiteStore(t)
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

func TestSQLiteGetRecipe_Found(t *testing.T) {
	s := newTestSQLiteStore(t)

	recipe, err := s.GetRecipe("recipe-starlight-tonic")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if recipe.Name != "Starlight Tonic" {
		t.Errorf("expected 'Starlight Tonic', got %s", recipe.Name)
	}
	if recipe.Difficulty != model.DifficultyNovice {
		t.Errorf("expected novice difficulty, got %v", recipe.Difficulty)
	}
	// Verify effects were loaded from the junction table.
	if len(recipe.Effects) == 0 {
		t.Error("expected at least one effect for Starlight Tonic")
	}
	// Verify ingredients were loaded from the junction table.
	if len(recipe.Ingredients) == 0 {
		t.Error("expected at least one ingredient for Starlight Tonic")
	}
}

func TestSQLiteGetRecipe_NotFound(t *testing.T) {
	s := newTestSQLiteStore(t)

	_, err := s.GetRecipe("recipe-nonexistent")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestSQLiteListRecipes_NoFilter(t *testing.T) {
	s := newTestSQLiteStore(t)

	recipes, err := s.ListRecipes(model.RecipeFilter{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(recipes) != 8 {
		t.Fatalf("expected 8 seed recipes, got %d", len(recipes))
	}

	// Verify sorted by name.
	for i := 1; i < len(recipes); i++ {
		if recipes[i-1].Name > recipes[i].Name {
			t.Errorf("recipes not sorted: %s before %s", recipes[i-1].Name, recipes[i].Name)
		}
	}

	// Verify each recipe has effects and ingredients loaded.
	for _, r := range recipes {
		if len(r.Effects) == 0 {
			t.Errorf("recipe %s has no effects", r.Name)
		}
		if len(r.Ingredients) == 0 {
			t.Errorf("recipe %s has no ingredients", r.Name)
		}
	}
}

func TestSQLiteListRecipes_FilterByDifficulty(t *testing.T) {
	s := newTestSQLiteStore(t)
	diff := model.DifficultyMaster
	filter := model.RecipeFilter{Difficulty: &diff}

	recipes, err := s.ListRecipes(filter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(recipes) == 0 {
		t.Fatal("expected at least one master recipe")
	}
	for _, r := range recipes {
		if r.Difficulty != model.DifficultyMaster {
			t.Errorf("expected master difficulty, got %v for %s", r.Difficulty, r.Name)
		}
	}
}

// ── BrewStore Tests ─────────────────────────────────────────────────

func TestSQLiteSaveAndGetBrew(t *testing.T) {
	s := newTestSQLiteStore(t)

	brew := model.Brew{
		ID:              "brew-test-001",
		IngredientsUsed: []string{"ing-solar-crystal", "ing-drift-dust"},
		Result:          model.BrewResultSuccess,
		PotionName:      "Test Elixir",
		Effects:         []string{"night vision", "mild euphoria"},
		Notes:           "A test brew.",
		BrewedAt:        time.Now().UTC().Truncate(time.Second), // Truncate for comparison
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
	if len(retrieved.IngredientsUsed) != 2 {
		t.Errorf("expected 2 ingredients, got %d", len(retrieved.IngredientsUsed))
	}
	if len(retrieved.Effects) != 2 {
		t.Errorf("expected 2 effects, got %d", len(retrieved.Effects))
	}
	if retrieved.Notes != "A test brew." {
		t.Errorf("expected 'A test brew.', got %s", retrieved.Notes)
	}
}

func TestSQLiteSaveBrew_WithRecipeID(t *testing.T) {
	s := newTestSQLiteStore(t)

	recipeID := "recipe-starlight-tonic"
	brew := model.Brew{
		ID:              "brew-test-recipe",
		RecipeID:        &recipeID,
		IngredientsUsed: []string{"ing-solar-crystal", "ing-luminous-spore"},
		Result:          model.BrewResultSuccess,
		PotionName:      "Starlight Tonic",
		Effects:         []string{"night vision"},
		BrewedAt:        time.Now().UTC().Truncate(time.Second),
	}

	if err := s.SaveBrew(brew); err != nil {
		t.Fatalf("SaveBrew error: %v", err)
	}

	retrieved, err := s.GetBrew("brew-test-recipe")
	if err != nil {
		t.Fatalf("GetBrew error: %v", err)
	}
	if retrieved.RecipeID == nil {
		t.Fatal("expected non-nil RecipeID")
	}
	if *retrieved.RecipeID != recipeID {
		t.Errorf("expected recipeID %s, got %s", recipeID, *retrieved.RecipeID)
	}
}

func TestSQLiteSaveBrew_NilRecipeID(t *testing.T) {
	s := newTestSQLiteStore(t)

	brew := model.Brew{
		ID:              "brew-experimental",
		RecipeID:        nil, // experimental brew — no recipe
		IngredientsUsed: []string{"ing-solar-crystal"},
		Result:          model.BrewResultPartial,
		PotionName:      "Mystery Goo",
		Effects:         []string{},
		BrewedAt:        time.Now().UTC().Truncate(time.Second),
	}

	if err := s.SaveBrew(brew); err != nil {
		t.Fatalf("SaveBrew error: %v", err)
	}

	retrieved, err := s.GetBrew("brew-experimental")
	if err != nil {
		t.Fatalf("GetBrew error: %v", err)
	}
	if retrieved.RecipeID != nil {
		t.Errorf("expected nil RecipeID, got %v", *retrieved.RecipeID)
	}
}

func TestSQLiteGetBrew_NotFound(t *testing.T) {
	s := newTestSQLiteStore(t)

	_, err := s.GetBrew("brew-nonexistent")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestSQLiteListBrews_ReverseChronological(t *testing.T) {
	s := newTestSQLiteStore(t)

	// Save 3 brews with distinct timestamps.
	baseTime := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	for i, id := range []string{"brew-1", "brew-2", "brew-3"} {
		brew := model.Brew{
			ID:              id,
			IngredientsUsed: []string{"ing-solar-crystal"},
			Result:          model.BrewResultSuccess,
			PotionName:      "Test " + id,
			Effects:         []string{},
			BrewedAt:        baseTime.Add(time.Duration(i) * time.Minute),
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

	// Newest first.
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

func TestSQLiteListBrews_EmptyStore(t *testing.T) {
	s := newTestSQLiteStore(t)

	brews, err := s.ListBrews()
	if err != nil {
		t.Fatalf("ListBrews error: %v", err)
	}
	if len(brews) != 0 {
		t.Errorf("expected 0 brews, got %d", len(brews))
	}
}

// ── Interaction Tests ───────────────────────────────────────────────

func TestSQLiteLoadInteractions(t *testing.T) {
	s := newTestSQLiteStore(t)

	interactions, err := s.LoadInteractions()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 5×5 matrix = 25 interactions.
	if len(interactions) != 25 {
		t.Fatalf("expected 25 interactions, got %d", len(interactions))
	}
}

func TestSQLiteLoadInteractionMap(t *testing.T) {
	s := newTestSQLiteStore(t)

	matrix, err := s.LoadInteractionMap()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify a known interaction: solar + void = volatile.
	if result, ok := matrix[model.ElementSolar][model.ElementVoid]; !ok {
		t.Error("expected solar→void interaction to exist")
	} else if result != model.InteractionVolatile {
		t.Errorf("expected volatile for solar+void, got %v", result)
	}

	// Verify solar + solar = neutral.
	if result := matrix[model.ElementSolar][model.ElementSolar]; result != model.InteractionNeutral {
		t.Errorf("expected neutral for solar+solar, got %v", result)
	}
}

// ── Migration Idempotency Test ──────────────────────────────────────

func TestSQLiteMigrationIdempotency(t *testing.T) {
	// Running migrations twice should not error (schema_migrations prevents re-run).
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	migrationsDir := testMigrationsDir(t)

	// First run — creates everything.
	s1, err := NewSQLiteStore(dbPath, migrationsDir)
	if err != nil {
		t.Fatalf("first init failed: %v", err)
	}
	s1.Close()

	// Second run — should be a no-op (migrations already applied).
	s2, err := NewSQLiteStore(dbPath, migrationsDir)
	if err != nil {
		t.Fatalf("second init failed (idempotency broken): %v", err)
	}
	defer s2.Close()

	// Verify data is still intact.
	planets, err := s2.ListPlanets()
	if err != nil {
		t.Fatalf("listing planets after re-init: %v", err)
	}
	if len(planets) != 6 {
		t.Errorf("expected 6 planets after re-init, got %d", len(planets))
	}
}
