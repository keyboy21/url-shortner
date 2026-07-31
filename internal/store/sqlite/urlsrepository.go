package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/keyboy21/url-shortner/internal/apperror"
	"github.com/keyboy21/url-shortner/internal/model"
	modernsqlite "modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"
)

type UrlRepository struct {
	db *sqlx.DB
}

func (r UrlRepository) SaveUrl(ctx context.Context, u *model.Url) (*model.Url, error) {
	createdUrl := &model.Url{}
	const query = `INSERT INTO urls (url, alias)
	VALUES ($1,$2)
	RETURNING id, url, alias, created_at`

	if err := r.db.GetContext(ctx, createdUrl, query, u.Url, u.Alias); err != nil {
		return nil, classifyConstraintError(err)
	}
	return createdUrl, nil
}

func (r UrlRepository) FindById(ctx context.Context, id int) (*model.Url, error) {
	url := &model.Url{}

	const query = `SELECT id, url, alias, created_at FROM urls WHERE id = $1`

	if err := r.db.GetContext(ctx, url, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.ErrUrlNotFound
		}
		return nil, err
	}

	return url, nil
}

func (r UrlRepository) FindByUrl(ctx context.Context, url string) (*model.Url, error) {
	u := &model.Url{}

	const query = `SELECT id, url, alias, created_at FROM urls WHERE url = $1`
	if err := r.db.GetContext(ctx, u, query, url); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.ErrUrlNotFound
		}
		return nil, err
	}

	return u, nil
}

func (r UrlRepository) FindByAlias(ctx context.Context, alias string) (*model.Url, error) {
	u := &model.Url{}

	const query = `SELECT id, url, alias, created_at FROM urls WHERE alias = $1`
	if err := r.db.GetContext(ctx, u, query, alias); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.ErrUrlNotFound
		}
		return nil, err
	}

	return u, nil
}

func (r UrlRepository) DeleteByAlias(ctx context.Context, alias string) (*model.Url, error) {
	u := &model.Url{}

	const query = `DELETE FROM urls
	WHERE alias = $1
	RETURNING id, url, alias, created_at`
	if err := r.db.GetContext(ctx, u, query, alias); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.ErrUrlNotFound
		}

		return nil, err
	}

	return u, nil
}

func classifyConstraintError(err error) error {
	var sqliteErr *modernsqlite.Error
	if !errors.As(err, &sqliteErr) || sqliteErr.Code() != sqlite3.SQLITE_CONSTRAINT_UNIQUE {
		return err
	}

	switch {
	case strings.Contains(sqliteErr.Error(), "urls.alias"):
		return apperror.ErrAliasAlreadyUsed
	case strings.Contains(sqliteErr.Error(), "urls.url"):
		return apperror.ErrUrlAlreadyExists
	default:
		return err
	}
}
