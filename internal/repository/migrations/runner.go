package migrations

import (
	"context"
	"errors"
	"fmt"
	"log"

	"dnd_schedule/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/golang-migrate/migrate/v4"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5"
)

// Run applies all pending migrations inside a single transaction per step.
func Run(ctx context.Context, config *config.DBConfig) (*pgxpool.Pool, error) {
	dbConfig, err := pgxpool.ParseConfig(config.GetConnString())
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(
		ctx,
		dbConfig,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create database connection pool: %w", err)
	}
	m, err := migrate.New(
		"file://migrations/migrations_scripts",
		config.GetConnString())
	if err != nil {
		return nil, err
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Database migration succeeded")

	return pool, nil
}
