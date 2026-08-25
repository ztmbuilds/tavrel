package api

import (
	"context"
	"net/http"
	"noname/internal/store"
	"strconv"
)

func newListOptions[C any](r *http.Request) store.ListOptions[C] {
	limit := 50

	if value := r.URL.Query().Get("limit"); value != "" {
		if num, err := strconv.Atoi(value); err == nil && num > 0 {

			limit = num
		}
	}

	options := store.ListOptions[C]{
		Limit:  limit,
		Cursor: nil,
	}

	encodedCursor := r.URL.Query().Get("cursor")

	cursor, ok := decodeCursor[C](encodedCursor)
	if ok {
		options.Cursor = &cursor
	}

	return options

}

func writeListResponse[T, C any](w http.ResponseWriter, message string, page store.Page[T, C]) {
	meta := &Meta{}
	if cursor, ok := encodeCursor(page.NextCursor); ok {
		meta.NextCursor = &cursor
	}

	writeJson(w, http.StatusOK, &Response{
		Message: message,
		Data:    page.Items,
		Meta:    meta,
	})

}

func listHandler[T, C, O any](buildOptionsFn func(*http.Request) O, listFn func(context.Context, O) (store.Page[T, C], error), msg string) AppHandler {
	return func(w http.ResponseWriter, r *http.Request) error {

		options := buildOptionsFn(r)

		page, err := listFn(r.Context(), options)

		if err != nil {
			return Internal(err)
		}

		writeListResponse(w, msg, page)

		return nil

	}
}
