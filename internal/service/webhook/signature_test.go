package webhook

import (
	"encoding/json"
	"testing"
)

func TestGenerateSecret(t *testing.T) {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}

	plain, hash, ciphertext, err := GenerateSecret(key)
	if err != nil {
		t.Fatalf("GenerateSecret failed: %v", err)
	}
	if len(plain) < 6 || plain[:6] != "whsec_" {
		t.Errorf("expected whsec_ prefix, got %s", plain)
	}
	if len(hash) != 64 {
		t.Errorf("expected 64-char hash, got %d", len(hash))
	}
	if len(ciphertext) == 0 {
		t.Error("expected non-empty ciphertext")
	}

	decrypted, err := DecryptSecret(ciphertext, key)
	if err != nil {
		t.Fatalf("DecryptSecret failed: %v", err)
	}
	if decrypted != plain {
		t.Errorf("decrypted %q != plain %q", decrypted, plain)
	}
}

func TestDecryptSecretWrongKey(t *testing.T) {
	key1 := make([]byte, 32)
	key2 := make([]byte, 32)
	key1[0] = 1
	key2[0] = 2

	_, _, ciphertext, err := GenerateSecret(key1)
	if err != nil {
		t.Fatalf("GenerateSecret failed: %v", err)
	}

	_, err = DecryptSecret(ciphertext, key2)
	if err == nil {
		t.Error("expected error with wrong key")
	}
}

func TestComputeSignature(t *testing.T) {
	secret := "whsec_abc123"
	ts := "1234567890"
	body := []byte(`{"type":"test"}`)

	sig := ComputeSignature(secret, ts, body)
	if sig == "" {
		t.Error("expected non-empty signature")
	}
	if sig[:7] != "sha256=" {
		t.Errorf("expected sha256= prefix, got %s", sig[:7])
	}

	sig2 := ComputeSignature(secret, ts, body)
	if sig != sig2 {
		t.Error("signature not deterministic")
	}

	sig3 := ComputeSignature(secret+"x", ts, body)
	if sig == sig3 {
		t.Error("different secret produced same signature")
	}
}

func TestEventMatches(t *testing.T) {
	tests := []struct {
		subscribed []string
		eventType  string
		want       bool
	}{
		{[]string{"image.uploaded", "image.deleted"}, "image.uploaded", true},
		{[]string{"image.uploaded", "image.deleted"}, "image.processed", false},
		{[]string{}, "image.uploaded", false},
		{nil, "image.uploaded", false},
		{[]string{"*"}, "image.uploaded", false},
	}

	for _, tt := range tests {
		got := EventMatches(tt.subscribed, tt.eventType)
		if got != tt.want {
			t.Errorf("EventMatches(%v, %s) = %v, want %v", tt.subscribed, tt.eventType, got, tt.want)
		}
	}
}

func TestNormalizeEvents(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
		err   bool
	}{
		{"valid array", `["image.uploaded", "image.deleted"]`, []string{"image.uploaded", "image.deleted"}, false},
		{"empty array", `[]`, nil, false},
		{"null", `null`, nil, false},
		{"empty string", ``, nil, false},
		{"invalid json", `{`, nil, true},
		{"single event", `["image.uploaded"]`, []string{"image.uploaded"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NormalizeEvents(json.RawMessage(tt.input))
			if tt.err && err == nil {
				t.Error("expected error")
			}
			if !tt.err && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !tt.err {
				if len(got) != len(tt.want) {
					t.Errorf("got %v, want %v", got, tt.want)
				} else {
					for i := range got {
						if got[i] != tt.want[i] {
							t.Errorf("got[%d] = %s, want %s", i, got[i], tt.want[i])
						}
					}
				}
			}
		})
	}
}
