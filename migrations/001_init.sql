-- ============================================================================
-- Cosmic Potions — SQLite Schema + Seed Data
-- ============================================================================
-- This migration creates all tables and inserts seed data in a single file.
-- The migration runner tracks which migrations have been applied in a
-- schema_migrations table, so this only runs once per database file.
--
-- SQLITE CONCEPT: Foreign Keys
-- SQLite supports foreign keys but doesn't enforce them by default — you
-- must run `PRAGMA foreign_keys = ON` per connection. Our Go store does
-- this when opening the database. Foreign keys ensure referential integrity:
-- you can't insert an ingredient referencing a planet that doesn't exist.
--
-- SQLITE CONCEPT: TEXT for Everything
-- Unlike PostgreSQL (which has enums, arrays, UUID types), SQLite has only
-- five storage classes: NULL, INTEGER, REAL, TEXT, BLOB. Our custom types
-- (Element, Biome, Rarity) are stored as TEXT and validated by the Go layer,
-- not the database. This is a common SQLite pattern.

-- ── Schema Tracking ─────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS schema_migrations (
    version    INTEGER PRIMARY KEY,
    applied_at TEXT NOT NULL DEFAULT (datetime('now'))
);

-- ── Planets ─────────────────────────────────────────────────────────

CREATE TABLE planets (
    id            TEXT PRIMARY KEY,
    name          TEXT NOT NULL,
    system        TEXT NOT NULL,
    biome         TEXT NOT NULL,
    description   TEXT NOT NULL,
    discovered_at TEXT NOT NULL  -- ISO 8601 timestamp stored as TEXT
);

-- ── Ingredients ─────────────────────────────────────────────────────

CREATE TABLE ingredients (
    id           TEXT PRIMARY KEY,
    planet_id    TEXT NOT NULL REFERENCES planets(id),
    name         TEXT NOT NULL,
    element      TEXT NOT NULL,
    potency      INTEGER NOT NULL CHECK (potency BETWEEN 1 AND 10),
    rarity       TEXT NOT NULL,
    description  TEXT NOT NULL,
    flavor_notes TEXT NOT NULL  -- JSON array stored as TEXT, e.g. '["bitter","sweet"]'
);

-- ── Recipes ─────────────────────────────────────────────────────────

CREATE TABLE recipes (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL,
    description TEXT NOT NULL,
    difficulty  TEXT NOT NULL
);

CREATE TABLE recipe_effects (
    recipe_id TEXT NOT NULL REFERENCES recipes(id),
    effect    TEXT NOT NULL,
    PRIMARY KEY (recipe_id, effect)
);

CREATE TABLE recipe_ingredients (
    recipe_id     TEXT NOT NULL REFERENCES recipes(id),
    ingredient_id TEXT NOT NULL REFERENCES ingredients(id),
    quantity      INTEGER NOT NULL DEFAULT 1,
    PRIMARY KEY (recipe_id, ingredient_id)
);

-- ── Brews ───────────────────────────────────────────────────────────

CREATE TABLE brews (
    id          TEXT PRIMARY KEY,
    recipe_id   TEXT REFERENCES recipes(id),  -- nullable for experimental brews
    result      TEXT NOT NULL,
    potion_name TEXT NOT NULL,
    notes       TEXT NOT NULL DEFAULT '',
    brewed_at   TEXT NOT NULL  -- ISO 8601 timestamp
);

CREATE TABLE brew_ingredients (
    brew_id       TEXT NOT NULL REFERENCES brews(id),
    ingredient_id TEXT NOT NULL,
    PRIMARY KEY (brew_id, ingredient_id)
);

CREATE TABLE brew_effects (
    brew_id TEXT NOT NULL REFERENCES brews(id),
    effect  TEXT NOT NULL,
    PRIMARY KEY (brew_id, effect)
);

-- ── Interactions ────────────────────────────────────────────────────

CREATE TABLE interactions (
    element1 TEXT NOT NULL,
    element2 TEXT NOT NULL,
    result   TEXT NOT NULL,
    PRIMARY KEY (element1, element2)
);

-- ============================================================================
-- Seed Data
-- ============================================================================

-- ── Planets ─────────────────────────────────────────────────────────

