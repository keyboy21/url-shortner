package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/keyboy21/url-shortner/internal/lib"
	"github.com/keyboy21/url-shortner/internal/service"
	"go.uber.org/zap"
)

const (
	maxRequestBodySize = 1 << 20
	urlCollectionPath  = "/api/v1/urls"
)

type UrlHandler struct {
	service *service.UrlService
	logger  *zap.SugaredLogger
}

type createUrlRequest struct {
	Url   string `json:"url"`
	Alias string `json:"alias,omitempty"`
}

func NewUrlHandler(service *service.UrlService, logger *zap.SugaredLogger) *UrlHandler {
	return &UrlHandler{
		service: service,
		logger:  logger,
	}
}

func (h *UrlHandler) Routes() http.Handler {
	router := chi.NewRouter()
	router.Post(urlCollectionPath, h.Create)
	router.Get(urlCollectionPath+"/{alias}", h.Get)
	router.Delete(urlCollectionPath+"/{alias}", h.Delete)
	router.Get("/{alias}", h.Redirect)
	return router
}

func (h *UrlHandler) Create(w http.ResponseWriter, r *http.Request) {
	var request createUrlRequest
	if err := decodeJson(w, r, &request); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			writeError(w, http.StatusRequestEntityTooLarge, "request body is too large")
			return
		}
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	created, err := h.service.Create(r.Context(), request.Url, request.Alias)
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}

	w.Header().Set("Location", urlCollectionPath+"/"+created.Alias)
	writeJson(w, http.StatusCreated, created)
}

func (h *UrlHandler) Get(w http.ResponseWriter, r *http.Request) {
	found, err := h.service.FindByAlias(r.Context(), chi.URLParam(r, "alias"))
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}

	writeJson(w, http.StatusOK, found)
}

func (h *UrlHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.service.DeleteByAlias(r.Context(), chi.URLParam(r, "alias")); err != nil {
		h.writeServiceError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *UrlHandler) Redirect(w http.ResponseWriter, r *http.Request) {
	found, err := h.service.FindByAlias(r.Context(), chi.URLParam(r, "alias"))
	if err != nil {
		if errors.Is(err, apperror.ErrUrlNotFound) || errors.Is(err, apperror.ErrInvalidAlias) {
			http.NotFound(w, r)
			return
		}
		h.writeServiceError(w, r, err)
		return
	}

	http.Redirect(w, r, found.Url, http.StatusFound)
}

func (h *UrlHandler) writeServiceError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, apperror.ErrInvalidUrl):
		writeError(w, http.StatusUnprocessableEntity, "URL must be an absolute http or https URL")
	case errors.Is(err, apperror.ErrInvalidAlias):
		writeError(w, http.StatusUnprocessableEntity, "alias must contain 3 to 64 letters, numbers, hyphens, or underscores")
	case errors.Is(err, apperror.ErrUrlNotFound):
		writeError(w, http.StatusNotFound, "URL not found")
	case errors.Is(err, apperror.ErrUrlAlreadyExists):
		writeError(w, http.StatusConflict, "URL already exists")
	case errors.Is(err, apperror.ErrAliasAlreadyUsed):
		writeError(w, http.StatusConflict, "alias already used")
	default:
		h.logger.Errorw(
			"request failed",
			"error", err,
			"request_id", chimiddleware.GetReqID(r.Context()),
		)
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}

func decodeJson(w http.ResponseWriter, r *http.Request, destination any) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodySize)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(destination); err != nil {
		return err
	}

	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("request body must contain one JSON object")
		}
		return err
	}

	return nil
}
