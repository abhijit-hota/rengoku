package handlers

import (
	"api/common"
	DB "api/db"
	"api/utils"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func AddBookmark(ctx *gin.Context) {
	var body DB.Bookmark
	db := DB.GetDB()

	if err := ctx.BindJSON(&body); err != nil {
		return
	}
	if err := common.GetMetadata(body.URL, &body.Meta); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	tx, err := db.Begin()
	utils.Must(err)

	stmt := "INSERT INTO meta (title, description, favicon) VALUES (?, ?, ?)"
	info, err := tx.Exec(stmt, body.Meta.Title, body.Meta.Description, body.Meta.Favicon)
	metaID, _ := info.LastInsertId()

	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid URL."})
		return
	}

	now := time.Now().Unix()
	body.Created = now
	body.LastUpdated = now

	stmt = "INSERT INTO links (url, meta_id, created, last_updated) VALUES (?, ?, ?, ?)"
	info, err = tx.Exec(stmt, body.URL, metaID, now, now)
	utils.Must(err)

	if len(body.Tags) == 0 {
		tx.Commit()

		ctx.JSON(http.StatusOK, body)
		return
	}

	linkID, _ := info.LastInsertId()

	statement, err := tx.Prepare("INSERT OR IGNORE INTO tags (name, created, last_updated) VALUES (?, ?, ?)")
	utils.Must(err)
	defer statement.Close()

	for _, tag := range body.Tags {
		statement.Exec(tag, now, now)
	}

	query := fmt.Sprintf("SELECT id FROM tags WHERE name IN (%s)", strings.TrimRight(strings.Repeat("?,", len(body.Tags)), ","))
	tagIDs, err := tx.Query(query, utils.ToGenericArray(body.Tags)...)
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
	ctx.JSON(http.StatusOK, body)
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

	rows, err := db.Query(`SELECT links.id, links.url, links.created, links.last_updated, group_concat(tags.name), meta.title, meta.favicon, meta.description
							FROM links 
							JOIN meta ON meta.id = links.meta_id 
							JOIN links_tags ON links_tags.link_id = links.id 
							JOIN tags ON tags.id = links_tags.tag_id GROUP BY links.id`)
	utils.Must(err)
	defer rows.Close()

	bookmarks := make([]DB.Bookmark, 0)
	for rows.Next() {
		var bm DB.Bookmark
		var tagStr string
		err = rows.Scan(
			&bm.ID,
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

func DeleteBookmarkTag(ctx *gin.Context) {
	db := DB.GetDB()
	var uri struct {
		IdUri
		TagId int `uri:"tagId" binding:"required"`
	}

	if err := ctx.BindUri(&uri); err != nil {
		return
	}

	stmt := "DELETE FROM links_tags WHERE link_id = ? AND tag_id = ?"
	info, _ := db.Exec(stmt, uri.ID, uri.TagId)
	numDeleted, _ := info.RowsAffected()

	ctx.JSON(http.StatusOK, gin.H{"deleted": numDeleted == 1})
}

func AddBookmarkTag(ctx *gin.Context) {
	db := DB.GetDB()
	var uri IdUri
	var newTag struct {
		Name string `json:"name" form:"name" binding:"required"`
	}

	if err := ctx.BindUri(&uri); err != nil {
		return
	}
	if err := ctx.Bind(&newTag); err != nil {
		return
	}

	tx, _ := db.Begin()
	now := time.Now().Unix()

	stmt := "INSERT OR IGNORE INTO tags (name, created, last_updated) VALUES (?, ?, ?)"
	_, err := tx.Exec(stmt, newTag.Name, now, now)
	utils.Must(err)

	query := "SELECT id FROM tags WHERE name = ?"
	tag := tx.QueryRow(query, newTag.Name)
	var tagID int
	tag.Scan(&tagID)
	utils.Must(err)

	stmt = "INSERT OR IGNORE INTO links_tags (tag_id, link_id) VALUES (?, ?)"
	info, err := tx.Exec(stmt, tagID, uri.ID)
	utils.Must(err)
	updatedLinks, _ := info.RowsAffected()

	tx.Commit()
	ctx.JSON(http.StatusOK, gin.H{"added": updatedLinks == 1})
}

func DeleteBookmark(ctx *gin.Context) {
	db := DB.GetDB()
	var uri IdUri

	if err := ctx.BindUri(&uri); err != nil {
		return
	}

	tx, _ := db.Begin()

	stmt := "DELETE FROM links WHERE id = ?"
	info, _ := tx.Exec(stmt, uri.ID)
	numDeleted, _ := info.RowsAffected()

	stmt = "DELETE FROM links_tags WHERE link_id = ?"
	tx.Exec(stmt, uri.ID)

	utils.Must(tx.Commit())

	ctx.JSON(http.StatusOK, gin.H{"deleted": numDeleted == 1})
}
