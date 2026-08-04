package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/keyboy21/url-shortner/internal/apiserver/handler"
	apperror "github.com/keyboy21/url-shortner/internal/lib"
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

	testCases := []struct {
		name         string
		payload      string
		expectedCode int
	}{
		{
			name:         "valid",
			payload:      `{"url":"https://go.dev","alias":"golang"}`,
			expectedCode: http.StatusCreated,
		},
		{
			name:         "invalid json",
			payload:      `{"url":""`,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "body too large",
			payload:      `{"url":"` + strings.Repeat("a", 1<<20) + `"}`,
			expectedCode: http.StatusRequestEntityTooLarge,
		},
		{
			name:         "duplicate alias",
			payload:      `{"url":"https://go.dev","alias":"golang"}`,
			expectedCode: http.StatusConflict,
		}}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/v1/urls", strings.NewReader(test.payload))
			res := httptest.NewRecorder()
			app.router.ServeHTTP(res, req)

			require.Equal(t, test.expectedCode, res.Code)

			var body handler.Response
			require.NoError(t, json.Unmarshal(res.Body.Bytes(), &body))
			assert.Equal(t, test.expectedCode, body.Status)
			if test.expectedCode == http.StatusCreated {
				assert.Equal(t, "/api/v1/urls/golang", res.Header().Get("Location"))
				assert.Contains(t, res.Body.String(), `"alias":"golang"`)
				created, _ := app.service.FindByAlias(context.Background(), "golang")
				assert.Equal(t, "https://go.dev", created.Url)
			}

		})
	}

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
