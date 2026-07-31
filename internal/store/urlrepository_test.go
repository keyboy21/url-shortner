package store

import (
	"context"
	"testing"

	"github.com/keyboy21/url-shortner/internal/apperror"
	"github.com/keyboy21/url-shortner/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUrl_SaveUrl(t *testing.T) {
	db := testutils.NewSQLiteDB(t)
	s := New(db)
	u := testutils.CreateTestURL()

	created, err := s.Url().SaveUrl(context.Background(), u)
	require.NoError(t, err)
	require.NotNil(t, created)

	assert.Greater(t, created.Id, 0)
	assert.Equal(t, u.Url, created.Url)
	assert.Equal(t, u.Alias, created.Alias)
	assert.False(t, created.CreatedAt.IsZero())
}

func TestUrl_FindById(t *testing.T) {
	db := testutils.NewSQLiteDB(t)
	s := New(db)

	created, err := s.Url().SaveUrl(context.Background(), testutils.CreateTestURL())
	require.NoError(t, err)
	require.NotNil(t, created)

	found, err := s.Url().FindById(context.Background(), created.Id)
	require.NoError(t, err)
	require.NotNil(t, found)

	assert.Equal(t, created.Id, found.Id)
	assert.Equal(t, created.Url, found.Url)
	assert.Equal(t, created.Alias, found.Alias)
	assert.Equal(t, created.CreatedAt, found.CreatedAt)
}

func TestUrl_FindByUrl(t *testing.T) {
	db := testutils.NewSQLiteDB(t)
	s := New(db)

	created, err := s.Url().SaveUrl(context.Background(), testutils.CreateTestURL())
	require.NoError(t, err)
	require.NotNil(t, created)

	found, err := s.Url().FindByUrl(context.Background(), created.Url)

	require.NoError(t, err)
	require.NotNil(t, found)

	assert.Equal(t, created.Id, found.Id)
	assert.Equal(t, created.Url, found.Url)
	assert.Equal(t, created.Alias, found.Alias)
	assert.Equal(t, created.CreatedAt, found.CreatedAt)
}

func TestUrl_FindByAlias(t *testing.T) {
	db := testutils.NewSQLiteDB(t)
	s := New(db)

	created, err := s.Url().SaveUrl(context.Background(), testutils.CreateTestURL())
	require.NoError(t, err)
	require.NotNil(t, created)

	found, err := s.Url().FindByAlias(context.Background(), created.Alias)

	require.NoError(t, err)
	require.NotNil(t, found)

	assert.Equal(t, created.Id, found.Id)
	assert.Equal(t, created.Url, found.Url)
	assert.Equal(t, created.Alias, found.Alias)
	assert.Equal(t, created.CreatedAt, found.CreatedAt)
}

func TestUrl_DeleteByAlias(t *testing.T) {
	db := testutils.NewSQLiteDB(t)
	s := New(db)

	created, err := s.Url().SaveUrl(context.Background(), testutils.CreateTestURL())
	require.NoError(t, err)
	require.NotNil(t, created)

	deleted, err := s.Url().DeleteByAlias(context.Background(), created.Alias)

	require.NoError(t, err)
	require.NotNil(t, deleted)

	assert.Equal(t, created.Id, deleted.Id)
	assert.Equal(t, created.Url, deleted.Url)
	assert.Equal(t, created.Alias, deleted.Alias)
	assert.Equal(t, created.CreatedAt, deleted.CreatedAt)

	found, err := s.Url().FindById(context.Background(), deleted.Id)

	assert.Nil(t, found)
	require.ErrorIs(t, err, apperror.ErrUrlNotFound)
}

func TestUrl_SaveUrl_DuplicateAlias(t *testing.T) {
	db := testutils.NewSQLiteDB(t)
	s := New(db)

	first := testutils.CreateTestURL()
	_, err := s.Url().SaveUrl(context.Background(), first)
	require.NoError(t, err)

	second := testutils.CreateTestURL()
	second.Url = "https://go.dev"
	_, err = s.Url().SaveUrl(context.Background(), second)
	require.ErrorIs(t, err, apperror.ErrAliasAlreadyUsed)
}
