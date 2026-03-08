package mw

import (
	"log"
	"net/http"
	"time"
)

// ── Request Logger Middleware ────────────────────────────────────────
//
// GO CONCEPT: Custom vs chi's Built-In Logger
// chi ships with middleware.Logger, which we used initially. We now replace
// it with a custom logger for two reasons:
//  1. Control over the log format (we want request ID in every log line)
//  2. Learning opportunity — writing middleware from scratch is the best
//     way to understand the pattern
//
// GO CONCEPT: The Middleware Signature
// Go HTTP middleware has exactly one shape:
//
//   func(next http.Handler) http.Handler
//
// It receives the "next" handler in the chain, and returns a new handler
// that wraps it. When the middleware is applied with r.Use(RequestLogger),
// chi calls it at startup to build the handler chain. At request time, the
// wrapper runs, does its work, and calls next.ServeHTTP(w, r).
//
// This is pure function composition — no magic, no decorators, no framework
// hooks. Each middleware is a function that transforms one handler into another.

// RequestLogger logs every request with method, path, status code, and duration.
// It also logs the request ID (set by chi's RequestID middleware, which runs
// earlier in the chain).
//
// GO CONCEPT: responseWriter Wrapper (Capturing Status Code)
// http.ResponseWriter doesn't expose the status code after WriteHeader is
// called — there's no w.StatusCode field. To log the status, we wrap the
// ResponseWriter in a struct that intercepts WriteHeader and records the
// code. This wrapper pattern is extremely common in Go middleware.
func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap the ResponseWriter to capture the status code.
		ww := &statusWriter{ResponseWriter: w, status: http.StatusOK}

		// Call the next handler. Everything after this line runs AFTER
		// the handler has written its response.
		next.ServeHTTP(ww, r)

		// Log the completed request.
		// GO CONCEPT: log.Printf
		// Go's built-in log package is simple (timestamp + message, no levels).
		// For a game API, this is sufficient. Production services often use
		// structured logging (slog, zerolog, zap) for JSON output that log
		// aggregators can parse. We'll keep it simple here.
		duration := time.Since(start)
		log.Printf("%s %s → %d (%s)",
			r.Method,
			r.URL.Path,
			ww.status,
			duration.Round(time.Millisecond),
		)
	})
}

// statusWriter wraps http.ResponseWriter to capture the HTTP status code.
//
// GO CONCEPT: Embedding for Interface Satisfaction
// By embedding http.ResponseWriter (without a field name), statusWriter
// automatically satisfies the http.ResponseWriter interface — it "inherits"
// all methods from the embedded type. We only override WriteHeader to
// intercept the status code; all other methods (Write, Header) delegate
// to the original ResponseWriter automatically.
//
// This is Go's composition model — instead of inheritance (class StatusWriter
// extends ResponseWriter), you embed the interface and override what you need.
type statusWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

// WriteHeader captures the status code before delegating to the real writer.
//
// GO CONCEPT: Method Override via Embedding
// Because statusWriter embeds http.ResponseWriter, it already has a WriteHeader
// method (from the embedded type). Defining our own WriteHeader "shadows" the
// embedded version — calls to ww.WriteHeader() hit this method, not the
// original. This is how you override specific behavior while keeping everything
// else from the embedded type.
func (w *statusWriter) WriteHeader(code int) {
	if !w.wroteHeader {
		w.status = code
		w.wroteHeader = true
	}
	w.ResponseWriter.WriteHeader(code)
}
