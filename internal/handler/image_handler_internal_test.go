package handler

import "testing"

// Strategy configs carry non-string values (for example "use_path_style": true
// for R2/MinIO). Decoding into map[string]string failed on those documents, so
// a direct-URL strategy silently produced no base URL and image links could not
// be built. Keep extractStrategyURL aligned with service.linkMode.
func TestExtractStrategyURLParsesConfigsWithNonStringValues(t *testing.T) {
	cases := []struct {
		name         string
		strategyType string
		configs      string
		want         string
	}{
		{"plain url", "s3", `{"url":"https://cdn.example.com/"}`, "https://cdn.example.com"},
		{"url with path style bool", "s3", `{"url":"https://cdn.example.com","use_path_style":true}`, "https://cdn.example.com"},
		{"url with numeric value", "oss", `{"url":"https://cdn.example.com","part_size":10485760}`, "https://cdn.example.com"},
		{"kodo domain fallback", "kodo", `{"domain":"https://cdn.example.com/","is_private":false}`, "https://cdn.example.com"},
		{"proxy mode yields no direct url", "s3", `{"url":"https://cdn.example.com","use_path_style":true,"link_mode":"proxy"}`, ""},
		{"local strategy", "local", `{"url":"/i"}`, ""},
		{"webdav strategy", "webdav", `{"url":"https://dav.example.com"}`, ""},
		{"invalid json", "s3", `{`, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := extractStrategyURL(tc.strategyType, []byte(tc.configs)); got != tc.want {
				t.Errorf("extractStrategyURL(%q, %s) = %q, want %q", tc.strategyType, tc.configs, got, tc.want)
			}
		})
	}
}
