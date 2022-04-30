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
		CREATE TABLE IF NOT EXISTS links (
			id INTEGER NOT NULL PRIMARY KEY, 
			title TEXT, 
			url TEXT, 
			created INTEGER, 
			last_updated INTEGER
		)
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
