package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/keyboy21/url-shortner/internal/apperror"
	"github.com/keyboy21/url-shortner/internal/model"
	"github.com/keyboy21/url-shortner/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeUrlRepository struct {
	saved       *model.Url
	saveErr     error
	found       *model.Url
	findErr     error
	deleted     *model.Url
	deleteErr   error
	deleteAlias string
}

func (f *fakeUrlRepository) FindById(context.Context, int) (*model.Url, error) {
	return f.found, f.findErr
}

func (f *fakeUrlRepository) FindByUrl(context.Context, string) (*model.Url, error) {
	return f.found, f.findErr
}

func (f *fakeUrlRepository) FindByAlias(context.Context, string) (*model.Url, error) {
	return f.found, f.findErr
}

func (f *fakeUrlRepository) SaveUrl(_ context.Context, url *model.Url) (*model.Url, error) {
	f.saved = url
	if f.saveErr != nil {
		return nil, f.saveErr
	}

	return &model.Url{
		Id:        1,
		Url:       url.Url,
		Alias:     url.Alias,
		CreatedAt: time.Now().UTC(),
	}, nil
}

func (f *fakeUrlRepository) DeleteByAlias(_ context.Context, alias string) (*model.Url, error) {
	f.deleteAlias = alias
	return f.deleted, f.deleteErr
}

func TestUrlService_Create(t *testing.T) {
	repository := &fakeUrlRepository{}
	urlService := service.NewUrlService(repository)

	created, err := urlService.Create(context.Background(), "https://go.dev", "golang")

	require.NoError(t, err)
	require.NotNil(t, created)
	assert.Equal(t, "https://go.dev", repository.saved.Url)
	assert.Equal(t, "golang", repository.saved.Alias)
	assert.Equal(t, repository.saved.Alias, created.Alias)
}

func TestUrlService_Create_GeneratesAlias(t *testing.T) {
	repository := &fakeUrlRepository{}
	urlService := service.NewUrlService(repository)

	created, err := urlService.Create(context.Background(), "https://go.dev", "")

	require.NoError(t, err)
	require.NotNil(t, created)
	assert.Len(t, created.Alias, 8)
	assert.NotEmpty(t, repository.saved.Alias)
}

func TestUrlService_Create_InvalidUrl(t *testing.T) {
	repository := &fakeUrlRepository{}
	urlService := service.NewUrlService(repository)

	created, err := urlService.Create(context.Background(), "not-a-url", "golang")

	assert.Nil(t, created)
	require.ErrorIs(t, err, apperror.ErrInvalidUrl)
	assert.Nil(t, repository.saved)
}

func TestUrlService_FindByAlias_NotFound(t *testing.T) {
	repository := &fakeUrlRepository{findErr: apperror.ErrUrlNotFound}
	urlService := service.NewUrlService(repository)

	found, err := urlService.FindByAlias(context.Background(), "missing")

	assert.Nil(t, found)
	require.ErrorIs(t, err, apperror.ErrUrlNotFound)
}

func TestUrlService_DeleteByAlias(t *testing.T) {
	repository := &fakeUrlRepository{}
	urlService := service.NewUrlService(repository)

	err := urlService.DeleteByAlias(context.Background(), "golang")

	require.NoError(t, err)
	assert.Equal(t, "golang", repository.deleteAlias)
}
