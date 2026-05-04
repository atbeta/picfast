package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestRoutePatternUsesChiPattern(t *testing.T) {
	r := chi.NewRouter()
	r.Get("/t/{hash}.png", func(w http.ResponseWriter, req *http.Request) {
		got := routePattern(req)
		if got != "/t/{hash}.png" {
			t.Fatalf("routePattern = %q, want /t/{hash}.png", got)
		}
	})

	r.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/t/abcdef.png", nil))
}

func TestRoutePatternUsesUnknownForUnmatchedRoutes(t *testing.T) {
	r := chi.NewRouter()
	r.NotFound(func(w http.ResponseWriter, req *http.Request) {
		got := routePattern(req)
		if got != "unknown" {
			t.Fatalf("routePattern = %q, want unknown", got)
		}
	})

	r.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/scan/random-path", nil))
}
