package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/pbeta/imgapi/internal/sqlc"
)

type AdminImageHandler struct {
	db *sqlc.Queries
}

func NewAdminImageHandler(db *sqlc.Queries) *AdminImageHandler {
	return &AdminImageHandler{db: db}
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

	// Apply filters in Go
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

	Paginated(w, rows, total, page, pageSize)
}

func (h *AdminImageHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		Fail(w, http.StatusBadRequest, "invalid id")
		return
	}

	img, err := h.db.GetImageByID(r.Context(), id)
	if err != nil {
		Fail(w, http.StatusNotFound, "image not found")
		return
	}

	// TODO: delete physical file from storage + thumbnail (Phase 3)

	if err := h.db.DeleteImage(r.Context(), id); err != nil {
		Fail(w, http.StatusInternalServerError, "failed to delete image")
		return
	}

	// Decrement counters
	if img.UserID.Valid {
		h.db.DecrementUserImageNum(r.Context(), img.UserID.Int64)
	}
	if img.AlbumID.Valid {
		h.db.DecrementAlbumImageNum(r.Context(), img.AlbumID.Int64)
	}

	SuccessMessage(w, "deleted")
}
