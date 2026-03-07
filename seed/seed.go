// Package seed provides the initial game data for Cosmic Potions.
//
// GO CONCEPT: Separate Package for Data
// Seed data lives in its own package so it can be imported cleanly by both
// the in-memory store and future SQLite migrations without creating circular
// dependencies. In Go, circular imports are a compile error (not just a warning),
// so package boundaries matter more than in TypeScript/Node.
package seed

import (
	"time"

	"github.com/LilyEssence/cosmic-potions/internal/model"
)

// GO CONCEPT: Package-Level Variables
// Variables declared outside of any function are "package-level" — they're
// initialized once when the package is first imported and live for the entire
// program lifetime. Think of them like module-level constants.
//
// We're using var (not const) because Go constants can only be primitive types
// (strings, numbers, bools). Structs, slices, and maps must use var.

// Planets returns the seed planets for the galaxy.
//
// GO CONCEPT: Functions That Return Data
// Rather than exposing raw package variables (which could be accidentally mutated),
// we use functions that return fresh slices. Each call returns a new copy, so
// callers can't corrupt the seed data.
func Planets() []model.Planet {
	// GO CONCEPT: Composite Literals
	// The `model.Planet{...}` syntax is a "composite literal" — Go's way of
	// constructing struct values inline. Unlike constructors in OOP languages,
	// structs are just values — no `new` keyword needed (though `new()` does exist).
	// Fields you don't specify get their "zero value" (0 for ints, "" for strings, etc.)
	return []model.Planet{
		{
			ID:           "planet-verdanthia",
			Name:         "Verdanthia",
			System:       "Emerald Reach",
			Biome:        model.BiomeFungal,
			Description:  "A world of towering mushroom forests and bioluminescent spore clouds. The air itself glows faintly green at dusk, and the ground pulses with mycorrhizal networks that connect every living thing.",
			DiscoveredAt: time.Date(2847, 3, 15, 0, 0, 0, 0, time.UTC),
		},
		{
			ID:           "planet-obsidian-drift",
			Name:         "The Obsidian Drift",
			System:       "Null Expanse",
			Biome:        model.BiomeVoid,
			Description:  "Not a planet so much as a wound in space — floating shards of dark matter orbit a gravitational anomaly. Harvesters must tether themselves to avoid being pulled into the nothing between the shards.",
			DiscoveredAt: time.Date(2901, 11, 2, 0, 0, 0, 0, time.UTC),
		},
		{
			ID:           "planet-heliox-prime",
			Name:         "Heliox Prime",
			System:       "Solari Belt",
			Biome:        model.BiomeVolcanic,
			Description:  "A world of molten rivers and solar flare geysers. The surface is a mosaic of cooling obsidian plates floating on magma seas. Ingredients here are forged in heat that melts starship hulls.",
			DiscoveredAt: time.Date(2756, 7, 22, 0, 0, 0, 0, time.UTC),
		},
		{
			ID:           "planet-crystara",
			Name:         "Crystara",
			System:       "Prism Nebula",
			Biome:        model.BiomeCrystalline,
			Description:  "Living crystal formations that grow, fracture, and sing in harmonic frequencies. Alchemists report hearing their deepest memories echoed back while harvesting here. The crystals remember everything.",
			DiscoveredAt: time.Date(2812, 1, 8, 0, 0, 0, 0, time.UTC),
		},
		{
			ID:           "planet-thalassia",
			Name:         "Thalassia",
			System:       "Deep Ring",
			Biome:        model.BiomeOceanic,
			Description:  "An entirely submerged world with no surface. Pressure-forged pearls nestle in abyssal caverns, and the water itself carries dissolved minerals that glow under UV light. Harvesting requires deep-dive suits rated for 400 atmospheres.",
			DiscoveredAt: time.Date(2834, 9, 30, 0, 0, 0, 0, time.UTC),
		},
		{
			ID:           "planet-pyraxis",
			Name:         "Pyraxis",
			System:       "Ember Chain",
			Biome:        model.BiomeVolcanic,
			Description:  "Where plasma storms scour glass deserts into razor-sharp dunes. Lightning strikes the sand so frequently that the surface is a lattice of fulgurite tubes — petrified lightning, endlessly shattered and reformed.",
			DiscoveredAt: time.Date(2889, 5, 14, 0, 0, 0, 0, time.UTC),
		},
	}
}

