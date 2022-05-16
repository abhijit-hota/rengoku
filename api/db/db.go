package db

import (
	"github.com/abhijit-hota/rengoku-server/utils"
	"database/sql"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func InitializeDB() (db *sql.DB) {
	var err error
	db, err = sql.Open("sqlite3", os.Getenv("DB_PATH"))
	utils.Must(err)

	t := `
CREATE TABLE IF NOT EXISTS links (
	id INTEGER NOT NULL PRIMARY KEY, 
	meta_id INTEGER,
	url TEXT NOT NULL UNIQUE,
	created INTEGER, 
	last_updated INTEGER,
	FOREIGN KEY (meta_id) REFERENCES meta(id)
);
CREATE TABLE IF NOT EXISTS meta (
	id INTEGER NOT NULL PRIMARY KEY,
	title TEXT,
	description TEXT,
	favicon TEXT
);
CREATE TABLE IF NOT EXISTS tags (
	id INTEGER NOT NULL PRIMARY KEY,
	name TEXT NOT NULL UNIQUE,
	created INTEGER,
	last_updated INTEGER
);
CREATE TABLE IF NOT EXISTS links_tags (
	tag_id INTEGER NOT NULL REFERENCES tags(id),
	link_id INTEGER NOT NULL REFERENCES links(id),
	UNIQUE(tag_id, link_id) ON CONFLICT IGNORE
);`

	_, err = db.Exec(t)

	utils.Must(err)
	return db
}

func GetDB() *sql.DB {
	if db == nil {
		db = InitializeDB()
	}
	return db
}
