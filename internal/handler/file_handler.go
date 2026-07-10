package handler

import (
	"log/slog"
	"net/http"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/atbeta/picfast/internal/domain"
	picmetrics "github.com/atbeta/picfast/internal/metrics"
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
	start := time.Now()
	result := "error"
	defer func() {
		picmetrics.ObserveImageServe("image", result, time.Since(start))
	}()

	key := chi.URLParam(r, "key")
	ext, params := parseProcessingParams(chi.URLParam(r, "ext"))

	img, err := h.db.GetImageByKey(r.Context(), key)
	if err != nil {
		result = "not_found"
		http.NotFound(w, r)
		return
	}

	// Extension must match
	if img.Extension != ext {
		result = "not_found"
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
			result = "denied"
			http.NotFound(w, r)
			return
		}
	}

	// Permission check
	if img.Permission == int16(domain.PermissionPrivate) {
		userID, ok := r.Context().Value(domain.ContextKeyUserID).(int64)
		role := domain.RoleUser
		if rVal, rOk := r.Context().Value(domain.ContextKeyRole).(domain.UserRole); rOk {
			role = rVal
		}
		if !ok || (img.UserID.Int64 != userID && role != domain.RoleAdmin) {
			result = "denied"
			slog.Info("private image access denied",
				"key", img.Key,
				"path", r.URL.Path,
				"authed", ok,
				"request_user_id", userID,
				"owner_user_id", img.UserID.Int64,
				"role", role,
			)
			http.NotFound(w, r)
			return
		}
	}

	// Load strategy and group config
	strategy, err := h.db.GetStrategyByID(r.Context(), img.StrategyID.Int64)
	if err != nil {
		result = "error"
		http.Error(w, "storage error", http.StatusInternalServerError)
		return
	}

	store, err := service.GetStorageForStrategy(strategy)
	if err != nil {
		result = "error"
		http.Error(w, "storage error", http.StatusInternalServerError)
		return
	}
	defer store.Close()

	pathname := img.Name
	if img.Path != "" && img.Path != "." {
		pathname = img.Path + "/" + img.Name
	}

	if strategy.StrategyType == "webdav" && service.IsDirectLinkMode(strategy.Configs) {
		if wds, ok := store.(interface{ HasPublicURL() bool }); ok && wds.HasPublicURL() {
			if img.Permission == int16(domain.PermissionPublic) && params.IsEmpty() {
				w.Header().Set("Cache-Control", "public, max-age=2592000")
				http.Redirect(w, r, store.URL(pathname), http.StatusFound)
				result = "success"
				return
			}
		}
	}

	data, err := store.Read(r.Context(), pathname)
	if err != nil {
		result = "not_found"
		http.NotFound(w, r)
		return
	}

	processed := false
	if !params.IsEmpty() {
		data, processed = service.ProcessImageOnTheFly(data, img.Extension, params)
	}

	// Determine content type: use output format only when processing succeeded,
	// otherwise the data may still be the original and the MIME must match.
	mimetype := img.Mimetype
	if processed && params.Format != "" {
		if ct := service.MimeTypeForFormat(params.Format); ct != "" {
			mimetype = ct
		}
	}

	// Cache headers: 30 days
	w.Header().Set("Content-Type", mimetype)
	if img.Permission == int16(domain.PermissionPrivate) {
		w.Header().Set("Cache-Control", "private, no-cache, no-store, must-revalidate")
	} else {
		w.Header().Set("Cache-Control", "public, max-age=2592000, s-maxage=60")
	}
	if params.IsEmpty() {
		w.Header().Set("ETag", `"`+img.Md5+`"`)
	}

	// Handle conditional request
	if params.IsEmpty() {
		if match := r.Header.Get("If-None-Match"); match != "" {
			if strings.Contains(match, img.Md5) {
				result = "success"
				w.WriteHeader(http.StatusNotModified)
				return
			}
		}
	}

	result = "success"
	w.Write(data)
}

