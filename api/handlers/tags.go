package handlers

import (
	"net/http"

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
		dbQuery += " WHERE name LIKE ?"
	}
	rows, err := db.Queryx(dbQuery, "%"+query.Str+"%")
	defer rows.Close()
	utils.Must(err)

	tags := make([]DB.Tag, 0)
	for rows.Next() {
		tag := DB.Tag{}
		rows.StructScan(&tag)
		tags = append(tags, tag)
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

	var tag DB.Tag
	stmt := "INSERT INTO tags (name) VALUES (?) RETURNING *"
	row := db.QueryRowx(stmt, req.Name)
	err := row.StructScan(&tag)
	if err != nil {
		if DB.IsUniqueErr(err) {
			ctx.JSON(http.StatusBadRequest, gin.H{"code": "NAME_ALREADY_PRESENT"})
			return
		}
		panic(err)
	}

	ctx.JSON(http.StatusOK, tag)
}

func CreateBulkTags(ctx *gin.Context) {
	db := DB.GetDB()
	var req struct {
		Names []string `json:"names" binding:"required"`
	}
	if err := ctx.BindJSON(&req); err != nil {
		return
	}

	stmt := "INSERT OR IGNORE INTO tags (name) VALUES (?) RETURNING *"

	var tags []DB.Tag
	for _, name := range req.Names {
		var tag DB.Tag
		row := db.QueryRowx(stmt, name)
		utils.Must(row.StructScan(&tag))
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

	tx := db.MustBegin()

	statement := "UPDATE tags SET name = ? WHERE id = ? AND name != ?"

	_, err := tx.Exec(statement, req.Name, uri.ID, req.Name)
	if err != nil {
		if DB.IsUniqueErr(err) {
			ctx.JSON(http.StatusBadRequest, gin.H{"code": "NAME_ALREADY_PRESENT"})
			return
		}
		panic(err)
	}

	var tag DB.Tag
	row := tx.QueryRowx("SELECT * FROM tags WHERE id = ?", uri.ID)
	utils.Must(row.StructScan(&tag))

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
	info := db.MustExec(statement, uri.ID)
	numDeleted, _ := info.RowsAffected()

	ctx.JSON(http.StatusOK, gin.H{"deleted": numDeleted == 1})
}
