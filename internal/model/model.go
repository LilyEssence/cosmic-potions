// Package model defines the core domain types for Cosmic Potions.
//
// GO CONCEPT: Packages
// Every Go file starts with a "package" declaration. A package is Go's unit of
// code organization — similar to a module/namespace in other languages. All files
// in the same directory must declare the same package name.
//
// The package name is how other code imports and references these types:
//
//	import "github.com/LilyEssence/cosmic-potions/internal/model"
//	planet := model.Planet{Name: "Verdanthia"}
//
// The "internal/" path is special in Go: packages under internal/ can only be
// imported by code within the same module. It's Go's built-in way of saying
// "these are private implementation details, not a public API."
package model

import "time"

// GO CONCEPT: Structs
// A struct is Go's primary way to group related data — like a TypeScript interface
// or a class without methods. Unlike classes, structs have no inheritance. Instead,
// Go uses composition (embedding one struct inside another) and interfaces.
//
// GO CONCEPT: Struct Tags (the `json:"..."` bits)
// The backtick strings after each field are "struct tags" — metadata that libraries
// can read at runtime. The `json:"name"` tag tells Go's JSON encoder/decoder what
// key name to use. Without tags, it would use the Go field name (which is PascalCase
// by convention, but APIs typically want camelCase or snake_case).
//
// GO CONCEPT: Exported vs Unexported
// In Go, capitalization IS the access modifier:
//   - Planet (capital P) → exported (public) — visible outside the package
//   - planet (lowercase p) → unexported (private) — only visible within the package
// This applies to struct fields too: Name is public, name would be private.

// Planet represents an origin world where ingredients are harvested.
type Planet struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	System       string    `json:"system"`
	Biome        Biome     `json:"biome"`
	Description  string    `json:"description"`
	DiscoveredAt time.Time `json:"discoveredAt"`
}

// Biome represents the environmental type of a planet.
//
// GO CONCEPT: Custom Types
// Go lets you create named types based on existing ones. `type Biome string`
// creates a new type called Biome that is backed by a string. Even though it's
// "just a string" under the hood, the compiler treats it as a distinct type —
// you can't accidentally pass a random string where a Biome is expected without
// an explicit conversion. This gives you type safety without the overhead of a
// full struct.
type Biome string

// GO CONCEPT: Constants
// `const` declares compile-time constants. The `iota`-less approach below simply
// assigns string values. These act like a TypeScript string union type or an enum,
// but Go doesn't enforce that only these values can be used — validation is your
// responsibility (we'll add that in the brewing engine).
const (
	BiomeVolcanic    Biome = "volcanic"
	BiomeCrystalline Biome = "crystalline"
	BiomeOceanic     Biome = "oceanic"
	BiomeFungal      Biome = "fungal"
	BiomeVoid        Biome = "void"
)

// Ingredient represents a harvestable material from a planet.
type Ingredient struct {
	ID          string   `json:"id"`
	PlanetID    string   `json:"planetId"`
	Name        string   `json:"name"`
	Element     Element  `json:"element"`
	Potency     int      `json:"potency"` // 1-10
	Rarity      Rarity   `json:"rarity"`
	Description string   `json:"description"`
	FlavorNotes []string `json:"flavorNotes"` // e.g., "bitter", "luminous", "electric"
}

// Element represents the elemental classification of an ingredient.
// The five elements interact with each other (synergy, neutral, or volatile).
type Element string

const (
	ElementSolar       Element = "solar"
	ElementVoid        Element = "void"
	ElementCrystalline Element = "crystalline"
	ElementOrganic     Element = "organic"
	ElementPlasma      Element = "plasma"
)

// Rarity indicates how rare an ingredient is.
type Rarity string

const (
	RarityCommon    Rarity = "common"
	RarityUncommon  Rarity = "uncommon"
	RarityRare      Rarity = "rare"
	RarityLegendary Rarity = "legendary"
)

