package storage

import (
	"encoding/json"
	"testing"
)

func baseS3Config(t *testing.T, extra string) json.RawMessage {
	t.Helper()
	raw := `{"endpoint":"https://s3.example.com","bucket":"b","access_key":"ak","secret_key":"sk"`
	if extra != "" {
		raw += "," + extra
	}
	raw += "}"
	return json.RawMessage(raw)
}

func TestS3CompatiblePathStyleDefaults(t *testing.T) {
	tests := []struct {
		typ      string
		extra    string
		wantPath bool
	}{
		{"s3", "", true},
		{"tos", "", false},
		{"obs", "", false},
		{"s3", `"use_path_style":false`, false},
		{"tos", `"use_path_style":true`, true},
	}

	for _, tt := range tests {
		t.Run(tt.typ+"/"+tt.extra, func(t *testing.T) {
			s, err := New(tt.typ, baseS3Config(t, tt.extra))
			if err != nil {
				t.Fatalf("New(%s) error: %v", tt.typ, err)
			}
			s3s, ok := s.(*S3Storage)
			if !ok {
				t.Fatalf("New(%s) returned %T, want *S3Storage", tt.typ, s)
			}
			if got := s3s.client.Options().UsePathStyle; got != tt.wantPath {
				t.Fatalf("UsePathStyle = %v, want %v", got, tt.wantPath)
			}
		})
	}
}

func TestS3CompatibleValidator(t *testing.T) {
	for _, typ := range []string{"s3", "tos", "obs"} {
		if err := ValidateConfig(typ, baseS3Config(t, "")); err != nil {
			t.Fatalf("ValidateConfig(%s) error: %v", typ, err)
		}
		bad := json.RawMessage(`{"endpoint":"","bucket":"b","access_key":"ak","secret_key":"sk"}`)
		if err := ValidateConfig(typ, bad); err == nil {
			t.Fatalf("ValidateConfig(%s) expected error for missing endpoint", typ)
		}
	}
}
