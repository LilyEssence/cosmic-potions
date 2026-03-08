package store

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ── Migration Runner ────────────────────────────────────────────────
//
// GO CONCEPT: database/sql Package
// Go's standard library includes `database/sql`, which provides a generic
// interface for SQL databases — similar to how Java's JDBC works. You write
// code against database/sql, and a DRIVER (like modernc.org/sqlite) plugs in
// underneath. This means switching databases mostly involves changing the
// driver import and connection string, not rewriting queries.
//
// GO CONCEPT: Blank Import (_ "driver")
// SQL drivers register themselves using an init() function. You import the
// driver with a blank identifier `_ "modernc.org/sqlite"` which runs init()
// but doesn't give you a named reference. The driver registers itself with
// database/sql, and you access it via sql.Open("sqlite", ...).
//
// We do the blank import in sqlite.go (the file that opens the connection),
// not here. This file just uses the *sql.DB handle it receives.

// RunMigrations reads SQL files from the migrations directory and applies any
// that haven't been run yet. Each migration's version number is extracted from
// its filename (e.g., 001_init.sql → version 1).
//
// GO CONCEPT: Error Wrapping with fmt.Errorf
// Go 1.13+ introduced error wrapping: `fmt.Errorf("context: %w", err)` creates
// a new error that wraps the original. Callers can unwrap with errors.Is() or
// errors.As(). The %w verb (not %v!) preserves the error chain. This is how you
// add context without losing the original error — like adding stack frames in
// other languages, but explicitly.
func RunMigrations(db *sql.DB, migrationsDir string) error {
	// Ensure the schema_migrations tracking table exists.
	// CREATE TABLE IF NOT EXISTS is idempotent — safe to run every startup.
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version    INTEGER PRIMARY KEY,
			applied_at TEXT NOT NULL DEFAULT (datetime('now'))
		)
	`)
	if err != nil {
		return fmt.Errorf("creating schema_migrations table: %w", err)
	}

	// Read all .sql files from the migrations directory.
	//
	// GO CONCEPT: filepath.Glob
	// Glob returns filenames matching a pattern, similar to shell globbing.
	// It's deterministic but NOT guaranteed to return sorted results on all
	// operating systems, so we sort explicitly.
	pattern := filepath.Join(migrationsDir, "*.sql")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("reading migrations directory: %w", err)
	}
	sort.Strings(files) // Ensures 001 runs before 002, etc.

	for _, file := range files {
		// Extract version number from filename: "001_init.sql" → 1
		base := filepath.Base(file) // "001_init.sql"
		parts := strings.SplitN(base, "_", 2)
		if len(parts) < 2 {
			log.Printf("⚠️  Skipping non-migration file: %s", base)
			continue
		}

		var version int
		// GO CONCEPT: fmt.Sscanf
		// Sscanf parses a string according to a format, like C's sscanf.
		// %d scans a decimal integer. It returns the number of items
		// successfully scanned. We use it to extract "001" → int 1.
		if _, err := fmt.Sscanf(parts[0], "%d", &version); err != nil {
			log.Printf("⚠️  Skipping file with non-numeric prefix: %s", base)
			continue
		}

		// Check if this migration has already been applied.
		var exists bool
		err := db.QueryRow("SELECT 1 FROM schema_migrations WHERE version = ?", version).Scan(&exists)
		if err == nil {
			// Already applied — skip.
			log.Printf("  ✓ Migration %03d already applied", version)
			continue
		}
		// If err is sql.ErrNoRows, the migration hasn't been applied yet.
		// Any other error is unexpected.
		if err != sql.ErrNoRows {
			return fmt.Errorf("checking migration %d: %w", version, err)
		}

		// Read and execute the migration SQL.
		//
		// GO CONCEPT: os.ReadFile
		// Reads an entire file into a []byte. For SQL migrations (small files),
		// this is simpler than streaming. For large files, you'd use os.Open()
		// and read in chunks.
		content, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("reading migration file %s: %w", base, err)
		}

		// GO CONCEPT: Transactions (db.Begin / tx.Commit / tx.Rollback)
		// A transaction groups multiple SQL statements into an atomic unit.
		// Either ALL statements succeed (Commit) or NONE of them take effect
		// (Rollback). This ensures a half-applied migration doesn't corrupt
		// the database. The defer tx.Rollback() is a safety net — if we
		// return early due to an error, the transaction is rolled back.
		// Calling Rollback after Commit is a no-op (safe to defer).
		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("starting transaction for migration %d: %w", version, err)
		}
		defer tx.Rollback() // No-op after Commit; safety net for errors.

		// ExecContext isn't needed here since we're not using contexts for
		// migration timeouts (migrations run at startup, not under request
		// deadlines). Exec runs the entire SQL file as one batch.
		if _, err := tx.Exec(string(content)); err != nil {
			return fmt.Errorf("executing migration %d: %w", version, err)
		}

		// Record that this migration has been applied.
		if _, err := tx.Exec("INSERT INTO schema_migrations (version) VALUES (?)", version); err != nil {
			return fmt.Errorf("recording migration %d: %w", version, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("committing migration %d: %w", version, err)
		}

		log.Printf("  ✅ Applied migration %03d: %s", version, base)
	}

	return nil
}
