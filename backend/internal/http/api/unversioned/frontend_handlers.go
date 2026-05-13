package unversioned

import (
	"io/fs"
	"net/http"
	"strings"

	"github.com/auto-dns/pihole-cluster-admin/internal/frontend"
	"github.com/go-chi/chi"
)

func registerFrontend(r chi.Router, d Deps) {
	sub, err := fs.Sub(frontend.Files, "dist")
	if err != nil {
		d.Logger.Warn().Err(err).Msg("No embedded frontend found; skipping static file routes")
		return
	}

	// Serve all frontend paths with SPA fallback
	fileServer := http.FileServer(http.FS(sub))
	r.Handle("/*", spaHandler(sub, fileServer))
}

func spaHandler(sub fs.FS, fileServer http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// If it's a likely static asset, set long cache headers (Vite filenames are content-hashed).
		if isLikelyAsset(path) {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		}

		// Try to open the requested path from the embedded FS
		trimmed := strings.TrimPrefix(path, "/")
		if trimmed == "" {
			trimmed = "index.html"
		}
		if f, err := sub.Open(trimmed); err == nil {
			_ = f.Close()
			fileServer.ServeHTTP(w, r)
			return
		}

		// SPA route → serve index.html directly to avoid http.FileServer's
		// built-in redirect of /index.html → / which would lose the path.
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		content, err := fs.ReadFile(sub, "index.html")
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(content)
	})
}

func isLikelyAsset(p string) bool {
	// Covers typical Vite outputs: /assets/... or /static/...
	return strings.HasPrefix(p, "/assets/") || strings.HasPrefix(p, "/static/")
}
