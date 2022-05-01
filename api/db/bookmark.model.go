package db

type Meta struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Favicon     []byte `json:"favicon"`
}

type Bookmark struct {
	Meta        Meta   `json:"meta"`
	URL         string `json:"url" binding:"required"`
	Created     int64  `json:"created"`
	LastUpdated int64  `json:"last_updated"`
}
