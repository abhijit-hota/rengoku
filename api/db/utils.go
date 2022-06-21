package db

import (
	"modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"
)

const (
	ALREADY_EXISTS = iota
)

func IsUniqueErr(err error) bool {
	e, ok := err.(*sqlite.Error)
	return ok && e.Code() == sqlite3.SQLITE_CONSTRAINT_UNIQUE
}
