package sqlite

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/keyboy21/url-shortner/internal/store"
	_ "modernc.org/sqlite"
)

type Store struct {
	db            *sqlx.DB
	urlRepository *UrlRepository
}

func ConnectSqlite(dbPath string) (*sqlx.DB, error) {
	db, err := sqlx.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("can not open sqliteDb: %w", err)
	}

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("can not verify connection: %w", err)
	}

	return db, nil
}

func New(db *sqlx.DB) *Store {
	return &Store{
		db: db,
		urlRepository: &UrlRepository{
			db: db,
		},
	}
}

func (s *Store) Url() store.UrlRepository {
	return s.urlRepository
}

func (s *Store) Close() error {
	if err := s.db.Close(); err != nil {
		return fmt.Errorf("failed to close database: %w", err)
	}
	return nil
}
