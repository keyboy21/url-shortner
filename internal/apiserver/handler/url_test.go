package handler_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/keyboy21/url-shortner/internal/apiserver/handler"
	"github.com/keyboy21/url-shortner/internal/apperror"
	"github.com/keyboy21/url-shortner/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type fakeUrlService struct {
	created      *model.Url
	createErr    error
	found        *model.Url
	findErr      error
	deleteErr    error
	createdUrl   string
	createdAlias string
	deletedAlias string
}

// boilerplates for handler UrlService
func (f *fakeUrlService) Create(_ context.Context, url, alias string) (*model.Url, error) {
	f.createdUrl = url
	f.createdAlias = alias
	return f.created, f.createErr
}

func (f *fakeUrlService) FindByAlias(context.Context, string) (*model.Url, error) {
	return f.found, f.findErr
}

func (f *fakeUrlService) DeleteByAlias(_ context.Context, alias string) error {
	f.deletedAlias = alias
	return f.deleteErr
}

func TestUrlHandler_Create(t *testing.T) {
	urlService := &fakeUrlService{
		created: &model.Url{
			Id:        1,
			Url:       "https://go.dev",
			Alias:     "golang",
			CreatedAt: time.Now().UTC(),
		},
	}
	router := testRouter(urlService)
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/urls",
		strings.NewReader(`{"url":"https://go.dev","alias":"golang"}`),
	)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	require.Equal(t, http.StatusCreated, response.Code)
	assert.Equal(t, "/api/v1/urls/golang", response.Header().Get("Location"))
	assert.Contains(t, response.Body.String(), `"alias":"golang"`)
	assert.Equal(t, "https://go.dev", urlService.createdUrl)
	assert.Equal(t, "golang", urlService.createdAlias)
}

func TestUrlHandler_Create_InvalidJson(t *testing.T) {
	router := testRouter(&fakeUrlService{})
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/urls",
		strings.NewReader(`{"url":`),
	)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	assert.Equal(t, http.StatusBadRequest, response.Code)
}

func TestUrlHandler_Create_DuplicateAlias(t *testing.T) {
	router := testRouter(&fakeUrlService{createErr: apperror.ErrAliasAlreadyUsed})
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/urls",
		strings.NewReader(`{"url":"https://go.dev","alias":"golang"}`),
	)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	assert.Equal(t, http.StatusConflict, response.Code)
}

func TestUrlHandler_Get_NotFound(t *testing.T) {
	router := testRouter(&fakeUrlService{findErr: apperror.ErrUrlNotFound})
	request := httptest.NewRequest(http.MethodGet, "/api/v1/urls/missing", nil)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	assert.Equal(t, http.StatusNotFound, response.Code)
}

func TestUrlHandler_Delete(t *testing.T) {
	urlService := &fakeUrlService{}
	router := testRouter(urlService)
	request := httptest.NewRequest(http.MethodDelete, "/api/v1/urls/golang", nil)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	assert.Equal(t, http.StatusNoContent, response.Code)
	assert.Empty(t, response.Body.String())
	assert.Equal(t, "golang", urlService.deletedAlias)
}

func TestUrlHandler_Redirect(t *testing.T) {
	router := testRouter(&fakeUrlService{
		found: &model.Url{
			Url:   "https://go.dev",
			Alias: "golang",
		},
	})
	request := httptest.NewRequest(http.MethodGet, "/golang", nil)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	assert.Equal(t, http.StatusFound, response.Code)
	assert.Equal(t, "https://go.dev", response.Header().Get("Location"))
}

func TestUrlHandler_UnexpectedError(t *testing.T) {
	router := testRouter(&fakeUrlService{findErr: errors.New("database unavailable")})
	request := httptest.NewRequest(http.MethodGet, "/api/v1/urls/golang", nil)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	assert.Equal(t, http.StatusInternalServerError, response.Code)
	assert.NotContains(t, response.Body.String(), "database unavailable")
}

func testRouter(urlService handler.UrlService) http.Handler {
	urlHandler := handler.NewUrlHandler(urlService, zap.NewNop().Sugar())
	router := chi.NewRouter()
	router.Post("/api/v1/urls", urlHandler.Create)
	router.Get("/api/v1/urls/{alias}", urlHandler.Get)
	router.Delete("/api/v1/urls/{alias}", urlHandler.Delete)
	router.Get("/{alias}", urlHandler.Redirect)
	return router
}
