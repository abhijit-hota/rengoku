package handlers

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/abhijit-hota/rengoku/server/common"
	DB "github.com/abhijit-hota/rengoku/server/db"
	"github.com/abhijit-hota/rengoku/server/utils"
	"github.com/jmoiron/sqlx"

	"github.com/gin-gonic/gin"
)

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

	tx := db.MustBegin()

	stmt := "INSERT INTO links (url) VALUES (?) RETURNING *"
	err = tx.Get(&body.Bookmark, stmt, body.URL)

	if err != nil {
		if DB.IsUniqueErr(err) {
			utils.Must(tx.Rollback())
			ctx.JSON(http.StatusBadRequest, gin.H{"cause": "NAME_ALREADY_PRESENT"})
			return
		}
		panic(err)
	}

	meta := make(chan DB.Meta)
	go func() {
		meta <- *utils.MustGet(common.GetMetadata(body.URL))
	}()

	config := utils.GetConfig()

	urlActions := config.URLActions
	shouldSaveOffline := config.ShouldSaveOffline
	for _, urlAction := range urlActions {
		if urlAction.Match(body.URL) {
			body.TagIds = append(body.TagIds, urlAction.Tags...)
			body.FolderIds = append(body.FolderIds, urlAction.Folders...)
			shouldSaveOffline = urlAction.ShouldSaveOffline
		}
	}
	if shouldSaveOffline {
		filename := strconv.Itoa(int(body.ID))
		go common.SavePage(body.URL, filename)
	}

	var bm BookmarkRes
	bm.Bookmark = body.Bookmark
	bm.Tags = []DB.Tag{}

	for _, tagId := range body.TagIds {
		tx.MustExec("INSERT INTO links_tags (tag_id, link_id) VALUES (?, ?)", tagId, body.ID)

		var tag DB.Tag
		err = tx.Get(&tag, "SELECT * FROM tags WHERE id = ?", tagId)
		utils.Must(err)

		bm.Tags = append(bm.Tags, tag)
	}

	for _, folderId := range body.FolderIds {
		tx.MustExec("INSERT INTO links_folders (folder_id, link_id) VALUES (?, ?)", folderId, body.ID)
	}

	stmt = "INSERT INTO meta (title, description, favicon, link_id) VALUES (:title, :description, :favicon, :link_id)"
	bm.Meta = <-meta
	bm.FixFavicon()
	bm.Meta.LinkID = body.ID
	_, err = tx.NamedExec(stmt, bm.Meta)
	utils.Must(err)

	tx.Commit()
	ctx.JSON(http.StatusOK, bm)
}

const (
	title = "title"
	date  = "date"
	asc   = "asc"
	desc  = "desc"
)

var sortColumnMap = map[string]string{
	title: "meta.title",
	date:  "links.created_at",
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

	// Pagination
	Page int `form:"page"`
}

