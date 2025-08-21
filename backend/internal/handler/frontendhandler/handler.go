package frontendhandler

import (
	"io/fs"
	"net/http"
	"strings"

	"github.com/auto-dns/pihole-cluster-admin/internal/frontend"
	"github.com/go-chi/chi"
	"github.com/rs/zerolog"
)

type Handler struct {
	logger zerolog.Logger
}

func NewHandler(logger zerolog.Logger) *Handler {
	return &Handler{logger: logger}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()

	sub, err := fs.Sub(frontend.Files, "internal/frontend/dist")
	if err != nil {
		h.logger.Warn().Msg("No embedded frontend found; skipping static file routes")
		return r
	}

	// Serve all frontend paths with SPA fallback
	fileServer := http.FileServer(http.FS(sub))
	r.Handle("/*", spaHandler(sub, fileServer, h.logger))

	return r
}

func spaHandler(sub fs.FS, fileServer http.Handler, logger zerolog.Logger) http.Handler {
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

		// SPA route → rewrite to /index.html and serve via FileServer
		// (avoid ServeContent/ReadSeeker issues entirely)
		w.Header().Set("Cache-Control", "no-cache")
		r2 := r.Clone(r.Context())
		r2.URL.Path = "/index.html"
		fileServer.ServeHTTP(w, r2)
	})
}

func isLikelyAsset(p string) bool {
	// Covers typical Vite outputs: /assets/... or /static/...
	return strings.HasPrefix(p, "/assets/") || strings.HasPrefix(p, "/static/")
}
