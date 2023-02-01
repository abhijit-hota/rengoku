package db

import (
	"database/sql"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

const (
	// migrationsSourceURL is the URL to the folder containing the migrations
	migrationsSourceURL = "file://db/migrations"

	// databaseName is the name of the database for logging
	databaseName = "rengoku.sqlite.db"
)

// migrateToLatest runs the database migrations to the latest version.
func migrateToLatest(db *sql.DB) error {

	// Create sqliteDBInstance driver from existing sqlite DB client
	sqliteDBInstance, err := sqlite.WithInstance(db, new(sqlite.Config))
	if err != nil {
		return fmt.Errorf("failed to read migration target %w", err)
	}

	// Create a migrationClient using the migrations folder as source and dbURL
	migrationClient, err := migrate.NewWithDatabaseInstance(
		migrationsSourceURL, databaseName, sqliteDBInstance,
	)
	if err != nil {
		return fmt.Errorf("failed to create migration instance %w", err)
	}
	// Defer closing the migrationClient
	defer func(m *migrate.Migrate) {
		srcErr, dbErr := m.Close()
		if srcErr != nil {
			fmt.Printf("failed to close migration source: %v", srcErr)
		}

		if dbErr != nil {
			fmt.Printf("failed to close migration target: %v", dbErr)
		}

	}(migrationClient)

	// Run the up migrations
	if err := migrationClient.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migrate: failed to run migration: %w", err)
	}

	// Get the current version of the database and print it
	version, _, _ := migrationClient.Version()
	fmt.Printf("[INFO]: Successfully migrated DB to latest version: %v", version)

	return nil
}
