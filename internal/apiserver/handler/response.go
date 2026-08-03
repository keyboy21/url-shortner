package handler

import (
	"net/http"

	"github.com/go-chi/render"
)

type Response struct {
	Status int    `json:"status"`
	Data   any    `json:"data,omitempty"`
	Error  string `json:"error,omitempty"`
}

func ErrorResponse(w http.ResponseWriter, r *http.Request, status int, message string) {
	writeResponse(w, r, status, Response{
		Status: status,
		Error:  message,
	})
}

func SuccessResponse(w http.ResponseWriter, r *http.Request, status int, value any) {
	if status == http.StatusNoContent {
		render.NoContent(w, r)
		return
	}

	writeResponse(w, r, status, Response{
		Status: status,
		Data:   value,
	})
}

func writeResponse(w http.ResponseWriter, r *http.Request, status int, response Response) {
	render.Status(r, status)
	render.JSON(w, r, response)
}
