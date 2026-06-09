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
)

type FlatUploadHandler struct {
	upload         *service.UploadService
	baseURL        string
	maxUploadBytes int64
}

type flatUploadResponse struct {
	URL          string `json:"url"`
	ThumbnailURL string `json:"thumbnail_url,omitempty"`
	Markdown     string `json:"markdown,omitempty"`
	BBCode       string `json:"bbcode,omitempty"`
	HTML         string `json:"html,omitempty"`
}

func NewFlatUploadHandler(upload *service.UploadService, baseURL string, maxUploadBytes int64) *FlatUploadHandler {
	if maxUploadBytes <= 0 {
		maxUploadBytes = defaultMaxUploadBytes
	}
	return &FlatUploadHandler{upload: upload, baseURL: baseURL, maxUploadBytes: maxUploadBytes}
}

func (h *FlatUploadHandler) Upload(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	observeUpload := func(result, reason string, bytes int64) {
		picmetrics.ObserveUpload("flat", result, reason, bytes, time.Since(start))
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
		FileSize: header.Size,
		UserID:   userID,
		ClientIP: clientip.FromRequest(r),
	})
	if err != nil {
		observeUpload("error", picmetrics.ClassifyUploadError(err), int64(len(fileData)))
		Fail(w, http.StatusBadRequest, err.Error())
		return
	}
	observeUpload("success", picmetrics.ReasonNone, result.OriginalSizeBytes)

	imageURL := h.baseURL + "/i/" + result.Image.Key + "." + result.Image.Extension
	thumbURL := h.baseURL + "/t/" + result.Image.Md5 + ".png"

	resp := flatUploadResponse{
		URL:          imageURL,
		ThumbnailURL: thumbURL,
		Markdown:     "![" + result.Image.OriginName + "](" + imageURL + ")",
		BBCode:       "[img]" + imageURL + "[/img]",
		HTML:         `<img src="` + imageURL + `" alt="` + result.Image.OriginName + `" />`,
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(resp)
}
