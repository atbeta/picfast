package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/atbeta/picfast/internal/domain"
	"github.com/atbeta/picfast/internal/service/moderation"
	"github.com/atbeta/picfast/internal/sqlc"
	"github.com/go-chi/chi/v5"
)

type ModerationHandler struct {
	db *sqlc.Queries
}

func NewModerationHandler(db *sqlc.Queries) *ModerationHandler {
	return &ModerationHandler{db: db}
}

// ListPending returns images awaiting moderation review.
func (h *ModerationHandler) ListPending(w http.ResponseWriter, r *http.Request) {
	page, pageSize := parsePagination(r)

	images, err := h.db.ListPendingImages(r.Context(), sqlc.ListPendingImagesParams{
		Limit:  pageSize,
		Offset: (page - 1) * pageSize,
	})
	if err != nil {
		Fail(w, http.StatusInternalServerError, "failed to list pending images")
		return
	}

	total, _ := h.db.CountPendingImages(r.Context())

	items := make([]ImageListItem, len(images))
	for i, img := range images {
		items[i] = ImageListItem{
			ID:         img.ID,
			Key:        img.Key,
			OriginName: img.OriginName,
			SizeBytes:  img.SizeBytes,
			Mimetype:   img.Mimetype,
			Extension:  img.Extension,
			Width:      img.Width,
			Height:     img.Height,
			Permission: img.Permission,
			AlbumID:    domain.PgInt8PtrVal(img.AlbumID),
			CreatedAt:  img.CreatedAt,
		}
	}

	Paginated(w, items, total, page, pageSize)
}

// Approve approves a pending image.
func (h *ModerationHandler) Approve(w http.ResponseWriter, r *http.Request) {
	adminID, ok := r.Context().Value(domain.ContextKeyUserID).(int64)
	if !ok {
		Fail(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		Fail(w, http.StatusBadRequest, "invalid image id")
		return
	}

	if err := moderation.UpdateImageModeration(r.Context(), h.db, id, moderation.StatusApproved, adminID, "approved by admin"); err != nil {
		Fail(w, http.StatusInternalServerError, "failed to approve image")
		return
	}

	SuccessMessage(w, "image approved")
}

// Reject rejects a pending image.
func (h *ModerationHandler) Reject(w http.ResponseWriter, r *http.Request) {
	adminID, ok := r.Context().Value(domain.ContextKeyUserID).(int64)
	if !ok {
		Fail(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		Fail(w, http.StatusBadRequest, "invalid image id")
		return
	}

	var reqBody struct {
		Reason string `json:"reason"`
	}
	if err := decodeJSON(r, &reqBody); err != nil {
		// reason is optional
	}

	if err := moderation.UpdateImageModeration(r.Context(), h.db, id, moderation.StatusRejected, adminID, reqBody.Reason); err != nil {
		Fail(w, http.StatusInternalServerError, "failed to reject image")
		return
	}

	SuccessMessage(w, "image rejected")
}

func decodeJSON(r *http.Request, v any) error {
	return json.NewDecoder(r.Body).Decode(v)
}
