package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/pbeta/imgapi/internal/domain"
	"github.com/pbeta/imgapi/internal/service"
	"github.com/pbeta/imgapi/internal/sqlc"
)

type ImageHandler struct {
	db      *sqlc.Queries
	upload  *service.UploadService
	deleter *service.DeleteService
	baseURL string
}

func NewImageHandler(db *sqlc.Queries, upload *service.UploadService, deleter *service.DeleteService, baseURL string) *ImageHandler {
	return &ImageHandler{db: db, upload: upload, deleter: deleter, baseURL: baseURL}
}

func (h *ImageHandler) Upload(w http.ResponseWriter, r *http.Request) {
	// 32MB max memory for multipart
	r.Body = http.MaxBytesReader(w, r.Body, 50<<20)

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		Fail(w, http.StatusBadRequest, "failed to parse multipart form")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		Fail(w, http.StatusBadRequest, "file is required")
		return
	}
	defer file.Close()

	fileData, err := io.ReadAll(file)
	if err != nil {
		Fail(w, http.StatusInternalServerError, "failed to read file")
		return
	}

	// Resolve optional auth
	var userID *int64
	if uid, ok := r.Context().Value(domain.ContextKeyUserID).(int64); ok {
		userID = &uid
	}

	// Optional params
	var strategyID *int64
	if sid := r.FormValue("strategy_id"); sid != "" {
		if v, err := strconv.ParseInt(sid, 10, 64); err == nil {
			strategyID = &v
		}
	}

	var albumID *int64
	if aid := r.FormValue("album_id"); aid != "" {
		if v, err := strconv.ParseInt(aid, 10, 64); err == nil {
			albumID = &v
		}
	}

	var perm *int16
	if p := r.FormValue("permission"); p != "" {
		if v, err := strconv.ParseInt(p, 10, 16); err == nil {
			sv := int16(v)
			perm = &sv
		}
	}

	result, err := h.upload.Store(r.Context(), service.UploadParams{
		FileData:   fileData,
		FileName:   header.Filename,
		FileSize:   header.Size,
		StrategyID: strategyID,
		AlbumID:    albumID,
		Permission: perm,
		UserID:     userID,
		ClientIP:   r.RemoteAddr,
	})
	if err != nil {
		Fail(w, http.StatusBadRequest, err.Error())
		return
	}

	Created(w, imageResponse(result.Image, result.Links))
}

func (h *ImageHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(domain.ContextKeyUserID).(int64)
	if !ok {
		Fail(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	page, pageSize := parsePagination(r)

	images, err := h.db.ListImagesByUser(r.Context(), sqlc.ListImagesByUserParams{
		UserID: domain.PgInt8(userID),
		Limit:  pageSize,
		Offset: (page - 1) * pageSize,
	})
	if err != nil {
		Fail(w, http.StatusInternalServerError, "failed to list images")
		return
	}

	total, _ := h.db.CountImagesByUser(r.Context(), domain.PgInt8(userID))

	items := make([]ImageListItem, len(images))
	for i, img := range images {
		links := h.buildLinks(img)
		items[i] = ImageListItem{
			ID:           img.ID,
			Key:          img.Key,
			OriginName:   img.OriginName,
			SizeBytes:    img.SizeBytes,
			Mimetype:     img.Mimetype,
			Extension:    img.Extension,
			Width:        img.Width,
			Height:       img.Height,
			Permission:   img.Permission,
			AlbumID:      domain.PgInt8PtrVal(img.AlbumID),
			URL:          links.URL,
			ThumbnailURL: links.ThumbnailURL,
			Links:        links,
			CreatedAt:    img.CreatedAt,
		}
	}

	Paginated(w, items, total, page, pageSize)
}

func (h *ImageHandler) Get(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")

	img, err := h.db.GetImageByKey(r.Context(), key)
	if err != nil {
		Fail(w, http.StatusNotFound, "image not found")
		return
	}

	// Check permission
	if img.Permission == int16(domain.PermissionPrivate) {
		userID, ok := r.Context().Value(domain.ContextKeyUserID).(int64)
		if !ok || img.UserID.Int64 != userID {
			Fail(w, http.StatusNotFound, "image not found")
			return
		}
	}

	links := h.buildLinks(img)
	Success(w, imageResponse(img, links))
}

func (h *ImageHandler) Delete(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")

	userID, ok := r.Context().Value(domain.ContextKeyUserID).(int64)
	if !ok {
		Fail(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	img, err := h.db.GetImageByKey(r.Context(), key)
	if err != nil {
		Fail(w, http.StatusNotFound, "image not found")
		return
	}

	if img.UserID.Int64 != userID {
		Fail(w, http.StatusForbidden, "not your image")
		return
	}

	if err := h.deleter.DeleteImage(r.Context(), img.ID); err != nil {
		Fail(w, http.StatusInternalServerError, "failed to delete image")
		return
	}

	SuccessMessage(w, "deleted")
}

type updateImageRequest struct {
	AlbumID    *int64 `json:"album_id"`
	Permission *int16 `json:"permission"`
}

func (h *ImageHandler) Update(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")

	userID, ok := r.Context().Value(domain.ContextKeyUserID).(int64)
	if !ok {
		Fail(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	img, err := h.db.GetImageByKey(r.Context(), key)
	if err != nil {
		Fail(w, http.StatusNotFound, "image not found")
		return
	}

	if img.UserID.Int64 != userID {
		Fail(w, http.StatusForbidden, "not your image")
		return
	}

	var req updateImageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Fail(w, http.StatusBadRequest, "invalid request body")
		return
	}

	albumID := img.AlbumID
	perm := img.Permission
	if req.AlbumID != nil {
		albumID = domain.PgInt8(*req.AlbumID)
	}
	if req.Permission != nil {
		perm = *req.Permission
	}

	updated, err := h.db.UpdateImage(r.Context(), sqlc.UpdateImageParams{
		ID:         img.ID,
		AlbumID:    albumID,
		Permission: perm,
	})
	if err != nil {
		Fail(w, http.StatusInternalServerError, "failed to update image")
		return
	}

	links := h.buildLinks(updated)
	Success(w, imageResponse(updated, links))
}

func (h *ImageHandler) buildLinks(img sqlc.Image) domain.ImageLinks {
	url := h.baseURL + "/i/" + img.Key + "." + img.Extension
	thumbURL := h.baseURL + "/t/" + img.Md5 + ".png"

	return domain.ImageLinks{
		URL:          url,
		HTML:         fmt.Sprintf(`<img src="%s" alt="%s" />`, url, img.OriginName),
		BBCode:       fmt.Sprintf("[img]%s[/img]", url),
		Markdown:     fmt.Sprintf("![%s](%s)", img.OriginName, url),
		ThumbnailURL: thumbURL,
	}
}

func imageResponse(img sqlc.Image, links domain.ImageLinks) ImageResponse {
	return ImageResponse{
		ID:         img.ID,
		Key:        img.Key,
		OriginName: img.OriginName,
		SizeBytes:  img.SizeBytes,
		Mimetype:   img.Mimetype,
		Extension:  img.Extension,
		Width:      img.Width,
		Height:     img.Height,
		Md5:        img.Md5,
		Sha1:       img.Sha1,
		Permission: img.Permission,
		AlbumID:    domain.PgInt8PtrVal(img.AlbumID),
		Links:      links,
		CreatedAt:  img.CreatedAt,
	}
}
