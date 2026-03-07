// Package brewing contains the brew engine — the core business logic for
// combining ingredients into potions.
//
// This is the architectural centerpiece of Cosmic Potions. It's deliberately
// separated from HTTP handlers and data storage so it can be:
//   - Tested in isolation with deterministic randomness
//   - Reused if we ever add a CLI or websocket interface
//   - Reasoned about without thinking about HTTP or databases
//
// This will be implemented in Step 5.
package brewing
