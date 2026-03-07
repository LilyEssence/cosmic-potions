// Package handler contains the HTTP handlers that map API routes to
// business logic.
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
// This will be implemented in Step 6.
package handler
