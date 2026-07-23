package storage

import "github.com/keyboy21/url-shortner/internal/model"

type UrlRepository interface {
	FindById(int) (*model.Urls, error)
	FindByUrl(string) (*model.Urls, error)
	FindByAlias(string) (*model.Urls, error)
	SaveUrl(url, alias string) error
}

type Store interface {
	Url() UrlRepository
	Close() error
}
