package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
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
	defer tx.Rollback()

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

	metaChan := make(chan DB.Meta)
	go func() {
		meta, err := common.GetMetadata(body.URL)
		if err != nil {
			if err.Error() != "not html" {
				ctx.JSON(http.StatusInternalServerError, gin.H{"message": "Something went wrong."})
			}
			close(metaChan)
			return
		}
		metaChan <- *meta
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
	lastSaved := make(chan int64)

	if shouldSaveOffline {
		now := time.Now().Unix()
		filename := fmt.Sprint(body.ID) + "_" + fmt.Sprint(now)
		go func() {
			err := common.SavePage(body.URL, filename)
			// TODO: error handling
			if err != nil {
				lastSaved <- 0
				return
			}
			lastSaved <- now
		}()
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
	bm.Meta = <-metaChan
	bm.FixFavicon()
	bm.Meta.LinkID = body.ID
	_, err = tx.NamedExec(stmt, bm.Meta)
	utils.Must(err)

	if shouldSaveOffline {
		if savedAt := <-lastSaved; savedAt != 0 {
			bm.LastSavedOffline = savedAt
			stmt = "UPDATE links SET last_saved_offline = :last_saved_offline WHERE id = :id"

			_, err = tx.NamedExec(stmt, bm)
			utils.Must(err)
		}
	}

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

func UpdateBookmark(ctx *gin.Context) {
	var uri IdUri

	if err := ctx.BindUri(&uri); err != nil {
		return
	}
	linkId := uri.ID

	db := DB.GetDB()
	tx := db.MustBegin()
	defer tx.Rollback()

	type UpdateReq struct {
		Title       string  `json:"title,omitempty"`
		Description string  `json:"description,omitempty"`
		TagIds      []int64 `json:"tag_ids,omitempty"`
		FolderIds   []int64 `json:"folder_ids,omitempty"`
	}
	updateReq := UpdateReq{}
	utils.Must(tx.Get(&updateReq, "SELECT title, description FROM meta WHERE link_id = ?", linkId))
	utils.Must(tx.Select(&updateReq.TagIds, "SELECT tag_id FROM links_tags WHERE link_id = ?", linkId))
	utils.Must(tx.Select(&updateReq.FolderIds, "SELECT folder_id FROM links_folders WHERE link_id = ?", linkId))

	if err := ctx.BindJSON(&updateReq); err != nil {
		return
	}

	tx.MustExec("UPDATE meta SET title = ?, description = ? WHERE link_id = ?", updateReq.Title, updateReq.Description, linkId)

	type IDPair struct {
		FolderId int64 `db:"folder_id"`
		TagId    int64 `db:"tag_id"`
		LinkId   int64 `db:"link_id"`
	}

	stmt := "DELETE FROM links_tags WHERE link_id = ?"
	tx.MustExec(stmt, linkId)
	if len(updateReq.TagIds) > 0 {
		tagPairs := make([]IDPair, 0)
		for _, tagId := range updateReq.TagIds {
			tagPairs = append(tagPairs, IDPair{
				TagId:  tagId,
				LinkId: linkId,
			})
		}
		stmt = "INSERT OR IGNORE INTO links_tags(tag_id, link_id) VALUES (:tag_id, :link_id)"
		_, err := tx.NamedExec(stmt, tagPairs)
		utils.Must(err)
	}

	stmt = "DELETE FROM links_folders WHERE link_id = ?"
	tx.MustExec(stmt, linkId)
	if len(updateReq.FolderIds) > 0 {
		folderPairs := make([]IDPair, 0)
		for _, folderId := range updateReq.FolderIds {
			folderPairs = append(folderPairs, IDPair{
				FolderId: folderId,
				LinkId:   linkId,
			})
		}
		stmt = "INSERT OR IGNORE INTO links_folders(folder_id, link_id) VALUES (:folder_id, :link_id)"
		_, err := tx.NamedExec(stmt, folderPairs)
		utils.Must(err)
	}

	var updatedBookmark struct {
		UpdateReq
		LastUpdated int64 `json:"last_updated" db:"last_updated"`
	}
	updatedBookmark.UpdateReq = updateReq
	utils.Must(tx.Get(&updatedBookmark.LastUpdated, "SELECT last_updated FROM links WHERE id = ?", linkId))

	tx.Commit()
	ctx.JSON(http.StatusOK, updatedBookmark)
}

func DeleteBookmark(ctx *gin.Context) {
	db := DB.GetDB()
	var uri IdUri

	if err := ctx.BindUri(&uri); err != nil {
		return
	}

	var lastSavedOffline int64
	err := db.Get(&lastSavedOffline, "DELETE FROM links WHERE id = ? RETURNING last_saved_offline", uri.ID)

	if err != nil {
		//TODO: Log error
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "Something went wrong"})
	}

	config := utils.GetConfig()

	if config.ShouldDeleteOffline {
		err := os.Remove(os.Getenv("SAVE_OFFLINE_PATH") + fmt.Sprint(uri.ID) + "_" + fmt.Sprint(lastSavedOffline))
		utils.Must(err)
		// TODO: Some kind of error handling
	}

	ctx.JSON(http.StatusOK, gin.H{"deleted": true})
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
		LinkIds []int64 `json:"link_ids" binding:"required"`
		TagIds  []int64 `json:"tag_ids" binding:"required"`
	}
	if err := ctx.BindJSON(&body); err != nil {
		return
	}

	str := "INSERT OR IGNORE INTO links_tags(tag_id, link_id) VALUES (:tag_id, :link_id)"

	type IDPair struct {
		TagId  int64 `db:"tag_id"`
		LinkId int64 `db:"link_id"`
	}

	allPairs := make([]IDPair, 0)

	for _, linkId := range body.LinkIds {
		for _, tagId := range body.TagIds {
			allPairs = append(allPairs, IDPair{
				TagId:  tagId,
				LinkId: linkId,
			})
		}
	}
	info, err := db.NamedExec(str, allPairs)
	utils.Must(err)

	ctx.JSON(http.StatusOK, gin.H{"updated": utils.MustGet(info.RowsAffected())})
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

	savedTs := time.Now().Unix()

	if err := common.SavePage(bm.URL, fmt.Sprint(uri.ID)+"_"+fmt.Sprint(savedTs)); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"saved": false})
		return
	}

	tx.MustExec("UPDATE links SET last_saved_offline = ? WHERE id = ?", savedTs, uri.ID)

	utils.Must(tx.Commit())
	ctx.JSON(http.StatusOK, gin.H{"saved": true, "lastSavedOffline": savedTs})
}

