package db

import (
	"bingo/api/utils"
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func InitializeDB() (db *sql.DB) {
	var err error
	db, err = sql.Open("sqlite3", "../links.db")
	utils.Handle(err)

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS links (
			id INTEGER NOT NULL PRIMARY KEY, 
			title text, 
			url text, 
			created integer, 
			last_updated integer
		)
		`)
	utils.Handle(err)
	return db
}

func GetDB() *sql.DB {
	if db == nil {
		db = InitializeDB()
	}
	return db
}
