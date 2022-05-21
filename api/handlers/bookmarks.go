package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/abhijit-hota/rengoku/server/common"
	DB "github.com/abhijit-hota/rengoku/server/db"
	"github.com/abhijit-hota/rengoku/server/utils"

	"github.com/gin-gonic/gin"
)

type BookmarkReq struct {
	DB.Bookmark
	TagIds    []int64  `json:"tags"`
	FolderIds []string `json:"folders"`
}
type BookmarkRes struct {
	DB.Bookmark
	Tags []DB.Tag `json:"tags"`
}

func AddBookmark(ctx *gin.Context) {
	db := DB.GetDB()
	var body BookmarkReq
	var err error

	if err := ctx.BindJSON(&body); err != nil {
		return
	}

	body.URL, err = utils.NormalizeURL(body.URL)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	meta := make(chan DB.Meta)
	go func() {
		meta <- *utils.MustGet(common.GetMetadata(body.URL))
	}()

	tx, err := db.Begin()
	utils.Must(err)

	now := time.Now().Unix()
	body.Created = now
	body.LastUpdated = now

	stmt := "INSERT INTO links (url, created, last_updated) VALUES (?, ?, ?)"
	linkInsertionInfo, err := tx.Exec(stmt, body.URL, now, now)
	if err != nil && strings.HasPrefix(err.Error(), "UNIQUE constraint failed") {
		utils.Must(tx.Rollback())
		ctx.JSON(http.StatusBadRequest, gin.H{"code": "NAME_ALREADY_PRESENT"})
		return
	}
	linkID, _ := linkInsertionInfo.LastInsertId()
	body.ID = linkID

	urlActions := utils.GetConfig().URLActions

	shouldSaveOffline := utils.GetConfig().ShouldSaveOffline
	for _, urlAction := range urlActions {
		if urlAction.Match(body.URL) {
			body.TagIds = append(body.TagIds, urlAction.Tags...)
			body.FolderIds = append(body.FolderIds, urlAction.Folders...)
			shouldSaveOffline = urlAction.ShouldSaveOffline
		}
	}
	if shouldSaveOffline {
		go common.SavePage(body.URL, int(linkID))
	}

	var bm BookmarkRes
	bm.Bookmark = body.Bookmark
	bm.Tags = []DB.Tag{}

	for _, tagId := range body.TagIds {
		_, err := tx.Exec("INSERT INTO links_tags (tag_id, link_id) VALUES (?, ?)", tagId, linkID)
		utils.Must(err)

		row := tx.QueryRow("SELECT * FROM tags WHERE id = ?", tagId)
		var tag DB.Tag
		row.Scan(&tag.ID, &tag.Name, &tag.Created, &tag.LastUpdated)
		bm.Tags = append(bm.Tags, tag)
	}

	for _, folderId := range body.FolderIds {
		_, err := tx.Exec("INSERT INTO links_folders (folder_id, link_id) VALUES (?, ?)", folderId, linkID)
		utils.Must(err)
	}

	stmt = "INSERT INTO meta (title, description, favicon, link_id) VALUES (?, ?, ?, ?)"
	bm.Meta = <-meta
	bm.NormalizeFavicon()
	_, err = tx.Exec(stmt, bm.Meta.Title, bm.Meta.Description, bm.Meta.Favicon, linkID)
	utils.Must(err)

	tx.Commit()
	ctx.JSON(http.StatusOK, bm)
}

const (
	Title = "title"
	Date  = "date"
	Asc   = "asc"
	Desc  = "desc"
	Tags  = "tags"
)

var sortColumnMap = map[string]string{
	Title: "meta.title",
	Date:  "links.created",
}

