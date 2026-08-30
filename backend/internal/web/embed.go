package web

import (
	"embed"
	"io/fs"
	"net/http"
	"path"
	"strings"
)

// Files are refreshed by `make build` after the frontend production build.
// The checked-in fallback keeps direct `go build` useful from a clean checkout.
//
//go:embed dist/*
var files embed.FS

func Handler(api http.Handler) http.Handler {
	static, _ := fs.Sub(files, "dist")
	serve := http.FileServer(http.FS(static))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" || r.URL.Path == "/ready" || strings.HasPrefix(r.URL.Path, "/api/") {
			api.ServeHTTP(w, r)
			return
		}
		name := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
		if name == "" {
			name = "index.html"
		}
		if _, err := fs.Stat(static, name); err != nil {
			r.URL.Path = "/index.html"
		}
		w.Header().Set("Cache-Control", "no-cache")
		serve.ServeHTTP(w, r)
	})
}
