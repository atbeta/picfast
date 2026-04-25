package handler

import (
	"encoding/json"
	"net/http"
)

type Response struct {
	Status  bool        `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func Success(w http.ResponseWriter, data interface{}) {
	writeJSON(w, http.StatusOK, Response{
		Status:  true,
		Message: "success",
		Data:    data,
	})
}

func SuccessMessage(w http.ResponseWriter, message string) {
	writeJSON(w, http.StatusOK, Response{
		Status:  true,
		Message: message,
		Data:    struct{}{},
	})
}

func Created(w http.ResponseWriter, data interface{}) {
	writeJSON(w, http.StatusCreated, Response{
		Status:  true,
		Message: "created",
		Data:    data,
	})
}

func Fail(w http.ResponseWriter, code int, message string) {
	writeJSON(w, code, Response{
		Status:  false,
		Message: message,
		Data:    struct{}{},
	})
}

func FailWithErrors(w http.ResponseWriter, code int, message string, errors map[string][]string) {
	writeJSON(w, code, Response{
		Status:  false,
		Message: message,
		Data:    errors,
	})
}

func Paginated(w http.ResponseWriter, data interface{}, total int64, page int, pageSize int) {
	Success(w, map[string]interface{}{
		"items": data,
		"total": total,
		"page":  page,
		"size":  pageSize,
	})
}

func writeJSON(w http.ResponseWriter, code int, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}
