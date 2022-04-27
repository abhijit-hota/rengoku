package main

import (
	"database/sql"
	"errors"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	_ "github.com/mattn/go-sqlite3"
	"golang.design/x/hotkey"
	"golang.design/x/hotkey/mainthread"
)

func handle(err error) {
	if err != nil {
		panic(err)
	}
}

var db *sql.DB

func GetTitleFromURL(url string) (title string, err error) {
	url = strings.TrimSpace(url)
	if !(strings.HasPrefix(url, "https://") || strings.HasPrefix(url, "https://")) {
		url = "https://" + url
	}
	res, err := http.Get(url)
	if err != nil {
		return "", errors.New("Invalid URL.")
	}

	defer res.Body.Close()
	resBytes, err := io.ReadAll(res.Body)
	str := string(resBytes)
	re := regexp.MustCompile("<title>(.+)</title>")

	title = re.FindStringSubmatch(str)[1]
	return title, nil
}

func SaveToDB(url string) {
	title, err := GetTitleFromURL(url)

	if err != nil {
		log.Println(err)
		return
	}

	stmt, err := db.Prepare("INSERT INTO LINKS (url, title, created, last_updated) VALUES (?, ?, ?, ?)")
	handle(err)
	defer stmt.Close()

	now := time.Now().Unix()
	_, err = stmt.Exec(url, title, now, now)
	handle(err)

	log.Println("Saved URL.")
}

func InitializeDB() {
	var err error

	db, err = sql.Open("sqlite3", "./links.db")
	handle(err)

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS links (
			id INTEGER NOT NULL PRIMARY KEY, 
			title text, 
			url text, 
			created integer, 
			last_updated integer
		)
	`)
	handle(err)
}

func ListenForHotKey() {
	hk := hotkey.New([]hotkey.Modifier{hotkey.ModCtrl, hotkey.ModShift}, hotkey.Key(0x2f))
	err := hk.Register()
	if err != nil {
		return
	}

	for range hk.Keydown() {
		url, err := clipboard.ReadAll()
		handle(err)
		SaveToDB(url)
	}
}
func main() {
	InitializeDB()
	defer db.Close()

	mainthread.Init(ListenForHotKey)
}
