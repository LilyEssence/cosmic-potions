package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/LilyEssence/cosmic-potions/internal/brewing"
	"github.com/LilyEssence/cosmic-potions/internal/model"
	"github.com/LilyEssence/cosmic-potions/internal/store"
	"github.com/LilyEssence/cosmic-potions/seed"
)

// ── Handler Test Setup ──────────────────────────────────────────────
//
// GO CONCEPT: httptest Package
// Go's standard library includes net/http/httptest, which provides:
//   - httptest.NewRequest() → creates a fake *http.Request without a real server
//   - httptest.NewRecorder() → captures what the handler writes (status, headers, body)
//
// Together with chi's test utilities, we can test handlers without starting
// a real HTTP server. This is fast, isolated, and deterministic.
//
// GO CONCEPT: Test Fixtures
// newTestRouter() creates a fully wired router for handler tests. It uses
// the real MemoryStore (loaded with seed data) and a deterministic brew
// engine (via fakeRand). This is an "integration at the handler level" —
// real stores, fake randomness.

// fakeRand for handler tests — deterministic brew outcomes.
type fakeRandHandler struct {
	floatVal float64
	intVal   int
}

func (f *fakeRandHandler) Float64() float64 { return f.floatVal }
func (f *fakeRandHandler) Intn(n int) int {
	if n == 0 {
		return 0
	}
	return f.intVal % n
}

// newTestRouter creates a router with all routes registered, using real
// seed data and a deterministic randomizer.
func newTestRouter() chi.Router {
	r := chi.NewRouter()
	memStore := store.NewMemoryStore()
	engine := brewing.NewBrewEngine(
		seed.Interactions(),
		seed.Effects(),
		&fakeRandHandler{floatVal: 0.0, intVal: 0}, // always success
	)
	RegisterRoutes(r, Deps{
		Planets:      memStore,
		Ingredients:  memStore,
		Recipes:      memStore,
		Brews:        memStore,
		Engine:       engine,
		Effects:      seed.Effects(),
		Interactions: seed.InteractionsList(),
	})
	return r
}

// parseBody is a test helper that decodes a JSON response body.
func parseBody(t *testing.T, body []byte) map[string]interface{} {
	t.Helper()
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("failed to parse JSON body: %v\nbody: %s", err, string(body))
	}
	return result
}

// ── Health Check ────────────────────────────────────────────────────

func TestHealthEndpoint(t *testing.T) {
	r := newTestRouter()

	req := httptest.NewRequest("GET", "/api/health", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("health: expected 200, got %d", w.Code)
	}

	body := parseBody(t, w.Body.Bytes())
	if body["status"] != "ok" {
		t.Errorf("expected status=ok, got %v", body["status"])
	}
}

// ── Planet Handlers ─────────────────────────────────────────────────

