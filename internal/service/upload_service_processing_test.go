package service

import (
	"bytes"
	"testing"

	"github.com/atbeta/picfast/internal/config"
	"github.com/atbeta/picfast/internal/domain"
)

func TestResolveImageProcessingConfigDefaults(t *testing.T) {
	cfg := resolveImageProcessingConfig(config.AppConfig{AllowUserImageProcessing: true}, nil)
	if cfg.Quality != 85 || cfg.StripExif != true || cfg.Format != "" || cfg.EnableWatermark {
		t.Fatalf("unexpected defaults: %+v", cfg)
	}
}

func TestResolveImageProcessingConfigRespectsUserSettings(t *testing.T) {
	enable := true
	strip := false
	quality := 72
	format := "webp"
	watermark := &domain.WatermarkConfig{Text: "PicFast"}
	cfg := resolveImageProcessingConfig(config.AppConfig{AllowUserImageProcessing: true}, &domain.UserImageProcessingSettings{
		ImageSaveQuality:  &quality,
		ImageSaveFormat:   &format,
		IsStripExif:       &strip,
		IsEnableWatermark: &enable,
		WatermarkConfigs:  watermark,
	})
	if cfg.Quality != 72 || cfg.Format != "webp" || cfg.StripExif != false || !cfg.EnableWatermark {
		t.Fatalf("unexpected user config: %+v", cfg)
	}
	if cfg.WatermarkConfigs == nil || cfg.WatermarkConfigs.Text != "PicFast" {
		t.Fatalf("unexpected watermark config: %+v", cfg.WatermarkConfigs)
	}
}

func TestResolveImageProcessingConfigIgnoresUserWhenDisabled(t *testing.T) {
	quality := 50
	cfg := resolveImageProcessingConfig(config.AppConfig{AllowUserImageProcessing: false}, &domain.UserImageProcessingSettings{
		ImageSaveQuality: &quality,
	})
	if cfg.Quality != 85 {
		t.Fatalf("quality = %d, want 85", cfg.Quality)
	}
}

func TestResolveImageProcessingConfigSkipsWhenEnabled(t *testing.T) {
	cfg := resolveImageProcessingConfig(config.AppConfig{SkipImageProcessing: true}, nil)
	if !cfg.SkipProcessing {
		t.Fatal("expected SkipProcessing to be true")
	}
	// user settings should be overridden even when user processing is allowed
	quality := 50
	cfg2 := resolveImageProcessingConfig(config.AppConfig{AllowUserImageProcessing: true, SkipImageProcessing: true}, &domain.UserImageProcessingSettings{
		ImageSaveQuality: &quality,
	})
	if !cfg2.SkipProcessing {
		t.Fatal("expected SkipProcessing to be true even with user settings")
	}
}

// minimalPNG is a valid 1×1 white PNG.
var minimalPNG = []byte{
	0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, // signature
	0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52, // IHDR chunk
	0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01, // 1×1
	0x08, 0x02, 0x00, 0x00, 0x00, 0x90, 0x77, 0x53, // RGBA + CRC
	0xDE, 0x00, 0x00, 0x00, 0x0C, 0x49, 0x44, 0x41, // IDAT chunk
	0x54, 0x08, 0xD7, 0x63, 0xF8, 0xCF, 0xC0, 0x00, // compressed data
	0x00, 0x00, 0x02, 0x00, 0x01, 0xE2, 0x21, 0xBC, // ...
	0x33, 0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4E, // IEND chunk
	0x44, 0xAE, 0x42, 0x60, 0x82,
}

func TestProcessUploadImageSkipsAllProcessing(t *testing.T) {
	cfg := imageProcessingConfig{
		SkipProcessing: true,
		Quality:        1,
		StripExif:      true,
	}
	result := processUploadImage(minimalPNG, "png", cfg)
	if !bytes.Equal(result.data, minimalPNG) {
		t.Fatal("expected raw bytes unchanged with SkipProcessing")
	}
	if result.processed {
		t.Fatal("expected processed=false with SkipProcessing")
	}
	// dimensions should be read from the image header
	if result.width != 1 || result.height != 1 {
		t.Fatalf("expected dimensions 1x1, got %dx%d", result.width, result.height)
	}
}

func TestReadImageDimensions(t *testing.T) {
	w, h, err := ReadImageDimensions(minimalPNG)
	if err != nil {
		t.Fatalf("unexpected error for valid PNG: %v", err)
	}
	if w != 1 || h != 1 {
		t.Fatalf("expected 1x1, got %dx%d", w, h)
	}

	_, _, err = ReadImageDimensions([]byte("not an image"))
	if err == nil {
		t.Fatal("expected error for invalid data")
	}
}
