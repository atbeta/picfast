package handler

import (
	"net/http"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/atbeta/picfast/internal/domain"
	"github.com/atbeta/picfast/internal/service"
	"github.com/atbeta/picfast/internal/sqlc"
	"github.com/go-chi/chi/v5"
)

var md5HashRegex = regexp.MustCompile(`^[a-f0-9]{32}$`)

type FileHandler struct {
	db       *sqlc.Queries
	baseURL  string
	thumbDir string
}

func NewFileHandler(db *sqlc.Queries, baseURL, thumbDir string) *FileHandler {
	return &FileHandler{db: db, baseURL: baseURL, thumbDir: thumbDir}
}

func (h *FileHandler) ServeImage(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	ext := strings.TrimPrefix(filepath.Ext(r.URL.Path), ".")

	img, err := h.db.GetImageByKey(r.Context(), key)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// Extension must match
	if img.Extension != ext {
		http.NotFound(w, r)
		return
	}

	// Moderation check: pending/rejected images are only visible to owner or admin
	if img.ModerationStatus != "approved" && img.ModerationStatus != "" {
		userID, ok := r.Context().Value(domain.ContextKeyUserID).(int64)
		role := domain.RoleUser
		if rVal, rOk := r.Context().Value(domain.ContextKeyRole).(domain.UserRole); rOk {
			role = rVal
		}
		if !ok || (img.UserID.Int64 != userID && role != domain.RoleAdmin) {
			http.NotFound(w, r)
			return
		}
	}

	// Permission check
	if img.Permission == int16(domain.PermissionPrivate) {
		userID, ok := r.Context().Value(domain.ContextKeyUserID).(int64)
		if !ok || img.UserID.Int64 != userID {
			http.NotFound(w, r)
			return
		}
	}

	// Load strategy and group config
	strategy, err := h.db.GetStrategyByID(r.Context(), img.StrategyID.Int64)
	if err != nil {
		http.Error(w, "storage error", http.StatusInternalServerError)
		return
	}

	store, err := service.GetStorageForStrategy(strategy)
	if err != nil {
		http.Error(w, "storage error", http.StatusInternalServerError)
		return
	}
	defer store.Close()

	pathname := img.Name
	if img.Path != "" && img.Path != "." {
		pathname = img.Path + "/" + img.Name
	}
	data, err := store.Read(r.Context(), pathname)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// Cache headers: 30 days
	w.Header().Set("Content-Type", img.Mimetype)
	w.Header().Set("Cache-Control", "public, max-age=2592000")
	w.Header().Set("ETag", `"`+img.Md5+`"`)

	// Handle conditional request
	if match := r.Header.Get("If-None-Match"); match != "" {
		if strings.Contains(match, img.Md5) {
			w.WriteHeader(http.StatusNotModified)
			return
		}
	}

	w.Write(data)
}

func (h *FileHandler) ServeThumbnail(w http.ResponseWriter, r *http.Request) {
	md5Hash := chi.URLParam(r, "hash")
	if !md5HashRegex.MatchString(md5Hash) {
		http.NotFound(w, r)
		return
	}

	filePath := filepath.Join(h.thumbDir, md5Hash+".png")
	http.ServeFile(w, r, filePath)
}