INSERT INTO planets (id, name, system, biome, description, discovered_at) VALUES
('planet-verdanthia',    'Verdanthia',        'Emerald Reach', 'fungal',      'A world of towering mushroom forests and bioluminescent spore clouds. The air itself glows faintly green at dusk, and the ground pulses with mycorrhizal networks that connect every living thing.', '2847-03-15T00:00:00Z'),
('planet-obsidian-drift','The Obsidian Drift', 'Null Expanse',  'void',        'Not a planet so much as a wound in space — floating shards of dark matter orbit a gravitational anomaly. Harvesters must tether themselves to avoid being pulled into the nothing between the shards.', '2901-11-02T00:00:00Z'),
('planet-heliox-prime',  'Heliox Prime',      'Solari Belt',   'volcanic',    'A world of molten rivers and solar flare geysers. The surface is a mosaic of cooling obsidian plates floating on magma seas. Ingredients here are forged in heat that melts starship hulls.',     '2756-07-22T00:00:00Z'),
('planet-crystara',      'Crystara',          'Prism Nebula',  'crystalline', 'Living crystal formations that grow, fracture, and sing in harmonic frequencies. Alchemists report hearing their deepest memories echoed back while harvesting here. The crystals remember everything.',                              '2812-01-08T00:00:00Z'),
('planet-thalassia',     'Thalassia',         'Deep Ring',     'oceanic',     'An entirely submerged world with no surface. Pressure-forged pearls nestle in abyssal caverns, and the water itself carries dissolved minerals that glow under UV light. Harvesting requires deep-dive suits rated for 400 atmospheres.', '2834-09-30T00:00:00Z'),
('planet-pyraxis',       'Pyraxis',           'Ember Chain',   'volcanic',    'Where plasma storms scour glass deserts into razor-sharp dunes. Lightning strikes the sand so frequently that the surface is a lattice of fulgurite tubes — petrified lightning, endlessly shattered and reformed.',                      '2889-05-14T00:00:00Z');

-- ── Ingredients ─────────────────────────────────────────────────────

INSERT INTO ingredients (id, planet_id, name, element, potency, rarity, description, flavor_notes) VALUES
-- Verdanthia (Fungal)
('ing-luminous-spore',   'planet-verdanthia',     'Luminous Sporecap',      'organic',     4,  'common',    'A gently glowing mushroom cap that releases clouds of bioluminescent spores when disturbed. Smells like warm bread and ozone.',                                                        '["luminous","earthy","sweet"]'),
('ing-mycelial-thread',  'planet-verdanthia',     'Mycelial Heartthread',   'organic',     7,  'rare',      'A strand of the planet''s central mycorrhizal network, carefully extracted without severing the connection. Pulses with faint bio-electricity.',                                     '["electric","bitter","alive"]'),
('ing-dewdrop-fungus',   'planet-verdanthia',     'Morning Dew Fungus',     'organic',     2,  'common',    'Tiny translucent mushrooms that only fruit at dawn, collecting dew in their concave caps. Each drop contains trace enzymes unique to Verdanthia.',                                   '["sweet","delicate","cool"]'),
-- The Obsidian Drift (Void)
('ing-null-shard',       'planet-obsidian-drift', 'Null Shard',             'void',        8,  'rare',      'A sliver of crystallized dark matter that absorbs light. Holding it feels like gripping a hole in reality. Alchemists report brief out-of-body experiences during extraction.',      '["absence","cold","numbing"]'),
('ing-drift-dust',       'planet-obsidian-drift', 'Drift Dust',             'void',        3,  'uncommon',  'Fine particles that float between the shards, seemingly unaffected by gravity. Collected with electrostatic nets. Faintly hums at frequencies below human hearing.',               '["hollow","electric","weightless"]'),
('ing-entropy-bloom',    'planet-obsidian-drift', 'Entropy Bloom',          'void',        9,  'legendary', 'A flower-like structure that grows on the boundary between void shards. It appears to age backwards — blooming into a bud, then a seed, then nothing. Each one exists for exactly 7 hours.', '["temporal","bitter","infinite"]'),
-- Heliox Prime (Volcanic/Solar)
('ing-solar-crystal',    'planet-heliox-prime',   'Solar Crystal',          'solar',       6,  'uncommon',  'Formed when solar flares strike cooling magma at exactly the right angle. Warm to the touch even in deep space. Emits a faint golden glow that intensifies near other solar ingredients.', '["warm","luminous","sharp"]'),
('ing-magma-pearl',      'planet-heliox-prime',   'Magma Pearl',            'plasma',      5,  'common',    'Round nodules formed in magma currents, polished by centuries of flowing lava. Each one contains a tiny pocket of superheated gas that pops satisfyingly when cracked.',             '["smoky","spicy","dense"]'),
('ing-flare-essence',    'planet-heliox-prime',   'Flare Essence',          'solar',       9,  'legendary', 'Bottled solar flare plasma, captured at the peak of a geyser eruption. The containment vial must be replaced every 48 hours or the essence burns through. Blindingly bright.',       '["blinding","electric","overwhelming"]'),
-- Crystara (Crystalline)
('ing-resonance-prism',  'planet-crystara',       'Resonance Prism',        'crystalline', 6,  'uncommon',  'A naturally occurring prism that splits not just light but sound. Tap it and you hear three distinct tones in harmony. Used to tune other ingredients before brewing.',               '["harmonic","clear","bright"]'),
('ing-memory-quartz',    'planet-crystara',       'Memory Quartz',          'crystalline', 8,  'rare',      'Crystals that absorb and replay the memories of those who touch them. Harvesters wear thick gloves — not for protection, but for privacy. The quartz doesn''t discriminate.',       '["nostalgic","bittersweet","luminous"]'),
('ing-singing-dust',     'planet-crystara',       'Singing Dust',           'crystalline', 3,  'common',    'Powdered crystal that vibrates at a constant frequency. Pour it on a surface and it self-organizes into geometric patterns. Children on Crystara use it like glitter.',             '["tingling","sweet","resonant"]'),
-- Thalassia (Oceanic)
('ing-abyssal-pearl',    'planet-thalassia',      'Abyssal Pearl',          'organic',     7,  'rare',      'Forged at 400 atmospheres of pressure, these pearls contain dissolved minerals that glow an eerie blue under UV light. Each one takes a century to form.',                          '["briny","luminous","deep"]'),
('ing-pressure-salt',    'planet-thalassia',      'Pressure Salt',          'crystalline', 4,  'common',    'Crystallized minerals from the deepest trenches, compressed into impossibly dense cubes. A single grain seasons a gallon. Dissolves with a satisfying snap.',                      '["sharp","briny","electric"]'),
('ing-tidal-essence',    'planet-thalassia',      'Tidal Essence',          'plasma',      5,  'uncommon',  'Liquid captured during the convergence of Thalassia''s three moons, when tidal forces create underwater plasma discharges. Sloshes with visible electric arcs.',                   '["electric","salty","turbulent"]'),
-- Pyraxis (Volcanic/Plasma)
('ing-fulgurite-spike',  'planet-pyraxis',        'Fulgurite Spike',        'plasma',      7,  'uncommon',  'Petrified lightning — a glass tube formed when plasma strikes sand at millions of degrees. Each spike is unique, branching like frozen roots. Crackles faintly when held.',         '["electric","sharp","smoky"]'),
('ing-storm-glass',      'planet-pyraxis',        'Storm Glass',            'plasma',      10, 'legendary', 'A sphere of volcanic glass with a permanent plasma storm trapped inside. The lightning never stops. It''s unclear whether the storm was captured or whether it grew there. Either answer is unsettling.', '["electric","overwhelming","alive"]'),
('ing-glass-sand',       'planet-pyraxis',        'Glass Sand',             'solar',       2,  'common',    'Fine grains of naturally fused silica that still hold residual heat from plasma strikes. Warm to the touch and slightly magnetic. Sparkles like powdered gold.',                   '["warm","gritty","luminous"]');

