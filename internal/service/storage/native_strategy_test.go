package storage

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestNativeStrategyValidatorsAcceptRequiredConfigs(t *testing.T) {
	tests := []struct {
		name string
		typ  string
		cfg  string
	}{
		{
			name: "kodo",
			typ:  "kodo",
			cfg:  `{"access_key":"ak","secret_key":"sk","bucket":"bucket","domain":"https://cdn.example.com","zone":"z0"}`,
		},
		{
			name: "oss",
			typ:  "oss",
			cfg:  `{"endpoint":"https://oss-cn-hangzhou.aliyuncs.com","access_key":"ak","secret_key":"sk","bucket":"bucket","url":"https://cdn.example.com"}`,
		},
		{
			name: "cos",
			typ:  "cos",
			cfg:  `{"bucket_url":"https://bucket-1250000000.cos.ap-guangzhou.myqcloud.com","secret_id":"sid","secret_key":"sk","url":"https://cdn.example.com"}`,
		},
		{
			name: "webdav",
			typ:  "webdav",
			cfg:  `{"endpoint":"https://dav.example.com/uploads","username":"user","password":"pass","url":"https://cdn.example.com"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateConfig(tt.typ, json.RawMessage(tt.cfg)); err != nil {
				t.Fatalf("ValidateConfig(%s) error = %v", tt.typ, err)
			}
		})
	}
}

func TestNativeStrategyValidatorsRejectMissingRequiredConfigs(t *testing.T) {
	tests := []struct {
		typ string
		cfg string
	}{
		{typ: "kodo", cfg: `{"access_key":"ak","secret_key":"sk","domain":"https://cdn.example.com"}`},
		{typ: "oss", cfg: `{"endpoint":"https://oss-cn-hangzhou.aliyuncs.com","access_key":"ak","secret_key":"sk"}`},
		{typ: "cos", cfg: `{"bucket_url":"https://bucket-1250000000.cos.ap-guangzhou.myqcloud.com","secret_id":"sid"}`},
		{typ: "webdav", cfg: `{"endpoint":"https://dav.example.com/uploads","username":"user"}`},
	}

	for _, tt := range tests {
		t.Run(tt.typ, func(t *testing.T) {
			if err := ValidateConfig(tt.typ, json.RawMessage(tt.cfg)); err == nil {
				t.Fatalf("ValidateConfig(%s) expected error", tt.typ)
			}
		})
	}
}

func TestNativeStrategyURLsUseConfiguredPublicBase(t *testing.T) {
	tests := []struct {
		typ string
		cfg string
	}{
		{typ: "kodo", cfg: `{"access_key":"ak","secret_key":"sk","bucket":"bucket","domain":"https://cdn.example.com","zone":"z0"}`},
		{typ: "oss", cfg: `{"endpoint":"https://oss-cn-hangzhou.aliyuncs.com","access_key":"ak","secret_key":"sk","bucket":"bucket","url":"https://cdn.example.com"}`},
		{typ: "cos", cfg: `{"bucket_url":"https://bucket-1250000000.cos.ap-guangzhou.myqcloud.com","secret_id":"sid","secret_key":"sk","url":"https://cdn.example.com"}`},
		{typ: "webdav", cfg: `{"endpoint":"https://dav.example.com/uploads","username":"user","password":"pass","url":"https://cdn.example.com"}`},
	}

	for _, tt := range tests {
		t.Run(tt.typ, func(t *testing.T) {
			store, err := New(tt.typ, json.RawMessage(tt.cfg))
			if err != nil {
				t.Fatalf("New(%s) error = %v", tt.typ, err)
			}
			t.Cleanup(func() { _ = store.Close() })

			got := store.URL("/2026/05/cat.png")
			if got != "https://cdn.example.com/2026/05/cat.png" {
				t.Fatalf("URL() = %q", got)
			}
		})
	}
}

func TestWebDAVStorageURLFallsBackToEndpoint(t *testing.T) {
	store, err := New("webdav", json.RawMessage(`{"endpoint":"https://dav.example.com/uploads/","username":"user","password":"pass"}`))
	if err != nil {
		t.Fatalf("New(webdav) error = %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })

	got := store.URL("2026/05/cat.png")
	if !strings.HasPrefix(got, "https://dav.example.com/uploads/") {
		t.Fatalf("URL() = %q, want endpoint prefix", got)
	}
}
