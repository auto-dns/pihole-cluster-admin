package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/auto-dns/pihole-cluster-admin/internal/config"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "modernc.org/sqlite" // Pure Go SQLite driver
)

func NewDatabase(cfg config.DatabaseConfig) (*sql.DB, error) {
	if err := os.MkdirAll(filepath.Dir(cfg.Path), 0755); err != nil {
		return nil, fmt.Errorf("create database directory: %w", err)
	}

	db, err := sql.Open("sqlite", cfg.Path)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	db.SetMaxOpenConns(1)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	if err := runMigrations(db, cfg.MigrationsPath); err != nil {
		return nil, fmt.Errorf("run migrations: %w", err)
	}

	return db, nil
}

func runMigrations(db *sql.DB, migrationPath string) error {
	absPath, err := filepath.Abs(migrationPath)
	if err != nil {
		return fmt.Errorf("resolve migration path: %w", err)
	}

	driver, err := sqlite3.WithInstance(db, &sqlite3.Config{})
	if err != nil {
		return fmt.Errorf("create migration driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://"+absPath,
		"sqlite", driver)
	if err != nil {
		return fmt.Errorf("create migrator: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("apply migrations: %w", err)
	}
	return nil
}
