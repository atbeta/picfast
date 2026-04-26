package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/atbeta/picfast/internal/domain"
	"github.com/atbeta/picfast/internal/service"
	"github.com/atbeta/picfast/internal/sqlc"
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

	total, _ := h.db.CountAllImages(r.Context())

	type imageItem struct {
		sqlc.ListAllImagesRow
		URL          string            `json:"url"`
		ThumbnailURL string            `json:"thumbnail_url"`
		Links        domain.ImageLinks `json:"links"`
	}

	items := make([]imageItem, len(rows))
	for i, img := range rows {
		url := fmt.Sprintf("%s/i/%s.%s", h.baseURL, img.Key, img.Extension)
		thumbURL := fmt.Sprintf("%s/t/%s.png", h.baseURL, img.Md5)
		items[i] = imageItem{
			ListAllImagesRow: img,
			URL:              url,
			ThumbnailURL:     thumbURL,
			Links: domain.ImageLinks{
				URL:          url,
				HTML:         fmt.Sprintf(`<img src="%s" alt="%s" />`, url, img.OriginName),
				BBCode:       fmt.Sprintf("[img]%s[/img]", url),
				Markdown:     fmt.Sprintf("![%s](%s)", img.OriginName, url),
				ThumbnailURL: thumbURL,
			},
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

	if err := h.deleter.DeleteImage(r.Context(), id); err != nil {
		Fail(w, http.StatusInternalServerError, "failed to delete image")
		return
	}

	SuccessMessage(w, "deleted")
}
