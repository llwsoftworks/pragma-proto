package database

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DB wraps a pgxpool connection pool.
type DB struct {
	Pool *pgxpool.Pool
}

// Connect initialises the PostgreSQL connection pool and verifies connectivity.
func Connect(ctx context.Context, dsn string) (*DB, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("database: parse config: %w", err)
	}

	cfg.MaxConns = 25
	cfg.MinConns = 5
	cfg.MaxConnLifetime = 30 * time.Minute
	cfg.MaxConnIdleTime = 5 * time.Minute
	cfg.HealthCheckPeriod = 1 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("database: new pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("database: ping: %w", err)
	}

	return &DB{Pool: pool}, nil
}

// Close releases pool resources.
func (db *DB) Close() {
	db.Pool.Close()
}

// Healthy returns nil if the database is reachable.
func (db *DB) Healthy(ctx context.Context) error {
	return db.Pool.Ping(ctx)
}

// RunMigrations applies all .sql migration files in the given directory in order.
// It creates a simple schema_migrations tracking table to avoid re-running.
func (db *DB) RunMigrations(ctx context.Context, migrationsDir string) error {
	// Ensure migrations tracking table exists.
	_, err := db.Pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version    TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ DEFAULT NOW()
		)
	`)
	if err != nil {
		return fmt.Errorf("database: create migrations table: %w", err)
	}

	// Load already-applied versions.
	rows, err := db.Pool.Query(ctx, `SELECT version FROM schema_migrations ORDER BY version`)
	if err != nil {
		return fmt.Errorf("database: query applied migrations: %w", err)
	}
	applied := map[string]bool{}
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			return err
		}
		applied[v] = true
	}
	rows.Close()

	// Find migration files.
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("database: read migrations dir: %w", err)
	}

	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			files = append(files, e.Name())
		}
	}
	sort.Strings(files)

	for _, file := range files {
		version := strings.TrimSuffix(file, ".sql")
		if applied[version] {
			continue
		}

		content, err := os.ReadFile(filepath.Join(migrationsDir, file))
		if err != nil {
			return fmt.Errorf("database: read migration %s: %w", file, err)
		}

		tx, err := db.Pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("database: begin tx for %s: %w", file, err)
		}

		if _, err := tx.Exec(ctx, string(content)); err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("database: apply migration %s: %w", file, err)
		}

		if _, err := tx.Exec(ctx, `INSERT INTO schema_migrations (version) VALUES ($1)`, version); err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("database: record migration %s: %w", file, err)
		}

		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("database: commit migration %s: %w", file, err)
		}
	}

	return nil
}

// SetTenantID sets the PostgreSQL session variable used by RLS policies.
// Call this within each request transaction: SET LOCAL app.current_school_id = '...'.
func SetTenantID(ctx context.Context, pool *pgxpool.Pool, schoolID string) error {
	_, err := pool.Exec(ctx, `SET LOCAL app.current_school_id = $1`, schoolID)
	return err
}
