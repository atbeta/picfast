package service

import (
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
