package brewing

import (
	"fmt"

	"github.com/LilyEssence/cosmic-potions/internal/model"
)

// ── Potion Name Generation ──────────────────────────────────────────
//
// GO CONCEPT: Maps of Slices
// Go maps can hold any value type, including slices. We use
// map[model.BrewResult][]string to associate each brew outcome with a pool
// of adjectives to randomly pick from. This is cleaner than a giant switch
// statement and easy to extend.

// resultAdjectives provides flavor words based on how the brew turned out.
var resultAdjectives = map[model.BrewResult][]string{
	model.BrewResultSuccess: {
		"Radiant", "Perfect", "Brilliant", "Masterful",
		"Pure", "Exquisite", "Sublime", "Flawless",
	},
	model.BrewResultPartial: {
		"Cloudy", "Unstable", "Flickering", "Murky",
		"Imperfect", "Hazy", "Wavering", "Dim",
	},
	model.BrewResultCatastrophicFailure: {
		"Cursed", "Chaotic", "Abysmal", "Ruined",
		"Catastrophic", "Disastrous", "Wretched", "Doomed",
	},
}

// elementNouns provides potion type names based on the dominant element.
var elementNouns = map[model.Element][]string{
	model.ElementSolar: {
		"Sunfire Elixir", "Dawn Draught", "Solar Serum",
		"Starlight Tonic", "Sun Essence", "Corona Philter",
	},
	model.ElementVoid: {
		"Null Essence", "Void Tincture", "Shadow Brew",
		"Darkness Draught", "Abyss Philter", "Emptiness Elixir",
	},
	model.ElementCrystalline: {
		"Crystal Philter", "Prism Potion", "Resonance Draught",
		"Gem Elixir", "Lattice Tonic", "Facet Serum",
	},
	model.ElementOrganic: {
		"Living Tonic", "Bloom Elixir", "Spore Serum",
		"Growth Draught", "Nature Philter", "Verdant Brew",
	},
	model.ElementPlasma: {
		"Storm Concentrate", "Lightning Draught", "Plasma Brew",
		"Thunder Elixir", "Spark Serum", "Arc Philter",
	},
}

// generatePotionName creates a thematic name from the brew result and
// dominant element. Uses the Randomizer to pick from name pools, so names
// are deterministic in tests but varied in production.
func generatePotionName(dominant model.Element, result model.BrewResult, rand Randomizer) string {
	adjectives := resultAdjectives[result]
	nouns := elementNouns[dominant]

	if len(adjectives) == 0 || len(nouns) == 0 {
		return "Mysterious Concoction"
	}

	adj := adjectives[rand.Intn(len(adjectives))]
	noun := nouns[rand.Intn(len(nouns))]

	return fmt.Sprintf("%s %s", adj, noun)
}

// ── Brew Notes Generation ───────────────────────────────────────────
//
// Flavor text describing what happened during the brew. These add personality
// to API responses and give the frontend fun text to display.

var successNotes = []string{
	"The potion glows with a steady, confident light. A textbook brew.",
	"Ingredients harmonize beautifully. The mixture settles into a luminous liquid.",
	"A satisfying hiss as the last ingredient dissolves. The color is remarkable.",
	"The cauldron hums with resonance. Every element found its place perfectly.",
	"Vapor rises in slow spirals. The scent is indescribable — but pleasant.",
	"A flash of chromatic light marks the final reaction. The result is pristine.",
}

var partialNotes = []string{
	"The brew bubbles uncertainly but stabilizes. Some effects are weaker than expected.",
	"A faint pop — something didn't quite take. The result is drinkable but imperfect.",
	"The mixture clouds briefly before partially clearing. Not your best work.",
	"One ingredient resists integration. The potion works, but with diminished potency.",
	"An uneven glow suggests incomplete catalysis. Serviceable, if inelegant.",
	"The expected color is... close. Close enough. Probably fine.",
}

var catastrophicNotes = []string{
	"The mixture turns black, screams briefly, and dissolves the container.",
	"A blinding flash — when your vision clears, the ingredients have become a small, angry cloud.",
	"The potion explodes in a shower of sparks. The ceiling has a new scorch mark.",
	"Everything liquefied instantly into a substance that smells like burnt ambition.",
	"The cauldron cracks. The liquid evaporates. The table is slightly on fire. So it goes.",
	"A vacuum implosion collapses the mixture into a single, disappointed droplet.",
	"The ingredients combine into a foam that grows, overflows, and absorbs your notes.",
	"Reality briefly inverts in a 3-foot radius. Your eyebrows are now on backwards.",
}

// generateNotes picks a flavor text note based on the brew outcome.
// For volatile brews that succeed, the notes acknowledge the risk taken.
func generateNotes(result model.BrewResult, synergies, volatiles int, rand Randomizer) string {
	var pool []string
	switch result {
	case model.BrewResultSuccess:
		pool = successNotes
	case model.BrewResultPartial:
		pool = partialNotes
	case model.BrewResultCatastrophicFailure:
		pool = catastrophicNotes
	}

	if len(pool) == 0 {
		return "Something happened."
	}

	note := pool[rand.Intn(len(pool))]

	// Add a bonus line for risky volatile successes — the player earned it
	if result == model.BrewResultSuccess && volatiles > 0 {
		note += " Against all odds, the volatile elements stabilized."
	}

	// Add a note about strong synergy
	if result == model.BrewResultSuccess && synergies >= 2 {
		note += " The synergies between elements produced exceptional results."
	}

	return note
}
