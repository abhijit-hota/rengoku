package handlers

import (
	DB "api/db"
	"api/utils"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type Node struct {
	Children Tree  `json:"children"`
	Links    []int `json:"links"`
}
type Tree map[string]*Node

func GetLinkTree(ctx *gin.Context) {
	db := DB.GetDB()

	rows, err := db.Query(`SELECT links_tags.link_id, tags.path FROM links_tags JOIN tags ON tags.id = links_tags.tag_id;`)
	utils.Must(err)
	defer rows.Close()

	linktree := make(Tree)

	for rows.Next() {
		var linkID int
		var path string

		err = rows.Scan(&linkID, &path)
		utils.Must(err)

		pathArr := strings.Split(path, "/")
		depth := len(pathArr) - 1
		cursor := linktree

		for index, tag := range pathArr {
			if cursor[tag] == nil {
				cursor[tag] = &Node{make(Tree), make([]int, 0)}
			}
			if index == depth {
				cursor[tag].Links = append(cursor[tag].Links, linkID)
			}
			cursor = cursor[tag].Children
		}
	}
	err = rows.Err()
	utils.Must(err)
	ctx.JSON(http.StatusOK, linktree)
}

type IdUri struct {
	ID int64 `uri:"id" binding:"required"`
}

type NameRequest struct {
	Name string `json:"name" form:"name" binding:"required"`
}

func GetAllTags(ctx *gin.Context) {
	db := DB.GetDB()

	rows, err := db.Query(`SELECT * FROM tags`)
	defer rows.Close()
	utils.Must(err)

	tags := make([]DB.Tag, 0)
	for rows.Next() {
		var id int64
		var tag string
		var created int64
		var lastUpdated int64

		rows.Scan(
			&id,
			&tag,
			&created,
			&lastUpdated,
		)
		tags = append(tags, DB.Tag{ID: id, Name: tag, Created: created, LastUpdated: lastUpdated})
	}
	if rows.Err() != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "INTERNAL_ERROR"})
	}
	ctx.JSON(http.StatusOK, tags)
}

func CreateTag(ctx *gin.Context) {
	db := DB.GetDB()

	var req NameRequest
	if err := ctx.Bind(&req); err != nil {
		return
	}

	now := time.Now().Unix()

	stmt := "INSERT INTO tags (name, created, last_updated) VALUES (?, ?, ?)"
	res, err := db.Exec(stmt, req.Name, now, now)
	if err != nil && strings.HasPrefix(err.Error(), "UNIQUE constraint failed") {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": "NAME_ALREADY_PRESENT"})
		return
	}
	utils.Must(err)

	tag := DB.Tag{Created: now, LastUpdated: now, Name: req.Name}
	tag.ID, _ = res.LastInsertId()

	ctx.JSON(http.StatusOK, tag)
}

func UpdateTagName(ctx *gin.Context) {
	db := DB.GetDB()

	var uri IdUri
	if err := ctx.BindUri(&uri); err != nil {
		return
	}

	var req NameRequest
	if err := ctx.Bind(&req); err != nil {
		return
	}

	tx, err := db.Begin()
	utils.Must(err)

	statement := "UPDATE tags SET name = ?, last_updated = ? WHERE id = ? AND name != ?"
	now := time.Now().Unix()

	info, err := tx.Exec(statement, req.Name, now, uri.ID, req.Name)
	if err != nil && strings.HasPrefix(err.Error(), "UNIQUE constraint failed") {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": "NAME_ALREADY_PRESENT"})
		return
	}
	utils.Must(err)

	numUpdated, _ := info.RowsAffected()
	fmt.Println(numUpdated)

	var tag DB.Tag
	updatedTag := tx.QueryRow("SELECT * FROM tags WHERE id = ?", uri.ID)
	updatedTag.Scan(&tag.ID, &tag.Name, &tag.Created, &tag.LastUpdated)

	tx.Commit()
	ctx.JSON(http.StatusOK, tag)
}

func DeleteTag(ctx *gin.Context) {
	db := DB.GetDB()

	var uri IdUri
	if err := ctx.BindUri(&uri); err != nil {
		return
	}

	tx, _ := db.Begin()

	statement := "DELETE FROM tags WHERE id = ?"
	info, err := tx.Exec(statement, uri.ID)
	utils.Must(err)
	numDeleted, _ := info.RowsAffected()

	statement = "DELETE FROM links_tags WHERE tag_id = ?"
	_, err = tx.Exec(statement, uri.ID)
	utils.Must(err)

	ctx.JSON(http.StatusOK, gin.H{"deleted": numDeleted == 1})
}
