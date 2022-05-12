package handlers

import (
	DB "bingo/api/db"
	"bingo/api/utils"
	"net/http"
	"strings"

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