// Ingredients returns the seed ingredients spread across planets and elements.
func Ingredients() []model.Ingredient {
	return []model.Ingredient{
		// ── Verdanthia (Fungal) ──────────────────────────────────────
		{
			ID:          "ing-luminous-spore",
			PlanetID:    "planet-verdanthia",
			Name:        "Luminous Sporecap",
			Element:     model.ElementOrganic,
			Potency:     4,
			Rarity:      model.RarityCommon,
			Description: "A gently glowing mushroom cap that releases clouds of bioluminescent spores when disturbed. Smells like warm bread and ozone.",
			FlavorNotes: []string{"luminous", "earthy", "sweet"},
		},
		{
			ID:          "ing-mycelial-thread",
			PlanetID:    "planet-verdanthia",
			Name:        "Mycelial Heartthread",
			Element:     model.ElementOrganic,
			Potency:     7,
			Rarity:      model.RarityRare,
			Description: "A strand of the planet's central mycorrhizal network, carefully extracted without severing the connection. Pulses with faint bio-electricity.",
			FlavorNotes: []string{"electric", "bitter", "alive"},
		},
		{
			ID:          "ing-dewdrop-fungus",
			PlanetID:    "planet-verdanthia",
			Name:        "Morning Dew Fungus",
			Element:     model.ElementOrganic,
			Potency:     2,
			Rarity:      model.RarityCommon,
			Description: "Tiny translucent mushrooms that only fruit at dawn, collecting dew in their concave caps. Each drop contains trace enzymes unique to Verdanthia.",
			FlavorNotes: []string{"sweet", "delicate", "cool"},
		},

		// ── The Obsidian Drift (Void) ───────────────────────────────
		{
			ID:          "ing-null-shard",
			PlanetID:    "planet-obsidian-drift",
			Name:        "Null Shard",
			Element:     model.ElementVoid,
			Potency:     8,
			Rarity:      model.RarityRare,
			Description: "A sliver of crystallized dark matter that absorbs light. Holding it feels like gripping a hole in reality. Alchemists report brief out-of-body experiences during extraction.",
			FlavorNotes: []string{"absence", "cold", "numbing"},
		},
		{
			ID:          "ing-drift-dust",
			PlanetID:    "planet-obsidian-drift",
			Name:        "Drift Dust",
			Element:     model.ElementVoid,
			Potency:     3,
			Rarity:      model.RarityUncommon,
			Description: "Fine particles that float between the shards, seemingly unaffected by gravity. Collected with electrostatic nets. Faintly hums at frequencies below human hearing.",
			FlavorNotes: []string{"hollow", "electric", "weightless"},
		},
		{
			ID:          "ing-entropy-bloom",
			PlanetID:    "planet-obsidian-drift",
			Name:        "Entropy Bloom",
			Element:     model.ElementVoid,
			Potency:     9,
			Rarity:      model.RarityLegendary,
			Description: "A flower-like structure that grows on the boundary between void shards. It appears to age backwards — blooming into a bud, then a seed, then nothing. Each one exists for exactly 7 hours.",
			FlavorNotes: []string{"temporal", "bitter", "infinite"},
		},

		// ── Heliox Prime (Volcanic/Solar) ───────────────────────────
		{
			ID:          "ing-solar-crystal",
			PlanetID:    "planet-heliox-prime",
			Name:        "Solar Crystal",
			Element:     model.ElementSolar,
			Potency:     6,
			Rarity:      model.RarityUncommon,
			Description: "Formed when solar flares strike cooling magma at exactly the right angle. Warm to the touch even in deep space. Emits a faint golden glow that intensifies near other solar ingredients.",
			FlavorNotes: []string{"warm", "luminous", "sharp"},
		},
		{
			ID:          "ing-magma-pearl",
			PlanetID:    "planet-heliox-prime",
			Name:        "Magma Pearl",
			Element:     model.ElementPlasma,
			Potency:     5,
			Rarity:      model.RarityCommon,
			Description: "Round nodules formed in magma currents, polished by centuries of flowing lava. Each one contains a tiny pocket of superheated gas that pops satisfyingly when cracked.",
			FlavorNotes: []string{"smoky", "spicy", "dense"},
		},
		{
			ID:          "ing-flare-essence",
			PlanetID:    "planet-heliox-prime",
			Name:        "Flare Essence",
			Element:     model.ElementSolar,
			Potency:     9,
			Rarity:      model.RarityLegendary,
			Description: "Bottled solar flare plasma, captured at the peak of a geyser eruption. The containment vial must be replaced every 48 hours or the essence burns through. Blindingly bright.",
			FlavorNotes: []string{"blinding", "electric", "overwhelming"},
		},

		// ── Crystara (Crystalline) ──────────────────────────────────
		{
			ID:          "ing-resonance-prism",
			PlanetID:    "planet-crystara",
			Name:        "Resonance Prism",
			Element:     model.ElementCrystalline,
			Potency:     6,
			Rarity:      model.RarityUncommon,
			Description: "A naturally occurring prism that splits not just light but sound. Tap it and you hear three distinct tones in harmony. Used to tune other ingredients before brewing.",
			FlavorNotes: []string{"harmonic", "clear", "bright"},
		},
		{
			ID:          "ing-memory-quartz",
			PlanetID:    "planet-crystara",
			Name:        "Memory Quartz",
			Element:     model.ElementCrystalline,
			Potency:     8,
			Rarity:      model.RarityRare,
			Description: "Crystals that absorb and replay the memories of those who touch them. Harvesters wear thick gloves — not for protection, but for privacy. The quartz doesn't discriminate.",
			FlavorNotes: []string{"nostalgic", "bittersweet", "luminous"},
		},
		{
			ID:          "ing-singing-dust",
			PlanetID:    "planet-crystara",
			Name:        "Singing Dust",
			Element:     model.ElementCrystalline,
			Potency:     3,
			Rarity:      model.RarityCommon,
			Description: "Powdered crystal that vibrates at a constant frequency. Pour it on a surface and it self-organizes into geometric patterns. Children on Crystara use it like glitter.",
			FlavorNotes: []string{"tingling", "sweet", "resonant"},
		},

		// ── Thalassia (Oceanic) ─────────────────────────────────────
		{
			ID:          "ing-abyssal-pearl",
			PlanetID:    "planet-thalassia",
			Name:        "Abyssal Pearl",
			Element:     model.ElementOrganic,
			Potency:     7,
			Rarity:      model.RarityRare,
			Description: "Forged at 400 atmospheres of pressure, these pearls contain dissolved minerals that glow an eerie blue under UV light. Each one takes a century to form.",
			FlavorNotes: []string{"briny", "luminous", "deep"},
		},
		{
			ID:          "ing-pressure-salt",
			PlanetID:    "planet-thalassia",
			Name:        "Pressure Salt",
			Element:     model.ElementCrystalline,
			Potency:     4,
			Rarity:      model.RarityCommon,
			Description: "Crystallized minerals from the deepest trenches, compressed into impossibly dense cubes. A single grain seasons a gallon. Dissolves with a satisfying snap.",
			FlavorNotes: []string{"sharp", "briny", "electric"},
		},
		{
			ID:          "ing-tidal-essence",
			PlanetID:    "planet-thalassia",
			Name:        "Tidal Essence",
			Element:     model.ElementPlasma,
			Potency:     5,
			Rarity:      model.RarityUncommon,
			Description: "Liquid captured during the convergence of Thalassia's three moons, when tidal forces create underwater plasma discharges. Sloshes with visible electric arcs.",
			FlavorNotes: []string{"electric", "salty", "turbulent"},
		},

		// ── Pyraxis (Volcanic/Plasma) ───────────────────────────────
		{
			ID:          "ing-fulgurite-spike",
			PlanetID:    "planet-pyraxis",
			Name:        "Fulgurite Spike",
			Element:     model.ElementPlasma,
			Potency:     7,
			Rarity:      model.RarityUncommon,
			Description: "Petrified lightning — a glass tube formed when plasma strikes sand at millions of degrees. Each spike is unique, branching like frozen roots. Crackles faintly when held.",
			FlavorNotes: []string{"electric", "sharp", "smoky"},
		},
		{
			ID:          "ing-storm-glass",
			PlanetID:    "planet-pyraxis",
			Name:        "Storm Glass",
			Element:     model.ElementPlasma,
			Potency:     10,
			Rarity:      model.RarityLegendary,
			Description: "A sphere of volcanic glass with a permanent plasma storm trapped inside. The lightning never stops. It's unclear whether the storm was captured or whether it grew there. Either answer is unsettling.",
			FlavorNotes: []string{"electric", "overwhelming", "alive"},
		},
		{
			ID:          "ing-glass-sand",
			PlanetID:    "planet-pyraxis",
			Name:        "Glass Sand",
			Element:     model.ElementSolar,
			Potency:     2,
			Rarity:      model.RarityCommon,
			Description: "Fine grains of naturally fused silica that still hold residual heat from plasma strikes. Warm to the touch and slightly magnetic. Sparkles like powdered gold.",
			FlavorNotes: []string{"warm", "gritty", "luminous"},
		},
	}
}

