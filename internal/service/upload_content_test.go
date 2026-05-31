package service

import "testing"

func TestContentMatchesExtension(t *testing.T) {
	tests := []struct {
		ext   string
		data  []byte
		match bool
	}{
		{"png", []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}, true},
		{"png", []byte{0xFF, 0xD8, 0xFF}, false},
		{"jpg", []byte{0xFF, 0xD8, 0xFF, 0xE0}, true},
		{"gif", []byte("GIF89a"), true},
		{"webp", append(append([]byte("RIFF"), []byte{0, 0, 0, 0}...), []byte("WEBP")...), true},
		{"svg", []byte(`<svg xmlns="http://www.w3.org/2000/svg"></svg>`), true},
		{"svg", []byte("not svg"), false},
		{"ico", []byte{0x00, 0x00, 0x01, 0x00}, true},
	}

	for _, tt := range tests {
		got := contentMatchesExtension(tt.ext, tt.data)
		if got != tt.match {
			t.Fatalf("ext=%s match=%v want %v", tt.ext, got, tt.match)
		}
	}
}

func TestValidateUploadContentRejectsMismatch(t *testing.T) {
	err := validateUploadContent("png", []byte{0xFF, 0xD8, 0xFF})
	if err == nil {
		t.Fatal("expected mismatch error")
	}
}
