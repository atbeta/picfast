package httputil

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type Response struct {
	Status  bool        `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func Success(w http.ResponseWriter, data interface{}) {
	WriteJSON(w, http.StatusOK, Response{
		Status:  true,
		Message: "success",
		Data:    data,
	})
}

func SuccessMessage(w http.ResponseWriter, message string) {
	WriteJSON(w, http.StatusOK, Response{
		Status:  true,
		Message: message,
		Data:    struct{}{},
	})
}

func Created(w http.ResponseWriter, data interface{}) {
	WriteJSON(w, http.StatusCreated, Response{
		Status:  true,
		Message: "created",
		Data:    data,
	})
}

func Fail(w http.ResponseWriter, code int, message string) {
	WriteJSON(w, code, Response{
		Status:  false,
		Message: message,
		Data:    struct{}{},
	})
}

func FailWithErrors(w http.ResponseWriter, code int, message string, errors map[string][]string) {
	WriteJSON(w, code, Response{
		Status:  false,
		Message: message,
		Data:    errors,
	})
}

func Paginated(w http.ResponseWriter, data interface{}, total int64, page, pageSize int32) {
	totalPages := int32(0)
	if pageSize > 0 && total > 0 {
		totalPages = int32((total + int64(pageSize) - 1) / int64(pageSize))
	}
	Success(w, map[string]interface{}{
		"items":       data,
		"total":       total,
		"page":        page,
		"size":        pageSize,
		"total_pages": totalPages,
	})
}

func WriteJSON(w http.ResponseWriter, code int, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Warn("failed to write json response", "error", err)
	}
}
