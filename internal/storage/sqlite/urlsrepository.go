package sqlite

import (
	"database/sql"

	"github.com/keyboy21/url-shortner/internal/model"
)

type UrlRepository struct {
	db *sql.DB
}

func (u UrlRepository) SaveUrl(url, alias string) error {
	return nil
}

func (u UrlRepository) FindById(id int) (*model.Urls, error) {
	return nil, nil
}

func (u UrlRepository) FindByUrl(url string) (*model.Urls, error) {
	return nil, nil
}

func (u UrlRepository) FindByAlias(alias string) (*model.Urls, error) {
	return nil, nil
}
