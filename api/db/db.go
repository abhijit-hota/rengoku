package db

import (
	"bingo/api/utils"
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func InitializeDB() (db *sql.DB) {
	var err error
	conf := utils.GetConfig()
	db, err = sql.Open("sqlite3", conf.DatabasePath)
	utils.Must(err)

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS bookmarks (
			id INTEGER NOT NULL PRIMARY KEY, 
			meta_id INTEGER,
			url TEXT,
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
			path TEXT NOT NULL UNIQUE,
			is_root INTEGER,
			created INTEGER,
			last_updated INTEGER
		);
		CREATE TABLE IF NOT EXISTS bookmarks_tags (
			tag_id INTEGER NOT NULL REFERENCES tags(id),
			bookmark_id INTEGER NOT NULL REFERENCES bookmarks(id),
			UNIQUE(tag_id, bookmark_id) ON CONFLICT IGNORE
		);
	`)
	utils.Must(err)
	return db
}

func GetDB() *sql.DB {
	if db == nil {
		db = InitializeDB()
	}
	return db
}