func RefetchMetadata(ctx *gin.Context) {
	var uri IdUri
	if err := ctx.BindUri(&uri); err != nil {
		return
	}

	db := DB.GetDB()
	tx := db.MustBegin()
	defer tx.Rollback()

	var bm BookmarkRes
	utils.Must(
		tx.Get(&bm.URL, `SELECT url from links WHERE links.id = ?`, uri.ID),
	)

	meta, err := common.GetMetadata(bm.URL)
	if err != nil {
		if err.Error() != "not html" {
			ctx.JSON(http.StatusInternalServerError, gin.H{"message": "Something went wrong."})
			return
		}
	} else {
		bm.Meta = *meta
	}
	bm.Meta.LinkID = uri.ID
	bm.FixFavicon()

	_, err = tx.NamedExec(
		"UPDATE meta SET title = :title, description = :description, favicon = :favicon WHERE link_id = :link_id",
		bm.Meta,
	)
	utils.Must(err)

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
		Folder string
	}

	bookmarks := make([]ExtBookmark, len(parsedLinks))
	var wg sync.WaitGroup

	for i, v := range parsedLinks {
		normalisedTags := make([]DB.Tag, 0)
		for _, tag := range v.Tags {
			normalisedTags = append(normalisedTags, DB.Tag{
				ID:   -1,
				Name: tag,
			})
		}

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
				Tags: normalisedTags,
			},
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
	defer tx.Rollback()

	for bmIdx := range bookmarks {
		bm := &bookmarks[bmIdx]

		stmt := "INSERT OR IGNORE INTO links (url) VALUES (?)"
		tx.MustExec(stmt, bm.URL)

		stmt = "SELECT * FROM links WHERE url = ?"

		tx.Get(&bm.Bookmark, stmt, bm.URL)
		linkID := bm.Bookmark.ID

		stmt = "INSERT OR IGNORE INTO meta (link_id, title, description, favicon) VALUES (?, ?, ?, ?)"
		tx.MustExec(stmt, linkID, bm.Meta.Title, bm.Meta.Description, bm.Meta.Favicon)
		for tagIdx := range bm.Tags {
			tag := &bm.Tags[tagIdx]

			stmt := "INSERT OR IGNORE INTO tags (name) VALUES (?)"
			tx.MustExec(stmt, tag.Name)

			stmt = "SELECT id FROM tags WHERE name = ?"
			tx.Get(&tag.ID, stmt, tag.Name)

			stmt = "INSERT OR IGNORE INTO links_tags (link_id, tag_id) VALUES (?, ?)"
			tx.MustExec(stmt, linkID, tag.ID)
		}

		if bm.Folder == "" {
			continue
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
