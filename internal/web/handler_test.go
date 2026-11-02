package web

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"
)

const (
	testIndexHTML = "<!doctype html><html><body>app</body></html>"
	testAssetJS   = "console.log(1)"
)

func testFS() fstest.MapFS {
	return fstest.MapFS{
		"index.html":        {Data: []byte(testIndexHTML)},
		"favicon.svg":       {Data: []byte("<svg></svg>")},
		"assets/app-abc.js": {Data: []byte(testAssetJS)},
	}
}

func TestHandler(t *testing.T) {
	handler := newHandler(testFS())

	tests := []struct {
		name        string
		method      string
		target      string
		wantStatus  int
		wantBody    string
		wantHeaders map[string]string
	}{
		{
			name:       "root serves shell",
			method:     http.MethodGet,
			target:     "/",
			wantStatus: http.StatusOK,
			wantBody:   testIndexHTML,
			wantHeaders: map[string]string{
				"Cache-Control": "no-cache",
				"Content-Type":  "text/html; charset=utf-8",
			},
		},
		{
			name:       "deep link falls back to shell",
			method:     http.MethodGet,
			target:     "/traces/abc123",
			wantStatus: http.StatusOK,
			wantBody:   testIndexHTML,
			wantHeaders: map[string]string{
				"Cache-Control": "no-cache",
			},
		},
		{
			name:       "hashed asset gets immutable cache header",
			method:     http.MethodGet,
			target:     "/assets/app-abc.js",
			wantStatus: http.StatusOK,
			wantBody:   testAssetJS,
			wantHeaders: map[string]string{
				"Cache-Control": "public, max-age=31536000, immutable",
			},
		},
		{
			name:       "index.html canonicalizes to root",
			method:     http.MethodGet,
			target:     "/index.html",
			wantStatus: http.StatusMovedPermanently,
			wantHeaders: map[string]string{
				"Location": "./",
			},
		},
		{
			name:       "unhashed file served as-is",
			method:     http.MethodGet,
			target:     "/favicon.svg",
			wantStatus: http.StatusOK,
			wantBody:   "<svg></svg>",
		},
		{
			name:       "non-GET gets 405",
			method:     http.MethodPost,
			target:     "/",
			wantStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.target, nil)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
			if tt.wantBody != "" && rec.Body.String() != tt.wantBody {
				t.Fatalf("body = %q, want %q", rec.Body.String(), tt.wantBody)
			}
			for key, want := range tt.wantHeaders {
				if got := rec.Header().Get(key); got != want {
					t.Fatalf("header %s = %q, want %q", key, got, want)
				}
			}
		})
	}
}

func TestHandlerUnbuilt(t *testing.T) {
	handler := newHandler(fstest.MapFS{})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusServiceUnavailable)
	}
	// Generic body only: no repo internals (paths, commands) leak to clients.
	if want := "service unavailable\n"; rec.Body.String() != want {
		t.Fatalf("body = %q, want %q", rec.Body.String(), want)
	}
}
