package api

import (
	"net/http"
)

type AppHandler func(w http.ResponseWriter, r *http.Request) error

func Make(h AppHandler) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		err := h(w, r)

		if err != nil {
			writeError(w, err)
			return
		}

	}
}
