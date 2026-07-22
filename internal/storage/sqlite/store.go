package sqlite

import (
	"database/sql"

	_ "modernc.org/sqlite"
)

type Storage struct {
	db             *sql.DB
	UrlsRepository *UrlsRepository
}

func New(db *sql.DB) *Storage {
	return &Storage{
		db: db,
	}
}
