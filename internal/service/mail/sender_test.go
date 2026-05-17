package mail

import (
	"context"
	"testing"
)

func TestNoopSender(t *testing.T) {
	t.Run("ready", func(t *testing.T) {
		for _, tt := range []struct {
			ready  bool
			expect bool
		}{
			{true, true},
			{false, false},
		} {
			if got := NewNoopSender(tt.ready).Ready(); got != tt.expect {
				t.Errorf("Ready() = %v, want %v", got, tt.expect)
			}
		}
	})

	t.Run("send", func(t *testing.T) {
		sender := NewNoopSender(true)
		err := sender.Send(context.Background(), Message{
			ToEmail: "test@example.com",
			ToName:  "Test User",
			Subject: "Test",
			Text:    "Hello",
		})
		if err != nil {
			t.Errorf("Send() error = %v, want nil", err)
		}
	})
}

func TestNewSenderNilConfig(t *testing.T) {
	sender := NewSender(nil)
	if sender.Ready() {
		t.Error("NewSender(nil) ready = true, want false")
	}
}

func TestFormatAddress(t *testing.T) {
	tests := []struct {
		name  string
		email string
		want  string
	}{
		{"", "test@example.com", "test@example.com"},
		{"User", "test@example.com", "\"User\" <test@example.com>"},
		{"User Name", "test@example.com", "\"User Name\" <test@example.com>"},
	}

	for _, tt := range tests {
		t.Run(tt.email, func(t *testing.T) {
			got := formatAddress(tt.name, tt.email)
			if got != tt.want {
				t.Errorf("formatAddress(%q, %q) = %q, want %q", tt.name, tt.email, got, tt.want)
			}
		})
	}
}
