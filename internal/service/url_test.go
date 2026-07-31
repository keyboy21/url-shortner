package service_test

import (
	"context"
	"testing"

	"github.com/keyboy21/url-shortner/internal/apperror"
	"github.com/keyboy21/url-shortner/internal/service"
	"github.com/keyboy21/url-shortner/internal/store/sqlite"
	"github.com/keyboy21/url-shortner/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUrlService_Create(t *testing.T) {
	urlService := newTestService(t)

	created, err := urlService.Create(context.Background(), "https://go.dev", "golang")

	require.NoError(t, err)
	require.NotNil(t, created)
	assert.Equal(t, "https://go.dev", created.Url)
	assert.Equal(t, "golang", created.Alias)

	found, err := urlService.FindByAlias(context.Background(), "golang")
	require.NoError(t, err)
	assert.Equal(t, created.Id, found.Id)
}

func TestUrlService_Create_GeneratesAlias(t *testing.T) {
	urlService := newTestService(t)

	created, err := urlService.Create(context.Background(), "https://go.dev", "")

	require.NoError(t, err)
	require.NotNil(t, created)
	assert.Len(t, created.Alias, 8)

	found, err := urlService.FindByAlias(context.Background(), created.Alias)
	require.NoError(t, err)
	assert.Equal(t, created.Id, found.Id)
}

func TestUrlService_Create_InvalidUrl(t *testing.T) {
	urlService := newTestService(t)

	created, err := urlService.Create(context.Background(), "not-a-url", "golang")

	assert.Nil(t, created)
	require.ErrorIs(t, err, apperror.ErrInvalidUrl)

	found, err := urlService.FindByAlias(context.Background(), "golang")
	assert.Nil(t, found)
	require.ErrorIs(t, err, apperror.ErrUrlNotFound)
}

func TestUrlService_FindByAlias_NotFound(t *testing.T) {
	urlService := newTestService(t)

	found, err := urlService.FindByAlias(context.Background(), "missing")

	assert.Nil(t, found)
	require.ErrorIs(t, err, apperror.ErrUrlNotFound)
}

func TestUrlService_DeleteByAlias(t *testing.T) {
	urlService := newTestService(t)
	_, err := urlService.Create(context.Background(), "https://go.dev", "golang")
	require.NoError(t, err)

	err = urlService.DeleteByAlias(context.Background(), "golang")

	require.NoError(t, err)

	found, err := urlService.FindByAlias(context.Background(), "golang")
	assert.Nil(t, found)
	require.ErrorIs(t, err, apperror.ErrUrlNotFound)
}

func newTestService(t *testing.T) *service.UrlService {
	t.Helper()

	store := sqlite.NewStore(testutils.NewSQLiteDB(t))
	return service.NewUrlService(store.Url())
}
