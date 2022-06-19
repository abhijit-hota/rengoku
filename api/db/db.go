package db

import (
	"os"

	"github.com/abhijit-hota/rengoku/server/utils"
	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

var db *sqlx.DB

func InitializeDB() (db *sqlx.DB) {
	var err error
	dsn := "file:" + os.Getenv("DB_PATH") + "?_pragma=foreign_keys(1)"
	db, err = sqlx.Open("sqlite", dsn)
	utils.Must(err)

	t := `
CREATE TABLE IF NOT EXISTS links (
	id INTEGER NOT NULL PRIMARY KEY, 
	url TEXT NOT NULL UNIQUE,
	created INTEGER, 
	last_updated INTEGER,
	last_saved_offline INTEGER DEFAULT 0
);
CREATE TABLE IF NOT EXISTS meta (
	id INTEGER NOT NULL PRIMARY KEY,
	link_id INTEGER NOT NULL UNIQUE REFERENCES links(id) ON DELETE CASCADE,
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
	tag_id INTEGER NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
	link_id INTEGER NOT NULL REFERENCES links(id) ON DELETE CASCADE,
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
	folder_id TEXT NOT NULL REFERENCES folders(id) ON DELETE CASCADE,
	link_id INTEGER NOT NULL REFERENCES links(id) ON DELETE CASCADE,
	UNIQUE(folder_id, link_id) ON CONFLICT IGNORE
);`

	_, err = db.Exec(t)
	utils.Must(err)
	return db
}

func GetDB() *sqlx.DB {
	if db == nil {
		db = InitializeDB()
	}
	return db
}
