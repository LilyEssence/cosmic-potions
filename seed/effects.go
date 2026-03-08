package seed

import "github.com/LilyEssence/cosmic-potions/internal/model"

// Effects returns the master list of all discoverable potion effects for the
// Cosmic Alchemist's Codex. There are 28 effects across four tiers.
//
// This list defines what "100% completion" looks like. The Go API serves it
// via GET /api/effects so the frontend can render the codex grid. The brewing
// engine uses effect IDs to link its rules to these definitions.
func Effects() []model.Effect {
	return []model.Effect{
		// ── Common Effects (10) ──────────────────────────────────────
		// Easy to discover — appear with basic element combinations and
		// low potency thresholds. Most players find these in their first
		// few brews.
		{
			ID:          "effect-night-vision",
			Name:        "Night Vision",
			Tier:        model.EffectTierCommon,
			Description: "Enhanced ability to see in complete darkness. Stars appear brighter and shadows become navigable.",
		},
		{
			ID:          "effect-mild-euphoria",
			Name:        "Mild Euphoria",
			Tier:        model.EffectTierCommon,
			Description: "A pleasant, warm sense of contentment that lasts for hours. Side effects include spontaneous humming.",
		},
		{
			ID:          "effect-bioluminescence",
			Name:        "Bioluminescence",
			Tier:        model.EffectTierCommon,
			Description: "Skin emits a soft, ambient glow in the drinker's chosen color. Popular at interstellar galas.",
		},
		{
			ID:          "effect-enhanced-perception",
			Name:        "Enhanced Perception",
			Tier:        model.EffectTierCommon,
			Description: "Heightened awareness of fine details — read small text from across a room, spot micro-fractures in hulls.",
		},
		{
			ID:          "effect-inner-calm",
			Name:        "Inner Calm",
			Tier:        model.EffectTierCommon,
			Description: "Deep tranquility and mental clarity. The drinker becomes unflappable for several hours.",
		},
		{
			ID:          "effect-mild-warmth",
			Name:        "Mild Warmth",
			Tier:        model.EffectTierCommon,
			Description: "Comfortable internal heat that persists regardless of external temperature. Perfect for ice worlds.",
		},
		{
			ID:          "effect-skin-regeneration",
			Name:        "Skin Regeneration",
			Tier:        model.EffectTierCommon,
			Description: "Accelerated healing of surface wounds. Scars fade, burns cool, and paper cuts become a distant memory.",
		},
		{
			ID:          "effect-enhanced-focus",
			Name:        "Enhanced Focus",
			Tier:        model.EffectTierCommon,
			Description: "Laser-sharp concentration on a singular task. Distractions become literally invisible.",
		},
		{
			ID:          "effect-energy-boost",
			Name:        "Energy Boost",
			Tier:        model.EffectTierCommon,
			Description: "A burst of clean, sustained energy without jitters. Like caffeine but from the cosmos.",
		},
		{
			ID:          "effect-basic-shielding",
			Name:        "Basic Shielding",
			Tier:        model.EffectTierCommon,
			Description: "A faint protective aura that absorbs minor impacts. Won't stop a plasma bolt, but great for hail.",
		},

		// ── Uncommon Effects (8) ─────────────────────────────────────
		// Require specific element pairings or 3+ ingredient combinations.
		// Players discover these by experimenting with cross-element brews.
		{
			ID:          "effect-cosmic-hearing",
			Name:        "Cosmic Hearing",
			Tier:        model.EffectTierUncommon,
			Description: "Ability to perceive deep-space transmissions as audible sound. Most hear music in the cosmic background radiation.",
		},
		{
			ID:          "effect-zero-gravity-tolerance",
			Name:        "Zero-Gravity Tolerance",
			Tier:        model.EffectTierUncommon,
			Description: "Complete comfort and coordination in null gravity. No nausea, no disorientation, no floating into walls.",
		},
		{
			ID:          "effect-pressure-resistance",
			Name:        "Pressure Resistance",
			Tier:        model.EffectTierUncommon,
			Description: "Body withstands extreme pressure differentials. Essential for deep-ocean worlds and gas giant moons.",
		},
		{
			ID:          "effect-gravity-resistance",
			Name:        "Gravity Resistance",
			Tier:        model.EffectTierUncommon,
			Description: "Reduced effect of gravitational pull on the drinker's body. Walking becomes bouncing on high-G worlds.",
		},
		{
			ID:          "effect-bone-density-boost",
			Name:        "Bone Density Boost",
			Tier:        model.EffectTierUncommon,
			Description: "Skeletal structure becomes incredibly resilient. Useful for miners, explorers, and the clumsy.",
		},
		{
			ID:          "effect-harmonic-sensitivity",
			Name:        "Harmonic Sensitivity",
			Tier:        model.EffectTierUncommon,
			Description: "Perception of crystalline frequency patterns invisible to normal senses. The universe hums.",
		},
		{
			ID:          "effect-time-dilation",
			Name:        "Time Dilation Perception",
			Tier:        model.EffectTierUncommon,
			Description: "Subjective time slows dramatically — ten minutes feel like an hour. Popular with artists and overthinkers.",
		},
		{
			ID:          "effect-temporary-translucency",
			Name:        "Temporary Translucency",
			Tier:        model.EffectTierUncommon,
			Description: "Body becomes partially transparent. Useful for stealth, terrifying for medical professionals.",
		},

		// ── Rare Effects (6) ─────────────────────────────────────────
		// Require synergy interactions between specific element pairs, plus
		// high combined potency. These reward players who understand which
		// elements harmonize.
		{
			ID:          "effect-structural-sight",
			Name:        "Structural Sight",
			Tier:        model.EffectTierRare,
			Description: "See through solid matter to the underlying framework. Walls reveal their studs, planets reveal their cores.",
		},
		{
			ID:          "effect-electrical-manipulation",
			Name:        "Electrical Manipulation",
			Tier:        model.EffectTierRare,
			Description: "Control of nearby electrical fields. Charge devices by touching them, dim lights for dramatic effect.",
		},
		{
			ID:          "effect-plasma-generation",
			Name:        "Plasma Generation",
			Tier:        model.EffectTierRare,
			Description: "Ability to generate small plasma discharges from fingertips. Mostly used for campfire lighting.",
		},
		{
			ID:          "effect-heat-invulnerability",
			Name:        "Heat Invulnerability",
			Tier:        model.EffectTierRare,
			Description: "Complete immunity to extreme temperatures. Walk on lava flows and sunbathe on stellar surfaces.",
		},
		{
			ID:          "effect-entropy-reversal",
			Name:        "Entropy Reversal",
			Tier:        model.EffectTierRare,
			Description: "Local objects briefly reverse their decay. Rust retreats, cracks seal, wilted flowers stand tall.",
		},
		{
			ID:          "effect-telepathic-link",
			Name:        "Telepathic Link",
			Tier:        model.EffectTierRare,
			Description: "Brief mental connection with nearby beings. Thoughts arrive as feelings rather than words.",
		},

		// ── Legendary Effects (4) ────────────────────────────────────
		// The hardest to discover. Require volatile element combinations
		// that succeed — the player must brave catastrophic failure risk
		// to unlock these. Completing all four is a badge of mastery.
		{
			ID:          "effect-temporal-echo",
			Name:        "Temporal Echo",
			Tier:        model.EffectTierLegendary,
			Description: "Brief ability to perceive moments from the past layered onto the present. Walls remember who touched them.",
		},
		{
			ID:          "effect-cosmic-omniscience",
			Name:        "Cosmic Omniscience",
			Tier:        model.EffectTierLegendary,
			Description: "A fleeting, overwhelming awareness of everything in local space. Stars whisper their names to you.",
		},
		{
			ID:          "effect-stellar-rebirth",
			Name:        "Stellar Rebirth",
			Tier:        model.EffectTierLegendary,
			Description: "Complete cellular regeneration in seconds. Scars vanish, fatigue evaporates, years feel undone.",
		},
		{
			ID:          "effect-void-walking",
			Name:        "Void Walking",
			Tier:        model.EffectTierLegendary,
			Description: "Ability to step briefly into the space between dimensions. The void is cold, silent, and vast.",
		},
	}
}
