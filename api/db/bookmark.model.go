package db

type Tag struct {
	Path   string `json:"path" binding:"required"`
	IsRoot bool   `json:"isRoot"`
}

type Meta struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Favicon     string `json:"favicon"`
}

type Bookmark struct {
	Meta        Meta   `json:"meta"`
	URL         string `json:"url" binding:"required"`
	Created     int64  `json:"created"`
	LastUpdated int64  `json:"last_updated"`
	Tags        []Tag  `json:"tags"`
}
