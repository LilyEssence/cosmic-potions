// Package handler contains the HTTP handlers that map API routes to
// business logic. It is the bridge between HTTP (requests, responses, status
// codes) and the domain layer (store interfaces, brew engine).
//
// GO CONCEPT: Handler Pattern
// In Go, HTTP handlers are just functions (or methods) with the signature:
//
//	func(w http.ResponseWriter, r *http.Request)
//
// There's no controller class, no decorator, no framework magic. Each handler
// receives a request, does work, and writes a response. This simplicity means
// handlers are easy to test using Go's built-in httptest package.
//
// Organization:
//   - response.go    — JSON response helpers and error mapping
//   - routes.go      — Centralized route registration (the API surface at a glance)
//   - planets.go     — GET /api/planets, GET /api/planets/{id}
//   - ingredients.go — GET /api/ingredients, GET /api/ingredients/{id}
//   - recipes.go     — GET /api/recipes, GET /api/recipes/{id}
//   - brew.go        — POST /api/brew, POST /api/brew/check, GET /api/brews
//   - effects.go     — GET /api/effects (codex master list)
//   - interactions.go — GET /api/interactions (element matrix)
package handler
