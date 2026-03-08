// Package main is the entry point for the Cosmic Potions API server.
//
// GO CONCEPT: The main Package
// Every Go executable must have a `package main` with a `func main()`. This is
// the program's entry point — like the top-level code in a Node.js script. There's
// no framework magic or dependency injection container; you wire everything together
// explicitly in main(). This is intentional — Go favors explicitness over magic.
//
// GO CONCEPT: cmd/ Convention
// Putting main.go under cmd/server/ is a Go convention for projects that might
// have multiple executables. Each subdirectory under cmd/ is a separate binary.
// The server directory makes it clear this binary is the HTTP server.
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/LilyEssence/cosmic-potions/internal/brewing"
	"github.com/LilyEssence/cosmic-potions/internal/handler"
	"github.com/LilyEssence/cosmic-potions/internal/store"
	"github.com/LilyEssence/cosmic-potions/seed"
)

func main() {
	// ── Determine Port ──────────────────────────────────────────────
	// GO CONCEPT: os.Getenv
	// No dotenv library needed for basic env vars — Go's standard library
	// has os.Getenv(). Railway sets PORT automatically; we default to 8080
	// for local development.
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// ── Create Router ───────────────────────────────────────────────
	// GO CONCEPT: Third-Party Packages
	// chi is a lightweight router that builds on Go's standard net/http.
	// Unlike Express (which wraps everything), chi routers ARE http.Handlers —
	// they work with the standard library, not against it. This is a Go
	// philosophy: compose small tools rather than adopt large frameworks.
	r := chi.NewRouter()

	// ── Middleware ───────────────────────────────────────────────────
	// chi includes built-in middleware. We'll add custom middleware later,
	// but these handle the basics for now.
	r.Use(middleware.RequestID) // Adds X-Request-Id header
	r.Use(middleware.RealIP)    // Extracts real IP from proxy headers
	r.Use(middleware.Logger)    // Logs method, path, status, duration
	r.Use(middleware.Recoverer) // Catches panics, returns 500 instead of crashing

	// ── Dependencies ────────────────────────────────────────────────
	//
	// GO CONCEPT: Explicit Wiring (No DI Framework)
	// Go has no Spring, no Nest.js, no dependency injection container. You
	// create instances and pass them to constructors manually. This is more
	// typing than @Inject() decorators, but the entire dependency graph is
	// visible right here in main() — no magic, no surprises, no "how did
	// this get here?" moments.
	//
	// The pattern is:
	//   1. Create low-level dependencies (store, engine)
	//   2. Pass them to higher-level consumers (handlers)
	//   3. Register the handlers on the router
	//
	// Notice that memStore is assigned to four different interface fields.
	// It's ONE struct satisfying FOUR interfaces — the handler for each
	// resource only sees the methods it needs.

	// Store: in-memory, pre-loaded with seed data (planets, ingredients, recipes).
	memStore := store.NewMemoryStore()

	// Brew engine: pure business logic, no I/O. Receives the element interaction
	// matrix, effect definitions, and a randomizer for non-deterministic outcomes.
	engine := brewing.NewBrewEngine(
		seed.Interactions(),
		seed.Effects(),
		brewing.NewDefaultRandomizer(),
	)

	// ── Routes ──────────────────────────────────────────────────────
	// handler.RegisterRoutes creates all API handlers from the provided
	// dependencies and mounts 12 endpoints under /api. The Deps struct
	// makes it explicit what the handler layer requires.
	handler.RegisterRoutes(r, handler.Deps{
		Planets:      memStore,
		Ingredients:  memStore,
		Recipes:      memStore,
		Brews:        memStore,
		Engine:       engine,
		Effects:      seed.Effects(),
		Interactions: seed.InteractionsList(),
	})

	log.Printf("📋 Registered API routes: planets, ingredients, recipes, brew, effects, interactions")

	// ── Create Server ───────────────────────────────────────────────
	// GO CONCEPT: Struct Initialization
	// We configure the server as a struct with named fields. Go doesn't have
	// constructors — you just set fields directly. The & gives us a pointer
	// to the struct (more on pointers later, but for now: the server needs
	// to be shared between the "start" and "shutdown" code paths).
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// ── Graceful Shutdown ───────────────────────────────────────────
	// GO CONCEPT: Goroutines and Channels
	// This is the most "Go" part of the program. Here's what's happening:
	//
	// 1. We create a CHANNEL (quit) that carries os.Signal values. Channels
	//    are Go's primary way for goroutines to communicate safely. Think of
	//    a channel as a typed, thread-safe pipe.
	//
	// 2. signal.Notify tells the OS to send interrupt/terminate signals into
	//    our channel instead of killing the process immediately.
	//
	// 3. We start the server in a GOROUTINE (the `go` keyword). A goroutine
	//    is like a lightweight thread — it runs concurrently with main().
	//    Go can run millions of goroutines because they're managed by the Go
	//    runtime, not the OS. Each one starts with ~8KB of stack (vs ~1MB for
	//    OS threads).
	//
	// 4. main() then BLOCKS on `<-quit` — reading from the channel. This
	//    pauses main() until a signal arrives, while the server goroutine
	//    handles requests.
	//
	// 5. When Ctrl+C is pressed (or Railway sends SIGTERM during deploy),
	//    the signal flows through the channel, main() unblocks, and we
	//    gracefully shut down — finishing in-flight requests before exiting.
	//
	// In Node.js you'd use process.on('SIGTERM', ...) — the concept is
	// similar, but Go's channels make the control flow explicit rather than
	// callback-based.

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// GO CONCEPT: The `go` Keyword
	// Prefixing a function call with `go` runs it as a new goroutine.
	// This is how we start the server without blocking main().
	go func() {
		log.Printf("🚀 Cosmic Potions server starting on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			// GO CONCEPT: Error Handling
			// Go doesn't have try/catch. Instead, functions return errors as
			// values (usually the last return value). You check them explicitly
			// with `if err != nil`. This is verbose but intentional — Go wants
			// you to handle every error at the call site, not let them bubble
			// up silently as uncaught exceptions.
			//
			// http.ErrServerClosed is expected when we call Shutdown() below,
			// so we only log/exit for unexpected errors.
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Block until we receive a shutdown signal
	<-quit
	log.Println("🛑 Shutdown signal received, draining connections...")

	// GO CONCEPT: Context
	// A context.Context carries deadlines, cancellation signals, and metadata
	// across API boundaries and goroutines. Here, we create a context that
	// automatically cancels after 30 seconds — if the server can't finish
	// all in-flight requests in 30s, we force-quit.
	//
	// The `defer cancel()` ensures the context's resources are released when
	// this function returns. `defer` schedules a function call to run when the
	// enclosing function exits — like a finally block but for any function.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Gracefully shut down: stop accepting new connections, finish existing ones
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Forced shutdown: %v", err)
	}

	log.Println("✅ Server stopped gracefully")
}
