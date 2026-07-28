package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/atbeta/picfast/internal/clientip"
	"github.com/atbeta/picfast/internal/domain"
	picmetrics "github.com/atbeta/picfast/internal/metrics"
	"github.com/atbeta/picfast/internal/service"
	"github.com/atbeta/picfast/internal/sqlc"
)

type ShareXHandler struct {
	upload          *service.UploadService
	baseURL         string
	maxUploadBytes  int64
	db              *sqlc.Queries
	auditUploadLogs bool
}

type shareXResponse struct {
	URL          string `json:"url"`
	ThumbnailURL string `json:"thumbnail_url,omitempty"`
	DeletionURL  string `json:"deletion_url,omitempty"`
}

func NewShareXHandler(upload *service.UploadService, baseURL string, maxUploadBytes int64, db *sqlc.Queries, auditUploadLogs bool) *ShareXHandler {
	if maxUploadBytes <= 0 {
		maxUploadBytes = defaultMaxUploadBytes
	}
	return &ShareXHandler{upload: upload, baseURL: baseURL, maxUploadBytes: maxUploadBytes, db: db, auditUploadLogs: auditUploadLogs}
}

func (h *ShareXHandler) Upload(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	observeUpload := func(result, reason string, bytes int64) {
		picmetrics.ObserveUpload("sharex", result, reason, bytes, time.Since(start))
	}

	r.Body = http.MaxBytesReader(w, r.Body, h.maxUploadBytes)

	// multipartFormMemory sets the max in-memory buffer for form parsing;
	// payloads exceeding this are spilled to temp files on disk.
	const multipartFormMemory = 32 << 20
	if err := r.ParseMultipartForm(multipartFormMemory); err != nil {
		observeUpload("error", "invalid_file", 0)
		Fail(w, http.StatusBadRequest, multipartParseErrorMessage(err, h.maxUploadBytes))
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

	var userID *int64
	if uid, ok := r.Context().Value(domain.ContextKeyUserID).(int64); ok {
		userID = &uid
	}

	result, err := h.upload.Store(r.Context(), service.UploadParams{
		FileData: fileData,
		FileName: header.Filename,
		UserID:   userID,
		ClientIP: clientip.FromRequest(r),
	})
	if err != nil {
		observeUpload("error", picmetrics.ClassifyUploadError(err), int64(len(fileData)))
		Fail(w, http.StatusBadRequest, err.Error())
		return
	}
	observeUpload("success", picmetrics.ReasonNone, result.OriginalSizeBytes)

	if h.auditUploadLogs {
		details := map[string]any{
			"source": "sharex",
		}
		if userID == nil {
			details["guest"] = true
		}
		writeAuditLog(h.db, r, "image.upload", "image", result.Image.Key, result.Image.OriginName, details)
	}

	resp := shareXResponse{
		URL:          result.Links.URL,
		ThumbnailURL: result.Links.ThumbnailURL,
	}

	// ShareX expects URL fields at the top level of JSON response.
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *ShareXHandler) Config(w http.ResponseWriter, r *http.Request) {
	config := map[string]interface{}{
		"Version":         "15.0.0",
		"Name":            "PicFast",
		"DestinationType": "ImageUploader",
		"RequestMethod":   "POST",
		"RequestURL":      h.baseURL + "/api/v1/sharex/upload",
		"Headers": map[string]string{
			"Authorization": "{if:Authorization}",
		},
		"Body":         "MultipartFormData",
		"FileFormName": "file",
		"URL":          "{json:url}",
		"ThumbnailURL": "{json:thumbnail_url}",
	}

	// Keep exported config in native .sxcu shape (no envelope).
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="picfast.sxcu"`)
	_ = json.NewEncoder(w).Encode(config)
}
