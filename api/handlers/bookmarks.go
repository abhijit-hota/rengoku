package handlers

import (
	"api/common"
	DB "api/db"
	"api/utils"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/purell"
	"github.com/gin-gonic/gin"
)

func AddBookmark(ctx *gin.Context) {
	var body DB.Bookmark
	db := DB.GetDB()

	if err := ctx.BindJSON(&body); err != nil {
		return
	}

	link, err := purell.NormalizeURLString(body.URL, purell.FlagsSafe)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	body.URL = link

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

	stmt = "INSERT OR IGNORE INTO links (url, meta_id, created, last_updated) VALUES (?, ?, ?, ?)"
	info, err = tx.Exec(stmt, body.URL, metaID, now, now)
	utils.Must(err)

	num, _ := info.RowsAffected()
	if num == 0 {
		body.Tags = []string{}
		tx.Exec("DELETE FROM meta WHERE id = ?", metaID)
		tx.Commit()

		ctx.JSON(http.StatusBadRequest, gin.H{"message": "URL_ALREADY_PRESENT"})
		return
	}

	if len(body.Tags) == 0 {
		body.Tags = []string{}
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

const (
	Title       = "title"
	DateCreated = "date_created"
	DateUpdated = "date_updated"
	Asc         = "asc"
	Desc        = "desc"
	Tags        = "tags"
)

var sortColumnMap = map[string]string{
	Title:       "meta.title",
	DateCreated: "links.created",
	DateUpdated: "links.last_updated",
}

type Query struct {
	// Sort queries
	SortBy string `form:"sort_by"` /* Title || Date */
	Order  string `form:"order"`   /* Asc || Desc */

	// Filter queries
	FilterBy string  `form:"filter_by"` /* Tags || DateRange */
	Tags     []int64 `form:"tags[]"`
}

func GetBookmarks(ctx *gin.Context) {
	db := DB.GetDB()
	var queryParams Query

	if err := ctx.ShouldBind(&queryParams); err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	dbQuery := `SELECT links.id, links.url, links.created, links.last_updated, GROUP_CONCAT(IFNULL(tags.name,"")), meta.title, meta.favicon, meta.description
				FROM links 
				LEFT JOIN meta ON meta.id = links.meta_id 
				LEFT JOIN links_tags ON links_tags.link_id = links.id 
				LEFT JOIN tags ON tags.id = links_tags.tag_id`

	if queryParams.FilterBy == Tags && len(queryParams.Tags) > 0 {
		dbQuery += fmt.Sprintf("\nWHERE tags.id IN (%s)", strings.TrimRight(strings.Repeat("?,", len(queryParams.Tags)), ","))
	}
	dbQuery += "\nGROUP BY links.id"

	sortByColumn := sortColumnMap[queryParams.SortBy]
	if sortByColumn != "" {
		order := Asc
		if queryParams.Order == Desc {
			order = Desc
		}
		order = strings.ToUpper(order)
		dbQuery += fmt.Sprintf("\nORDER BY %s %s", sortByColumn, order)
	}
	preparedQuery, _ := db.Prepare(dbQuery)
	rows, err := preparedQuery.Query(utils.ToGenericArray(queryParams.Tags)...)
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
		if tagStr == "" {
			bm.Tags = []string{}
		} else {
			bm.Tags = strings.Split(tagStr, ",")
		}

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
