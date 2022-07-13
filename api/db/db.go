package db

import (
	"os"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

var db *sqlx.DB

func InitializeDB() (db *sqlx.DB) {
	dsn := "file:" + os.Getenv("DB_PATH") + "?_pragma=foreign_keys(1)"
	db = sqlx.MustOpen("sqlite", dsn)

	schema := `--sql
CREATE TABLE IF NOT EXISTS links (
	id INTEGER NOT NULL PRIMARY KEY, 
	url TEXT NOT NULL UNIQUE,
	created_at INTEGER DEFAULT (unixepoch()), 
	last_updated INTEGER DEFAULT (unixepoch()),
	last_saved_offline INTEGER DEFAULT 0
);
CREATE TABLE IF NOT EXISTS meta (
	id INTEGER NOT NULL PRIMARY KEY,
	link_id INTEGER NOT NULL UNIQUE REFERENCES links(id) ON DELETE CASCADE,
	title TEXT,
	description TEXT,
	favicon TEXT
);
CREATE TRIGGER IF NOT EXISTS on_meta_update UPDATE ON meta
	BEGIN
		UPDATE links SET last_updated = (unixepoch()) WHERE id = OLD.link_id;
	END;
CREATE TABLE IF NOT EXISTS tags (
	id INTEGER NOT NULL PRIMARY KEY,
	name TEXT NOT NULL UNIQUE,
	created_at INTEGER DEFAULT (unixepoch()),
	last_updated INTEGER DEFAULT (unixepoch())
);
CREATE TRIGGER IF NOT EXISTS update_tags_timestamp UPDATE OF name ON tags 
	BEGIN
		UPDATE tags SET last_updated = (unixepoch()) WHERE id = OLD.id;
	END;
CREATE TABLE IF NOT EXISTS links_tags (
	tag_id INTEGER NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
	link_id INTEGER NOT NULL REFERENCES links(id) ON DELETE CASCADE,
	UNIQUE(tag_id, link_id) ON CONFLICT IGNORE
);
CREATE TRIGGER IF NOT EXISTS on_tag_add INSERT ON links_tags 
	BEGIN
		UPDATE links SET last_updated = (unixepoch()) WHERE id = NEW.link_id;
	END;
CREATE TRIGGER IF NOT EXISTS on_tag_remove DELETE ON links_tags 
	BEGIN
		UPDATE links SET last_updated = (unixepoch()) WHERE id = OLD.link_id;
	END;
CREATE TABLE IF NOT EXISTS folders (
	id INTEGER NOT NULL PRIMARY KEY,
	name TEXT NOT NULL,
	parent_id TEXT,
	created_at INTEGER DEFAULT (unixepoch()),
	last_updated INTEGER DEFAULT (unixepoch()),
	UNIQUE(name, parent_id)
);
CREATE INDEX IF NOT EXISTS folder_parent_id ON folders(parent_id);
CREATE TRIGGER IF NOT EXISTS update_folders_timestamp UPDATE OF name, parent_id ON folders 
	BEGIN
		UPDATE folders SET last_updated = (unixepoch()) WHERE id = OLD.id;
	END;
CREATE TABLE IF NOT EXISTS links_folders (
	folder_id TEXT NOT NULL REFERENCES folders(id) ON DELETE CASCADE,
	link_id INTEGER NOT NULL REFERENCES links(id) ON DELETE CASCADE,
	UNIQUE(folder_id, link_id) ON CONFLICT IGNORE
);
CREATE TRIGGER IF NOT EXISTS on_folder_add INSERT ON links_folders 
	BEGIN
		UPDATE links SET last_updated = (unixepoch()) WHERE id = NEW.link_id;
	END;
CREATE TRIGGER IF NOT EXISTS on_folder_remove DELETE ON links_folders 
	BEGIN
		UPDATE links SET last_updated = (unixepoch()) WHERE id = OLD.link_id;
	END;
`
	db.MustExec(schema)
	return db
}

func GetDB() *sqlx.DB {
	if db == nil {
		db = InitializeDB()
	}
	return db
}
