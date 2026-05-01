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

func TestRuntimeSnapshotsFollowSetterUpdates(t *testing.T) {
	cfg := &Config{}
	setter := NewSetter(cfg)

	setter.SetAppName("PicFast Pro")
	setter.SetBaseURL("https://img.example.com")
	setter.SetAllowGuestUpload(true)
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
