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
			h.fileServer.ServeHTTP(w, r)
			return
		}
	}

	// SPA fallback: serve index.html
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(h.index)
}
