package db

import (
	"fmt"
	"os"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

type DB struct {
	sqlx.DB
	Initialized bool
}

var db DB

func InitializeDB() (db DB) {
	dbPath := os.Getenv("DB_PATH") + "?_pragma=foreign_keys(1)"

	dsn := "file:" + dbPath
	db.DB = *sqlx.MustOpen("sqlite", dsn)

	if err := migrateToLatest(db.DB.DB); err != nil {
		panic(fmt.Errorf("not migrated: %w", err))
	}

	db.Initialized = true
	return db
}

func GetDB() DB {
	if !db.Initialized {
		db = InitializeDB()
	}
	return db
}
