package service

import (
	"encoding/json"
	"testing"
)

func TestLinkBuilderBuildImageLinks(t *testing.T) {
	builder := LinkBuilder{BaseURL: "https://img.example.com"}

	links := builder.BuildImageLinks("abc123", "png", "md5hash", "cat.png")

	if links.URL != "https://img.example.com/i/abc123.png" {
		t.Fatalf("URL = %q", links.URL)
	}
	if links.ThumbnailURL != "https://img.example.com/t/md5hash.png" {
		t.Fatalf("ThumbnailURL = %q", links.ThumbnailURL)
	}
	if links.HTML != `<img src="https://img.example.com/i/abc123.png" alt="cat.png" />` {
		t.Fatalf("HTML = %q", links.HTML)
	}
	if links.BBCode != "[img]https://img.example.com/i/abc123.png[/img]" {
		t.Fatalf("BBCode = %q", links.BBCode)
	}
	if links.Markdown != "![cat.png](https://img.example.com/i/abc123.png)" {
		t.Fatalf("Markdown = %q", links.Markdown)
	}
}

// TestLinkBuilderBuildImageLinksWithStrategyURL pins the contract that
// StrategyURL (when non-empty) is used verbatim as the image URL.
//
// Callers are responsible for clearing StrategyURL when the storage
// strategy is served through the picfast /i/{key}.{ext} proxy route
// (currently: local and webdav). See upload_service.go which performs
// that clearing. If a future refactor inverts this priority (e.g. always
// prefer the proxy URL), this test will fail and force a deliberate update.
func TestLinkBuilderBuildImageLinksWithStrategyURL(t *testing.T) {
	builder := LinkBuilder{
		BaseURL:     "https://img.example.com",
		StrategyURL: "https://bucket.cos.ap-shanghai.myqcloud.com/2026/06/02/abc.png",
	}

	links := builder.BuildImageLinks("abc123", "png", "md5hash", "cat.png")

	if links.URL != "https://bucket.cos.ap-shanghai.myqcloud.com/2026/06/02/abc.png" {
		t.Fatalf("URL = %q, want strategy URL to be used verbatim", links.URL)
	}
	// ThumbnailURL is always derived from BaseURL regardless of StrategyURL.
	if links.ThumbnailURL != "https://img.example.com/t/md5hash.png" {
		t.Fatalf("ThumbnailURL = %q, want proxy URL", links.ThumbnailURL)
	}
}

func TestIsProxyLinkMode(t *testing.T) {
	tests := []struct {
		name    string
		configs json.RawMessage
		want    bool
	}{
		{"nil configs", nil, false},
		{"empty object", json.RawMessage(`{}`), false},
		{"explicit direct", json.RawMessage(`{"link_mode":"direct"}`), false},
		{"proxy", json.RawMessage(`{"link_mode":"proxy"}`), true},
		{"other fields only", json.RawMessage(`{"bucket":"b","url":"https://cdn.example.com"}`), false},
		{"proxy with other fields", json.RawMessage(`{"bucket":"b","link_mode":"proxy"}`), true},
		{"invalid json", json.RawMessage(`not-json`), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsProxyLinkMode(tt.configs); got != tt.want {
				t.Fatalf("IsProxyLinkMode(%s) = %v, want %v", tt.configs, got, tt.want)
			}
		})
	}
}