// Recipe represents a known potion formula.
type Recipe struct {
	ID          string             `json:"id"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Difficulty  Difficulty         `json:"difficulty"`
	Effects     []string           `json:"effects"`
	Ingredients []RecipeIngredient `json:"ingredients"`
}

// RecipeIngredient specifies an ingredient and how much of it a recipe needs.
type RecipeIngredient struct {
	IngredientID string `json:"ingredientId"`
	Quantity     int    `json:"quantity"`
}

// Difficulty represents the skill level required for a recipe.
type Difficulty string

const (
	DifficultyNovice Difficulty = "novice"
	DifficultyAdept  Difficulty = "adept"
	DifficultyMaster Difficulty = "master"
)

// Brew represents the result of attempting to make a potion.
//
// GO CONCEPT: Pointer Fields
// The `*string` type for RecipeID means "pointer to a string" — it can be nil.
// Regular Go strings can't be nil (their zero value is ""), but sometimes you need
// to distinguish "no value" from "empty string". A pointer lets RecipeID be null
// in JSON when the brew was experimental (not following a known recipe).
//
// In TypeScript terms: `string` in Go ≈ `string` (always has a value),
// while `*string` in Go ≈ `string | null`.
type Brew struct {
	ID              string     `json:"id"`
	RecipeID        *string    `json:"recipeId"`        // nil if experimental
	IngredientsUsed []string   `json:"ingredientsUsed"` // ingredient IDs
	Result          BrewResult `json:"result"`
	PotionName      string     `json:"potionName"`
	Effects         []string   `json:"effects"`
	Notes           string     `json:"notes"`
	BrewedAt        time.Time  `json:"brewedAt"`
}

// BrewResult describes the outcome of a brewing attempt.
type BrewResult string

const (
	BrewResultSuccess             BrewResult = "success"
	BrewResultPartial             BrewResult = "partial"
	BrewResultCatastrophicFailure BrewResult = "catastrophic_failure"
)

// Effect represents a discoverable potion effect in the Cosmic Alchemist's Codex.
// Players fill their codex by discovering effects through experimentation.
// The Go API serves the master effect list; player progress lives in frontend localStorage.
type Effect struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Tier        EffectTier `json:"tier"`
	Description string     `json:"description"`
}

// EffectTier indicates how difficult an effect is to discover through brewing.
// The tiers form a natural progression — common effects appear almost accidentally,
// while legendary effects demand mastery of dangerous volatile element combinations.
type EffectTier string

const (
	EffectTierCommon    EffectTier = "common"
	EffectTierUncommon  EffectTier = "uncommon"
	EffectTierRare      EffectTier = "rare"
	EffectTierLegendary EffectTier = "legendary"
)

// Interaction defines how two elements react when combined.
type Interaction struct {
	Element1 Element           `json:"element1"`
	Element2 Element           `json:"element2"`
	Result   InteractionResult `json:"result"`
}

// InteractionResult describes the outcome type when two elements meet.
type InteractionResult string

const (
	InteractionSynergy  InteractionResult = "synergy"
	InteractionNeutral  InteractionResult = "neutral"
	InteractionVolatile InteractionResult = "volatile"
)

// GO CONCEPT: Filter Structs
// Rather than passing many optional parameters to a function (Go has no optional
// params or function overloading), a common pattern is to pass a "filter" or
// "options" struct. Pointer fields (*Element, *Rarity) let us distinguish "not
// specified" (nil) from a specific value.

// IngredientFilter holds optional criteria for filtering ingredient queries.
type IngredientFilter struct {
	PlanetID *string  `json:"planetId,omitempty"`
	Element  *Element `json:"element,omitempty"`
	Rarity   *Rarity  `json:"rarity,omitempty"`
}

// RecipeFilter holds optional criteria for filtering recipe queries.
// Follows the same pointer-for-optional pattern as IngredientFilter.
type RecipeFilter struct {
	Difficulty *Difficulty `json:"difficulty,omitempty"`
}

// PlanetWithIngredients bundles a planet with all its harvestable ingredients.
// This is a convenience type for the "get planet detail" endpoint that returns
// everything in one response rather than requiring two API calls.
type PlanetWithIngredients struct {
	Planet      Planet       `json:"planet"`
	Ingredients []Ingredient `json:"ingredients"`
}