type Query struct {
	// Sort queries
	SortBy string `form:"sort_by"` /* Title || Date */
	Order  string `form:"order"`   /* Asc || Desc */

	// Filter queries
	Folder int64   `form:"folder,omitempty"`
	Tags   []int64 `form:"tags[]"`

	// Search
	Search string `form:"search"`
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
				LEFT JOIN meta ON meta.link_id = links.id 
				LEFT JOIN links_tags ON links_tags.link_id = links.id 
				LEFT JOIN tags ON tags.id = links_tags.tag_id`

	prefix := "WHERE"
	if queryParams.Search != "" {
		dbQuery += fmt.Sprintf("\nWHERE meta.title LIKE '%%%[1]v%%' OR links.url LIKE '%%%[1]v%%' OR tags.name LIKE '%%%[1]v%%'", queryParams.Search)
		prefix = "AND"
	}
	if len(queryParams.Tags) > 0 {
		dbQuery += fmt.Sprintf("\n%s tags.id IN (%s)", prefix, utils.GetMultiParam(len(queryParams.Tags)))
		prefix = "AND"
	}
	if queryParams.Folder > 0 {
		dbQuery += fmt.Sprintf("\n%s folders.id = (%v)", prefix, queryParams.Folder)
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
	preparedQuery, err := db.Prepare(dbQuery)
	utils.Must(err)

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

func DeleteBookmarkProperty(ctx *gin.Context) {
	db := DB.GetDB()
	var uri struct {
		IdUri
		Property   string `uri:"property" binding:"required"`
		PropertyId int    `uri:"propertyId" binding:"required"`
	}

	if err := ctx.BindUri(&uri); err != nil {
		return
	}

	if uri.Property != "tag" && uri.Property != "folder" {
		ctx.AbortWithStatus(400)
		return
	}

	stmt := fmt.Sprintf("DELETE FROM links_%[1]vs WHERE link_id = ? AND %[1]v_id = ?", uri.Property)
	info, _ := db.Exec(stmt, uri.ID, uri.PropertyId)
	numDeleted, _ := info.RowsAffected()

	ctx.JSON(http.StatusOK, gin.H{"deleted": numDeleted == 1})
}

func AddBookmarkProperty(ctx *gin.Context) {
	db := DB.GetDB()
	var uri struct {
		IdUri
		Property string `uri:"property"`
	}

	var newProperty struct {
		Id int `json:"id" form:"id" binding:"required"`
	}

	if err := ctx.BindUri(&uri); err != nil {
		return
	}
	if err := ctx.Bind(&newProperty); err != nil {
		return
	}
	if uri.Property != "tag" && uri.Property != "folder" {
		ctx.AbortWithStatus(400)
		return
	}

	stmt := fmt.Sprintf("INSERT OR IGNORE INTO links_%[1]vs (%[1]v_id, link_id) VALUES (?, ?)", uri.Property)
	info, err := db.Exec(stmt, newProperty.Id, uri.ID)
	utils.Must(err)
	updatedLinks, _ := info.RowsAffected()

	ctx.JSON(http.StatusOK, gin.H{"added": updatedLinks == 1})
}

func DeleteBookmark(ctx *gin.Context) {
	db := DB.GetDB()
	var uri IdUri

	if err := ctx.BindUri(&uri); err != nil {
		return
	}

	info, err := db.Exec("DELETE FROM links WHERE id = ?", uri.ID)
	utils.Must(err)
	numDeleted, _ := info.RowsAffected()

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

	placeholder := utils.GetMultiParam(len(body.Ids))

	stmt := "DELETE FROM links WHERE id IN (" + placeholder + ")"
	info, _ := db.Exec(stmt, utils.ToGenericArray(body.Ids)...)
	numDeleted, _ := info.RowsAffected()

	ctx.JSON(http.StatusOK, gin.H{"deleted": numDeleted})
}

func BulkAddBookmarkTags(ctx *gin.Context) {
	db := DB.GetDB()

	var body struct {
		LinkIds []int `json:"link_ids" binding:"required"`
		TagIds  []int `json:"tag_ids" binding:"required"`
	}
	if err := ctx.Bind(&body); err != nil {
		return
	}

	numTotalPairs := len(body.LinkIds) * len(body.TagIds)
	fmt.Println("calc len:", numTotalPairs)

	str := "INSERT OR IGNORE INTO links_tags(tag_id, link_id) VALUES " + utils.RepeatWithSeparator("(?, ?),", numTotalPairs, ",")
	stmt, err := db.Prepare(str)
	utils.Must(err)

	allPairs := []int{}

	for _, linkId := range body.LinkIds {
		for _, tagId := range body.TagIds {
			allPairs = append(allPairs, tagId, linkId)
		}
	}

	info, err := stmt.Exec(utils.ToGenericArray(allPairs)...)
	utils.Must(err)
	updatedLinks, _ := info.RowsAffected()

	ctx.JSON(http.StatusOK, gin.H{"added": updatedLinks})
}
