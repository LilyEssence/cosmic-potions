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
	"path/filepath"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/LilyEssence/cosmic-potions/internal/brewing"
	"github.com/LilyEssence/cosmic-potions/internal/handler"
	mw "github.com/LilyEssence/cosmic-potions/internal/middleware"
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
	//
	// GO CONCEPT: Middleware Ordering
	// Middleware runs in the order it's registered. The order matters:
	//   1. RequestID — assigns an ID that later middleware can reference
	//   2. RealIP — extracts the client IP before logging reads it
	//   3. CORS — must run before any handler writes response headers
	//   4. RequestLogger — logs after the handler completes (captures status)
	//   5. Recoverer — catches panics from any handler, returns 500
	//
	// Each middleware wraps the next one like nesting dolls:
	//   RequestID → RealIP → CORS → Logger → Recoverer → Handler
	//
	// GO CONCEPT: Named Import Alias
	// We import our middleware package as `mw` to avoid collision with chi's
	// `middleware` package. Both are in scope, distinguished by their alias:
	//   middleware.RequestID  — from chi
	//   mw.CORS(...)         — from our package

	// chi built-ins: request ID, real IP extraction, panic recovery.
	r.Use(middleware.RequestID) // Adds X-Request-Id header
	r.Use(middleware.RealIP)    // Extracts real IP from proxy headers

	// Custom CORS: allows the React frontend (dev and prod) to call the API.
	// GO CONCEPT: Configuring Middleware with Environment
	// In production, you might read allowed origins from an env var. For now,
	// we hardcode both the local dev server and the production domain. This
	// keeps the game playable locally and when deployed on Railway.
	r.Use(mw.CORS([]string{
		"http://localhost:5173", // Vite dev server
		"https://lixie.art",     // Production frontend
		"https://www.lixie.art", // www variant
	}))

	// Custom request logger: method, path, status, duration.
	// Replaces chi's middleware.Logger for consistent format and control.
	r.Use(mw.RequestLogger)

	// Panic recovery: catches panics in handlers, returns 500 instead of crashing.
	r.Use(middleware.Recoverer)

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
	// Notice that the store is assigned to four different interface fields.
	// It's ONE struct satisfying FOUR interfaces — the handler for each
	// resource only sees the methods it needs.

	// ── Store Selection ─────────────────────────────────────────────
	//
	// GO CONCEPT: Environment-Driven Configuration
	// STORE_TYPE toggles between "memory" (default) and "sqlite". This lets
	// us use the in-memory store for fast local development and the SQLite
	// store for persistent deployment on Railway. Both satisfy the same
	// four interfaces — the rest of the code doesn't know the difference.
	//
	// GO CONCEPT: Interface Variables
	// We declare the store as four interface variables, then assign the
	// concrete implementation based on the env var. This is Go's version
	// of runtime polymorphism — no abstract classes needed.
	var (
		planets     store.PlanetStore
		ingredients store.IngredientStore
		recipes     store.RecipeStore
		brews       store.BrewStore

		// These are loaded from seed package by default, but overridden
		// by SQLite when using the database store.
		interactionMap  = seed.Interactions()
		effectList      = seed.Effects()
		interactionList = seed.InteractionsList()
	)

	storeType := os.Getenv("STORE_TYPE")
	switch storeType {
	case "sqlite":
		// ── SQLite Store ────────────────────────────────────────────
		// DB_PATH controls where the SQLite file lives. Railway mounts
		// persistent storage; locally, we default to a file in the project.
		dbPath := os.Getenv("DB_PATH")
		if dbPath == "" {
			dbPath = "cosmic-potions.db"
		}

		migrationsDir := os.Getenv("MIGRATIONS_DIR")
		if migrationsDir == "" {
			migrationsDir = "migrations"
		}

		// Ensure the parent directory exists. Railway mounts volumes at
		// specific paths, but the directory inside the container might not
		// exist yet on first deploy.
		if dir := filepath.Dir(dbPath); dir != "." && dir != "" {
			if err := os.MkdirAll(dir, 0755); err != nil {
				log.Fatalf("Failed to create database directory %s: %v", dir, err)
			}
		}

		log.Printf("📦 Using SQLite store: %s", dbPath)
		sqliteStore, err := store.NewSQLiteStore(dbPath, migrationsDir)
		if err != nil {
			log.Fatalf("Failed to initialize SQLite store: %v", err)
		}
		defer sqliteStore.Close()

		planets = sqliteStore
		ingredients = sqliteStore
		recipes = sqliteStore
		brews = sqliteStore

		// Load interactions from the database instead of the seed package.
		// This ensures the database is the single source of truth when using SQLite.
		dbInteractionMap, err := sqliteStore.LoadInteractionMap()
		if err != nil {
			log.Fatalf("Failed to load interactions from database: %v", err)
		}
		interactionMap = dbInteractionMap

		dbInteractionList, err := sqliteStore.LoadInteractions()
		if err != nil {
			log.Fatalf("Failed to load interaction list from database: %v", err)
		}
		interactionList = dbInteractionList

		// Effects are still loaded from the seed package — they're static
		// game data that isn't stored in SQLite (no effects table).

	default:
		// ── Memory Store ────────────────────────────────────────────
		// Default: fast, no persistence. Pre-loaded with seed data.
		log.Printf("📦 Using in-memory store (set STORE_TYPE=sqlite for persistence)")
		memStore := store.NewMemoryStore()
		planets = memStore
		ingredients = memStore
		recipes = memStore
		brews = memStore
	}

	// Brew engine: pure business logic, no I/O. Receives the element interaction
	// matrix, effect definitions, and a randomizer for non-deterministic outcomes.
	engine := brewing.NewBrewEngine(
		interactionMap,
		effectList,
		brewing.NewDefaultRandomizer(),
	)

	// ── Routes ──────────────────────────────────────────────────────
	// handler.RegisterRoutes creates all API handlers from the provided
	// dependencies and mounts 12 endpoints under /api. The Deps struct
	// makes it explicit what the handler layer requires.
	handler.RegisterRoutes(r, handler.Deps{
		Planets:      planets,
		Ingredients:  ingredients,
		Recipes:      recipes,
		Brews:        brews,
		Engine:       engine,
		Effects:      effectList,
		Interactions: interactionList,
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
