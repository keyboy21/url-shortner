package store

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

type Store struct {
	db            *sqlx.DB
	urlRepository *UrlRepository
}

func New(db *sqlx.DB) *Store {
	return &Store{
		db: db,
		urlRepository: &UrlRepository{
			db: db,
		},
	}
}

func (s *Store) Url() UrlInterface {
	return s.urlRepository
}

func (s *Store) Close() error {
	if err := s.db.Close(); err != nil {
		return fmt.Errorf("failed to close database: %w", err)
	}
	return nil
}
