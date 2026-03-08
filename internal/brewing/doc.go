// Package brewing contains the brew engine — the core business logic for
// combining ingredients into potions.
//
// This is the architectural centerpiece of Cosmic Potions. It's deliberately
// separated from HTTP handlers and data storage so it can be:
//   - Tested in isolation with deterministic randomness
//   - Reused if we ever add a CLI or websocket interface
//   - Reasoned about without thinking about HTTP or databases
//
// The engine provides three public operations:
//   - CheckCompatibility: preview element interactions without brewing
//   - CalculateEffects: determine which codex effects a brew could discover
//   - Brew: the full brewing flow (compatibility → outcome roll → effects → result)
//
// Randomness is injected via the Randomizer interface, making all outcomes
// deterministic in tests while unpredictable in production.
package brewing
