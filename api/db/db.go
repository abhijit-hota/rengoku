package db

import (
	"database/sql"
	"os"

	"github.com/abhijit-hota/rengoku/server/utils"
	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func InitializeDB() (db *sql.DB) {
	var err error
	db, err = sql.Open("sqlite3", os.Getenv("DB_PATH")+"?foreign_keys=1")
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
);
CREATE TABLE IF NOT EXISTS folders (
	id INTEGER NOT NULL PRIMARY KEY,
	name TEXT NOT NULL,
	path TEXT,
	created INTEGER,
	last_updated INTEGER,
	UNIQUE(name, path)
);
CREATE INDEX IF NOT EXISTS folder_path ON folders(path);
CREATE TABLE IF NOT EXISTS links_folders (
	folder_id TEXT NOT NULL REFERENCES folders(id),
	link_id INTEGER NOT NULL REFERENCES links(id),
	UNIQUE(folder_id, link_id) ON CONFLICT IGNORE
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
