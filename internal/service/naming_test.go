package service

import (
	"strings"
	"testing"
)

func TestGeneratePathname(t *testing.T) {
	result := GeneratePathname("{Y}/{m}/{d}", "{md5-16}", "png", "abcd1234efgh5678", 42)
	if !strings.Contains(result, ".png") {
		t.Fatalf("expected .png in result, got %s", result)
	}
	if !strings.Contains(result, "/") {
		t.Fatalf("expected path separator, got %s", result)
	}
}

func TestComputeMD5(t *testing.T) {
	hash := ComputeMD5([]byte("hello"))
	if hash == "" {
		t.Fatal("expected non-empty hash")
	}
	if len(hash) != 32 {
		t.Fatalf("expected md5 length 32, got %d", len(hash))
	}
}

func TestGenerateImageKey(t *testing.T) {
	tests := []struct {
		name   string
		length int
	}{
		{"length 4", 4},
		{"length 6", 6},
		{"length 8", 8},
		{"length 12", 12},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key1 := GenerateImageKey(tt.length)
			key2 := GenerateImageKey(tt.length)
			if len(key1) != tt.length {
				t.Fatalf("expected key length %d, got %d", tt.length, len(key1))
			}
			if key1 == key2 {
				t.Fatal("expected unique keys")
			}
		})
	}
}

func TestBaseKeyLength(t *testing.T) {
	tests := []struct {
		name        string
		totalImages int64
		minLength   int
		want        int
	}{
		// Default behaviour: minLength=4 follows the historical tier table.
		{"empty", 0, 4, 4},
		{"small", 100, 4, 4},
		{"near 4-5 boundary", 1679, 4, 4},
		{"at 4-5 boundary", 1680, 4, 5},
		{"mid 5", 30000, 4, 5},
		{"near 5-6 boundary", 60465, 4, 5},
		{"at 5-6 boundary", 60466, 4, 6},
		{"mid 6", 1000000, 4, 6},
		{"near 6-7 boundary", 2176781, 4, 6},
		{"at 6-7 boundary", 2176782, 4, 7},
		{"large", 100000000, 4, 8},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := BaseKeyLength(tt.totalImages, tt.minLength); got != tt.want {
				t.Errorf("BaseKeyLength(%d, %d) = %d, want %d", tt.totalImages, tt.minLength, got, tt.want)
			}
		})
	}
}

func TestBaseKeyLengthRespectsMinLength(t *testing.T) {
	tests := []struct {
		name        string
		totalImages int64
		minLength   int
		want        int
	}{
		// minLength above the natural tier wins on small libraries.
		{"min=5 small lib", 0, 5, 5},
		{"min=6 small lib", 1000, 6, 6},
		{"min=8 small lib", 0, 8, 8},
		{"min=10 tiny lib", 1, 10, 10},
		// minLength below the natural tier: tier still wins once it grows.
		{"min=4 tier=5", 30000, 4, 5},
		{"min=4 tier=8", 100000000, 4, 8},
		{"min=6 tier=7", 5000000, 6, 7},
		{"min=6 tier=8", 100000000, 6, 8},
		{"min=8 tier=8", 100000000, 8, 8},
		// Defensive clamping: out-of-range minLength is clamped to [4, 10].
		{"min=0 clamps to 4", 0, 0, 4},
		{"min=3 clamps to 4", 0, 3, 4},
		{"min=-100 clamps to 4", 100, -100, 4},
		{"min=11 clamps to 10", 0, 11, 10},
		{"min=999 clamps to 10", 0, 999, 10},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := BaseKeyLength(tt.totalImages, tt.minLength); got != tt.want {
				t.Errorf("BaseKeyLength(%d, %d) = %d, want %d", tt.totalImages, tt.minLength, got, tt.want)
			}
		})
	}
}

func TestClampImageKeyLength(t *testing.T) {
	tests := []struct {
		in, want int
	}{
		{MinImageKeyLength - 1, MinImageKeyLength},
		{0, MinImageKeyLength},
		{MinImageKeyLength, MinImageKeyLength},
		{5, 5},
		{8, 8},
		{MaxImageKeyLength, MaxImageKeyLength},
		{MaxImageKeyLength + 1, MaxImageKeyLength},
		{9999, MaxImageKeyLength},
	}
	for _, tt := range tests {
		if got := ClampImageKeyLength(tt.in); got != tt.want {
			t.Errorf("ClampImageKeyLength(%d) = %d, want %d", tt.in, got, tt.want)
		}
	}
}

func TestExpandRule(t *testing.T) {
	result := expandRule("{md5}", "abcd1234efgh5678", 1)
	if result != "abcd1234efgh5678" {
		t.Fatalf("expandRule failed: %s", result)
	}

	result2 := expandRule("{md5-16}", "abcd1234efgh5678", 1)
	if len(result2) != 16 {
		t.Fatalf("expandRule md5-16 length failed: %d", len(result2))
	}
}
