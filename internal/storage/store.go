package storage

import "github.com/keyboy21/url-shortner/internal/model"

type UrlRepository interface {
	FindById(int) (*model.Url, error)
	FindByUrl(string) (*model.Url, error)
	FindByAlias(string) (*model.Url, error)
	SaveUrl(u *model.Url) (*model.Url,error)
}

type Store interface {
	Url() UrlRepository
	Close() error
}
