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
	"modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"

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
	row := tx.QueryRowx(stmt, body.URL)
	err = row.StructScan(&body.Bookmark)

	if err != nil {
		e, ok := err.(*sqlite.Error)
		if ok && e.Code() == sqlite3.SQLITE_CONSTRAINT_UNIQUE {
			utils.Must(tx.Rollback())
			ctx.JSON(http.StatusBadRequest, gin.H{"code": "NAME_ALREADY_PRESENT"})
			return
		}
	}

	meta := make(chan DB.Meta)
	go func() {
		meta <- *utils.MustGet(common.GetMetadata(body.URL))
	}()

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
		go common.SavePage(body.URL, strconv.Itoa(int(body.ID)))
	}

	var bm BookmarkRes
	bm.Bookmark = body.Bookmark
	bm.Tags = []DB.Tag{}

	for _, tagId := range body.TagIds {
		_, err := tx.Exec("INSERT INTO links_tags (tag_id, link_id) VALUES (?, ?)", tagId, body.ID)
		utils.Must(err)

		row := tx.QueryRow("SELECT * FROM tags WHERE id = ?", tagId)
		var tag DB.Tag
		row.Scan(&tag.ID, &tag.Name, &tag.CreatedAt, &tag.LastUpdated)
		bm.Tags = append(bm.Tags, tag)
	}

	for _, folderId := range body.FolderIds {
		_, err := tx.Exec("INSERT INTO links_folders (folder_id, link_id) VALUES (?, ?)", folderId, body.ID)
		utils.Must(err)
	}

	stmt = "INSERT INTO meta (title, description, favicon, link_id) VALUES (?, ?, ?, ?)"
	bm.Meta = <-meta
	bm.NormalizeFavicon()
	_, err = tx.Exec(stmt, bm.Meta.Title, bm.Meta.Description, bm.Meta.Favicon, body.ID)
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
	Date:  "links.created_at",
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

	dbQuery := `SELECT links.id, links.url, links.created_at, links.last_updated, links.last_saved_offline,
					   IFNULL(GROUP_CONCAT(tags.id || "=" || tags.name), ""),
					   meta.title, meta.favicon, meta.description
				FROM links 
				LEFT JOIN meta ON meta.link_id = links.id 
				LEFT JOIN links_folders ON links_folders.link_id = links.id 
				LEFT JOIN folders ON folders.id = links_folders.folder_id
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
	// Will optimize when an issue arises
	dbQuery += "\nLIMIT 20 OFFSET " + strconv.Itoa(20*queryParams.Page)
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
			&bm.CreatedAt,
			&bm.LastUpdated,
			&bm.LastSavedOffline,
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
				if keyval == "" {
					bm.Tags = make([]DB.Tag, 0)
					break
				}
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

	ctx.JSON(http.StatusOK, gin.H{"data": bookmarks, "page": queryParams.Page})
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

func SaveBookmark(ctx *gin.Context) {
	var uri IdUri
	if err := ctx.BindUri(&uri); err != nil {
		return
	}

	db := DB.GetDB()

	var bm DB.Bookmark
	utils.Must(db.Get(&bm, "SELECT * FROM links WHERE id = ?", uri.ID))

	common.SavePage(bm.URL, fmt.Sprint(uri.ID))

	now := time.Now().Unix()
	db.MustExec("UPDATE links SET last_saved_offline = ? WHERE id = ?", now, uri.ID)

	ctx.JSON(http.StatusOK, gin.H{"saved": true})
}

func RefetchMetadata(ctx *gin.Context) {
	var uri IdUri
	if err := ctx.BindUri(&uri); err != nil {
		return
	}

	db := DB.GetDB()

	var bm BookmarkRes
	utils.Must(
		db.Get(
			&bm,
			`SELECT links.*, 
			meta.title "meta.title", 
			meta.favicon "meta.favicon", 
			meta.description "meta.description"
			FROM links LEFT JOIN meta ON meta.link_id = links.id
			WHERE links.id = ?`,
			uri.ID,
		),
	)

	bm.Meta = *utils.MustGet(common.GetMetadata(bm.URL))
	bm.NormalizeFavicon()

	// Todo: needs a trigger
	now := time.Now().Unix()

	tx := db.MustBegin()
	tx.MustExec(
		"UPDATE meta SET title = ?, description = ?, favicon = ? WHERE link_id = ?",
		bm.Meta.Title, bm.Meta.Description, bm.Meta.Favicon, bm.ID,
	)
	tx.MustExec("UPDATE links SET last_updated = ? WHERE id = ?", now, bm.ID)
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
				_bm.NormalizeFavicon()
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

		path := ""
		var folderID int
		for _, folder := range strings.Split(bm.Folder, common.FolderPathSeparator) {

			stmt = "INSERT OR IGNORE INTO folders (name, path) VALUES (?, ?)"
			tx.MustExec(stmt, folder, path)

			tx.Get(&folderID, "SELECT id FROM folders WHERE name = ? AND PATH = ?", folder, path)

			path += strconv.Itoa(folderID) + "/"
		}
		stmt = "INSERT OR IGNORE INTO links_folders (link_id, folder_id) VALUES (?, ?)"
		tx.MustExec(stmt, linkID, folderID)
	}
	utils.Must(tx.Commit())
	ctx.JSON(http.StatusOK, bookmarks)
}
