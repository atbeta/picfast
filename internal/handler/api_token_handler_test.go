package handler

import (
	"testing"
	"time"
)

func TestParseTokenExpiry(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{name: "empty means no expiry", input: "", want: false},
		{name: "never means no expiry", input: "never", want: false},
		{name: "go duration", input: "24h", want: true},
		{name: "days shorthand", input: "30d", want: true},
		{name: "years shorthand", input: "1y", want: true},
		{name: "invalid", input: "soon", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := parseTokenExpiry(tt.input)
			if ok != tt.want {
				t.Fatalf("ok = %v, want %v", ok, tt.want)
			}
			if ok && !got.After(time.Now()) {
				t.Fatalf("expiry = %s, want future time", got)
			}
		})
	}
}
