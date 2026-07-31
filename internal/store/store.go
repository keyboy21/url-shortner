package store

import (
	"context"

	"github.com/keyboy21/url-shortner/internal/model"
)

type UrlInterface interface {
	FindById(context.Context, int) (*model.Url, error)
	FindByUrl(context.Context, string) (*model.Url, error)
	FindByAlias(context.Context, string) (*model.Url, error)
	SaveUrl(context.Context, *model.Url) (*model.Url, error)
	DeleteByAlias(context.Context, string) (*model.Url, error)
}

type Interface interface {
	Url() UrlInterface
	Close() error
}
