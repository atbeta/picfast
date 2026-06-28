package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/atbeta/picfast/internal/domain"
	"github.com/atbeta/picfast/internal/events"
	"github.com/atbeta/picfast/internal/service"
	"github.com/atbeta/picfast/internal/service/moderation"
	"github.com/atbeta/picfast/internal/sqlc"
	"github.com/go-chi/chi/v5"
)

type ModerationHandler struct {
	db      *sqlc.Queries
	baseURL string
	emitter events.Emitter
}

func NewModerationHandler(db *sqlc.Queries, baseURL string, emitter events.Emitter) *ModerationHandler {
	return &ModerationHandler{db: db, baseURL: baseURL, emitter: emitter}
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

	total, err := h.db.CountPendingImages(r.Context())
	if err != nil {
		slog.Warn("failed to count pending images", "error", err)
		total = 0
	}

	items := make([]ImageListItem, len(images))
	linkBuilder := service.LinkBuilder{BaseURL: h.baseURL}
	for i, img := range images {
		links := linkBuilder.BuildImageLinks(img.Key, img.Extension, img.Md5, img.OriginName)
		items[i] = ImageListItem{
			ID:               img.ID,
			Key:              img.Key,
			OriginName:       img.OriginName,
			SizeBytes:        img.SizeBytes,
			Mimetype:         img.Mimetype,
			Extension:        img.Extension,
			Width:            img.Width,
			Height:           img.Height,
			Permission:       img.Permission,
			AlbumID:          domain.PgInt8PtrVal(img.AlbumID),
			URL:              links.URL,
			ThumbnailURL:     links.ThumbnailURL,
			Links:            links,
			ModerationStatus: img.ModerationStatus,
			CreatedAt:        img.CreatedAt,
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

	writeAuditLog(h.db, r, "admin.moderation.approve", "image", idStr, "", map[string]any{
		"moderation_status": "approved",
	})

	img, err := h.db.GetImageByID(r.Context(), id)
	if err != nil {
		slog.Warn("failed to fetch image for moderation event", "image_id", id, "error", err)
	} else {
		ev := events.BuildModerationReviewed(id, img.Key, string(moderation.StatusApproved), adminID, "approved by admin", domain.PgInt8PtrVal(img.UserID))
		ev.Actor = events.AdminActor(adminID)
		events.EmitAsync(h.emitter, ev)
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

	details := map[string]any{"moderation_status": "rejected"}
	if reqBody.Reason != "" {
		details["reason"] = reqBody.Reason
	}
	writeAuditLog(h.db, r, "admin.moderation.reject", "image", idStr, "", details)

	img, err := h.db.GetImageByID(r.Context(), id)
	if err != nil {
		slog.Warn("failed to fetch image for moderation event", "image_id", id, "error", err)
	} else {
		ev := events.BuildModerationReviewed(id, img.Key, string(moderation.StatusRejected), adminID, reqBody.Reason, domain.PgInt8PtrVal(img.UserID))
		ev.Actor = events.AdminActor(adminID)
		events.EmitAsync(h.emitter, ev)
	}

	SuccessMessage(w, "image rejected")
}

func decodeJSON(r *http.Request, v any) error {
	return json.NewDecoder(r.Body).Decode(v)
}