-- ── Recipes ─────────────────────────────────────────────────────────

INSERT INTO recipes (id, name, description, difficulty) VALUES
('recipe-starlight-tonic',    'Starlight Tonic',              'A luminous golden liquid that grants temporary enhanced vision in darkness. Side effects include seeing constellations on the inside of your eyelids for a few hours.',                                                'novice'),
('recipe-void-whisper',       'Void Whisper Elixir',          'An inky black draught that lets the drinker hear transmissions from deep space for a brief window. Most hear cosmic background radiation as music. Some hear words.',                                               'adept'),
('recipe-crystal-resonance',  'Crystal Resonance Draught',    'A shimmering prismatic potion that temporarily attunes the drinker to crystalline frequencies. Walls become semi-transparent, hidden structures reveal themselves.',                                                'novice'),
('recipe-pressure-forge',     'Pressure Forge Serum',         'A dense, heavy liquid that temporarily grants the drinker resistance to extreme pressure and gravity. Essential for deep-space salvage operations.',                                                               'adept'),
('recipe-temporal-dew',       'Temporal Morning Dew',         'A delicate, almost water-clear potion that slows the drinker''s perception of time. Ten minutes feel like an hour. Popular with artists and during boring diplomatic functions.',                                  'novice'),
('recipe-plasma-storm',       'Plasma Storm Concentrate',     'A violently crackling orange-white potion that must be consumed within seconds of brewing or it burns through any container. Grants brief but extraordinary electrical manipulation.',                              'master'),
('recipe-entropy-cascade',    'Entropy Cascade',              'The most dangerous known potion. It temporarily reverses local entropy — broken things repair, fires un-burn, wounds un-happen. The hangover involves living the next hour in reverse.',                            'master'),
('recipe-biolume-balm',       'Bioluminescent Balm',          'A soft green salve that makes the wearer''s skin glow faintly for several hours. Originally developed for Verdanthian cave navigation, now fashionable at interstellar galas.',                                    'novice');

