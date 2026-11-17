package api

import "net/http"

type errCode string

const (
	ErrNotFound        errCode = "NOT_FOUND"
	ErrInternal        errCode = "INTERNAL_SERVER_ERROR"
	ErrBadRequest      errCode = "BAD_REQUEST"
	ErrConflict        errCode = "CONFLICT"
	ErrUnauthorized    errCode = "UNAUTHORIZED"
	ErrForbidden       errCode = "FORBIDDEN"
	ErrPayloadTooLarge errCode = "PAYLOAD_TOO_LARGE"
)

type AppError struct {
	StatusCode int
	ErrCode    errCode
	Message    string
	Err        error
}

func (e *AppError) Error() string {
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func NotFound(msg string, err error) *AppError {

	return &AppError{
		Message:    msg,
		StatusCode: http.StatusNotFound,
		ErrCode:    ErrNotFound,
		Err:        err,
	}
}

func BadRequest(msg string, err error) *AppError {
	return &AppError{
		Message:    msg,
		StatusCode: http.StatusBadRequest,
		ErrCode:    ErrBadRequest,
		Err:        err,
	}
}

func Conflict(msg string, err error) *AppError {
	return &AppError{
		Message:    msg,
		StatusCode: http.StatusConflict,
		ErrCode:    ErrConflict,
		Err:        err,
	}
}

func PayloadTooLarge(msg string, err error) *AppError {
	return &AppError{
		Message:    msg,
		StatusCode: http.StatusRequestEntityTooLarge,
		ErrCode:    ErrPayloadTooLarge,
		Err:        err,
	}
}

func Unauthorized(msg string, err error) *AppError {
	return &AppError{
		Message:    msg,
		StatusCode: http.StatusUnauthorized,
		ErrCode:    ErrUnauthorized,
		Err:        err,
	}
}

func Forbidden(err error) *AppError {
	return &AppError{
		Message:    "Forbidden Action",
		StatusCode: http.StatusForbidden,
		ErrCode:    ErrForbidden,
		Err:        err,
	}
}

func Internal(err error) *AppError {
	return &AppError{
		Message:    "Internal Server Error",
		StatusCode: http.StatusInternalServerError,
		ErrCode:    ErrInternal,
		Err:        err,
	}
}
