package handlers

import (
	"bingo/api/common"
	DB "bingo/api/db"
	"bingo/api/utils"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

var config utils.Config

func init() {
	config = utils.GetConfig()
}

func AddBookmark(ctx *gin.Context) {
	var json DB.Bookmark
	db := DB.GetDB()

	if err := ctx.ShouldBindJSON(&json); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// hasAutotagRule = utils.Contains[]() config.AutoTagRules
	if err := common.GetMetadata(json.URL, &json.Meta); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	tx, err := db.Begin()
	utils.Must(err)

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

	statement, err = tx.Prepare("INSERT INTO links (url, meta_id, created, last_updated) VALUES (?, ?, ?, ?)")
	utils.Must(err)
	defer statement.Close()

	now := time.Now().Unix()
	json.Created = now
	json.LastUpdated = now

	info, err = statement.Exec(json.URL, metaID, now, now)
	utils.Must(err)

	if len(json.Tags) == 0 {
		tx.Commit()

		ctx.JSON(http.StatusOK, json)
		return
	}

	linkID, _ := info.LastInsertId()

	statement, err = tx.Prepare("INSERT OR IGNORE INTO tags (path, created, last_updated, is_root) VALUES (?, ?, ?, ?)")
	utils.Must(err)
	defer statement.Close()

	tags := make([]any, len(json.Tags))

	for i, tag := range json.Tags {
		isRoot := !strings.ContainsRune(tag, '/')
		statement.Exec(tag, now, now, isRoot)
		tags[i] = tag
	}

	query := fmt.Sprintf("SELECT id FROM tags WHERE path IN (%s)", strings.TrimRight(strings.Repeat("?,", len(tags)), ","))
	tagIDs, err := tx.Query(query, tags...)
	defer tagIDs.Close()
	utils.Must(err)

	statement, err = tx.Prepare("INSERT INTO links_tags (tag_id, link_id) VALUES (?, ?)")
	utils.Must(err)
	defer statement.Close()

	for tagIDs.Next() {
		var tagID int
		tagIDs.Scan(&tagID)
		statement.Exec(tagID, linkID)
	}
	err = tagIDs.Err()
	utils.Must(err)

	tx.Commit()
	ctx.JSON(http.StatusOK, json)
}
