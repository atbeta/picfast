package handler

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/atbeta/picfast/internal/domain"
	picmetrics "github.com/atbeta/picfast/internal/metrics"
	"github.com/atbeta/picfast/internal/service"
	"github.com/atbeta/picfast/internal/sqlc"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// extractStrategyURL resolves a direct storage URL from a strategy's configs for non-local strategies.
// Returns empty string for local storage (fallback to proxy URLs).
func extractStrategyURL(strategyType string, configs []byte) string {
	if strategyType == "" || strategyType == "local" || strategyType == "webdav" {
		return ""
	}
	var raw map[string]string
	if err := json.Unmarshal(configs, &raw); err != nil {
		return ""
	}
	base := raw["url"]
	if base == "" {
		base = raw["domain"] // Kodo uses "domain"
	}
	return strings.TrimRight(base, "/")
}

const defaultMaxUploadBytes = 50 << 20 // 50 MiB

type ImageHandler struct {
	db              *sqlc.Queries
	upload          *service.UploadService
	deleter         *service.DeleteService
	baseURL         string
	auditUploadLogs bool
	maxUploadBytes  int64
}

func NewImageHandler(db *sqlc.Queries, upload *service.UploadService, deleter *service.DeleteService, baseURL string, auditUploadLogs bool, maxUploadBytes int64) *ImageHandler {
	if maxUploadBytes <= 0 {
		maxUploadBytes = defaultMaxUploadBytes
	}
	return &ImageHandler{db: db, upload: upload, deleter: deleter, baseURL: baseURL, auditUploadLogs: auditUploadLogs, maxUploadBytes: maxUploadBytes}
}

func (h *ImageHandler) Upload(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	source := uploadSource(r)
	observeUpload := func(result, reason string, bytes int64) {
		picmetrics.ObserveUpload(source, result, reason, bytes, time.Since(start))
	}

	r.Body = http.MaxBytesReader(w, r.Body, h.maxUploadBytes)

	// multipartFormMemory sets the max in-memory buffer for form parsing;
	// payloads exceeding this are spilled to temp files on disk.
	const multipartFormMemory = 32 << 20
	if err := r.ParseMultipartForm(multipartFormMemory); err != nil {
		observeUpload("error", "invalid_file", 0)
		Fail(w, http.StatusBadRequest, "failed to parse multipart form")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		observeUpload("error", "invalid_file", 0)
		Fail(w, http.StatusBadRequest, "file is required")
		return
	}
	defer file.Close()

	fileData, err := io.ReadAll(file)
	if err != nil {
		observeUpload("error", "invalid_file", 0)
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

	var expiresAt *time.Time
	if exp := r.FormValue("expires_in"); exp != "" {
		duration, err := time.ParseDuration(exp)
		if err != nil {
			observeUpload("error", "invalid_file", int64(len(fileData)))
			Fail(w, http.StatusBadRequest, "invalid expires_in format")
			return
		}
		t := time.Now().Add(duration)
		expiresAt = &t
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
		ExpiresAt:  expiresAt,
	})
	if err != nil {
		observeUpload("error", picmetrics.ClassifyUploadError(err), int64(len(fileData)))
		Fail(w, http.StatusBadRequest, err.Error())
		return
	}
	observeUpload("success", picmetrics.ReasonNone, result.OriginalSizeBytes)

	if h.auditUploadLogs {
		details := map[string]any{
			"origin_name": result.Image.OriginName,
			"size_bytes":  result.Image.SizeBytes,
		}
		if userID == nil {
			details["guest"] = true
		}
		writeAuditLog(h.db, r, "image.upload", "image", result.Image.Key, result.Image.OriginName, details)
	}

	Created(w, imageResponse(result.Image, result.Links))
}

func uploadSource(r *http.Request) string {
	if rctx := chi.RouteContext(r.Context()); rctx != nil {
		switch rctx.RoutePattern() {
		case "/api/v1/upload":
			return "web"
		case "/api/v1/images":
			return "api"
		}
	}
	return "api"
}

