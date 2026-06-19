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

func SecurityHeaders(cfg *config.Config) func(next http.Handler) http.Handler {
	var (
		mu        sync.Mutex
		cachedCSP string
		cacheKey  string
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
			if key != cacheKey {
				cachedCSP = buildCSP(app)
				cacheKey = key
			}
			csp := cachedCSP
			mu.Unlock()
			w.Header().Set("Content-Security-Policy", csp)
			next.ServeHTTP(w, r)
		})
	}
}

func buildCSP(app config.AppConfig) string {
	scriptSrc := []string{"'self'", "'unsafe-inline'"}
	for _, src := range analyticsScriptSources(app.AnalyticsProvider, app.AnalyticsConfig) {
		scriptSrc = append(scriptSrc, src)
	}

	connectSrc := []string{"'self'"}
	for _, src := range analyticsConnectSources(app.AnalyticsProvider, app.AnalyticsConfig) {
		connectSrc = append(connectSrc, src)
	}

	return "default-src 'self'; " +
		"script-src " + strings.Join(scriptSrc, " ") + "; " +
		"script-src-attr 'none'; " +
		"style-src 'self' 'unsafe-inline'; " +
		"img-src 'self' data: blob: https: http:; " +
		"font-src 'self' data:; " +
		"connect-src " + strings.Join(connectSrc, " ") + "; " +
		"object-src 'none'; " +
		"frame-ancestors 'none'; " +
		"base-uri 'self'; " +
		"form-action 'self'"
}

func analyticsConnectSources(provider string, raw json.RawMessage) []string {
	switch provider {
	case "":
		return nil
	case "plausible":
		if origin := extractScriptOrigin(raw); origin != "" {
			return []string{origin}
		}
		return []string{"https://plausible.io"}
	case "umami":
		if origin := extractScriptOrigin(raw); origin != "" {
			return []string{origin}
		}
		return nil
	case "ga4":
		return []string{
			"https://www.googletagmanager.com",
			"https://www.google-analytics.com",
			"https://region1.google-analytics.com",
			"https://analytics.google.com",
		}
	case "baidu":
		return []string{"https://hm.baidu.com"}
	case "custom":
		return extractCustomScriptOrigins(raw)
	default:
		return nil
	}
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
