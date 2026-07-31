package testutils

import (
	"path/filepath"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/keyboy21/url-shortner/internal/model"
)

func NewSQLiteDB(t testing.TB) *sqlx.DB {
	t.Helper()

	db, err := sqlx.Open("sqlite", filepath.Join(t.TempDir(), "test.sqlite"))
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	if err := db.Ping(); err != nil {
		t.Fatalf("ping test database: %v", err)
	}

	_, err = db.Exec(`CREATE TABLE urls (
		id INTEGER PRIMARY KEY,
		url TEXT NOT NULL UNIQUE,
		alias TEXT NOT NULL UNIQUE,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		t.Fatalf("create test schema: %v", err)
	}

	return db
}

func CreateTestURL() *model.Url {
	return &model.Url{
		Url:   "https://github.com",
		Alias: "github",
	}
}
