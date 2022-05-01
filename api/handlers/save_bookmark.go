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

func getMetadataFromURL(url string, metadata *DB.Meta) (err error) {

	url = strings.TrimSpace(url)
	if !(strings.HasPrefix(url, "https://") || strings.HasPrefix(url, "https://")) {
		url = "https://" + url
	}
	// TODO check url before passing into http
	res, err := http.Get(url)
	if err != nil {
		return errors.New("Invalid URL.")
	}

	defer res.Body.Close()

	data, _ := io.ReadAll(res.Body)
	headRegex := regexp.MustCompile("<head>((.|\n|\r\n)+)</head>")

	head := string(headRegex.Find(data))
	head = strings.ReplaceAll(head, "\n", "")

	if metadata.Title == "" {
		titleRegex := regexp.MustCompile(`<title.*>(.+)<\/title>`)
		metaTitleRegex := regexp.MustCompile(`<meta.*?property="og:title".*?content="(.+?)".*?\/?>`)
		titleMatches := titleRegex.FindStringSubmatch(head)
		if len(titleMatches) == 0 {
			titleMatches = metaTitleRegex.FindStringSubmatch(head)
		}

		if len(titleMatches) == 0 {
			metadata.Title = ""
		} else {
			metadata.Title = titleMatches[1]
		}
	}

	if metadata.Description == "" {
		descriptionRegex := regexp.MustCompile(`<meta.*?(?:name="description"|property="og:description").*?content="(.*?)".*?\/>`)
		descMatches := descriptionRegex.FindStringSubmatch(head)
		if len(descMatches) == 0 {
			metadata.Description = ""
		} else {
			metadata.Description = descMatches[1]
		}
	}
	return nil
}

func AddBookmark(ctx *gin.Context) {
	var json DB.Bookmark

	if err := ctx.ShouldBindJSON(&json); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := getMetadataFromURL(json.URL, &json.Meta)

	statement, err := db.Prepare("INSERT INTO meta (title, description, favicon) VALUES (?, ?, ?)")
	utils.Must(err)
	defer statement.Close()

	info, err := statement.Exec(json.Meta.Title, json.Meta.Description, json.Meta.Favicon)
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
	json.Created = now
	json.LastUpdated = now

	_, err = statement.Exec(json.URL, metaID, now, now)
	utils.Must(err)

	ctx.JSON(http.StatusOK, json)
}
