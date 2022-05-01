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

func getMetadataFromURL(url string, availableMetadata DB.Meta) (meta *DB.Meta, err error) {

	url = strings.TrimSpace(url)
	if !(strings.HasPrefix(url, "https://") || strings.HasPrefix(url, "https://")) {
		url = "https://" + url
	}
	// TODO check url before passing into http
	res, err := http.Get(url)
	if err != nil {
		return nil, errors.New("Invalid URL.")
	}

	defer res.Body.Close()

	hm := new(DB.Meta)
	data, _ := io.ReadAll(res.Body)
	headRegex := regexp.MustCompile("<head>((.|\n|\r\n)+)</head>")

	head := string(headRegex.Find(data))
	head = strings.ReplaceAll(head, "\n", "")

	if availableMetadata.Title == "" {
		titleRegex := regexp.MustCompile(`<title.*>(.+)<\/title>`)
		metaTitleRegex := regexp.MustCompile(`<meta.*?property="og:title".*?content="(.+?)".*?\/?>`)
		titleMatches := titleRegex.FindStringSubmatch(head)
		if len(titleMatches) == 0 {
			titleMatches = metaTitleRegex.FindStringSubmatch(head)
		}

		if len(titleMatches) == 0 {
			hm.Title = ""
		} else {
			hm.Title = titleMatches[1]
		}
	} else {
		hm.Title = availableMetadata.Title
	}

	if availableMetadata.Description == "" {

		descriptionRegex := regexp.MustCompile(`<meta.*?(?:name="description"|property="og:description").*?content="(.*?)".*?\/>`)
		descMatches := descriptionRegex.FindStringSubmatch(head)
		if len(descMatches) == 0 {
			hm.Description = ""
		} else {
			hm.Description = descMatches[1]
		}
	} else {
		hm.Description = availableMetadata.Description
	}
	hm.Favicon = availableMetadata.Favicon
	return hm, nil
}

func AddBookmark(ctx *gin.Context) {
	var json DB.Bookmark
	if err := ctx.ShouldBindJSON(&json); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	meta, err := getMetadataFromURL(json.URL, json.Meta)
	statement, err := db.Prepare("INSERT INTO meta (title, description, favicon) VALUES (?, ?, ?)")
	utils.Must(err)
	defer statement.Close()
	info, err := statement.Exec(meta.Title, meta.Description, meta.Favicon)
	metaID, _ := info.LastInsertId()

	if err != nil {
		log.Println(err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid URL."})
		return
	}

	statement, err = db.Prepare("INSERT INTO bookmarks (url, meta, created, last_updated) VALUES (?, ?, ?, ?)")
	utils.Must(err)
	defer statement.Close()

	now := time.Now().Unix()
	_, err = statement.Exec(json.URL, metaID, now, now)
	utils.Must(err)

	ctx.String(http.StatusOK, "Saved URL.")
}
