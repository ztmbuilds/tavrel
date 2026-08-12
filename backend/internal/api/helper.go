package api

import (
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