// Recipes returns known potion formulas.
func Recipes() []model.Recipe {
	return []model.Recipe{
		{
			ID:          "recipe-starlight-tonic",
			Name:        "Starlight Tonic",
			Description: "A luminous golden liquid that grants temporary enhanced vision in darkness. Side effects include seeing constellations on the inside of your eyelids for a few hours.",
			Difficulty:  model.DifficultyNovice,
			Effects:     []string{"night vision", "enhanced perception", "mild euphoria"},
			Ingredients: []model.RecipeIngredient{
				{IngredientID: "ing-solar-crystal", Quantity: 1},
				{IngredientID: "ing-luminous-spore", Quantity: 2},
			},
		},
		{
			ID:          "recipe-void-whisper",
			Name:        "Void Whisper Elixir",
			Description: "An inky black draught that lets the drinker hear transmissions from deep space for a brief window. Most hear cosmic background radiation as music. Some hear words.",
			Difficulty:  model.DifficultyAdept,
			Effects:     []string{"cosmic hearing", "zero-gravity tolerance", "temporary translucency"},
			Ingredients: []model.RecipeIngredient{
				{IngredientID: "ing-null-shard", Quantity: 1},
				{IngredientID: "ing-drift-dust", Quantity: 2},
				{IngredientID: "ing-memory-quartz", Quantity: 1},
			},
		},
		{
			ID:          "recipe-crystal-resonance",
			Name:        "Crystal Resonance Draught",
			Description: "A shimmering prismatic potion that temporarily attunes the drinker to crystalline frequencies. Walls become semi-transparent, hidden structures reveal themselves.",
			Difficulty:  model.DifficultyNovice,
			Effects:     []string{"structural sight", "harmonic sensitivity"},
			Ingredients: []model.RecipeIngredient{
				{IngredientID: "ing-resonance-prism", Quantity: 1},
				{IngredientID: "ing-singing-dust", Quantity: 3},
			},
		},
		{
			ID:          "recipe-pressure-forge",
			Name:        "Pressure Forge Serum",
			Description: "A dense, heavy liquid that temporarily grants the drinker resistance to extreme pressure and gravity. Essential for deep-space salvage operations.",
			Difficulty:  model.DifficultyAdept,
			Effects:     []string{"pressure resistance", "gravity resistance", "bone density boost"},
			Ingredients: []model.RecipeIngredient{
				{IngredientID: "ing-abyssal-pearl", Quantity: 1},
				{IngredientID: "ing-pressure-salt", Quantity: 2},
				{IngredientID: "ing-magma-pearl", Quantity: 1},
			},
		},
		{
			ID:          "recipe-temporal-dew",
			Name:        "Temporal Morning Dew",
			Description: "A delicate, almost water-clear potion that slows the drinker's perception of time. Ten minutes feel like an hour. Popular with artists and during boring diplomatic functions.",
			Difficulty:  model.DifficultyNovice,
			Effects:     []string{"time dilation perception", "enhanced focus", "calm"},
			Ingredients: []model.RecipeIngredient{
				{IngredientID: "ing-dewdrop-fungus", Quantity: 3},
				{IngredientID: "ing-singing-dust", Quantity: 1},
			},
		},
		{
			ID:          "recipe-plasma-storm",
			Name:        "Plasma Storm Concentrate",
			Description: "A violently crackling orange-white potion that must be consumed within seconds of brewing or it burns through any container. Grants brief but extraordinary electrical manipulation.",
			Difficulty:  model.DifficultyMaster,
			Effects:     []string{"electrical manipulation", "plasma generation", "temporary invulnerability to heat"},
			Ingredients: []model.RecipeIngredient{
				{IngredientID: "ing-storm-glass", Quantity: 1},
				{IngredientID: "ing-fulgurite-spike", Quantity: 2},
				{IngredientID: "ing-flare-essence", Quantity: 1},
			},
		},
		{
			ID:          "recipe-entropy-cascade",
			Name:        "Entropy Cascade",
			Description: "The most dangerous known potion. It temporarily reverses local entropy — broken things repair, fires un-burn, wounds un-happen. The hangover involves living the next hour in reverse.",
			Difficulty:  model.DifficultyMaster,
			Effects:     []string{"local entropy reversal", "temporal echo", "reality instability"},
			Ingredients: []model.RecipeIngredient{
				{IngredientID: "ing-entropy-bloom", Quantity: 1},
				{IngredientID: "ing-null-shard", Quantity: 1},
				{IngredientID: "ing-memory-quartz", Quantity: 1},
				{IngredientID: "ing-tidal-essence", Quantity: 1},
			},
		},
		{
			ID:          "recipe-biolume-balm",
			Name:        "Bioluminescent Balm",
			Description: "A soft green salve that makes the wearer's skin glow faintly for several hours. Originally developed for Verdanthian cave navigation, now fashionable at interstellar galas.",
			Difficulty:  model.DifficultyNovice,
			Effects:     []string{"bioluminescence", "skin regeneration", "mild warmth"},
			Ingredients: []model.RecipeIngredient{
				{IngredientID: "ing-luminous-spore", Quantity: 2},
				{IngredientID: "ing-dewdrop-fungus", Quantity: 2},
				{IngredientID: "ing-glass-sand", Quantity: 1},
			},
		},
	}
}