INSERT INTO recipe_effects (recipe_id, effect) VALUES
('recipe-starlight-tonic',   'night vision'),
('recipe-starlight-tonic',   'enhanced perception'),
('recipe-starlight-tonic',   'mild euphoria'),
('recipe-void-whisper',      'cosmic hearing'),
('recipe-void-whisper',      'zero-gravity tolerance'),
('recipe-void-whisper',      'temporary translucency'),
('recipe-crystal-resonance', 'structural sight'),
('recipe-crystal-resonance', 'harmonic sensitivity'),
('recipe-pressure-forge',    'pressure resistance'),
('recipe-pressure-forge',    'gravity resistance'),
('recipe-pressure-forge',    'bone density boost'),
('recipe-temporal-dew',      'time dilation perception'),
('recipe-temporal-dew',      'enhanced focus'),
('recipe-temporal-dew',      'calm'),
('recipe-plasma-storm',      'electrical manipulation'),
('recipe-plasma-storm',      'plasma generation'),
('recipe-plasma-storm',      'temporary invulnerability to heat'),
('recipe-entropy-cascade',   'local entropy reversal'),
('recipe-entropy-cascade',   'temporal echo'),
('recipe-entropy-cascade',   'reality instability'),
('recipe-biolume-balm',      'bioluminescence'),
('recipe-biolume-balm',      'skin regeneration'),
('recipe-biolume-balm',      'mild warmth');

INSERT INTO recipe_ingredients (recipe_id, ingredient_id, quantity) VALUES
('recipe-starlight-tonic',   'ing-solar-crystal',    1),
('recipe-starlight-tonic',   'ing-luminous-spore',   2),
('recipe-void-whisper',      'ing-null-shard',       1),
('recipe-void-whisper',      'ing-drift-dust',       2),
('recipe-void-whisper',      'ing-memory-quartz',    1),
('recipe-crystal-resonance', 'ing-resonance-prism',  1),
('recipe-crystal-resonance', 'ing-singing-dust',     3),
('recipe-pressure-forge',    'ing-abyssal-pearl',    1),
('recipe-pressure-forge',    'ing-pressure-salt',    2),
('recipe-pressure-forge',    'ing-magma-pearl',      1),
('recipe-temporal-dew',      'ing-dewdrop-fungus',   3),
('recipe-temporal-dew',      'ing-singing-dust',     1),
('recipe-plasma-storm',      'ing-storm-glass',      1),
('recipe-plasma-storm',      'ing-fulgurite-spike',  2),
('recipe-plasma-storm',      'ing-flare-essence',    1),
('recipe-entropy-cascade',   'ing-entropy-bloom',    1),
('recipe-entropy-cascade',   'ing-null-shard',       1),
('recipe-entropy-cascade',   'ing-memory-quartz',    1),
('recipe-entropy-cascade',   'ing-tidal-essence',    1),
('recipe-biolume-balm',      'ing-luminous-spore',   2),
('recipe-biolume-balm',      'ing-dewdrop-fungus',   2),
('recipe-biolume-balm',      'ing-glass-sand',       1);

-- ── Interactions (5×5 element matrix) ───────────────────────────────

INSERT INTO interactions (element1, element2, result) VALUES
-- Solar row
('solar', 'solar',       'neutral'),
('solar', 'void',        'volatile'),
('solar', 'crystalline', 'synergy'),
('solar', 'organic',     'synergy'),
('solar', 'plasma',      'neutral'),
-- Void row
('void', 'solar',       'volatile'),
('void', 'void',        'neutral'),
('void', 'crystalline', 'neutral'),
('void', 'organic',     'volatile'),
('void', 'plasma',      'synergy'),
-- Crystalline row
('crystalline', 'solar',       'synergy'),
('crystalline', 'void',        'neutral'),
('crystalline', 'crystalline', 'neutral'),
('crystalline', 'organic',     'neutral'),
('crystalline', 'plasma',      'volatile'),
-- Organic row
('organic', 'solar',       'synergy'),
('organic', 'void',        'volatile'),
('organic', 'crystalline', 'neutral'),
('organic', 'organic',     'neutral'),
('organic', 'plasma',      'synergy'),
-- Plasma row
('plasma', 'solar',       'neutral'),
('plasma', 'void',        'synergy'),
('plasma', 'crystalline', 'volatile'),
('plasma', 'organic',     'synergy'),
('plasma', 'plasma',      'neutral');
