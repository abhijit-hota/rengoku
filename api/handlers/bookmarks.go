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

func AddBookmark(ctx *gin.Context) {
	var json DB.Bookmark
	db := DB.GetDB()

	if err := ctx.ShouldBindJSON(&json); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
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

	stmtStr := "INSERT OR IGNORE INTO tags (name, created, last_updated) VALUES"

	numValues := 0
	execValues := []any{}

	for _, tag := range json.Tags {
		execValues = append(execValues, tag, now, now)
	}
	stmtStr += strings.TrimRight(strings.Repeat("(?, ?, ?, ?),", numValues), ",")
	statement, _ = tx.Prepare(stmtStr)
	statement.Exec(execValues...)

	query := fmt.Sprintf("SELECT id FROM tags WHERE name IN (%s)", strings.TrimRight(strings.Repeat("?,", len(json.Tags)), ","))
	tagIDs, err := tx.Query(query, utils.ToGenericArray(json.Tags)...)
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

type Filter struct {
	Tags []int
	From int
	To   int
}

type Sort struct {
	AlphabeticTitle string
}

func GetBookmarks(ctx *gin.Context) {
	db := DB.GetDB()

	rows, err := db.Query(`SELECT links.url, links.created, links.last_updated, group_concat(tags.name), meta.title, meta.favicon, meta.description
							FROM links 
							JOIN meta ON meta.id = links.meta_id 
							JOIN links_tags ON links_tags.link_id = links.id 
							JOIN tags ON tags.id = links_tags.tag_id GROUP BY links.url`)
	utils.Must(err)
	defer rows.Close()

	bookmarks := make([]DB.Bookmark, 0)
	for rows.Next() {
		var bm DB.Bookmark
		var tagStr string
		err = rows.Scan(
			&bm.URL,
			&bm.Created,
			&bm.LastUpdated,
			&tagStr,
			&bm.Meta.Title,
			&bm.Meta.Favicon,
			&bm.Meta.Description,
		)
		utils.Must(err)
		bm.Tags = strings.Split(tagStr, ",")
		bookmarks = append(bookmarks, bm)
	}
	err = rows.Err()
	utils.Must(err)
	ctx.JSON(http.StatusOK, bookmarks)
}
