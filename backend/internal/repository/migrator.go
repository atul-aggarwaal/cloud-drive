package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// Runs DB UP migrations,if any
func RunDatabaseMigration(db *sql.DB) error {

	driver, err := postgres.WithInstance(db, &postgres.Config{})

	if err != nil {
		return fmt.Errorf("failed to create migration db instance driver %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance("file://internal/repository/migrations", "postgres", driver)

	if err != nil {
		return fmt.Errorf("failed to initialize migration controller instance %w", err)
	}

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			log.Println("No new migrations found")
			return nil
		}
		return fmt.Errorf("migration failed %w", err)
	}

	log.Println("Database migration complete")
	return nil
}
