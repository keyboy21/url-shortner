package sqlite

import (
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/keyboy21/url-shortner/internal/model"
	"github.com/keyboy21/url-shortner/internal/store"
)

type UrlRepository struct {
	db *sqlx.DB
}

func (r UrlRepository) SaveUrl(u *model.Url) (*model.Url, error) {
	createdUrl := &model.Url{}
	const query = `INSERT INTO urls (url, alias)
	VALUES ($1,$2)
	RETURNING id, url, alias, created_at`

	if err := r.db.Get(createdUrl, query, u.Url, u.Alias); err != nil {
		return nil, err
	}
	return createdUrl, nil
}

func (r UrlRepository) FindById(id int) (*model.Url, error) {
	url := &model.Url{}

	const query = `SELECT * FROM urls WHERE id = $1`

	if err := r.db.Get(url, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, store.ErrUrlNotFound
		}
		return nil, err
	}

	return url, nil
}

func (r UrlRepository) FindByUrl(url string) (*model.Url, error) {
	u := &model.Url{}

	if err := r.db.Get(u, "SELECT * FROM urls WHERE url = $1", url); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, store.ErrUrlNotFound
		}
		return nil, err
	}

	return u, nil
}

func (r UrlRepository) FindByAlias(alias string) (*model.Url, error) {
	u := &model.Url{}

	if err := r.db.Get(u, "SELECT * FROM urls WHERE alias = $1", alias); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, store.ErrUrlNotFound
		}
		return nil, err
	}

	return u, nil
}

func (r UrlRepository) DeleteUrl(url string) (*model.Url, error) {
	u := &model.Url{}

	if err := r.db.Get(u, "DELETE FROM urls WHERE url = $1 RETURNING id, url, alias, created_at", url); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, store.ErrUrlNotFound
		}

		return nil, err
	}

	return u, nil
}
