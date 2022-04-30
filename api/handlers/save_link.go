package handlers

import (
	DB "bingo/api/db"
	"bingo/api/utils"
	"database/sql"
	"errors"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

var db *sql.DB = DB.GetDB()

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

func SaveToDB(ctx *gin.Context) {
	raw, _ := io.ReadAll(ctx.Request.Body)
	url := string(raw)
	title, err := GetTitleFromURL(url)

	if err != nil {
		log.Println(err)
		return
	}

	stmt, err := db.Prepare("INSERT INTO LINKS (url, title, created, last_updated) VALUES (?, ?, ?, ?)")
	utils.Must(err)
	defer stmt.Close()

	now := time.Now().Unix()
	_, err = stmt.Exec(url, title, now, now)
	utils.Must(err)

	ctx.String(http.StatusOK, "Saved URL.")
}
