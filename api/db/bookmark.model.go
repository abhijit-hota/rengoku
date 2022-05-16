package db

import (
	"github.com/abhijit-hota/rengoku/server/utils"
	"net/url"
	"strings"
)

type Meta struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Favicon     string `json:"favicon"`
}

type Bookmark struct {
	ID          int64  `json:"id"`
	Meta        Meta   `json:"meta"`
	URL         string `json:"url" binding:"required"`
	Created     int64  `json:"created,omitempty"`
	LastUpdated int64  `json:"last_updated,omitempty"`
}

func (bm *Bookmark) NormalizeFavicon() {
	if bm.Meta.Favicon == "" {
		bm.Meta.Favicon = "favicon.ico"
	}
	favicon, err := url.Parse(bm.Meta.Favicon)
	utils.Must(err)
	if !favicon.IsAbs() {
		rootURL, _ := url.Parse(bm.URL)
		bm.Meta.Favicon = rootURL.Scheme + "://" + strings.TrimRight(rootURL.Hostname(), "/") + "/" + strings.TrimLeft(favicon.String(), "/")
	}
}

type Tag struct {
	ID          int64  `json:"id"`
	Name        string `json:"name" form:"name" binding:"required"`
	Created     int64  `json:"created,omitempty" form:"created"`
	LastUpdated int64  `json:"last_updated,omitempty" form:"last_updated"`
}