// Interactions returns the element interaction matrix.
//
// GO CONCEPT: Maps
// Go maps are like JavaScript/TypeScript objects or Map<K, V>. The type
// map[model.Element]map[model.Element]model.InteractionResult is a nested map:
// outer key is Element1, inner key is Element2, value is the result.
// Unlike TypeScript objects, Go maps must be explicitly initialized with make()
// or a composite literal — using a nil map for reads returns the zero value,
// but writing to a nil map panics.
func Interactions() map[model.Element]map[model.Element]model.InteractionResult {
	return map[model.Element]map[model.Element]model.InteractionResult{
		model.ElementSolar: {
			model.ElementSolar:      model.InteractionNeutral,
			model.ElementVoid:       model.InteractionVolatile,
			model.ElementCrystalline: model.InteractionSynergy,
			model.ElementOrganic:    model.InteractionSynergy,
			model.ElementPlasma:     model.InteractionNeutral,
		},
		model.ElementVoid: {
			model.ElementSolar:      model.InteractionVolatile,
			model.ElementVoid:       model.InteractionNeutral,
			model.ElementCrystalline: model.InteractionNeutral,
			model.ElementOrganic:    model.InteractionVolatile,
			model.ElementPlasma:     model.InteractionSynergy,
		},
		model.ElementCrystalline: {
			model.ElementSolar:      model.InteractionSynergy,
			model.ElementVoid:       model.InteractionNeutral,
			model.ElementCrystalline: model.InteractionNeutral,
			model.ElementOrganic:    model.InteractionNeutral,
			model.ElementPlasma:     model.InteractionVolatile,
		},
		model.ElementOrganic: {
			model.ElementSolar:      model.InteractionSynergy,
			model.ElementVoid:       model.InteractionVolatile,
			model.ElementCrystalline: model.InteractionNeutral,
			model.ElementOrganic:    model.InteractionNeutral,
			model.ElementPlasma:     model.InteractionSynergy,
		},
		model.ElementPlasma: {
			model.ElementSolar:      model.InteractionNeutral,
			model.ElementVoid:       model.InteractionSynergy,
			model.ElementCrystalline: model.InteractionVolatile,
			model.ElementOrganic:    model.InteractionSynergy,
			model.ElementPlasma:     model.InteractionNeutral,
		},
	}
}

// InteractionsList returns the interaction matrix as a flat slice of Interaction
// structs, which is more convenient for JSON serialization in API responses.
//
// GO CONCEPT: Multiple Return Styles
// We provide both map (fast lookups in the brew engine) and slice (easy to
// serialize) representations of the same data. This is idiomatic — functions
// are cheap, and having the right shape for each consumer is better than forcing
// one shape everywhere.
func InteractionsList() []model.Interaction {
	matrix := Interactions()
	var result []model.Interaction

	// GO CONCEPT: Range
	// `range` iterates over slices, maps, strings, and channels. For maps,
	// it yields key-value pairs. The iteration order of maps is intentionally
	// randomized in Go (to prevent code from depending on ordering) — if you
	// need consistent order, sort the keys first.
	for elem1, inner := range matrix {
		for elem2, outcome := range inner {
			result = append(result, model.Interaction{
				Element1: elem1,
				Element2: elem2,
				Result:   outcome,
			})
		}
	}
	return result
}
