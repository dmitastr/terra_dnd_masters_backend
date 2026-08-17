// Package migrations embeds and applies the SQL migrations for the SQLite
// database. Migration files live in migrations/files/, named NNNN_name.sql
// (e.g. 0001_create_demands.sql), and are applied in ascending version
// order. Applied versions are tracked in a schema_migrations table so Run
// is idempotent and safe to call on every process startup.
package migrations

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"sort"
	"strconv"
	"strings"

	"dnd_schedule/internal/config"
	"dnd_schedule/internal/repository/datasources"
	"github.com/sirupsen/logrus"
)

//go:embed files/*.sql
var migrationFiles embed.FS

// Migration is a single parsed, ready-to-apply migration file.
type Migration struct {
	Version int
	Name    string
	SQL     string
}

const createMigrationsTable = `
CREATE TABLE IF NOT EXISTS schema_migrations (
	version    INTEGER PRIMARY KEY,
	name       TEXT NOT NULL,
	applied_at TEXT NOT NULL DEFAULT (datetime('now'))
);`

type SQLiteRunner struct{}

func NewSQLiteRunner() IRunner {
	return &SQLiteRunner{}
}

// Run applies all pending embedded migrations, in version order, each in
// its own transaction. Already-applied versions (per schema_migrations)
// are skipped, so Run is safe to call on every startup.
func (s *SQLiteRunner) Run(ctx context.Context, config *config.DBConfig, log *logrus.Logger) (datasources.IDatasourceProvider, error) {
	db, err := sql.Open("sqlite", config.GetPath())
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	// SQLite allows only one writer at a time. Capping the pool at a
	// single connection avoids SQLITE_BUSY errors from Go's connection
	// pool racing itself, and busy_timeout handles external contention
	// (e.g. another process/tool with the file open).
	db.SetMaxOpenConns(1)

	pragmas := []string{
		"PRAGMA busy_timeout = 5000;",
		"PRAGMA journal_mode = WAL;",
		"PRAGMA synchronous = NORMAL;",
		"PRAGMA foreign_keys = ON;",
	}
	for _, p := range pragmas {
		if _, err := db.ExecContext(ctx, p); err != nil {
			db.Close()
			return nil, fmt.Errorf("set pragma %q: %w", p, err)
		}
	}

	if _, err := db.ExecContext(ctx, createMigrationsTable); err != nil {
		return nil, fmt.Errorf("create schema_migrations table: %w", err)
	}

	all, err := loadMigrations()
	if err != nil {
		return nil, fmt.Errorf("load migrations: %w", err)
	}

	applied, err := appliedVersions(ctx, db)
	if err != nil {
		return nil, fmt.Errorf("read applied migrations: %w", err)
	}

	for _, m := range all {
		if applied[m.Version] {
			continue
		}

		if err := applyMigration(ctx, db, m); err != nil {
			return nil, fmt.Errorf("apply migration %04d_%s: %w", m.Version, m.Name, err)
		}
	}

	dsProvider := datasources.NewSQLiteDatasourceProvider(db, log)

	return dsProvider, nil
}

func loadMigrations() ([]Migration, error) {
	entries, err := fs.ReadDir(migrationFiles, "files")
	if err != nil {
		return nil, fmt.Errorf("read migrations dir: %w", err)
	}

	out := make([]Migration, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}

		version, name, err := parseFilename(e.Name())
		if err != nil {
			return nil, fmt.Errorf("parse migration filename %q: %w", e.Name(), err)
		}

		content, err := migrationFiles.ReadFile("files/" + e.Name())
		if err != nil {
			return nil, fmt.Errorf("read migration file %q: %w", e.Name(), err)
		}

		out = append(out, Migration{Version: version, Name: name, SQL: string(content)})
	}

	sort.Slice(out, func(i, j int) bool { return out[i].Version < out[j].Version })

	// Guard against duplicate version prefixes, which would silently apply
	// in an unintended order.
	for i := 1; i < len(out); i++ {
		if out[i].Version == out[i-1].Version {
			return nil, fmt.Errorf("duplicate migration version %d (%s and %s)",
				out[i].Version, out[i-1].Name, out[i].Name)
		}
	}

	return out, nil
}

// parseFilename expects the form NNNN_name.sql, e.g. 0001_create_demands.sql.
func parseFilename(filename string) (int, string, error) {
	base := strings.TrimSuffix(filename, ".sql")

	parts := strings.SplitN(base, "_", 2)
	if len(parts) != 2 {
		return 0, "", fmt.Errorf("filename must be in format NNNN_name.sql, got %q", filename)
	}

	version, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, "", fmt.Errorf("invalid version prefix in %q: %w", filename, err)
	}

	return version, parts[1], nil
}

func appliedVersions(ctx context.Context, db *sql.DB) (map[int]bool, error) {
	rows, err := db.QueryContext(ctx, `SELECT version FROM schema_migrations;`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[int]bool)
	for rows.Next() {
		var v int
		if err := rows.Scan(&v); err != nil {
			return nil, err
		}
		applied[v] = true
	}

	return applied, rows.Err()
}

func applyMigration(ctx context.Context, db *sql.DB, m Migration) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck // no-op once committed

	if _, err := tx.ExecContext(ctx, m.SQL); err != nil {
		return fmt.Errorf("exec migration sql: %w", err)
	}

	if _, err := tx.ExecContext(ctx,
		`INSERT INTO schema_migrations (version, name) VALUES (?, ?);`,
		m.Version, m.Name,
	); err != nil {
		return fmt.Errorf("record migration: %w", err)
	}

	return tx.Commit()
}
