package storage

type Store interface {
	Urls() UrlsRepository
}
