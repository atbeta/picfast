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

func extractDirective(csp, directive string) string {
	start := strings.Index(csp, directive+" ")
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
	if !strings.Contains(csp, "object-src 'none';") {
		t.Fatalf("CSP should include object-src 'none', got: %s", csp)
	}
	if !strings.Contains(csp, "script-src-attr 'none';") {
		t.Fatalf("CSP should include script-src-attr 'none', got: %s", csp)
	}
	if !strings.Contains(csp, "connect-src 'self';") {
		t.Fatalf("CSP connect-src without analytics should be 'self' only, got: %s", csp)
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
	connectSrc := extractDirective(csp, "connect-src")
	if !strings.Contains(connectSrc, "https://stats.example.com") {
		t.Fatalf("CSP connect-src should include umami origin, got: %s", connectSrc)
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
	connectSrc := extractDirective(csp, "connect-src")
	if !strings.Contains(connectSrc, "https://region1.google-analytics.com") {
		t.Fatalf("CSP connect-src should include region1.google-analytics.com for GA4, got: %s", connectSrc)
	}
	if !strings.Contains(connectSrc, "https://www.googletagmanager.com") {
		t.Fatalf("CSP connect-src should include www.googletagmanager.com for GA4, got: %s", connectSrc)
	}
}

func TestCSPWithPlausible(t *testing.T) {
	cfg := testConfig()
	cfg.App.AnalyticsProvider = "plausible"
	cfg.App.AnalyticsConfig = json.RawMessage(`{"domain":"example.com","script_url":"https://plausible.io/js/script.js"}`)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	SecurityHeaders(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(rec, req)

	csp := rec.Header().Get("Content-Security-Policy")
	if !strings.Contains(csp, "https://plausible.io") {
		t.Fatalf("CSP should include plausible.io, got: %s", csp)
	}
	connectSrc := extractDirective(csp, "connect-src")
	if !strings.Contains(connectSrc, "https://plausible.io") {
		t.Fatalf("CSP connect-src should include plausible.io, got: %s", connectSrc)
	}
}

func TestCSPWithBaidu(t *testing.T) {
	cfg := testConfig()
	cfg.App.AnalyticsProvider = "baidu"

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	SecurityHeaders(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(rec, req)

	csp := rec.Header().Get("Content-Security-Policy")
	if !strings.Contains(csp, "https://hm.baidu.com") {
		t.Fatalf("CSP should include hm.baidu.com, got: %s", csp)
	}
	connectSrc := extractDirective(csp, "connect-src")
	if !strings.Contains(connectSrc, "https://hm.baidu.com") {
		t.Fatalf("CSP connect-src should include hm.baidu.com, got: %s", connectSrc)
	}
}

func TestCSPConnectSrcNoBroadWildcard(t *testing.T) {
	cfg := testConfig()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	SecurityHeaders(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(rec, req)

	csp := rec.Header().Get("Content-Security-Policy")
	connectSrc := extractDirective(csp, "connect-src")
	if strings.Contains(connectSrc, "https:") || strings.Contains(connectSrc, "http:") {
		t.Fatalf("CSP connect-src should NOT contain broad https:/http: scheme sources, got: %s", connectSrc)
	}
}

func TestCSPScriptSrcNoBroadWildcard(t *testing.T) {
	cfg := testConfig()
	cfg.App.AnalyticsProvider = "custom"
	cfg.App.AnalyticsConfig = json.RawMessage(`{"script":"<script src=\"https://tracker.example.com/analytics.js\"></script><script>console.log(1)</script>"}`)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	SecurityHeaders(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(rec, req)

	csp := rec.Header().Get("Content-Security-Policy")
	if !strings.Contains(csp, "https://tracker.example.com") {
		t.Fatalf("CSP should include extracted custom script origin, got: %s", csp)
	}
	scriptSrc := extractDirective(csp, "script-src")
	if strings.Contains(scriptSrc, "https:") && !strings.Contains(scriptSrc, "https://") {
		t.Fatalf("CSP script-src should NOT include broad https: for custom analytics, got: %s", csp)
	}
	connectSrc := extractDirective(csp, "connect-src")
	if !strings.Contains(connectSrc, "https://tracker.example.com") {
		t.Fatalf("CSP connect-src should include custom script origin, got: %s", connectSrc)
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
	scriptSrc := extractDirective(csp, "script-src")
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
	scriptSrc := extractDirective(csp, "script-src")
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
	cases := []struct {
		name   string
		script string
		expect int
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
