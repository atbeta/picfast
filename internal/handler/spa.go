package handler

import (
	"io/fs"
	"net/http"
	"path/filepath"
	"strings"
)

type SPAHandler struct {
	staticFS   fs.FS
	fileServer http.Handler
	index      []byte
}

func NewSPAHandler(staticFS fs.FS) *SPAHandler {
	idx, _ := fs.ReadFile(staticFS, "index.html")
	return &SPAHandler{
		staticFS:   staticFS,
		fileServer: http.FileServer(http.FS(staticFS)),
		index:      idx,
	}
}

func (h *SPAHandler) Serve(w http.ResponseWriter, r *http.Request) {
	clean := strings.TrimPrefix(r.URL.Path, "/")
	if clean == "" {
		clean = "index.html"
	}

	// If the path has a file extension, try serving it directly
	if ext := filepath.Ext(clean); ext != "" {
		if f, err := h.staticFS.Open(clean); err == nil {
			f.Close()
			// Vite fingerprints everything in /assets/ by content hash, so a
			// long-lived immutable cache is safe and avoids re-downloads.
			// Other top-level static files (favicon, icons) only need
			// revalidation, not aggressive caching.
			if strings.HasPrefix(clean, "assets/") {
				w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
			} else {
				w.Header().Set("Cache-Control", "no-cache")
			}
			h.fileServer.ServeHTTP(w, r)
			return
		}
	}

	// SPA fallback: serve index.html. Always revalidate so deploys pick up
	// the new asset hashes without users needing to force-refresh.
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(h.index)
}
