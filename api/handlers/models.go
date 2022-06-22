package handlers

import (
	"net/url"
	"strings"

	DB "github.com/abhijit-hota/rengoku/server/db"
	"github.com/abhijit-hota/rengoku/server/utils"
)

type BookmarkReq struct {
	DB.Bookmark
	TagIds    []int64 `json:"tags"`
	FolderIds []int64 `json:"folders"`
}
type BookmarkRes struct {
	DB.Bookmark
	DB.Meta `json:"meta"`
	Tags    []DB.Tag `json:"tags"`
}

func (bm *BookmarkRes) FixFavicon() {
	if bm.Meta.Favicon == "" {
		bm.Meta.Favicon = "favicon.ico"
	}
	favicon, err := url.Parse(bm.Meta.Favicon)
	utils.Must(err)
	if !favicon.IsAbs() {
		rootURL := utils.MustGet(url.Parse(bm.URL))
		bm.Meta.Favicon = rootURL.Scheme + "://" + strings.TrimRight(rootURL.Hostname(), "/") + "/" + strings.TrimLeft(favicon.String(), "/")
	}
}
