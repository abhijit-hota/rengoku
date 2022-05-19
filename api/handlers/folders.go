package handlers

import (
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	DB "github.com/abhijit-hota/rengoku/server/db"
	"github.com/abhijit-hota/rengoku/server/utils"
	"github.com/gin-gonic/gin"
)

type Tree struct {
	Id       int    `json:"id"`
	Name     string `json:"name"`
	Children []Tree `json:"children,omitempty"`
}

func GetLinkTree(ctx *gin.Context) {
	db := DB.GetDB()

	rows, err := db.Query(`SELECT id, name, path FROM folders`)
	utils.Must(err)
	defer rows.Close()

	linkTree := []Tree{}

	for rows.Next() {
		var linkID int
		var name string
		var path string

		err = rows.Scan(&linkID, &name, &path)
		utils.Must(err)

		pathArr := strings.Split(path+strconv.Itoa(linkID), "/")
		cursor := &linkTree

		depth := len(pathArr) - 1
		for index, idStr := range pathArr {
			id, _ := strconv.Atoi(idStr)

			foundIdx := utils.FindFunc(*cursor, func(node Tree) bool { return node.Id == id })
			if foundIdx == -1 {
				*cursor = append(*cursor, Tree{Id: id})
				foundIdx = len(*cursor) - 1
			}
			if depth == index {
				(*cursor)[foundIdx].Id = linkID
				(*cursor)[foundIdx].Name = name
			}
			cursor = &((*cursor)[foundIdx].Children)
		}
	}
	err = rows.Err()
	utils.Must(err)
	ctx.JSON(http.StatusOK, linkTree)
}

func GetRootFolders(ctx *gin.Context) {
	db := DB.GetDB()

	var query struct {
		Str string `form:"q"`
	}

	if err := ctx.BindQuery(&query); err != nil {
		return
	}

	dbQuery := "SELECT * FROM folders"
	if query.Str != "" {
		dbQuery += " WHERE name LIKE '%" + query.Str + "%'"
	}
	preparedStmt, err := db.Prepare(dbQuery)
	utils.Must(err)

	rows, err := preparedStmt.Query()
	defer rows.Close()
	utils.Must(err)

	folders := make([]DB.Folder, 0)
	for rows.Next() {
		var folder DB.Folder

		rows.Scan(
			&folder.ID,
			&folder.Name,
			&folder.Path,
			&folder.Created,
			&folder.LastUpdated,
		)
		folders = append(folders, folder)
	}
	if rows.Err() != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "INTERNAL_ERROR"})
	}
	ctx.JSON(http.StatusOK, folders)
}

var re = regexp.MustCompile(`(.*/|)(\d{1,})/$`)

func CreateFolder(ctx *gin.Context) {
	db := DB.GetDB()

	var req struct {
		NameRequest
		Path string `json:"path,omitempty" form:"path"`
	}
	if err := ctx.Bind(&req); err != nil {
		return
	}

	now := time.Now().Unix()

	split := re.FindStringSubmatch(req.Path)
	if len(split) > 0 {
		parentPath, immediateParent := split[1], split[2]

		query := "SELECT COUNT(*) FROM folders WHERE id = ? AND path = ?"
		result := db.QueryRow(query, immediateParent, parentPath)
		var numId int
		result.Scan(&numId)

		if numId != 1 {
			ctx.JSON(http.StatusBadRequest, gin.H{"message": "INVALID_FOLDER_PATH"})
			return
		}
	} else {
		req.Path = ""
	}

	stmt := "INSERT INTO folders (name, path, created, last_updated) VALUES (?, ?, ?, ?)"
	res, err := db.Exec(stmt, req.Name, req.Path, now, now)
	if err != nil && strings.HasPrefix(err.Error(), "UNIQUE constraint failed") || utils.MustGet(res.RowsAffected()) == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": "NAME_ALREADY_PRESENT"})
		return
	}
	utils.Must(err)
	latestId := utils.MustGet(res.LastInsertId())
	folder := DB.Folder{
		ID:          latestId,
		Created:     now,
		LastUpdated: now,
		Name:        req.Name,
		Path:        req.Path,
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
	if err := ctx.Bind(&req); err != nil {
		return
	}

	tx, err := db.Begin()
	utils.Must(err)

	statement := "UPDATE folders SET name = ?, last_updated = ? WHERE id = ? AND name != ?"
	now := time.Now().Unix()

	_, err = tx.Exec(statement, req.Name, now, uri.ID, req.Name)
	if err != nil && strings.HasPrefix(err.Error(), "UNIQUE constraint failed") {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": "NAME_ALREADY_PRESENT"})
		return
	}
	utils.Must(err)

	var folder DB.Folder
	updatedTag := tx.QueryRow("SELECT * FROM folders WHERE id = ?", uri.ID)
	updatedTag.Scan(&folder.ID, &folder.Name, &folder.Path, &folder.Created, &folder.LastUpdated)

	tx.Commit()
	ctx.JSON(http.StatusOK, folder)
}

func DeleteFolder(ctx *gin.Context) {
	db := DB.GetDB()

	var uri IdUri
	if err := ctx.BindUri(&uri); err != nil {
		return
	}

	tx, _ := db.Begin()

	statement := "DELETE FROM folders WHERE id = ?"
	info, err := tx.Exec(statement, uri.ID)
	utils.Must(err)
	numDeleted, _ := info.RowsAffected()

	statement = "DELETE FROM links_folders WHERE folder_id = ?"
	_, err = tx.Exec(statement, uri.ID)
	utils.Must(err)

	tx.Commit()
	ctx.JSON(http.StatusOK, gin.H{"deleted": numDeleted == 1})
}
