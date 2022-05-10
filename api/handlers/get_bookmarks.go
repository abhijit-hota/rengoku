package handlers

import (
	DB "bingo/api/db"
	"bingo/api/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func GetBookmarks(ctx *gin.Context) {
	db := DB.GetDB()

	rows, err := db.Query(`SELECT links.url, links.created, links.last_updated, group_concat(tags.path), meta.title, meta.favicon, meta.description
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
