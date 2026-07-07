package handlers

// NOTE: internal test package (`package handlers`, not `handlers_test`) —
// deliberate, so tests can reset the unexported package-level `loginLimiter`
// var between runs. If you'd rather keep tests external, export a
// `ResetLoginLimiterForTest()` func in auth.go instead and switch this back.

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var testPool *pgxpool.Pool

// TestMain runs against the existing darshan-postgres-1 container (assumes
// it's already up via docker compose). It drops/recreates darshan_test fresh
// on every run, then applies migrations/*.sql in filename order.
func TestMain(m *testing.M) {
	adminURL := getEnv("TEST_ADMIN_DATABASE_URL", "postgres://darshan:darshan@localhost:5433/darshan?sslmode=disable")
	testURL := getEnv("TEST_DATABASE_URL", "postgres://darshan:darshan@localhost:5433/darshan_test?sslmode=disable")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := resetTestDatabase(ctx, adminURL); err != nil {
		fmt.Fprintf(os.Stderr, "failed to reset test database: %v\n", err)
		os.Exit(1)
	}

	pool, err := pgxpool.New(ctx, testURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to connect to test database: %v\n", err)
		os.Exit(1)
	}
	testPool = pool

	if err := runMigrations(ctx, testPool, "../../migrations"); err != nil {
		fmt.Fprintf(os.Stderr, "failed to run migrations: %v\n", err)
		testPool.Close()
		os.Exit(1)
	}

	code := m.Run()

	testPool.Close()
	os.Exit(code)
}

func resetTestDatabase(ctx context.Context, adminURL string) error {
	adminPool, err := pgxpool.New(ctx, adminURL)
	if err != nil {
		return fmt.Errorf("connect to admin db: %w", err)
	}
	defer adminPool.Close()

	_, err = adminPool.Exec(ctx, `
		SELECT pg_terminate_backend(pid)
		FROM pg_stat_activity
		WHERE datname = 'darshan_test' AND pid <> pg_backend_pid()
	`)
	if err != nil {
		return fmt.Errorf("terminate existing connections: %w", err)
	}

	if _, err := adminPool.Exec(ctx, `DROP DATABASE IF EXISTS darshan_test`); err != nil {
		return fmt.Errorf("drop database: %w", err)
	}
	if _, err := adminPool.Exec(ctx, `CREATE DATABASE darshan_test`); err != nil {
		return fmt.Errorf("create database: %w", err)
	}
	return nil
}

// runMigrations applies every .sql file in dir, in filename order. No
// golang-migrate dependency — this codebase applies migrations via
// docker-entrypoint-initdb.d in dev, so tests need their own simple runner.
func runMigrations(ctx context.Context, pool *pgxpool.Pool, dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}

	var files []string
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".sql" {
			continue
		}
		files = append(files, e.Name())
	}
	sort.Strings(files)

	for _, name := range files {
		sqlBytes, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return fmt.Errorf("read migration %s: %w", name, err)
		}
		if _, err := pool.Exec(ctx, string(sqlBytes)); err != nil {
			return fmt.Errorf("apply migration %s: %w", name, err)
		}
	}
	return nil
}

// truncateAll clears app tables between tests without re-running migrations.
// Adjust table list if your schema has grown since.
func truncateAll(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	_, err := testPool.Exec(ctx, `
		TRUNCATE TABLE refresh_tokens, temples, users RESTART IDENTITY CASCADE
	`)
	if err != nil {
		t.Fatalf("truncateAll: %v", err)
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
