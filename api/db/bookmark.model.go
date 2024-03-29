package db

type Meta struct {
	ID          int64  `json:"id"`
	LinkID      int64  `json:"link_id,omitempty" db:"link_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Favicon     string `json:"favicon"`
}

type Bookmark struct {
	ID               int64  `json:"id"`
	URL              string `json:"url" binding:"required"`
	CreatedAt        int64  `json:"created_at,omitempty" db:"created_at"`
	LastUpdated      int64  `json:"last_updated,omitempty" db:"last_updated"`
	LastSavedOffline int64  `json:"last_saved_offline,omitempty" db:"last_saved_offline"`
}

type Tag struct {
	ID          int64  `json:"id"`
	Name        string `json:"name" form:"name" binding:"required"`
	CreatedAt   int64  `json:"created_at,omitempty" form:"created_at" db:"created_at"`
	LastUpdated int64  `json:"last_updated,omitempty" form:"last_updated" db:"last_updated"`
}

type Folder struct {
	ID          int64  `json:"id"`
	Name        string `json:"name" form:"name" binding:"required"`
	ParentID    int64  `json:"parent_id" form:"parent_id" db:"parent_id"`
	CreatedAt   int64  `json:"created_at,omitempty" form:"created_at" db:"created_at"`
	LastUpdated int64  `json:"last_updated,omitempty" form:"last_updated" db:"last_updated"`
}
