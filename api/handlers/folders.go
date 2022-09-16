package handlers

import (
	"encoding/json"
	"net/http"

	DB "github.com/abhijit-hota/rengoku/server/db"
	"github.com/abhijit-hota/rengoku/server/utils"
	"github.com/gin-gonic/gin"
)

type Tree struct {
	Id       int64  `json:"id"`
	Name     string `json:"name"`
	Children []Tree `json:"children,omitempty"`
}

func transform(leaves []DB.Folder, parentId int64) []Tree {
	trees := make([]Tree, 0)
	for _, leaf := range leaves {
		if leaf.ParentID == parentId {
			trees = append(trees, Tree{
				Id:       leaf.ID,
				Name:     leaf.Name,
				Children: transform(leaves, leaf.ID),
			})
		}
	}
	return trees
}

func GetLinkTree(ctx *gin.Context) {
	db := DB.GetDB()

	rows := make([]DB.Folder, 0)
	err := db.Select(&rows, `SELECT id, name, parent_id FROM folders`)
	utils.Must(err)

	linkTree := transform(rows, 0)

	utils.Must(err)
	ctx.JSON(http.StatusOK, linkTree)
}

func GetFolders(ctx *gin.Context) {
	db := DB.GetDB()

	var query struct {
		Str string `form:"q"`
	}

	if err := ctx.BindQuery(&query); err != nil {
		return
	}

	dbQuery := "SELECT * FROM folders"
	if query.Str != "" {
		dbQuery += " WHERE name LIKE ?"
	}

	rows, err := db.Queryx(dbQuery, "%"+query.Str+"%")
	utils.Must(err)
	defer rows.Close()

	folders := make([]DB.Folder, 0)
	for rows.Next() {
		var folder DB.Folder
		rows.StructScan(&folder)
		folders = append(folders, folder)
	}
	if rows.Err() != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "INTERNAL_ERROR"})
	}
	ctx.JSON(http.StatusOK, folders)
}

func CreateFolder(ctx *gin.Context) {
	db := DB.GetDB()

	var req struct {
		NameRequest
		ParentID json.Number `json:"parent_id" form:"parent_id"`
	}
	req.ParentID = json.Number("0")

	if err := ctx.BindJSON(&req); err != nil {
		return
	}

	if val, err := req.ParentID.Int64(); val != 0 && err != nil {
		var count int
		err := db.Get(&count, "SELECT COUNT(1) from folders WHERE id = ?", req.ParentID)

		if err != nil || count != 1 {
			ctx.JSON(http.StatusBadRequest, gin.H{"message": "INVALID_FOLDER_PATH"})
			return
		}
	}

	var folder DB.Folder

	stmt := "INSERT INTO folders (name, parent_id) VALUES (?, ?) RETURNING *"
	row := db.QueryRowx(stmt, req.Name, req.ParentID)
	err := row.StructScan(&folder)
	if err != nil {
		if DB.IsUniqueErr(err) {
			ctx.JSON(http.StatusBadRequest, gin.H{"cause": "NAME_ALREADY_PRESENT"})
			return
		}
		panic(err)
	}

	ctx.JSON(http.StatusOK, folder)
}

func UpdateFolderName(ctx *gin.Context) {
	db := DB.GetDB()

	var uri IdUri

	if err := ctx.BindUri(&uri); err != nil {
		return
	}

	var req NameRequest
	if err := ctx.BindJSON(&req); err != nil {
		return
	}

	tx := db.MustBegin()

	statement := "UPDATE folders SET name = ? WHERE id = ? AND name != ?"

	_, err := tx.Exec(statement, req.Name, uri.ID, req.Name)
	if err != nil {
		if DB.IsUniqueErr(err) {
			ctx.JSON(http.StatusBadRequest, gin.H{"cause": "NAME_ALREADY_PRESENT"})
			return
		}
		panic(err)
	}

	var folder DB.Folder
	query := "SELECT * FROM folders WHERE id = ?"
	tx.Get(&folder, query, uri.ID)

	tx.Commit()
	ctx.JSON(http.StatusOK, folder)
}

func DeleteFolder(ctx *gin.Context) {
	db := DB.GetDB()

	var uri IdUri
	if err := ctx.BindUri(&uri); err != nil {
		return
	}

	statement := "DELETE FROM folders WHERE id = ?"
	info := db.MustExec(statement, uri.ID)
	numDeleted, _ := info.RowsAffected()

	ctx.JSON(http.StatusOK, gin.H{"deleted": numDeleted == 1})
}
