package db

import (
	"context"
	"embed"
	"fmt"
	"log/slog"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

var Pool *pgxpool.Pool

func Connect(connectionString string) error {
	pool, err := pgxpool.New(context.Background(), connectionString)
	if err != nil {
		return fmt.Errorf("unable to connect to database: %w", err)
	}

	// Verify connection
	err = pool.Ping(context.Background())
	if err != nil {
		return fmt.Errorf("unable to ping database: %w", err)
	}

	Pool = pool
	slog.Info("Connected to PostgreSQL database")
	return nil
}

func RunMigrations() error {
	files := []string{
		"migrations/001_init_schema.sql",
		"migrations/002_seed_categories.sql",
		"migrations/003_add_analytics_insights.sql",
	}

	for _, file := range files {
		content, err := migrationFiles.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read migration %s: %w", file, err)
		}

		sql := strings.TrimSpace(string(content))
		if sql == "" {
			continue
		}

		_, err = Pool.Exec(context.Background(), sql)
		if err != nil {
			return fmt.Errorf("failed to run migration %s: %w", file, err)
		}

		slog.Info("Migration applied", slog.String("file", file))
	}

	return nil
}

func Close() {
	if Pool != nil {
		Pool.Close()
		slog.Info("Database connection closed")
	}
}
