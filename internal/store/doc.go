// Package store defines interfaces for data access and provides implementations.
//
// GO CONCEPT: Interfaces
// This package will contain:
//   - Store interfaces (PlanetStore, IngredientStore, etc.) — contracts
//     that define what operations are available without specifying how
//   - In-memory implementation (Step 4) — fast, no persistence
//   - SQLite implementation (Step 9) — persistent, production-ready
//
// Both implementations satisfy the same interfaces, so the rest of the code
// doesn't know or care which one is being used. This is Go's version of
// dependency injection — no framework needed, just interfaces.
//
// GO CONCEPT: Interface Satisfaction
// In Go, interfaces are satisfied IMPLICITLY. If a struct has all the methods
// an interface requires, it implements that interface — no "implements" keyword.
// This is called "structural typing" (vs Java/C#'s "nominal typing").
//
// This will be implemented in Step 4.
package store