func (h *ImageHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(domain.ContextKeyUserID).(int64)
	if !ok {
		Fail(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	page, pageSize := parsePagination(r)

	var albumID *int64
	if aid := r.FormValue("album_id"); aid != "" {
		if v, err := strconv.ParseInt(aid, 10, 64); err == nil {
			albumID = &v
		}
	}

	var dateFrom, dateTo *time.Time
	if df := r.FormValue("date_from"); df != "" {
		if t, err := time.Parse(time.RFC3339, df); err == nil {
			dateFrom = &t
		}
	}
	if dt := r.FormValue("date_to"); dt != "" {
		if t, err := time.Parse(time.RFC3339, dt); err == nil {
			dateTo = &t
		}
	}

	params := sqlc.ListImagesByUserParams{
		UserID:    domain.PgInt8(userID),
		AlbumID:   domain.PgInt8Ptr(albumID),
		Keyword:   domain.PgTextNonEmpty(r.FormValue("keyword")),
		Extension: domain.PgTextNonEmpty(r.FormValue("extension")),
		DateFrom:  domain.PgTimeWithZonePtr(dateFrom),
		DateTo:    domain.PgTimeWithZonePtr(dateTo),
		Limit:     pageSize,
		Offset:    (page - 1) * pageSize,
	}
	images, err := h.db.ListImagesByUser(r.Context(), params)
	if err != nil {
		Fail(w, http.StatusInternalServerError, "failed to list images")
		return
	}

	total, err := h.db.CountImagesByUser(r.Context(), sqlc.CountImagesByUserParams{
		UserID:    domain.PgInt8(userID),
		AlbumID:   domain.PgInt8Ptr(albumID),
		Keyword:   params.Keyword,
		Extension: params.Extension,
		DateFrom:  params.DateFrom,
		DateTo:    params.DateTo,
	})
	if err != nil {
		slog.Warn("failed to count images", "error", err, "user_id", userID)
		total = 0
	}

	items := make([]ImageListItem, len(images))
	strategyBaseCache := make(map[int64]string)
	for i, img := range images {
		var surl string
		if img.StrategyID.Valid {
			sid := img.StrategyID.Int64
			base, ok := strategyBaseCache[sid]
			if !ok {
				base = h.strategyBaseURL(r.Context(), img.StrategyID)
				strategyBaseCache[sid] = base
			}
			surl = joinStrategyURL(base, img.Path, img.Name)
		}
		links := h.buildLinks(img.Key, img.Extension, img.Md5, img.OriginName, surl)
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
			ModerationStatus: img.ModerationStatus,
			StrategyID:       domain.PgInt8PtrVal(img.StrategyID),
			StrategyName:     img.StrategyName.String,
			StrategyType:     img.StrategyType.String,
			Links:            links,
			CreatedAt:        img.CreatedAt,
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

	links := h.buildLinks(img.Key, img.Extension, img.Md5, img.OriginName, joinStrategyURL(h.strategyBaseURL(r.Context(), img.StrategyID), img.Path, img.Name))
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
	writeAuditLog(h.db, r, "image.delete", "image", key, img.OriginName, map[string]any{
		"image_id": img.ID,
	})

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

	// Handle album image count updates if album changed
	if img.AlbumID != updated.AlbumID {
		if img.AlbumID.Valid {
			if err := h.db.DecrementAlbumImageNum(r.Context(), img.AlbumID.Int64); err != nil {
				slog.Warn("failed to decrement album image count", "error", err, "album_id", img.AlbumID.Int64)
			}
		}
		if updated.AlbumID.Valid {
			if err := h.db.IncrementAlbumImageNum(r.Context(), updated.AlbumID.Int64); err != nil {
				slog.Warn("failed to increment album image count", "error", err, "album_id", updated.AlbumID.Int64)
			}
		}
	}
	writeAuditLog(h.db, r, "image.update", "image", key, updated.OriginName, map[string]any{
		"before_permission": img.Permission,
		"after_permission":  updated.Permission,
		"before_album_id":   domain.PgInt8PtrVal(img.AlbumID),
		"after_album_id":    domain.PgInt8PtrVal(updated.AlbumID),
	})

	links := h.buildLinks(updated.Key, updated.Extension, updated.Md5, updated.OriginName, joinStrategyURL(h.strategyBaseURL(r.Context(), updated.StrategyID), updated.Path, updated.Name))
	Success(w, imageResponse(updated, links))
}

type batchDeleteRequest struct {
	Keys []string `json:"keys"`
}

func (h *ImageHandler) BatchDelete(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(domain.ContextKeyUserID).(int64)
	if !ok {
		Fail(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req batchDeleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || len(req.Keys) == 0 {
		Fail(w, http.StatusBadRequest, "keys is required and must be a non-empty array")
		return
	}
	if len(req.Keys) > 200 {
		Fail(w, http.StatusBadRequest, "too many keys, max 200 per batch")
		return
	}

	var deleted, failed int
	for _, key := range req.Keys {
		img, err := h.db.GetImageByKey(r.Context(), key)
		if err != nil || img.UserID.Int64 != userID {
			failed++
			continue
		}
		if err := h.deleter.DeleteImage(r.Context(), img.ID); err != nil {
			slog.Warn("batch delete failed for image", "key", key, "error", err)
			failed++
			continue
		}
		deleted++
	}
	writeAuditLog(h.db, r, "image.batch_delete", "image", "", "", map[string]any{
		"count":   len(req.Keys),
		"deleted": deleted,
		"failed":  failed,
	})

	Success(w, map[string]int{"deleted": deleted, "failed": failed})
}

func (h *ImageHandler) buildLinks(key, extension, md5, originName string, strategyURL string) domain.ImageLinks {
	return service.LinkBuilder{BaseURL: h.baseURL, StrategyURL: strategyURL}.BuildImageLinks(key, extension, md5, originName)
}

func (h *ImageHandler) strategyBaseURL(ctx context.Context, strategyID pgtype.Int8) string {
	if !strategyID.Valid {
		return ""
	}
	strategy, err := h.db.GetStrategyByID(ctx, strategyID.Int64)
	if err != nil {
		return ""
	}
	return extractStrategyURL(strategy.StrategyType, strategy.Configs)
}

func joinStrategyURL(base, path, name string) string {
	if base == "" {
		return ""
	}
	pathname := path
	if name != "" && path != "" {
		pathname = path + "/" + name
	}
	return strings.TrimRight(base, "/") + "/" + strings.TrimLeft(pathname, "/")
}

func imageResponse(img sqlc.Image, links domain.ImageLinks) ImageResponse {
	return ImageResponse{
		ID:               img.ID,
		Key:              img.Key,
		OriginName:       img.OriginName,
		SizeBytes:        img.SizeBytes,
		Mimetype:         img.Mimetype,
		Extension:        img.Extension,
		Width:            img.Width,
		Height:           img.Height,
		Md5:              img.Md5,
		Sha1:             img.Sha1,
		Permission:       img.Permission,
		AlbumID:          domain.PgInt8PtrVal(img.AlbumID),
		ModerationStatus: img.ModerationStatus,
		Links:            links,
		CreatedAt:        img.CreatedAt,
	}
}
