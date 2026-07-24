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
			if strings.HasPrefix(clean, "assets/") {
				w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
			} else if clean == "index.html" {
				w.Header().Set("Cache-Control", "no-store")
			} else {
				w.Header().Set("Cache-Control", "no-cache")
			}
			h.fileServer.ServeHTTP(w, r)
			return
		}
	}

	// SPA fallback: serve index.html. no-store prevents the browser from
	// using a stale index.html that references now-deleted JS chunks after a
	// deploy. The hashed /assets/ files are safe to cache immutably.
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(h.index)
}
