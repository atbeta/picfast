package handler

import (
	"net/http"

	"github.com/atbeta/picfast/internal/handler/httputil"
)

type Response = httputil.Response

func Success(w http.ResponseWriter, data interface{}) {
	httputil.Success(w, data)
}

func SuccessMessage(w http.ResponseWriter, message string) {
	httputil.SuccessMessage(w, message)
}

func Created(w http.ResponseWriter, data interface{}) {
	httputil.Created(w, data)
}

func Fail(w http.ResponseWriter, code int, message string) {
	httputil.Fail(w, code, message)
}

func FailWithErrors(w http.ResponseWriter, code int, message string, errors map[string][]string) {
	httputil.FailWithErrors(w, code, message, errors)
}

func Paginated(w http.ResponseWriter, data interface{}, total int64, page, pageSize int32) {
	httputil.Paginated(w, data, total, page, pageSize)
}
