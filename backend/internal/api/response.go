package api

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
)

type apiError struct {
	Code    errCode `json:"code"`
	Message string  `json:"message"`
}
type Meta struct {
	NextCursor *string `json:"nextCursor"`
	Total      *int64  `json:"total,omitempty"`
}

type Response struct {
	Message string    `json:"message,omitempty"`
	Data    any       `json:"data,omitempty"`
	Meta    *Meta     `json:"meta"`
	Error   *apiError `json:"error,omitempty"`
}

func writeJson(w http.ResponseWriter, statusCode int, payload *Response) {

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	err := json.NewEncoder(w).Encode(payload)

	if err != nil {
		slog.Error("Error writing response", "error", err)
	}
}

func writeError(w http.ResponseWriter, err error) {

	var e *AppError

	if ok := errors.As(err, &e); ok {

		writeJson(w, e.StatusCode, &Response{
			Error: &apiError{
				Code:    e.ErrCode,
				Message: e.Message,
			},
		})

		if e.ErrCode == ErrInternal {
			slog.Error("Internal Server Error", "error", e.Err)

		}
		return

	}

	writeJson(w, http.StatusInternalServerError, &Response{
		Error: &apiError{
			Code:    ErrInternal,
			Message: "Internal Server Error",
		},
	})

	slog.Error("Internal Server Error", "error", err)

}
