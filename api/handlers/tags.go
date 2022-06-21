package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	DB "github.com/abhijit-hota/rengoku/server/db"
	"github.com/abhijit-hota/rengoku/server/utils"

	"github.com/gin-gonic/gin"
)

type IdUri struct {
	ID int64 `uri:"id" binding:"required"`
}

type NameRequest struct {
	Name string `json:"name" form:"name" binding:"required"`
}
type BulkNameRequest struct {
	Names []string `json:"names" binding:"required"`
}

func GetAllTags(ctx *gin.Context) {
	db := DB.GetDB()

	var query struct {
		Str string `form:"q"`
	}

	if err := ctx.BindQuery(&query); err != nil {
		return
	}

	dbQuery := "SELECT * FROM tags"
	if query.Str != "" {
		dbQuery += " WHERE name LIKE '%" + query.Str + "%'"
	}
	preparedStmt, err := db.Prepare(dbQuery)
	utils.Must(err)

	rows, err := preparedStmt.Query()
	defer rows.Close()
	utils.Must(err)

	tags := make([]DB.Tag, 0)
	for rows.Next() {
		var id int64
		var tag string
		var createdAt int64
		var lastUpdated int64

		rows.Scan(
			&id,
			&tag,
			&createdAt,
			&lastUpdated,
		)
		tags = append(tags, DB.Tag{ID: id, Name: tag, CreatedAt: createdAt, LastUpdated: lastUpdated})
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

	stmt := "INSERT INTO tags (name, created_at, last_updated) VALUES (?, ?, ?)"
	res, err := db.Exec(stmt, req.Name, now, now)
	if err != nil && strings.HasPrefix(err.Error(), "UNIQUE constraint failed") {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": "NAME_ALREADY_PRESENT"})
		return
	}
	utils.Must(err)

	tag := DB.Tag{CreatedAt: now, LastUpdated: now, Name: req.Name}
	tag.ID, _ = res.LastInsertId()

	ctx.JSON(http.StatusOK, tag)
}

func CreateBulkTags(ctx *gin.Context) {
	db := DB.GetDB()

	var req BulkNameRequest
	if err := ctx.BindJSON(&req); err != nil {
		return
	}

	now := time.Now().Unix()

	stmt := "INSERT OR IGNORE INTO tags (name, created_at, last_updated) VALUES (?, ?, ?)"

	var tags []DB.Tag
	for _, name := range req.Names {
		res, _ := db.Exec(stmt, name, now, now)

		tag := DB.Tag{CreatedAt: now, LastUpdated: now, Name: name}
		tag.ID, _ = res.LastInsertId()
		tags = append(tags, tag)
	}

	ctx.JSON(http.StatusOK, tags)
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
	updatedTag.Scan(&tag.ID, &tag.Name, &tag.CreatedAt, &tag.LastUpdated)

	tx.Commit()
	ctx.JSON(http.StatusOK, tag)
}

func DeleteTag(ctx *gin.Context) {
	db := DB.GetDB()

	var uri IdUri
	if err := ctx.BindUri(&uri); err != nil {
		return
	}

	statement := "DELETE FROM tags WHERE id = ?"
	info, err := db.Exec(statement, uri.ID)
	utils.Must(err)
	numDeleted, _ := info.RowsAffected()

	ctx.JSON(http.StatusOK, gin.H{"deleted": numDeleted == 1})
}
