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
	key1 := GenerateImageKey()
	key2 := GenerateImageKey()
	if len(key1) != 6 {
		t.Fatalf("expected key length 6, got %d", len(key1))
	}
	if key1 == key2 {
		t.Fatal("expected unique keys")
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
