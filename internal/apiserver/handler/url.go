package handler

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	apperror "github.com/keyboy21/url-shortner/internal/lib"
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
	var req createUrlRequest
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodySize)
	if err := render.DecodeJSON(r.Body, &req); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			ErrorResponse(w, r, http.StatusRequestEntityTooLarge, "request body is too large")
			return
		}
		ErrorResponse(w, r, http.StatusBadRequest, "invalid JSON body")
		return
	}

	created, err := h.service.Create(r.Context(), req.Url, req.Alias)
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}

	w.Header().Set("Location", urlCollectionPath+"/"+created.Alias)
	SuccessResponse(w, r, http.StatusCreated, created)
}

func (h *UrlHandler) Get(w http.ResponseWriter, r *http.Request) {
	found, err := h.service.FindByAlias(r.Context(), chi.URLParam(r, "alias"))
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}

	SuccessResponse(w, r, http.StatusOK, found)
}

func (h *UrlHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.service.DeleteByAlias(r.Context(), chi.URLParam(r, "alias")); err != nil {
		h.writeServiceError(w, r, err)
		return
	}

	SuccessResponse(w, r, http.StatusNoContent, nil)
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
		ErrorResponse(w, r, http.StatusUnprocessableEntity, "URL must be an absolute http or https URL")
	case errors.Is(err, apperror.ErrInvalidAlias):
		ErrorResponse(w, r, http.StatusUnprocessableEntity, "alias must contain 3 to 64 letters, numbers, hyphens, or underscores")
	case errors.Is(err, apperror.ErrUrlNotFound):
		ErrorResponse(w, r, http.StatusNotFound, "URL not found")
	case errors.Is(err, apperror.ErrUrlAlreadyExists):
		ErrorResponse(w, r, http.StatusConflict, "URL already exists")
	case errors.Is(err, apperror.ErrAliasAlreadyUsed):
		ErrorResponse(w, r, http.StatusConflict, "alias already used")
	default:
		h.logger.Errorw(
			"request failed",
			"error", err,
			"request_id", chimiddleware.GetReqID(r.Context()),
		)
		ErrorResponse(w, r, http.StatusInternalServerError, "internal server error")
	}
}