// parseProcessingParams extracts the file extension and image processing
// parameters from the URL ext segment.  The format is:
//
//	{ext}[@w_{width},h_{height},q_{quality},f_{format}]
//
// Examples:
//
//	"jpg"                → ext="jpg", no processing
//	"jpg@w_300"          → ext="jpg", resize width 300
//	"png@w_800,h_600"    → ext="png", fit within 800×600
//	"jpg@q_80,f_webp"    → ext="jpg", quality 80, output webp
//	"jpg@w_300,q_60,f_webp" → all combined
func parseProcessingParams(raw string) (ext string, params service.ProcessingParams) {
	baseExt, variant, hasVariant := strings.Cut(strings.ToLower(strings.TrimSpace(raw)), "@")
	ext = strings.TrimPrefix(baseExt, ".")
	if !hasVariant || variant == "" {
		return ext, params
	}

	for _, part := range strings.Split(variant, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		k, v, ok := strings.Cut(part, "_")
		if !ok || v == "" {
			continue
		}
		switch k {
		case "w":
			if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 10000 {
				params.Width = n
			}
		case "h":
			if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 10000 {
				params.Height = n
			}
		case "q":
			if n, err := strconv.Atoi(v); err == nil && n >= 1 && n <= 100 {
				params.Quality = n
			}
		case "f":
			f := strings.ToLower(v)
			if f == "jpg" {
				f = "jpeg"
			}
			if f == "jpeg" || f == "png" || f == "webp" || f == "gif" {
				params.Format = f
			}
		}
	}

	return ext, params
}

func (h *FileHandler) ServeThumbnail(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	result := "error"
	defer func() {
		picmetrics.ObserveImageServe("thumbnail", result, time.Since(start))
	}()

	md5Hash := chi.URLParam(r, "hash")
	if !md5HashRegex.MatchString(md5Hash) {
		result = "not_found"
		http.NotFound(w, r)
		return
	}

	images, err := h.db.GetImagesByMD5(r.Context(), md5Hash)
	if err != nil || len(images) == 0 {
		result = "not_found"
		http.NotFound(w, r)
		return
	}

	// Permission check: if ALL images with this MD5 are private, check if user owns at least one
	isPublic := false
	for _, img := range images {
		if img.Permission == int16(domain.PermissionPublic) {
			isPublic = true
			break
		}
	}

	if !isPublic {
		userID, ok := r.Context().Value(domain.ContextKeyUserID).(int64)
		role := domain.RoleUser
		if rVal, rOk := r.Context().Value(domain.ContextKeyRole).(domain.UserRole); rOk {
			role = rVal
		}
		if !ok {
			result = "denied"
			slog.Info("private thumbnail access denied",
				"hash", md5Hash,
				"path", r.URL.Path,
				"authed", false,
			)
			http.NotFound(w, r)
			return
		}

		ownsOne := false
		for _, img := range images {
			if img.UserID.Int64 == userID {
				ownsOne = true
				break
			}
		}
		if !ownsOne && role != domain.RoleAdmin {
			result = "denied"
			slog.Info("private thumbnail access denied",
				"hash", md5Hash,
				"path", r.URL.Path,
				"authed", true,
				"request_user_id", userID,
				"role", role,
				"reason", "not_owner",
			)
			http.NotFound(w, r)
			return
		}
	}

	if !isPublic {
		w.Header().Set("Cache-Control", "private, no-cache, no-store, must-revalidate")
	} else {
		w.Header().Set("Cache-Control", "public, max-age=2592000, s-maxage=60")
	}

	filePath := filepath.Join(h.thumbDir, md5Hash+".png")
	ww := &serveStatusWriter{ResponseWriter: w, statusCode: http.StatusOK}
	http.ServeFile(ww, r, filePath)
	if ww.statusCode == http.StatusNotFound {
		result = "not_found"
		return
	}
	if ww.statusCode >= http.StatusInternalServerError {
		result = "error"
		return
	}
	result = "success"
}

type serveStatusWriter struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

func (w *serveStatusWriter) WriteHeader(code int) {
	if !w.written {
		w.statusCode = code
		w.written = true
	}
	w.ResponseWriter.WriteHeader(code)
}

func (w *serveStatusWriter) Write(b []byte) (int, error) {
	if !w.written {
		w.written = true
	}
	return w.ResponseWriter.Write(b)
}
