package handler

import "testing"

func TestParseImageVariant(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantExt    string
		wantWidth  int
		wantStyled bool
	}{
		{name: "plain ext", input: "jpg", wantExt: "jpg", wantWidth: 0, wantStyled: false},
		{name: "width variant", input: "jpg@w_300", wantExt: "jpg", wantWidth: 300, wantStyled: true},
		{name: "invalid variant ignored", input: "jpg@foo", wantExt: "jpg", wantWidth: 0, wantStyled: true},
		{name: "invalid width ignored", input: "jpg@w_0", wantExt: "jpg", wantWidth: 0, wantStyled: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ext, width, styled := parseImageVariant(tt.input)
			if ext != tt.wantExt || width != tt.wantWidth || styled != tt.wantStyled {
				t.Fatalf("parseImageVariant(%q) = (%q,%d,%v), want (%q,%d,%v)",
					tt.input, ext, width, styled, tt.wantExt, tt.wantWidth, tt.wantStyled)
			}
		})
	}
}

