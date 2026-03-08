package mw

import "net/http"

// ── CORS Middleware ─────────────────────────────────────────────────
//
// GO CONCEPT: CORS in Go (No Express cors() Equivalent)
// Go doesn't have a built-in CORS middleware. Most Go projects either use a
// small third-party package (like rs/cors) or write their own — it's just a
// few headers. We write our own here to keep dependencies minimal and to
// understand exactly what's happening.
//
// CORS (Cross-Origin Resource Sharing) controls which websites can call our
// API from JavaScript. Without it, a browser at localhost:5173 (our React dev
// server) would refuse to fetch from localhost:8080 (our Go API) because
// they're different origins.
//
// GO CONCEPT: Closure-Based Configuration
// The CORS function takes an allowed origins list and returns a middleware
// function. This is a "closure" pattern — the returned middleware captures
// the `allowedOrigins` variable from its enclosing scope. Each time the
// middleware runs, it has access to the origins list without needing a struct.
//
// In TypeScript, this would be:
//   const cors = (origins: string[]) => (req, res, next) => { ... }
//
// In Go, it's the same concept with explicit types:
//   func CORS(origins []string) func(http.Handler) http.Handler

// CORS returns middleware that sets Cross-Origin Resource Sharing headers.
// The allowedOrigins parameter lists the exact origins permitted to call the API.
// If the request's Origin header matches any of them, the response allows the
// cross-origin request.
//
// For local development: "http://localhost:5173" (Vite dev server)
// For production: "https://lixie.art"
func CORS(allowedOrigins []string) func(http.Handler) http.Handler {
	// Build a set for O(1) lookups instead of scanning a slice every request.
	//
	// GO CONCEPT: Maps as Sets
	// Go has no built-in Set type. The idiomatic substitute is map[T]bool
	// (or map[T]struct{} for zero-memory values). Checking membership is
	// `if set[key] { ... }` — a missing key returns false (the bool zero value).
	originSet := make(map[string]bool, len(allowedOrigins))
	for _, o := range allowedOrigins {
		originSet[o] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			if originSet[origin] {
				// GO CONCEPT: Header().Set vs Header().Add
				// Set replaces any existing values for the header key.
				// Add appends a value (useful for multi-value headers like
				// Set-Cookie). For CORS headers, Set is correct — there
				// should be exactly one value per header.
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Request-Id")
				w.Header().Set("Access-Control-Max-Age", "86400") // 24 hours — browsers cache preflight results
			}

			// GO CONCEPT: Preflight Requests
			// Browsers send an OPTIONS request before certain cross-origin
			// requests (POST with JSON body, custom headers, etc.) to ask
			// "is this allowed?" The server must respond with CORS headers
			// and a 204 No Content — no body, just headers.
			//
			// If we don't handle OPTIONS, the browser never sends the real
			// request and the user sees a CORS error in the console.
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
