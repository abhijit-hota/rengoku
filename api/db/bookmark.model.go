package db

type Meta struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Favicon     string `json:"favicon"`
}

type Bookmark struct {
	Meta        Meta     `json:"meta"`
	URL         string   `json:"url" binding:"required"`
	Created     int64    `json:"created"`
	LastUpdated int64    `json:"last_updated"`
	Tags        []string `json:"tags"`
}

type Tag struct {
	ID          int64  `json:"id"`
	Name        string `json:"name" binding:"required"`
	Created     int64  `json:"created"`
	LastUpdated int64  `json:"last_updated"`
}