func GetBookmarks(ctx *gin.Context) {
	db := DB.GetDB()
	var queryParams Query

	if err := ctx.ShouldBind(&queryParams); err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	dbQuery := `
SELECT links.id, links.url, links.created_at, links.last_updated, links.last_saved_offline,
	   IFNULL(GROUP_CONCAT(tags.id || "=" || tags.name), ""),
	   meta.title, meta.favicon, meta.description, COUNT(1) OVER() AS full_count
FROM links 
LEFT JOIN meta ON meta.link_id = links.id 
LEFT JOIN links_folders ON links_folders.link_id = links.id 
LEFT JOIN folders ON folders.id = links_folders.folder_id
LEFT JOIN links_tags ON links_tags.link_id = links.id 
LEFT JOIN tags ON tags.id = links_tags.tag_id
`

	prefix := "WHERE"
	if queryParams.Search != "" {
		dbQuery += "\nWHERE meta.title LIKE :query OR links.url LIKE :query OR tags.name LIKE :query"
		prefix = "AND"
	}
	if len(queryParams.Tags) > 0 {
		dbQuery += "\n" + prefix + " tags.id IN (:tagIDs)"
		prefix = "AND"
	}
	if queryParams.Folder > 0 {
		dbQuery += "\n" + prefix + " folders.id = (:folderID)"
	}
	dbQuery += "\nGROUP BY links.id"

	sortByColumn := sortColumnMap[queryParams.SortBy]
	if sortByColumn != "" {
		order := asc
		if queryParams.Order == desc {
			order = desc
		}
		order = strings.ToUpper(order)
		dbQuery += "\nORDER BY" + " " + sortByColumn + " " + order
	}
	// Will optimize when an issue arises
	dbQuery += "\nLIMIT 20 OFFSET " + strconv.Itoa(20*queryParams.Page)

	arg := map[string]interface{}{
		"query":    "%" + queryParams.Search + "%",
		"tagIDs":   queryParams.Tags,
		"folderID": queryParams.Folder,
	}
	query, args, err := sqlx.Named(dbQuery, arg)
	query, args, err = sqlx.In(query, args...)

	utils.Must(err)

	rows, err := db.Queryx(query, args...)
	utils.Must(err)
	defer rows.Close()

	bookmarks := make([]BookmarkRes, 0)
	var fullCount int
	for rows.Next() {
		var bm BookmarkRes
		var tagStr string
		err = rows.Scan(
			&bm.Bookmark.ID,
			&bm.URL,
			&bm.CreatedAt,
			&bm.LastUpdated,
			&bm.LastSavedOffline,
			&tagStr,
			&bm.Meta.Title,
			&bm.Meta.Favicon,
			&bm.Meta.Description,
			&fullCount,
		)
		utils.Must(err)
		if tagStr == "" {
			bm.Tags = []DB.Tag{}
		} else {
			keyVals := strings.Split(tagStr, ",")
			for _, keyval := range keyVals {
				str := strings.Split(keyval, "=")
				tag := DB.Tag{
					ID:   int64(utils.MustGet(strconv.Atoi(str[0]))),
					Name: str[1],
				}
				bm.Tags = append(bm.Tags, tag)
			}
		}
		bookmarks = append(bookmarks, bm)
	}
	utils.Must(rows.Err())

	ctx.JSON(http.StatusOK, gin.H{"data": bookmarks, "page": queryParams.Page, "total": fullCount})
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

	stmt := ""
	switch uri.Property {
	case "tag":
		stmt = "DELETE FROM links_tags WHERE link_id = ? AND tag_id = ?"
	case "folder":
		stmt = "DELETE FROM links_folders WHERE link_id = ? AND folder_id = ?"
	default:
		ctx.AbortWithStatus(400)
		return
	}

	info := db.MustExec(stmt, uri.ID, uri.PropertyId)
	numDeleted := utils.MustGet(info.RowsAffected())

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

	stmt := ""
	switch uri.Property {
	case "tag":
		stmt = "INSERT OR IGNORE INTO links_tags (tag_id, link_id) VALUES (?, ?)"
	case "folder":
		stmt = "INSERT OR IGNORE INTO links_folders (folder_id, link_id) VALUES (?, ?)"
	default:
		ctx.AbortWithStatus(400)
		return
	}

	info := db.MustExec(stmt, newProperty.Id, uri.ID)
	updatedLinks := utils.MustGet(info.RowsAffected())

	ctx.JSON(http.StatusOK, gin.H{"added": updatedLinks == 1})
}

func DeleteBookmark(ctx *gin.Context) {
	db := DB.GetDB()
	var uri IdUri

	if err := ctx.BindUri(&uri); err != nil {
		return
	}

	info := db.MustExec("DELETE FROM links WHERE id = ?", uri.ID)
	numDeleted := utils.MustGet(info.RowsAffected())

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

	query, args, err := sqlx.In("DELETE FROM links WHERE id IN (?)", body.Ids)
	utils.Must(err)

	info := db.MustExec(query, args...)
	numDeleted := utils.MustGet(info.RowsAffected())

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

	str := "INSERT OR IGNORE INTO links_tags(tag_id, link_id) VALUES (:tag_id, :link_id)"

	allPairs := []map[string]int{}

	for _, linkId := range body.LinkIds {
		for _, tagId := range body.TagIds {
			allPairs = append(allPairs, map[string]int{
				"tag_id":  tagId,
				"link_id": linkId,
			})
		}
	}

	info, err := db.NamedExec(str, allPairs)
	utils.Must(err)
	updatedLinks := utils.MustGet(info.RowsAffected())

	ctx.JSON(http.StatusOK, gin.H{"added": updatedLinks})
}

