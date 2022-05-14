package db

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

type Tag struct {
	ID          int64  `json:"id"`
	Name        string `json:"name" form:"name" binding:"required"`
	Created     int64  `json:"created,omitempty" form:"created"`
	LastUpdated int64  `json:"last_updated,omitempty" form:"last_updated"`
}
