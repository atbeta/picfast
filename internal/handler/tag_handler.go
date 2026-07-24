package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/atbeta/picfast/internal/domain"
	"github.com/atbeta/picfast/internal/sqlc"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type TagHandler struct {
	db *sqlc.Queries
}

func NewTagHandler(db *sqlc.Queries) *TagHandler {
	return &TagHandler{db: db}
}

func (h *TagHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(domain.ContextKeyUserID).(int64)
	if !ok {
		Fail(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	tags, err := h.db.GetTagsByUserID(r.Context(), pgtype.Int8{Int64: userID, Valid: true})
	if err != nil {
		Fail(w, http.StatusInternalServerError, "failed to list tags")
		return
	}

	Success(w, tags)
}

func (h *TagHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(domain.ContextKeyUserID).(int64)
	if !ok {
		Fail(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req struct {
		Name string `json:"name"`
		Type string `json:"type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Fail(w, http.StatusBadRequest, "invalid request")
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		Fail(w, http.StatusBadRequest, "tag name is required")
		return
	}
	if len(req.Name) > 255 {
		Fail(w, http.StatusBadRequest, "tag name too long")
		return
	}
	req.Type = strings.TrimSpace(req.Type)
	if req.Type == "" {
		req.Type = "user"
	}

	tag, err := h.db.CreateTag(r.Context(), sqlc.CreateTagParams{
		UserID: pgtype.Int8{Int64: userID, Valid: true},
		Name:   req.Name,
		Type:   req.Type,
	})
	if err != nil {
		Fail(w, http.StatusInternalServerError, "failed to create tag")
		return
	}

	Success(w, tag)
}

func (h *TagHandler) AddToImage(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(domain.ContextKeyUserID).(int64)
	if !ok {
		Fail(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	imageID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		Fail(w, http.StatusBadRequest, "invalid image id")
		return
	}

	img, err := h.db.GetImageByID(r.Context(), imageID)
	if err != nil {
		Fail(w, http.StatusNotFound, "image not found")
		return
	}
	if !img.UserID.Valid || img.UserID.Int64 != userID {
		Fail(w, http.StatusForbidden, "not your image")
		return
	}

	var req struct {
		TagID int64 `json:"tag_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Fail(w, http.StatusBadRequest, "invalid request")
		return
	}

	tag, err := h.db.GetTagByID(r.Context(), req.TagID)
	if err != nil {
		Fail(w, http.StatusNotFound, "tag not found")
		return
	}
	if !tag.UserID.Valid || tag.UserID.Int64 != userID {
		Fail(w, http.StatusForbidden, "not your tag")
		return
	}

	if err := h.db.AddImageTag(r.Context(), sqlc.AddImageTagParams{
		ImageID: imageID,
		TagID:   req.TagID,
	}); err != nil {
		Fail(w, http.StatusInternalServerError, "failed to add tag")
		return
	}

	Success(w, nil)
}

func (h *TagHandler) RemoveFromImage(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(domain.ContextKeyUserID).(int64)
	if !ok {
		Fail(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	imageID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		Fail(w, http.StatusBadRequest, "invalid image id")
		return
	}

	img, err := h.db.GetImageByID(r.Context(), imageID)
	if err != nil {
		Fail(w, http.StatusNotFound, "image not found")
		return
	}
	if !img.UserID.Valid || img.UserID.Int64 != userID {
		Fail(w, http.StatusForbidden, "not your image")
		return
	}

	tagID, err := strconv.ParseInt(chi.URLParam(r, "tagId"), 10, 64)
	if err != nil {
		Fail(w, http.StatusBadRequest, "invalid tag id")
		return
	}

	tag, err := h.db.GetTagByID(r.Context(), tagID)
	if err != nil {
		Fail(w, http.StatusNotFound, "tag not found")
		return
	}
	if !tag.UserID.Valid || tag.UserID.Int64 != userID {
		Fail(w, http.StatusForbidden, "not your tag")
		return
	}

	if err := h.db.RemoveImageTag(r.Context(), sqlc.RemoveImageTagParams{
		ImageID: imageID,
		TagID:   tagID,
	}); err != nil {
		Fail(w, http.StatusInternalServerError, "failed to remove tag")
		return
	}

	Success(w, nil)
}

func (h *TagHandler) GetImageTags(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(domain.ContextKeyUserID).(int64)
	if !ok {
		Fail(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	imageID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		Fail(w, http.StatusBadRequest, "invalid image id")
		return
	}

	img, err := h.db.GetImageByID(r.Context(), imageID)
	if err != nil {
		Fail(w, http.StatusNotFound, "image not found")
		return
	}
	if img.Permission != int16(domain.PermissionPublic) && (!img.UserID.Valid || img.UserID.Int64 != userID) {
		Fail(w, http.StatusNotFound, "image not found")
		return
	}

	tags, err := h.db.GetImageTags(r.Context(), imageID)
	if err != nil && err != pgx.ErrNoRows {
		Fail(w, http.StatusInternalServerError, "failed to get image tags")
		return
	}
	if tags == nil {
		tags = []sqlc.Tag{}
	}

	Success(w, tags)
}
