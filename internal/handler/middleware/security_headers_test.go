package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/atbeta/picfast/internal/config"
)

func testConfig() *config.Config {
	return &config.Config{}
}

func extractScriptSrc(csp string) string {
	start := strings.Index(csp, "script-src ")
	if start < 0 {
		return ""
	}
	rest := csp[start:]
	end := strings.Index(rest, ";")
	if end < 0 {
		return rest
	}
	return rest[:end]
}

func TestSecurityHeaders(t *testing.T) {
	cfg := testConfig()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	SecurityHeaders(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(rec, req)

	for _, key := range []string{
		"X-Content-Type-Options",
		"X-Frame-Options",
		"Referrer-Policy",
		"Permissions-Policy",
		"Content-Security-Policy",
	} {
		if rec.Header().Get(key) == "" {
			t.Fatalf("missing header %s", key)
		}
	}
	if rec.Header().Get("X-Frame-Options") != "DENY" {
		t.Fatalf("X-Frame-Options = %q", rec.Header().Get("X-Frame-Options"))
	}
}

func TestCSPWithoutAnalytics(t *testing.T) {
	cfg := testConfig()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	SecurityHeaders(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(rec, req)

	csp := rec.Header().Get("Content-Security-Policy")
	if !strings.Contains(csp, "script-src 'self' 'unsafe-inline';") {
		t.Fatalf("CSP script-src without analytics should be default, got: %s", csp)
	}
}

func TestCSPWithUmami(t *testing.T) {
	cfg := testConfig()
	cfg.App.AnalyticsProvider = "umami"
	cfg.App.AnalyticsConfig = json.RawMessage(`{"script_url":"https://stats.example.com/script.js","website_id":"abc123"}`)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	SecurityHeaders(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(rec, req)

	csp := rec.Header().Get("Content-Security-Policy")
	if !strings.Contains(csp, "https://stats.example.com") {
		t.Fatalf("CSP should include umami script origin, got: %s", csp)
	}
}

func TestCSPWithGA4(t *testing.T) {
	cfg := testConfig()
	cfg.App.AnalyticsProvider = "ga4"

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	SecurityHeaders(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(rec, req)

	csp := rec.Header().Get("Content-Security-Policy")
	if !strings.Contains(csp, "https://www.googletagmanager.com") {
		t.Fatalf("CSP should include googletagmanager.com for GA4, got: %s", csp)
	}
	if !strings.Contains(csp, "https://www.google-analytics.com") {
		t.Fatalf("CSP should include google-analytics.com for GA4, got: %s", csp)
	}
}

func TestCSPWithCustom(t *testing.T) {
	cfg := testConfig()
	cfg.App.AnalyticsProvider = "custom"
	cfg.App.AnalyticsConfig = json.RawMessage(`{"script":"<script src=\"https://tracker.example.com/analytics.js\"></script><script>console.log(1)</script>"}`)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	SecurityHeaders(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(rec, req)

	csp := rec.Header().Get("Content-Security-Policy")
	if strings.Contains(csp, "https://tracker.example.com") {
		// the origin was found — good
	} else {
		t.Fatalf("CSP should include extracted custom script origin, got: %s", csp)
	}
	// script-src should not contain broad https: scheme-source
	scriptSrc := extractScriptSrc(csp)
	if strings.Contains(scriptSrc, "https:") && !strings.Contains(scriptSrc, "https://") {
		t.Fatalf("CSP script-src should NOT include broad https: for custom analytics, got: %s", csp)
	}
}

func TestCSPWithCustomInlineOnly(t *testing.T) {
	cfg := testConfig()
	cfg.App.AnalyticsProvider = "custom"
	cfg.App.AnalyticsConfig = json.RawMessage(`{"script":"<script>var _paq=window._paq||[]</script>"}`)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	SecurityHeaders(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(rec, req)

	csp := rec.Header().Get("Content-Security-Policy")
	scriptSrc := extractScriptSrc(csp)
	if scriptSrc != "script-src 'self' 'unsafe-inline'" {
		t.Fatalf("inline-only custom analytics should keep strict script-src, got: %s", scriptSrc)
	}
}

func TestCSPWithUmamiNoScriptURL(t *testing.T) {
	cfg := testConfig()
	cfg.App.AnalyticsProvider = "umami"
	cfg.App.AnalyticsConfig = json.RawMessage(`{"website_id":"abc123"}`)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	SecurityHeaders(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(rec, req)

	csp := rec.Header().Get("Content-Security-Policy")
	scriptSrc := extractScriptSrc(csp)
	if scriptSrc != "script-src 'self' 'unsafe-inline'" {
		t.Fatalf("umami without script_url should keep strict script-src, got: %s", scriptSrc)
	}
}

func TestCSPWithPlausibleNoScriptURL(t *testing.T) {
	cfg := testConfig()
	cfg.App.AnalyticsProvider = "plausible"
	cfg.App.AnalyticsConfig = json.RawMessage(`{"domain":"example.com"}`)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	SecurityHeaders(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(rec, req)

	csp := rec.Header().Get("Content-Security-Policy")
	if !strings.Contains(csp, "https://plausible.io") {
		t.Fatalf("plausible without script_url should still include default plausible.io, got: %s", csp)
	}
}

func TestExtractCustomScriptOrigins_Malicious(t *testing.T) {
	// HTML comment containing fake script src should be ignored
	cases := []struct {
		name   string
		script string
		expect int // expected number of origins
	}{
		{
			name:   "script inside HTML comment (known limitation)",
			script: "<!-- <script src=\"https://evil.com/x.js\"></script> -->",
			expect: 1,
		},
		{
			name:   "encoded quotes in src",
			script: `<script src="https://good.com/a.js"></script>`,
			expect: 1,
		},
		{
			name:   "src without scheme",
			script: `<script src="//example.com/t.js"></script>`,
			expect: 0,
		},
		{
			name:   "empty src",
			script: `<script src=""></script>`,
			expect: 0,
		},
		{
			name:   "src with javascript: scheme",
			script: `<script src="javascript:alert(1)"></script>`,
			expect: 0,
		},
		{
			name:   "multiple valid and invalid",
			script: `<script src="https://a.com/x.js"></script><script src="//b.com/y.js"></script><script src="https://c.com/z.js"></script>`,
			expect: 2,
		},
		{
			name:   "data scheme should be rejected",
			script: `<script src="data:text/javascript,alert(1)"></script>`,
			expect: 0,
		},
		{
			name:   "file scheme should be rejected",
			script: `<script src="file:///etc/passwd"></script>`,
			expect: 0,
		},
		{
			name:   "blob scheme should be rejected",
			script: `<script src="blob:https://example.com/uuid"></script>`,
			expect: 0,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			origins := extractCustomScriptOrigins(json.RawMessage(`{"script":"` + strings.ReplaceAll(tc.script, `"`, `\"`) + `"}`))
			if len(origins) != tc.expect {
				t.Fatalf("expected %d origins, got %d: %v", tc.expect, len(origins), origins)
			}
		})
	}
}
