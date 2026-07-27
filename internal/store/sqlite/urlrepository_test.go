package sqlite_test

import (
	"path/filepath"
	"testing"

	"github.com/keyboy21/url-shortner/internal/store"
	"github.com/keyboy21/url-shortner/internal/store/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUrl_SaveUrl(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.sqlite")
	db := sqlite.SetupTestDB(t, dbPath)
	s := sqlite.New(db)
	u := sqlite.CreateTestURL(t)

	created, err := s.Url().SaveUrl(u)
	assert.NoError(t, err)
	assert.NotNil(t, created)

	assert.Greater(t, created.Id, 0)
	assert.Equal(t, u.Url, created.Url)
	assert.Equal(t, u.Alias, created.Alias)
	assert.False(t, created.CreatedAt.IsZero())
}

func TestUrl_FindById(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.sqlite")
	db := sqlite.SetupTestDB(t, dbPath)
	s := sqlite.New(db)

	created, err := s.Url().SaveUrl(sqlite.CreateTestURL(t))
	require.NoError(t, err)
	require.NotNil(t, created)

	found, err := s.Url().FindById(created.Id)
	assert.NoError(t, err)
	assert.NotNil(t, found)

	assert.Equal(t, created.Id, found.Id)
	assert.Equal(t, created.Url, found.Url)
	assert.Equal(t, created.Alias, found.Alias)
	assert.Equal(t, created.CreatedAt, found.CreatedAt)
}

func TestUrl_FindByUrl(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.sqlite")
	db := sqlite.SetupTestDB(t, dbPath)
	s := sqlite.New(db)

	created, err := s.Url().SaveUrl(sqlite.CreateTestURL(t))
	require.NoError(t, err)
	require.NotNil(t, created)

	found, err := s.Url().FindByUrl(created.Url)

	assert.NoError(t, err)
	assert.NotNil(t, found)

	assert.Equal(t, created.Id, found.Id)
	assert.Equal(t, created.Url, found.Url)
	assert.Equal(t, created.Alias, found.Alias)
	assert.Equal(t, created.CreatedAt, found.CreatedAt)
}

func TestUrl_FindByAlias(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.sqlite")
	db := sqlite.SetupTestDB(t, dbPath)
	s := sqlite.New(db)

	created, err := s.Url().SaveUrl(sqlite.CreateTestURL(t))
	require.NoError(t, err)
	require.NotNil(t, created)

	found, err := s.Url().FindByAlias(created.Alias)

	assert.NoError(t, err)
	assert.NotNil(t, found)

	assert.Equal(t, created.Id, found.Id)
	assert.Equal(t, created.Url, found.Url)
	assert.Equal(t, created.Alias, found.Alias)
	assert.Equal(t, created.CreatedAt, found.CreatedAt)
}

func TestUrl_DeleteUrl(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.sqlite")
	db := sqlite.SetupTestDB(t, dbPath)
	s := sqlite.New(db)

	created, err := s.Url().SaveUrl(sqlite.CreateTestURL(t))
	require.NoError(t, err)
	require.NotNil(t, created)

	deleted, err := s.Url().DeleteUrl(created.Url)

	assert.NoError(t, err)
	assert.NotNil(t, deleted)

	assert.Equal(t, created.Id, deleted.Id)
	assert.Equal(t, created.Url, deleted.Url)
	assert.Equal(t, created.Alias, deleted.Alias)
	assert.Equal(t, created.CreatedAt, deleted.CreatedAt)

	found, err := s.Url().FindById(deleted.Id)

	assert.Nil(t, found)
	require.ErrorIs(t, err, store.ErrUrlNotFound)
}
