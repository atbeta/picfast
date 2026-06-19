package handler

import (
	"testing"

	"github.com/atbeta/picfast/internal/service"
)

func TestParseProcessingParams(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		wantE  string
		wantP  service.ProcessingParams
	}{
		{
			name:  "plain ext",
			input: "jpg",
			wantE: "jpg",
		},
		{
			name:  "width only",
			input: "jpg@w_300",
			wantE: "jpg",
			wantP: service.ProcessingParams{Width: 300},
		},
		{
			name:  "height only",
			input: "png@h_200",
			wantE: "png",
			wantP: service.ProcessingParams{Height: 200},
		},
		{
			name:  "width and height",
			input: "jpg@w_800,h_600",
			wantE: "jpg",
			wantP: service.ProcessingParams{Width: 800, Height: 600},
		},
		{
			name:  "quality only",
			input: "jpg@q_80",
			wantE: "jpg",
			wantP: service.ProcessingParams{Quality: 80},
		},
		{
			name:  "format only",
			input: "jpg@f_webp",
			wantE: "jpg",
			wantP: service.ProcessingParams{Format: "webp"},
		},
		{
			name:  "all params",
			input: "jpg@w_300,h_200,q_60,f_webp",
			wantE: "jpg",
			wantP: service.ProcessingParams{Width: 300, Height: 200, Quality: 60, Format: "webp"},
		},
		{
			name:  "jpg normalised to jpeg",
			input: "png@f_jpg",
			wantE: "png",
			wantP: service.ProcessingParams{Format: "jpeg"},
		},
		{
			name:  "invalid width ignored",
			input: "jpg@w_0,q_80",
			wantE: "jpg",
			wantP: service.ProcessingParams{Quality: 80},
		},
		{
			name:  "quality out of range ignored",
			input: "jpg@q_0,w_300",
			wantE: "jpg",
			wantP: service.ProcessingParams{Width: 300},
		},
		{
			name:  "quality over 100 ignored",
			input: "jpg@q_200",
			wantE: "jpg",
		},
		{
			name:  "width over 10000 ignored",
			input: "jpg@w_99999",
			wantE: "jpg",
		},
		{
			name:  "unknown key ignored",
			input: "jpg@w_300,x_abc",
			wantE: "jpg",
			wantP: service.ProcessingParams{Width: 300},
		},
		{
			name:  "missing value ignored",
			input: "jpg@w_,q_80",
			wantE: "jpg",
			wantP: service.ProcessingParams{Quality: 80},
		},
		{
			name:  "leading dot stripped",
			input: ".jpg@w_300",
			wantE: "jpg",
			wantP: service.ProcessingParams{Width: 300},
		},
		{
			name:  "empty variant",
			input: "jpg@",
			wantE: "jpg",
		},
		{
			name:  "whitespace trimmed",
			input: "  jpg@w_300  ",
			wantE: "jpg",
			wantP: service.ProcessingParams{Width: 300},
		},
		{
			name:  "invalid format ignored",
			input: "jpg@f_foo,w_300",
			wantE: "jpg",
			wantP: service.ProcessingParams{Width: 300},
		},
		{
			name:  "format avif not in whitelist",
			input: "jpg@f_avif",
			wantE: "jpg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ext, params := parseProcessingParams(tt.input)
			if ext != tt.wantE {
				t.Fatalf("ext = %q, want %q", ext, tt.wantE)
			}
			if params != tt.wantP {
				t.Fatalf("params = %+v, want %+v", params, tt.wantP)
			}
		})
	}
}
