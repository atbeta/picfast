package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestJobsEndpointNotRegisteredWithoutToken(t *testing.T) {
	env := newTestEnv(t)

	rec := doReq(env.Router, httptest.NewRequest(http.MethodPost, "/internal/jobs/run", nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("POST /internal/jobs/run without token status = %d, want 404", rec.Code)
	}
}

func TestJobsEndpointAuth(t *testing.T) {
	env := newTestEnv(t)
	env.Config.Jobs.Token = "test-jobs-token"
	env.rebuildRouter()

	for _, tc := range []struct {
		name   string
		header string
		want   int
	}{
		{"missing header", "", http.StatusUnauthorized},
		{"wrong token", "Bearer nope", http.StatusUnauthorized},
		{"not bearer", "Token test-jobs-token", http.StatusUnauthorized},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/internal/jobs/run", nil)
			if tc.header != "" {
				req.Header.Set("Authorization", tc.header)
			}
			rec := doReq(env.Router, req)
			if rec.Code != tc.want {
				t.Fatalf("status = %d, want %d", rec.Code, tc.want)
			}
		})
	}

	t.Run("correct token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/internal/jobs/run", nil)
		req.Header.Set("Authorization", "Bearer test-jobs-token")
		rec := doReq(env.Router, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200 (body=%s)", rec.Code, rec.Body.String())
		}
		var report struct {
			Webhooks struct {
				State string `json:"state"`
			} `json:"webhooks"`
			Cleanup struct {
				State string `json:"state"`
			} `json:"cleanup"`
		}
		if err := json.NewDecoder(rec.Body).Decode(&report); err != nil {
			t.Fatalf("decode report: %v", err)
		}
		if report.Cleanup.State != "ran" {
			t.Fatalf("cleanup state = %q, want ran", report.Cleanup.State)
		}
	})
}
