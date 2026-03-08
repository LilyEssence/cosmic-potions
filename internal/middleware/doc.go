// Package mw contains custom HTTP middleware for the Cosmic Potions server.
//
// GO CONCEPT: Middleware in Go
// Middleware in Go is just a function that takes an http.Handler and returns
// a new http.Handler. It's function composition — each middleware wraps the
// next handler, adding behavior (logging, auth, CORS) before/after the
// inner handler runs.
//
//	func MyMiddleware(next http.Handler) http.Handler {
//	    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//	        // do something before
//	        next.ServeHTTP(w, r)
//	        // do something after
//	    })
//	}
//
// This package provides:
//   - CORS: configurable cross-origin headers for frontend ↔ API communication
//   - RequestLogger: request logging with method, path, status, and duration
//
// chi's built-in middleware.RequestID and middleware.Recoverer are used directly
// from chi — we only write custom middleware when we need control over behavior.
package mw
