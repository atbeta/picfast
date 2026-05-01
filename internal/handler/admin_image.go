package handler

import (
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/atbeta/picfast/internal/domain"
	"github.com/atbeta/picfast/internal/service"
	"github.com/atbeta/picfast/internal/sqlc"
	"github.com/go-chi/chi/v5"
)

type AdminImageHandler struct {
	db      *sqlc.Queries
	deleter *service.DeleteService
	baseURL string
}

func NewAdminImageHandler(db *sqlc.Queries, deleter *service.DeleteService, baseURL string) *AdminImageHandler {
	return &AdminImageHandler{db: db, deleter: deleter, baseURL: baseURL}
}

func (h *AdminImageHandler) List(w http.ResponseWriter, r *http.Request) {
	page, pageSize := parsePagination(r)

	rows, err := h.db.ListAllImages(r.Context(), sqlc.ListAllImagesParams{
		Limit:  pageSize,
		Offset: (page - 1) * pageSize,
	})
	if err != nil {
		Fail(w, http.StatusInternalServerError, "failed to list images")
		return
	}

	keyword := r.URL.Query().Get("keyword")
	email := r.URL.Query().Get("email")
	ext := r.URL.Query().Get("extension")

	if keyword != "" || email != "" || ext != "" {
		filtered := make([]sqlc.ListAllImagesRow, 0)
		for _, img := range rows {
			if keyword != "" && !strings.Contains(strings.ToLower(img.OriginName), strings.ToLower(keyword)) {
				continue
			}
			if email != "" && !strings.Contains(strings.ToLower(img.UserEmail.String), strings.ToLower(email)) {
				continue
			}
			if ext != "" && img.Extension != ext {
				continue
			}
			filtered = append(filtered, img)
		}
		rows = filtered
	}

	total, err := h.db.CountAllImages(r.Context())
	if err != nil {
		slog.Warn("failed to count all images", "error", err)
		total = 0
	}

	type imageItem struct {
		sqlc.ListAllImagesRow
		URL          string            `json:"url"`
		ThumbnailURL string            `json:"thumbnail_url"`
		Links        domain.ImageLinks `json:"links"`
	}

	items := make([]imageItem, len(rows))
	linkBuilder := service.LinkBuilder{BaseURL: h.baseURL}
	for i, img := range rows {
		links := linkBuilder.BuildImageLinks(img.Key, img.Extension, img.Md5, img.OriginName)
		items[i] = imageItem{
			ListAllImagesRow: img,
			URL:              links.URL,
			ThumbnailURL:     links.ThumbnailURL,
			Links:            links,
		}
	}

	Paginated(w, items, total, page, pageSize)
}

func (h *AdminImageHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		Fail(w, http.StatusBadRequest, "invalid id")
		return
	}

	img, err := h.db.GetImageByID(r.Context(), id)
	if err != nil {
		slog.Warn("failed to load image before admin delete", "error", err, "image_id", id)
	}
	if err := h.deleter.DeleteImage(r.Context(), id); err != nil {
		Fail(w, http.StatusInternalServerError, "failed to delete image")
		return
	}
	writeAuditLog(h.db, r, "admin.image.delete", "image", strconv.FormatInt(id, 10), img.OriginName, nil)

	SuccessMessage(w, "deleted")
}
