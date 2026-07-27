package sqlite

import (
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/keyboy21/url-shortner/internal/storage"
	_ "modernc.org/sqlite"
)

type Store struct {
	db            *sqlx.DB
	urlRepository *UrlRepository
}

func ConnectSqlite(dbPath string) *sqlx.DB {
	db, err := sqlx.Open("sqlite", dbPath)
	if err != nil {
		log.Fatalf("can not open sqliteDb: %s", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("can not verify connection: %s", err)
	}

	return db
}

func New(db *sqlx.DB) *Store {
	return &Store{
		db: db,
		urlRepository: &UrlRepository{
			db: db,
		},
	}
}

func (s *Store) Url() storage.UrlRepository {
	return s.urlRepository
}

func (s *Store) Close() error {
	if err := s.db.Close(); err != nil {
		return fmt.Errorf("failed to close database: %w", err)
	}
	return nil
}
