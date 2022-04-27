package main

import (
	"database/sql"
	"io"
	"log"
	"net/http"
	"regexp"
	"time"

	"github.com/atotto/clipboard"
	_ "github.com/mattn/go-sqlite3"
	"golang.design/x/hotkey"
	"golang.design/x/hotkey/mainthread"
)

func handleErr(err error) {
	if err != nil {
		panic(err)
	}
}

var db *sql.DB

func GetTitleFromURL(text string) (title string) {
	res, err := http.Get(text)
	handleErr(err)

	defer res.Body.Close()
	resBytes, err := io.ReadAll(res.Body)
	str := string(resBytes)
	re := regexp.MustCompile("<title>(.+)</title>")

	title = re.FindStringSubmatch(str)[1]
	return
}

func SaveToDB(url string) {
	GetTitleFromURL(url)
	title := GetTitleFromURL(url)

	stmt, err := db.Prepare("INSERT INTO LINKS (url, title, created, last_updated) VALUES (?, ?, ?, ?)")
	handleErr(err)
	defer stmt.Close()

	now := time.Now().Unix()
	_, err = stmt.Exec(url, title, now, now)
	handleErr(err)

	log.Println("Saved URL.")
}

func InitializeDB() {
	var err error

	db, err = sql.Open("sqlite3", "./links.db")
	handleErr(err)

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS links (
			id INTEGER NOT NULL PRIMARY KEY, 
			title text, 
			url text, 
			created integer, 
			last_updated integer
		)
	`)
	handleErr(err)
}

func ListenForHotKey() {
	hk := hotkey.New([]hotkey.Modifier{hotkey.ModCtrl, hotkey.ModShift}, hotkey.Key(0x2f))
	err := hk.Register()
	if err != nil {
		return
	}

	for range hk.Keydown() {
		url, err := clipboard.ReadAll()
		if err != nil {
			panic(err)
		}
		SaveToDB(url)
	}
}
func main() {
	InitializeDB()
	defer db.Close()

	mainthread.Init(ListenForHotKey)
}
