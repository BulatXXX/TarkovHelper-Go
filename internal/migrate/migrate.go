package migrate

import (
	"database/sql"
	"embed"
	"fmt"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"

	_ "github.com/golang-migrate/migrate/v4/source/file" // не обязателен, но иногда полезен
	_ "github.com/jackc/pgx/v5/stdlib"                   // регистрирует sql driver "pgx"
)

//go:embed migrations/*.sql
var fs embed.FS

func Up(databaseURL string) error {
	if strings.HasPrefix(databaseURL, "postgresql://") {
		databaseURL = "postgres://" + strings.TrimPrefix(databaseURL, "postgresql://")
	}

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return fmt.Errorf("sql open: %w", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return fmt.Errorf("sql ping: %w", err)
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("postgres driver: %w", err)
	}

	src, err := iofs.New(fs, "migrations")
	if err != nil {
		return fmt.Errorf("iofs source: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", src, "postgres", driver)
	if err != nil {
		return fmt.Errorf("migrate init: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migrate up: %w", err)
	}
	return nil
}
