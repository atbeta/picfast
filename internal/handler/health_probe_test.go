package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthzReturnsOKWithoutDependencies(t *testing.T) {
	env := newTestEnv(t)

	rec := doReq(env.Router, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /healthz status = %d, want 200", rec.Code)
	}
	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode /healthz body: %v", err)
	}
	if body["status"] != "ok" {
		t.Fatalf("GET /healthz status field = %q, want ok", body["status"])
	}
}

func TestReadyzReturnsOKWhenDatabaseUp(t *testing.T) {
	env := newTestEnv(t)

	rec := doReq(env.Router, httptest.NewRequest(http.MethodGet, "/readyz", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /readyz status = %d, want 200", rec.Code)
	}
	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode /readyz body: %v", err)
	}
	if body["status"] != "ok" {
		t.Fatalf("GET /readyz status field = %q, want ok", body["status"])
	}
}