func SaveBookmark(ctx *gin.Context) {
	var uri IdUri
	if err := ctx.BindUri(&uri); err != nil {
		return
	}

	db := DB.GetDB()
	tx := db.MustBegin()

	var bm DB.Bookmark
	utils.Must(tx.Get(&bm, "SELECT * FROM links WHERE id = ?", uri.ID))

	common.SavePage(bm.URL, fmt.Sprint(uri.ID))

	tx.MustExec("UPDATE links SET last_saved_offline = ? WHERE id = ?", time.Now().Unix(), uri.ID)

	utils.Must(tx.Commit())
	ctx.JSON(http.StatusOK, gin.H{"saved": true})
}

func RefetchMetadata(ctx *gin.Context) {
	var uri IdUri
	if err := ctx.BindUri(&uri); err != nil {
		return
	}

	db := DB.GetDB()
	tx := db.MustBegin()

	var bm BookmarkRes
	utils.Must(
		tx.Get(&bm.URL, `SELECT url from links WHERE links.id = ?`, uri.ID),
	)

	bm.Meta = *utils.MustGet(common.GetMetadata(bm.URL))
	bm.Meta.LinkID = uri.ID
	bm.FixFavicon()

	tx.MustExec(
		"UPDATE meta SET title = :title, description = :description, favicon = :favicon WHERE link_id = :link_id",
		bm.Meta,
	)
	utils.Must(tx.Commit())

	ctx.JSON(http.StatusOK, bm.Meta)
}

func ImportBookmarks(ctx *gin.Context) {
	netscapeExport := utils.MustGet(ctx.FormFile("export"))
	file, _ := netscapeExport.Open()
	asBytes, _ := io.ReadAll(file)
	parsedLinks, _ := common.ParseNetscapeData(string(asBytes))

	_, shouldFetchMetadata := ctx.GetPostForm("fetchMeta")

	type ExtBookmark struct {
		BookmarkRes
		Tags   []string
		Folder string
	}
	bookmarks := make([]ExtBookmark, len(parsedLinks))
	var wg sync.WaitGroup

	for i, v := range parsedLinks {
		bm := ExtBookmark{
			BookmarkRes: BookmarkRes{
				Bookmark: DB.Bookmark{
					URL: v.Href,
				},
				Meta: DB.Meta{
					Title:       v.Title,
					Description: "",
					Favicon:     v.IconUri,
				},
			},
			Tags:   v.Tags,
			Folder: v.FolderPath,
		}
		if shouldFetchMetadata {
			wg.Add(1)
			go func(index int, _bm ExtBookmark) {
				defer wg.Done()
				meta, err := common.GetMetadata(_bm.URL)
				if err == nil {
					_bm.Meta = *meta
				}
				_bm.FixFavicon()
				bookmarks[index] = _bm
			}(i, bm)
		} else {
			bookmarks[i] = bm
		}
	}
	wg.Wait()

	db := DB.GetDB()
	tx := db.MustBegin()
	for _, bm := range bookmarks {
		stmt := "INSERT OR IGNORE INTO links (url) VALUES (?)"
		tx.MustExec(stmt, bm.URL)

		stmt = "SELECT id FROM links WHERE url = ?"
		var linkID int
		tx.Get(&linkID, stmt, bm.URL)

		stmt = "INSERT OR IGNORE INTO meta (link_id, title, description, favicon) VALUES (?, ?, ?, ?)"
		tx.MustExec(stmt, linkID, bm.Meta.Title, bm.Meta.Description, bm.Meta.Favicon)
		for _, tag := range bm.Tags {
			stmt := "INSERT OR IGNORE INTO tags (name) VALUES (?)"
			tx.MustExec(stmt, tag)

			stmt = "SELECT id FROM tags WHERE name = ?"
			var tagId int
			tx.Get(&tagId, stmt, tag)

			stmt = "INSERT OR IGNORE INTO links_tags (link_id, tag_id) VALUES (?, ?)"
			tx.MustExec(stmt, linkID, tagId)
		}

		var folderID, parentID int
		for _, folder := range strings.Split(bm.Folder, common.FolderPathSeparator) {

			stmt = "INSERT OR IGNORE INTO folders (name, parent_id) VALUES (?, ?)"
			tx.MustExec(stmt, folder, parentID)

			tx.Get(&folderID, "SELECT id FROM folders WHERE name = ? AND parent_id = ?", folder, parentID)

			parentID = folderID
		}
		stmt = "INSERT OR IGNORE INTO links_folders (link_id, folder_id) VALUES (?, ?)"
		tx.MustExec(stmt, linkID, folderID)
	}
	utils.Must(tx.Commit())
	ctx.JSON(http.StatusOK, bookmarks)
}
