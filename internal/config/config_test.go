package config

import (
	"sync"
	"testing"
	"time"
)

func TestLoadReadsAdminCredentialsFromEnv(t *testing.T) {
	tempDir := t.TempDir()
	t.Chdir(tempDir)

	t.Setenv("PICFAST_JWT_SECRET", "test-secret")
	t.Setenv("PICFAST_APP_ADMIN_EMAIL", "admin@example.com")
	t.Setenv("PICFAST_APP_ADMIN_PASSWORD", "super-secret")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.App.AdminEmail != "admin@example.com" {
		t.Fatalf("expected admin email from env, got %q", cfg.App.AdminEmail)
	}

	if cfg.App.AdminPassword != "super-secret" {
		t.Fatalf("expected admin password from env, got %q", cfg.App.AdminPassword)
	}
}

func TestLoadAllowsDefaultJWTSecretForQuickstart(t *testing.T) {
	tempDir := t.TempDir()
	t.Chdir(tempDir)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if !cfg.UsesDefaultJWTSecret() {
		t.Fatal("expected default JWT secret to be reported")
	}
}

func TestLoadReadsPprofEnabledFromEnv(t *testing.T) {
	tempDir := t.TempDir()
	t.Chdir(tempDir)

	t.Setenv("PICFAST_JWT_SECRET", "test-secret")
	t.Setenv("PICFAST_SERVER_PPROF_ENABLED", "true")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if !cfg.Server.EnablePprof {
		t.Fatal("expected pprof to be enabled from env")
	}
}

func TestLoadReadsImageKeyMinLengthFromEnv(t *testing.T) {
	tempDir := t.TempDir()
	t.Chdir(tempDir)

	t.Setenv("PICFAST_JWT_SECRET", "test-secret")
	t.Setenv("PICFAST_APP_IMAGE_KEY_MIN_LENGTH", "8")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if cfg.App.ImageKeyMinLength != 8 {
		t.Fatalf("ImageKeyMinLength = %d, want 8", cfg.App.ImageKeyMinLength)
	}
}

func TestImageKeyMinLengthDefaultIsFour(t *testing.T) {
	tempDir := t.TempDir()
	t.Chdir(tempDir)

	t.Setenv("PICFAST_JWT_SECRET", "test-secret")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if cfg.App.ImageKeyMinLength != 4 {
		t.Fatalf("ImageKeyMinLength = %d, want default 4", cfg.App.ImageKeyMinLength)
	}
}

func TestLoadReadsMetricsAddrFromEnv(t *testing.T) {
	tempDir := t.TempDir()
	t.Chdir(tempDir)

	t.Setenv("PICFAST_JWT_SECRET", "test-secret")
	t.Setenv("PICFAST_SERVER_METRICS_ADDR", ":9190")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.Server.MetricsAddr != ":9190" {
		t.Fatalf("MetricsAddr = %q, want :9190", cfg.Server.MetricsAddr)
	}
}

func TestLoadFallsBackToLegacyMetricsPort(t *testing.T) {
	tempDir := t.TempDir()
	t.Chdir(tempDir)

	t.Setenv("PICFAST_JWT_SECRET", "test-secret")
	t.Setenv("PICFAST_SERVER_METRICS_PORT", "9191")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.Server.MetricsAddr != "127.0.0.1:9191" {
		t.Fatalf("MetricsAddr = %q, want 127.0.0.1:9191", cfg.Server.MetricsAddr)
	}
}

func TestLoadReadsSiteMetadataFromEnv(t *testing.T) {
	tempDir := t.TempDir()
	t.Chdir(tempDir)

	t.Setenv("PICFAST_JWT_SECRET", "test-secret")
	t.Setenv("PICFAST_APP_SITE_DESCRIPTION", "Private image hosting")
	t.Setenv("PICFAST_APP_FAVICON_URL", "https://img.example.com/favicon.ico")
	t.Setenv("PICFAST_APP_ANALYTICS_PROVIDER", "umami")
	t.Setenv("PICFAST_APP_ANALYTICS_CONFIG", `{"website_id":"site-1"}`)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if cfg.App.SiteDescription != "Private image hosting" {
		t.Fatalf("SiteDescription = %q", cfg.App.SiteDescription)
	}
	if cfg.App.FaviconURL != "https://img.example.com/favicon.ico" {
		t.Fatalf("FaviconURL = %q", cfg.App.FaviconURL)
	}
	if cfg.App.AnalyticsProvider != "umami" {
		t.Fatalf("AnalyticsProvider = %q", cfg.App.AnalyticsProvider)
	}
	if string(cfg.App.AnalyticsConfig) != `{"website_id":"site-1"}` {
		t.Fatalf("AnalyticsConfig = %s", cfg.App.AnalyticsConfig)
	}
}

func TestRuntimeSnapshotsFollowSetterUpdates(t *testing.T) {
	cfg := &Config{}
	setter := NewSetter(cfg)

	setter.SetAppName("PicFast Pro")
	setter.SetBaseURL("https://img.example.com")
	setter.SetAllowGuestUpload(true)
	setter.SetGuestCapacityBytes(2048)
	setter.SetAllowUserImageProcessing(false)
	setter.SetDefaultImageTTL(time.Hour)

	server, app := cfg.RuntimeSnapshot()
	if server.BaseURL != "https://img.example.com" {
		t.Fatalf("BaseURL = %q", server.BaseURL)
	}
	if app.Name != "PicFast Pro" {
		t.Fatalf("Name = %q", app.Name)
	}
	if !app.AllowGuestUpload {
		t.Fatalf("AllowGuestUpload = false")
	}
	if app.GuestCapacityBytes != 2048 {
		t.Fatalf("GuestCapacityBytes = %d", app.GuestCapacityBytes)
	}
	if app.AllowUserImageProcessing {
		t.Fatalf("AllowUserImageProcessing = true")
	}
	if app.DefaultImageTTL != time.Hour {
		t.Fatalf("DefaultImageTTL = %s", app.DefaultImageTTL)
	}
}

func TestRuntimeSnapshotsCanRunConcurrentlyWithSetters(t *testing.T) {
	cfg := &Config{}
	setter := NewSetter(cfg)

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			setter.SetAllowRegistration(true)
			setter.SetAllowRegistration(false)
		}()
		go func() {
			defer wg.Done()
			_ = cfg.AppSnapshot()
			_, _ = cfg.RuntimeSnapshot()
		}()
	}
	wg.Wait()
}
