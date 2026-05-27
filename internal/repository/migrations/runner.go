package migrations

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

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
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	log.Println("Database connection pool established")

	m, err := migrate.New(
		"file://migrations/migrations_scripts",
		config.GetConnString())
	if err != nil {
		return nil, err
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return nil, err
	}

	log.Println("Database migration succeeded")

	if err := runFixtures(ctx, pool); err != nil {
		return nil, err
	}
	log.Println("Fixtures migration succeeded")

	return pool, nil
}

// runFixtures add fixtures to tables
func runFixtures(ctx context.Context, pool *pgxpool.Pool) error {
	fixtureSQL, err := os.ReadFile("migrations/fixtures/100001_add_test_data.up.sql")
	if err != nil {
		return err
	}

	_, err = pool.Exec(ctx, string(fixtureSQL))
	return err
}
