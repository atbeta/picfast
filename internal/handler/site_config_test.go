package handler_test

import (
	"net/http"
	"testing"

	"github.com/atbeta/picfast/internal/service"
)

func TestSiteConfigReportsImageCapabilities(t *testing.T) {
	env := newTestEnv(t)

	req := newJSONReq(t, http.MethodGet, "/api/v1/config", nil)
	rec := doReq(env.Router, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", rec.Code, rec.Body.String())
	}
	data := respDataMap(t, parseResp(t, rec))
	caps, ok := data["image_capabilities"].(map[string]interface{})
	if !ok {
		t.Fatalf("image_capabilities missing or wrong type: %T", data["image_capabilities"])
	}
	if caps["backend"] != service.Capabilities().Backend {
		t.Fatalf("backend = %v, want %q", caps["backend"], service.Capabilities().Backend)
	}
	if _, ok := caps["webp_output"].(bool); !ok {
		t.Fatalf("webp_output = %T, want bool", caps["webp_output"])
	}
	if _, ok := caps["watermark"].(bool); !ok {
		t.Fatalf("watermark = %T, want bool", caps["watermark"])
	}
}
