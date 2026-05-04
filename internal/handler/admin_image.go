package handler

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

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

	var dateFrom, dateTo *time.Time
	if df := r.URL.Query().Get("date_from"); df != "" {
		if t, err := time.Parse(time.RFC3339, df); err == nil {
			dateFrom = &t
		}
	}
	if dt := r.URL.Query().Get("date_to"); dt != "" {
		if t, err := time.Parse(time.RFC3339, dt); err == nil {
			dateTo = &t
		}
	}

	params := sqlc.ListAllImagesParams{
		Limit:     pageSize,
		Offset:    (page - 1) * pageSize,
		Keyword:   domain.PgTextNonEmpty(r.URL.Query().Get("keyword")),
		Email:     domain.PgTextNonEmpty(r.URL.Query().Get("email")),
		Extension: domain.PgTextNonEmpty(r.URL.Query().Get("extension")),
		DateFrom:  domain.PgTimeWithZonePtr(dateFrom),
		DateTo:    domain.PgTimeWithZonePtr(dateTo),
	}
	rows, err := h.db.ListAllImages(r.Context(), params)
	if err != nil {
		Fail(w, http.StatusInternalServerError, "failed to list images")
		return
	}

	total, err := h.db.CountAllImages(r.Context(), sqlc.CountAllImagesParams{
		Keyword:   params.Keyword,
		Email:     params.Email,
		Extension: params.Extension,
		DateFrom:  params.DateFrom,
		DateTo:    params.DateTo,
	})
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
