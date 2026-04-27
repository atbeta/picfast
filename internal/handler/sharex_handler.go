package handler

import (
	"io"
	"net/http"

	"github.com/atbeta/picfast/internal/domain"
	"github.com/atbeta/picfast/internal/service"
)

type ShareXHandler struct {
	upload  *service.UploadService
	baseURL string
}

type shareXResponse struct {
	URL          string `json:"url"`
	ThumbnailURL string `json:"thumbnail_url,omitempty"`
	DeletionURL  string `json:"deletion_url,omitempty"`
}

func NewShareXHandler(upload *service.UploadService, baseURL string) *ShareXHandler {
	return &ShareXHandler{upload: upload, baseURL: baseURL}
}

func (h *ShareXHandler) Upload(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 50<<20)

	if err := r.ParseMultipartForm(50 << 20); err != nil {
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

	var userID *int64
	if uid, ok := r.Context().Value(domain.ContextKeyUserID).(int64); ok {
		userID = &uid
	}

	result, err := h.upload.Store(r.Context(), service.UploadParams{
		FileData: fileData,
		FileName: header.Filename,
		FileSize: header.Size,
		UserID:   userID,
		ClientIP: r.RemoteAddr,
	})
	if err != nil {
		Fail(w, http.StatusBadRequest, err.Error())
		return
	}

	imageURL := h.baseURL + "/i/" + result.Image.Key + "." + result.Image.Extension
	thumbURL := h.baseURL + "/t/" + result.Image.Md5 + ".png"

	Success(w, shareXResponse{
		URL:          imageURL,
		ThumbnailURL: thumbURL,
	})
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

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", `attachment; filename="picfast.sxcu"`)
	Success(w, config)
}
