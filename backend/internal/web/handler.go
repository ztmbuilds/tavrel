package web

import (
	"bytes"
	"io/fs"
	"log/slog"
	"net/http"
	"path"
	"strings"
	"time"
)

const indexFile = "index.html"

func Handler() http.Handler {
	content, err := fs.Sub(embedded, "dist")
	if err != nil {
		return unbuilt()
	}
	return newHandler(content)
}

func newHandler(content fs.FS) http.Handler {
	if _, err := fs.Stat(content, indexFile); err != nil {
		return unbuilt()
	}

	files := http.FileServer(http.FS(content))
	// Read once: the shell is small and immutable for the process lifetime.
	indexBytes, err := fs.ReadFile(content, indexFile)
	if err != nil {
		return unbuilt()
	}
	serveIndex := serveIndexHandler(indexBytes)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		rel := strings.TrimPrefix(path.Clean("/"+r.URL.Path), "/")
		if strings.HasPrefix(rel, "assets/") {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
			files.ServeHTTP(w, r)
			return
		}
		if rel == "" {
			serveIndex(w, r)
			return
		}
		if _, err := fs.Stat(content, rel); err != nil {
			serveIndex(w, r)
			return
		}
		files.ServeHTTP(w, r)
	})
}

func unbuilt() http.Handler {
	slog.Warn("frontend dist is empty; run `make build-frontend` from the repo root")
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "service unavailable", http.StatusServiceUnavailable)
	})
}

// serveIndexHandler serves the SPA shell directly instead of via FileServer
// because FileServer canonically redirects */index.html to */, which would
// turn the SPA fallback into a redirect loop.
func serveIndexHandler(indexBytes []byte) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		http.ServeContent(w, r, indexFile, time.Time{}, bytes.NewReader(indexBytes))
	}
}
