package handlers

import (
	"bingo/api/common"
	DB "bingo/api/db"
	"bingo/api/utils"
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

var db *sql.DB = DB.GetDB()

func AddBookmark(ctx *gin.Context) {
	var json DB.Bookmark

	if err := ctx.ShouldBindJSON(&json); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := common.GetMetadata(json.URL, &json.Meta); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

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
