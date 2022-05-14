package handlers

import (
	"api/common"
	DB "api/db"
	"api/utils"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/purell"
	"github.com/gin-gonic/gin"
)

type BookmarkReq struct {
	DB.Bookmark
	Tags []int `json:"tags"`
}
type BookmarkRes struct {
	DB.Bookmark
	Tags []DB.Tag `json:"tags"`
}

func AddBookmark(ctx *gin.Context) {
	var body BookmarkReq

	db := DB.GetDB()

	if err := ctx.BindJSON(&body); err != nil {
		return
	}

	link, err := purell.NormalizeURLString(body.URL, purell.FlagsSafe)
	body.URL = link
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	tx, err := db.Begin()
	utils.Must(err)

	stmt := "SELECT COUNT(*) FROM links WHERE url = ?"
	existingCheck := tx.QueryRow(stmt, body.URL)
	var urlExists int
	existingCheck.Scan(&urlExists)

	if urlExists != 0 {
		body.Tags = []int{}
		tx.Commit()

		ctx.JSON(http.StatusBadRequest, gin.H{"message": "URL_ALREADY_PRESENT"})
		return
	}

	if err := common.GetMetadata(body.URL, &body.Meta); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// TODO: do this parallely? but what about meta id?
	stmt = "INSERT INTO meta (title, description, favicon) VALUES (?, ?, ?)"
	metaExecinfo, err := tx.Exec(stmt, body.Meta.Title, body.Meta.Description, body.Meta.Favicon)
	metaID, _ := metaExecinfo.LastInsertId()

	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid URL."})
		return
	}

	now := time.Now().Unix()
	body.Created = now
	body.LastUpdated = now

	stmt = "INSERT OR IGNORE INTO links (url, meta_id, created, last_updated) VALUES (?, ?, ?, ?)"
	linkInsertionInfo, err := tx.Exec(stmt, body.URL, metaID, now, now)
	utils.Must(err)

	urlActions := utils.GetConfig().URLActions

	shouldSaveOffline := utils.GetConfig().ShouldSaveOffline
	for _, urlAction := range urlActions {
		if urlAction.Match(body.URL) {
			body.Tags = append(body.Tags, urlAction.Tags...)
			shouldSaveOffline = urlAction.ShouldSaveOffline
		}
	}
	if shouldSaveOffline {
		go common.SavePage(body.URL)
	}
	fmt.Println(444, body.Tags)
	if len(body.Tags) == 0 {
		body.Tags = []int{}
		tx.Commit()

		ctx.JSON(http.StatusOK, body)
		return
	}

	query := fmt.Sprintf(
		"SELECT id, name FROM tags WHERE id IN (%s)",
		strings.TrimRight(strings.Repeat("?,", len(body.Tags)), ","),
	)
	statement, _ := tx.Prepare(query)
	tagIDs, err := statement.Query(utils.ToGenericArray(body.Tags)...)
	defer tagIDs.Close()
	utils.Must(err)

	statement, err = tx.Prepare("INSERT INTO links_tags (tag_id, link_id) VALUES (?, ?)")
	utils.Must(err)
	defer statement.Close()

	linkID, _ := linkInsertionInfo.LastInsertId()
	body.ID = linkID

	var res BookmarkRes
	res.Bookmark = body.Bookmark

	for tagIDs.Next() {
		var tag DB.Tag
		tagIDs.Scan(&tag.ID, &tag.Name)
		_, err := statement.Exec(tag.ID, linkID)
		utils.Must(err)

		fmt.Printf("%+v\n", tag)
		res.Tags = append(res.Tags, tag)
	}
	err = tagIDs.Err()
	utils.Must(err)

	tx.Commit()
	ctx.JSON(http.StatusOK, res)
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

	dbQuery := `SELECT links.id, links.url, links.created, links.last_updated, 
					   GROUP_CONCAT(IFNULL(tags.id,"") || "=" || IFNULL(tags.name,"")),
					   meta.title, meta.favicon, meta.description
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

	bookmarks := make([]BookmarkRes, 0)
	for rows.Next() {
		var bm BookmarkRes
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
		if tagStr == "=" {
			bm.Tags = []DB.Tag{}
		} else {
			keyVals := strings.Split(tagStr, ",")
			for _, keyval := range keyVals {
				str := strings.Split(keyval, "=")
				tagId, _ := strconv.Atoi(str[0])

				var tag DB.Tag
				tag.ID = int64(tagId)
				tag.Name = str[1]
				bm.Tags = append(bm.Tags, tag)
			}
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

func BulkDeleteBookmarks(ctx *gin.Context) {
	db := DB.GetDB()

	var body struct {
		Ids []int `json:"ids" binding:"required"`
	}
	if err := ctx.BindJSON(&body); err != nil {
		return
	}

	placeholder := strings.TrimRight(strings.Repeat("?,", len(body.Ids)), ",")

	tx, _ := db.Begin()

	stmt := "DELETE FROM links WHERE id IN (" + placeholder + ")"
	info, _ := tx.Exec(stmt, utils.ToGenericArray(body.Ids)...)
	numDeleted, _ := info.RowsAffected()

	stmt = "DELETE FROM links_tags WHERE link_id IN (" + placeholder + ")"
	tx.Exec(stmt, utils.ToGenericArray(body.Ids)...)

	utils.Must(tx.Commit())
	ctx.JSON(http.StatusOK, gin.H{"deleted": numDeleted})
}