func TestListPlanets(t *testing.T) {
	r := newTestRouter()

	req := httptest.NewRequest("GET", "/api/planets", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	body := parseBody(t, w.Body.Bytes())
	data, ok := body["data"].([]interface{})
	if !ok {
		t.Fatal("expected data to be an array")
	}
	if len(data) != 6 {
		t.Errorf("expected 6 planets, got %d", len(data))
	}
}

func TestGetPlanet_WithIngredients(t *testing.T) {
	r := newTestRouter()

	req := httptest.NewRequest("GET", "/api/planets/planet-verdanthia", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	body := parseBody(t, w.Body.Bytes())
	data, ok := body["data"].(map[string]interface{})
	if !ok {
		t.Fatal("expected data to be an object")
	}
	planet, ok := data["planet"].(map[string]interface{})
	if !ok {
		t.Fatal("expected planet key")
	}
	if planet["name"] != "Verdanthia" {
		t.Errorf("expected Verdanthia, got %v", planet["name"])
	}
	ingredients, ok := data["ingredients"].([]interface{})
	if !ok {
		t.Fatal("expected ingredients array")
	}
	if len(ingredients) == 0 {
		t.Error("expected ingredients for Verdanthia")
	}
}

func TestGetPlanet_NotFound(t *testing.T) {
	r := newTestRouter()

	req := httptest.NewRequest("GET", "/api/planets/planet-nonexistent", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

// ── Ingredient Handlers ─────────────────────────────────────────────

func TestListIngredients(t *testing.T) {
	r := newTestRouter()

	req := httptest.NewRequest("GET", "/api/ingredients", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	body := parseBody(t, w.Body.Bytes())
	data, ok := body["data"].([]interface{})
	if !ok {
		t.Fatal("expected data to be an array")
	}
	if len(data) < 10 {
		t.Errorf("expected at least 10 ingredients, got %d", len(data))
	}
}

func TestListIngredients_FilterByElement(t *testing.T) {
	r := newTestRouter()

	req := httptest.NewRequest("GET", "/api/ingredients?element=void", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	body := parseBody(t, w.Body.Bytes())
	data := body["data"].([]interface{})
	for _, item := range data {
		ing := item.(map[string]interface{})
		if ing["element"] != "void" {
			t.Errorf("expected void element, got %v", ing["element"])
		}
	}
}

func TestGetIngredient(t *testing.T) {
	r := newTestRouter()

	req := httptest.NewRequest("GET", "/api/ingredients/ing-solar-crystal", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestGetIngredient_NotFound(t *testing.T) {
	r := newTestRouter()

	req := httptest.NewRequest("GET", "/api/ingredients/ing-nonexistent", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

// ── Recipe Handlers ─────────────────────────────────────────────────

func TestListRecipes(t *testing.T) {
	r := newTestRouter()

	req := httptest.NewRequest("GET", "/api/recipes", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	body := parseBody(t, w.Body.Bytes())
	data := body["data"].([]interface{})
	if len(data) < 5 {
		t.Errorf("expected at least 5 recipes, got %d", len(data))
	}
}

func TestListRecipes_FilterByDifficulty(t *testing.T) {
	r := newTestRouter()

	req := httptest.NewRequest("GET", "/api/recipes?difficulty=novice", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	body := parseBody(t, w.Body.Bytes())
	data := body["data"].([]interface{})
	for _, item := range data {
		recipe := item.(map[string]interface{})
		if recipe["difficulty"] != "novice" {
			t.Errorf("expected novice difficulty, got %v", recipe["difficulty"])
		}
	}
}

func TestGetRecipe_NotFound(t *testing.T) {
	r := newTestRouter()

	req := httptest.NewRequest("GET", "/api/recipes/recipe-nonexistent", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

// ── Brew Handlers ───────────────────────────────────────────────────

func TestBrewEndpoint_Success(t *testing.T) {
	r := newTestRouter()

	body := `{"ingredientIds":["ing-solar-crystal","ing-luminous-spore"]}`
	req := httptest.NewRequest("POST", "/api/brew", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// fakeRand returns 0.0 → always success → 201 Created
	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d\nbody: %s", w.Code, w.Body.String())
	}

	resp := parseBody(t, w.Body.Bytes())
	data := resp["data"].(map[string]interface{})
	brew := data["brew"].(map[string]interface{})
	if brew["result"] != "success" {
		t.Errorf("expected success, got %v", brew["result"])
	}
	if brew["potionName"] == "" {
		t.Error("expected potion name")
	}
}

func TestBrewEndpoint_EmptyIngredients(t *testing.T) {
	r := newTestRouter()

	body := `{"ingredientIds":[]}`
	req := httptest.NewRequest("POST", "/api/brew", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestBrewEndpoint_InvalidJSON(t *testing.T) {
	r := newTestRouter()

	req := httptest.NewRequest("POST", "/api/brew", strings.NewReader("not json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestBrewEndpoint_UnknownIngredient(t *testing.T) {
	r := newTestRouter()

	body := `{"ingredientIds":["ing-nonexistent","ing-solar-crystal"]}`
	req := httptest.NewRequest("POST", "/api/brew", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestBrewEndpoint_TooManyIngredients(t *testing.T) {
	r := newTestRouter()

	body := `{"ingredientIds":["ing-solar-crystal","ing-luminous-spore","ing-drift-dust","ing-magma-pearl","ing-abyssal-pearl"]}`
	req := httptest.NewRequest("POST", "/api/brew", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected 422, got %d", w.Code)
	}
}

// ── Brew Check Handler ──────────────────────────────────────────────

func TestBrewCheckEndpoint(t *testing.T) {
	r := newTestRouter()

	body := `{"ingredientIds":["ing-solar-crystal","ing-luminous-spore"]}`
	req := httptest.NewRequest("POST", "/api/brew/check", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	resp := parseBody(t, w.Body.Bytes())
	data := resp["data"].(map[string]interface{})
	if _, ok := data["successChance"]; !ok {
		t.Error("expected successChance in response")
	}
}

func TestBrewCheckEndpoint_EmptyIngredients(t *testing.T) {
	r := newTestRouter()

	body := `{"ingredientIds":[]}`
	req := httptest.NewRequest("POST", "/api/brew/check", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

// ── Brew List Handler ───────────────────────────────────────────────

func TestBrewListEndpoint(t *testing.T) {
	r := newTestRouter()

	req := httptest.NewRequest("GET", "/api/brews", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	resp := parseBody(t, w.Body.Bytes())
	data := resp["data"].([]interface{})
	// No brews saved yet → empty array
	if len(data) != 0 {
		t.Errorf("expected empty brews list, got %d", len(data))
	}
}

// ── Effects Handler ─────────────────────────────────────────────────

func TestEffectsEndpoint(t *testing.T) {
	r := newTestRouter()

	req := httptest.NewRequest("GET", "/api/effects", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	resp := parseBody(t, w.Body.Bytes())
	data := resp["data"].([]interface{})
	if len(data) != 28 {
		t.Errorf("expected 28 effects, got %d", len(data))
	}

	// Verify each effect has required fields
	for i, item := range data {
		eff := item.(map[string]interface{})
		if eff["id"] == nil || eff["id"] == "" {
			t.Errorf("effect %d missing id", i)
		}
		if eff["name"] == nil || eff["name"] == "" {
			t.Errorf("effect %d missing name", i)
		}
		if eff["tier"] == nil || eff["tier"] == "" {
			t.Errorf("effect %d missing tier", i)
		}
	}
}

// ── Interactions Handler ────────────────────────────────────────────

func TestInteractionsEndpoint(t *testing.T) {
	r := newTestRouter()

	req := httptest.NewRequest("GET", "/api/interactions", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	resp := parseBody(t, w.Body.Bytes())
	data := resp["data"].([]interface{})
	// 5×5 matrix = 25 interactions
	if len(data) != 25 {
		t.Errorf("expected 25 interactions, got %d", len(data))
	}
}

// ── Response Wrapper Tests ──────────────────────────────────────────

func TestWriteJSON_ContentType(t *testing.T) {
	w := httptest.NewRecorder()
	writeJSON(w, http.StatusOK, map[string]string{"key": "value"})

	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected application/json, got %s", ct)
	}
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestWriteError_Format(t *testing.T) {
	w := httptest.NewRecorder()
	writeError(w, http.StatusNotFound, "planet not found")

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}

	resp := parseBody(t, w.Body.Bytes())
	errObj := resp["error"].(map[string]interface{})
	if errObj["code"].(float64) != 404 {
		t.Errorf("expected error code 404, got %v", errObj["code"])
	}
	if errObj["message"] != "planet not found" {
		t.Errorf("expected 'planet not found', got %v", errObj["message"])
	}
}

// Suppress unused import warnings — these types are used in test setup.
var _ model.Effect
var _ store.PlanetStore
