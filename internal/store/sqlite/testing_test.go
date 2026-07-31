package sqlite

import (
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/keyboy21/url-shortner/internal/model"
)

func SetupTestDB(t *testing.T, databaseUrl string) *sqlx.DB {
	t.Helper()

	db, err := sqlx.Open("sqlite", databaseUrl)
	if err != nil {
		t.Fatal(err)
	}

	if err := db.Ping(); err != nil {
		t.Fatal(err)
	}

	_, err = db.Exec(`CREATE TABLE urls (
		id INTEGER PRIMARY KEY,
		url TEXT NOT NULL UNIQUE,
		alias TEXT NOT NULL UNIQUE,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	);`)

	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = db.Close() })
	return db
}

func CreateTestURL(t *testing.T) *model.Url {
	t.Helper()

	return &model.Url{
		Url:   "https://github.com",
		Alias: "github",
	}
}
