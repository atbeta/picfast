package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/pbeta/imgapi/internal/domain"
	"github.com/pbeta/imgapi/internal/sqlc"
)

type AlbumHandler struct {
	db *sqlc.Queries
}

func NewAlbumHandler(db *sqlc.Queries) *AlbumHandler {
	return &AlbumHandler{db: db}
}

func (h *AlbumHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(domain.ContextKeyUserID).(int64)
	if !ok {
		Fail(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	page, pageSize := parsePagination(r)

	albums, err := h.db.ListAlbumsByUser(r.Context(), sqlc.ListAlbumsByUserParams{
		UserID: userID,
		Limit:  pageSize,
		Offset: (page - 1) * pageSize,
	})
	if err != nil {
		Fail(w, http.StatusInternalServerError, "failed to list albums")
		return
	}

	total, _ := h.db.CountAlbumsByUser(r.Context(), sqlc.CountAlbumsByUserParams{
		UserID: userID,
	})

	Paginated(w, albums, total, page, pageSize)
}

type createAlbumRequest struct {
	Name  string `json:"name"`
	Intro string `json:"intro"`
}

func (h *AlbumHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(domain.ContextKeyUserID).(int64)
	if !ok {
		Fail(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req createAlbumRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Fail(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" {
		Fail(w, http.StatusBadRequest, "name is required")
		return
	}

	album, err := h.db.CreateAlbum(r.Context(), sqlc.CreateAlbumParams{
		UserID: userID,
		Name:   req.Name,
		Intro:  req.Intro,
	})
	if err != nil {
		Fail(w, http.StatusInternalServerError, "failed to create album")
		return
	}

	h.db.IncrementUserAlbumNum(r.Context(), userID)

	Created(w, map[string]interface{}{
		"id":         album.ID,
		"name":       album.Name,
		"intro":      album.Intro,
		"image_num":  album.ImageNum,
		"created_at": album.CreatedAt,
	})
}

type updateAlbumRequest struct {
	Name  *string `json:"name"`
	Intro *string `json:"intro"`
}

func (h *AlbumHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(domain.ContextKeyUserID).(int64)
	if !ok {
		Fail(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		Fail(w, http.StatusBadRequest, "invalid id")
		return
	}

	album, err := h.db.GetAlbumByID(r.Context(), id)
	if err != nil {
		Fail(w, http.StatusNotFound, "album not found")
		return
	}

	if album.UserID != userID {
		Fail(w, http.StatusForbidden, "not your album")
		return
	}

	var req updateAlbumRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Fail(w, http.StatusBadRequest, "invalid request body")
		return
	}

	name := album.Name
	intro := album.Intro
	if req.Name != nil {
		name = *req.Name
	}
	if req.Intro != nil {
		intro = *req.Intro
	}

	updated, err := h.db.UpdateAlbum(r.Context(), sqlc.UpdateAlbumParams{
		ID:    id,
		Name:  name,
		Intro: intro,
	})
	if err != nil {
		Fail(w, http.StatusInternalServerError, "failed to update album")
		return
	}

	Success(w, map[string]interface{}{
		"id":         updated.ID,
		"name":       updated.Name,
		"intro":      updated.Intro,
		"image_num":  updated.ImageNum,
		"updated_at": updated.UpdatedAt,
	})
}

func (h *AlbumHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(domain.ContextKeyUserID).(int64)
	if !ok {
		Fail(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		Fail(w, http.StatusBadRequest, "invalid id")
		return
	}

	album, err := h.db.GetAlbumByID(r.Context(), id)
	if err != nil {
		Fail(w, http.StatusNotFound, "album not found")
		return
	}

	if album.UserID != userID {
		Fail(w, http.StatusForbidden, "not your album")
		return
	}

	if err := h.db.DeleteAlbum(r.Context(), id); err != nil {
		Fail(w, http.StatusInternalServerError, "failed to delete album")
		return
	}

	h.db.DecrementUserAlbumNum(r.Context(), userID)

	SuccessMessage(w, "deleted")
}
