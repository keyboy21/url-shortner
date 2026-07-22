package storage

import "github.com/keyboy21/url-shortner/internal/model"

type UrlsRepository interface {
	FindById(int) (*model.Urls, error)
	FindByUrl(string) (*model.Urls, error)
	Create(string) error
}
