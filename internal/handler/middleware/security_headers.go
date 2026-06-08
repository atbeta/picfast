package middleware

import (
	"encoding/json"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"

	"github.com/atbeta/picfast/internal/config"
)

// SecurityHeaders sets baseline HTTP security headers for HTML and API responses.
// It reads analytics configuration to allow external script sources in CSP.
// CSP string is cached and only rebuilt when the analytics config changes.
func SecurityHeaders(cfg *config.Config) func(next http.Handler) http.Handler {
	var (
		mu        sync.Mutex
		cachedCSP string
		cacheKey  string // provider + string(config)
	)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("X-Frame-Options", "DENY")
			w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
			w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")

			app := cfg.AppSnapshot()
			key := app.AnalyticsProvider + "\x00" + string(app.AnalyticsConfig)

			mu.Lock()
			csp := cachedCSP
			hit := key == cacheKey
			if !hit {
				csp = buildCSP(app)
				cachedCSP = csp
				cacheKey = key
			}
			mu.Unlock()
			w.Header().Set("Content-Security-Policy", csp)
			next.ServeHTTP(w, r)
		})
	}
}

func buildCSP(app config.AppConfig) string {
	scriptSrc := "'self' 'unsafe-inline'"
	for _, src := range analyticsScriptSources(app.AnalyticsProvider, app.AnalyticsConfig) {
		scriptSrc += " " + src
	}
	return "default-src 'self'; " +
		"script-src " + scriptSrc + "; " +
		"style-src 'self' 'unsafe-inline'; " +
		"img-src 'self' data: blob: https: http:; " +
		"font-src 'self' data:; " +
		"connect-src 'self' https: http:; " +
		"frame-ancestors 'none'; " +
		"base-uri 'self'; " +
		"form-action 'self'"
}

func analyticsScriptSources(provider string, raw json.RawMessage) []string {
	switch provider {
	case "":
		return nil
	case "plausible":
		sources := []string{"https://plausible.io"}
		if origin := extractScriptOrigin(raw); origin != "" {
			sources = append(sources, origin)
		}
		return sources
	case "umami":
		if origin := extractScriptOrigin(raw); origin != "" {
			return []string{origin}
		}
		return nil
	case "ga4":
		return []string{"https://www.googletagmanager.com", "https://www.google-analytics.com"}
	case "baidu":
		return []string{"https://hm.baidu.com"}
	case "custom":
		return extractCustomScriptOrigins(raw)
	default:
		return nil
	}
}

func extractScriptOrigin(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	var m map[string]interface{}
	if json.Unmarshal(raw, &m) != nil {
		return ""
	}
	scriptURL, _ := m["script_url"].(string)
	if scriptURL == "" {
		return ""
	}
	u, err := url.Parse(scriptURL)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return ""
	}
	return u.Scheme + "://" + u.Host
}

var scriptSrcRe = regexp.MustCompile(`<script\b[^>]*\bsrc\s*=\s*["']([^"']+)["'][^>]*>`)

func extractCustomScriptOrigins(raw json.RawMessage) []string {
	if len(raw) == 0 {
		return nil
	}
	var m map[string]interface{}
	if json.Unmarshal(raw, &m) != nil {
		return nil
	}
	scriptHTML, _ := m["script"].(string)
	if scriptHTML == "" {
		return nil
	}

	seen := make(map[string]bool)
	var origins []string
	for _, match := range scriptSrcRe.FindAllStringSubmatch(scriptHTML, -1) {
		if len(match) < 2 {
			continue
		}
		u, err := url.Parse(match[1])
		if err != nil || u.Scheme == "" || u.Host == "" {
			continue
		}
		origin := u.Scheme + "://" + strings.ToLower(u.Host)
		if !seen[origin] {
			seen[origin] = true
			origins = append(origins, origin)
		}
	}
	return origins
}
