package db

import (
	"net/url"
	"strings"

	"github.com/abhijit-hota/rengoku/server/utils"
)

type Meta struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Favicon     string `json:"favicon"`
}

type Bookmark struct {
	ID               int64  `json:"id"`
	Meta             Meta   `json:"meta"`
	URL              string `json:"url" binding:"required"`
	Created          int64  `json:"created,omitempty"`
	LastUpdated      int64  `json:"last_updated,omitempty" db:"last_updated"`
	LastSavedOffline int64  `json:"last_saved_offline,omitempty" db:"last_saved_offline"`
}

func (bm *Bookmark) NormalizeFavicon() {
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

type Tag struct {
	ID          int64  `json:"id"`
	Name        string `json:"name" form:"name" binding:"required"`
	Created     int64  `json:"created,omitempty" form:"created"`
	LastUpdated int64  `json:"last_updated,omitempty" form:"last_updated"`
}

type Folder struct {
	ID          int64  `json:"id"`
	Name        string `json:"name" form:"name" binding:"required"`
	Path        string `json:"path" form:"path"`
	Created     int64  `json:"created,omitempty" form:"created"`
	LastUpdated int64  `json:"last_updated,omitempty" form:"last_updated"`
}
