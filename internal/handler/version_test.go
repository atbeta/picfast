package handler_test

import (
	"net/http"
	"testing"
)

func TestVersionEndpoint(t *testing.T) {
	env := newTestEnv(t)

	req := newJSONReq(t, http.MethodGet, "/api/v1/version", nil)
	rec := doReq(env.Router, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", rec.Code, rec.Body.String())
	}
	data := respDataMap(t, parseResp(t, rec))
	if data["version"] == "" {
		t.Fatalf("version is empty: %#v", data)
	}
	if data["github_url"] == "" {
		t.Fatalf("github_url is empty: %#v", data)
	}
}
