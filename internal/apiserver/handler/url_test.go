package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/keyboy21/url-shortner/internal/apiserver/handler"
	"github.com/keyboy21/url-shortner/internal/lib"
	"github.com/keyboy21/url-shortner/internal/service"
	"github.com/keyboy21/url-shortner/internal/store"
	"github.com/keyboy21/url-shortner/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type testApp struct {
	router  http.Handler
	service *service.UrlService
	store   *store.Store
}

func TestUrlHandler_Create(t *testing.T) {
	app := newTestApp(t)
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/urls",
		strings.NewReader(`{"url":"https://go.dev","alias":"golang"}`),
	)
	response := httptest.NewRecorder()

	app.router.ServeHTTP(response, request)

	require.Equal(t, http.StatusCreated, response.Code)
	assert.Equal(t, "/api/v1/urls/golang", response.Header().Get("Location"))
	assert.Contains(t, response.Body.String(), `"alias":"golang"`)

	created, err := app.service.FindByAlias(context.Background(), "golang")
	require.NoError(t, err)
	assert.Equal(t, "https://go.dev", created.Url)
}

func TestUrlHandler_Create_InvalidJson(t *testing.T) {
	app := newTestApp(t)
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/urls",
		strings.NewReader(`{"url":`),
	)
	response := httptest.NewRecorder()

	app.router.ServeHTTP(response, request)

	require.Equal(t, http.StatusBadRequest, response.Code)
	assert.Equal(t, "application/json", response.Header().Get("Content-Type"))

	var body handler.Response
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
	assert.Equal(t, http.StatusBadRequest, body.Status)
	assert.Equal(t, "invalid JSON body", body.Error)
	assert.Nil(t, body.Data)
}

func TestUrlHandler_Create_BodyTooLarge(t *testing.T) {
	app := newTestApp(t)
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/urls",
		strings.NewReader(`{"url":"`+strings.Repeat("a", 1<<20)+`"}`),
	)
	response := httptest.NewRecorder()

	app.router.ServeHTTP(response, request)

	require.Equal(t, http.StatusRequestEntityTooLarge, response.Code)

	var body handler.Response
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
	assert.Equal(t, http.StatusRequestEntityTooLarge, body.Status)
	assert.Equal(t, "request body is too large", body.Error)
}

func TestUrlHandler_Create_DuplicateAlias(t *testing.T) {
	app := newTestApp(t)
	_, err := app.service.Create(context.Background(), "https://example.com", "golang")
	require.NoError(t, err)

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/urls",
		strings.NewReader(`{"url":"https://go.dev","alias":"golang"}`),
	)
	response := httptest.NewRecorder()

	app.router.ServeHTTP(response, request)

	assert.Equal(t, http.StatusConflict, response.Code)
}

func TestUrlHandler_Get_NotFound(t *testing.T) {
	app := newTestApp(t)
	request := httptest.NewRequest(http.MethodGet, "/api/v1/urls/missing", nil)
	response := httptest.NewRecorder()

	app.router.ServeHTTP(response, request)

	assert.Equal(t, http.StatusNotFound, response.Code)
}

func TestUrlHandler_Delete(t *testing.T) {
	app := newTestApp(t)
	_, err := app.service.Create(context.Background(), "https://go.dev", "golang")
	require.NoError(t, err)

	request := httptest.NewRequest(http.MethodDelete, "/api/v1/urls/golang", nil)
	response := httptest.NewRecorder()

	app.router.ServeHTTP(response, request)

	assert.Equal(t, http.StatusNoContent, response.Code)
	assert.Empty(t, response.Body.String())

	found, err := app.service.FindByAlias(context.Background(), "golang")
	assert.Nil(t, found)
	require.ErrorIs(t, err, apperror.ErrUrlNotFound)
}

func TestUrlHandler_Redirect(t *testing.T) {
	app := newTestApp(t)
	_, err := app.service.Create(context.Background(), "https://go.dev", "golang")
	require.NoError(t, err)

	request := httptest.NewRequest(http.MethodGet, "/golang", nil)
	response := httptest.NewRecorder()

	app.router.ServeHTTP(response, request)

	assert.Equal(t, http.StatusFound, response.Code)
	assert.Equal(t, "https://go.dev", response.Header().Get("Location"))
}

func TestUrlHandler_UnexpectedError(t *testing.T) {
	app := newTestApp(t)
	require.NoError(t, app.store.Close())

	request := httptest.NewRequest(http.MethodGet, "/api/v1/urls/golang", nil)
	response := httptest.NewRecorder()

	app.router.ServeHTTP(response, request)

	assert.Equal(t, http.StatusInternalServerError, response.Code)
	assert.NotContains(t, response.Body.String(), "database is closed")
}

func newTestApp(t *testing.T) *testApp {
	t.Helper()

	store := store.New(testutils.NewSQLiteDB(t))
	urlService := service.NewUrlService(store.Url())
	urlHandler := handler.NewUrlHandler(urlService, zap.NewNop().Sugar())

	return &testApp{
		router:  urlHandler.Routes(),
		service: urlService,
		store:   store,
	}
}
