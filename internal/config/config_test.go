package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

// unsetEnv removes an env var for the duration of the test, restoring any
// previous value afterwards. Use together with t.Chdir'd temp dirs so tests
// are not influenced by the ambient process environment.
func unsetEnv(t *testing.T, key string) {
	t.Helper()
	old, had := os.LookupEnv(key)
	if err := os.Unsetenv(key); err != nil {
		t.Fatalf("unsetenv %s: %v", key, err)
	}
	t.Cleanup(func() {
		if had {
			_ = os.Setenv(key, old)
		}
	})
}

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

func TestLoadReadsPortFromEnv(t *testing.T) {
	tempDir := t.TempDir()
	t.Chdir(tempDir)

	unsetEnv(t, "PICFAST_SERVER_PORT")
	t.Setenv("PICFAST_JWT_SECRET", "test-secret")
	t.Setenv("PORT", "9999")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.Server.Port != 9999 {
		t.Fatalf("Port = %d, want 9999 from $PORT", cfg.Server.Port)
	}
}

func TestLoadPicfastServerPortWinsOverPort(t *testing.T) {
	tempDir := t.TempDir()
	t.Chdir(tempDir)

	t.Setenv("PICFAST_JWT_SECRET", "test-secret")
	t.Setenv("PICFAST_SERVER_PORT", "1234")
	t.Setenv("PORT", "9999")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.Server.Port != 1234 {
		t.Fatalf("Port = %d, want 1234 from PICFAST_SERVER_PORT", cfg.Server.Port)
	}
}

func TestLoadPortDefaultsTo8080(t *testing.T) {
	tempDir := t.TempDir()
	t.Chdir(tempDir)

	unsetEnv(t, "PICFAST_SERVER_PORT")
	unsetEnv(t, "PORT")
	t.Setenv("PICFAST_JWT_SECRET", "test-secret")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.Server.Port != 8080 {
		t.Fatalf("Port = %d, want default 8080", cfg.Server.Port)
	}
}

func TestLoadMetricsEnabledDefaultsTrue(t *testing.T) {
	tempDir := t.TempDir()
	t.Chdir(tempDir)

	unsetEnv(t, "PICFAST_SERVER_METRICS_ENABLED")
	t.Setenv("PICFAST_JWT_SECRET", "test-secret")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if !cfg.Server.MetricsEnabled {
		t.Fatal("expected metrics_enabled to default to true")
	}
}

func TestLoadReadsMetricsEnabledFromEnv(t *testing.T) {
	tempDir := t.TempDir()
	t.Chdir(tempDir)

	t.Setenv("PICFAST_JWT_SECRET", "test-secret")
	t.Setenv("PICFAST_SERVER_METRICS_ENABLED", "false")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.Server.MetricsEnabled {
		t.Fatal("expected metrics_enabled to be disabled from env")
	}
}

func TestLoadAutoMigrateDefaultsTrue(t *testing.T) {
	tempDir := t.TempDir()
	t.Chdir(tempDir)

	unsetEnv(t, "PICFAST_DATABASE_AUTO_MIGRATE")
	t.Setenv("PICFAST_JWT_SECRET", "test-secret")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if !cfg.Database.AutoMigrate {
		t.Fatal("expected auto_migrate to default to true")
	}
}

func TestLoadReadsAutoMigrateFromEnv(t *testing.T) {
	tempDir := t.TempDir()
	t.Chdir(tempDir)

	t.Setenv("PICFAST_JWT_SECRET", "test-secret")
	t.Setenv("PICFAST_DATABASE_AUTO_MIGRATE", "false")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.Database.AutoMigrate {
		t.Fatal("expected auto_migrate to be disabled from env")
	}
}

func TestLoadStorageDefaultTypeDefaultsLocal(t *testing.T) {
	tempDir := t.TempDir()
	t.Chdir(tempDir)

	unsetEnv(t, "PICFAST_STORAGE_DEFAULT_TYPE")
	unsetEnv(t, "PICFAST_STORAGE_DEFAULT_CONFIG")
	t.Setenv("PICFAST_JWT_SECRET", "test-secret")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.Storage.DefaultType != "local" {
		t.Fatalf("DefaultType = %q, want local", cfg.Storage.DefaultType)
	}
	if cfg.Storage.DefaultConfig != nil {
		t.Fatalf("DefaultConfig = %s, want nil", cfg.Storage.DefaultConfig)
	}
}

func TestLoadReadsStorageDefaultsFromEnv(t *testing.T) {
	tempDir := t.TempDir()
	t.Chdir(tempDir)

	t.Setenv("PICFAST_JWT_SECRET", "test-secret")
	t.Setenv("PICFAST_STORAGE_DEFAULT_TYPE", "s3")
	t.Setenv("PICFAST_STORAGE_DEFAULT_CONFIG", `  {"endpoint":"https://s3.example.com","bucket":"pics"} `)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.Storage.DefaultType != "s3" {
		t.Fatalf("DefaultType = %q, want s3", cfg.Storage.DefaultType)
	}
	if string(cfg.Storage.DefaultConfig) != `{"endpoint":"https://s3.example.com","bucket":"pics"}` {
		t.Fatalf("DefaultConfig = %s, want trimmed JSON", cfg.Storage.DefaultConfig)
	}
}

func TestLoadStorageDefaultConfigFromYAMLMap(t *testing.T) {
	tempDir := t.TempDir()
	t.Chdir(tempDir)

	t.Setenv("PICFAST_JWT_SECRET", "test-secret")
	configYAML := `storage:
  default_type: s3
  default_config:
    endpoint: "https://account.r2.cloudflarestorage.com"
    region: auto
    bucket: pics
    access_key: AKID
    secret_key: SECRET
    use_path_style: true
`
	if err := os.WriteFile(filepath.Join(tempDir, "config.yaml"), []byte(configYAML), 0o644); err != nil {
		t.Fatalf("write config.yaml: %v", err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.Storage.DefaultType != "s3" {
		t.Fatalf("DefaultType = %q, want s3", cfg.Storage.DefaultType)
	}
	var parsed map[string]any
	if err := json.Unmarshal(cfg.Storage.DefaultConfig, &parsed); err != nil {
		t.Fatalf("DefaultConfig is not JSON: %v (%s)", err, cfg.Storage.DefaultConfig)
	}
	if parsed["endpoint"] != "https://account.r2.cloudflarestorage.com" {
		t.Fatalf("endpoint = %v", parsed["endpoint"])
	}
	if parsed["use_path_style"] != true {
		t.Fatalf("use_path_style = %v", parsed["use_path_style"])
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

func TestJobsConfigDefaults(t *testing.T) {
	unsetEnv(t, "PICFAST_JOBS_EMBEDDED")
	unsetEnv(t, "PICFAST_JOBS_TOKEN")
	t.Setenv("PICFAST_JWT_SECRET", "test-secret")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if !cfg.Jobs.Embedded {
		t.Fatal("jobs.embedded should default to true")
	}
	if cfg.Jobs.Token != "" {
		t.Fatalf("jobs.token should default to empty, got %q", cfg.Jobs.Token)
	}
}

func TestJobsConfigFromEnv(t *testing.T) {
	t.Setenv("PICFAST_JWT_SECRET", "test-secret")
	t.Setenv("PICFAST_JOBS_EMBEDDED", "false")
	t.Setenv("PICFAST_JOBS_TOKEN", "jobs-secret")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if cfg.Jobs.Embedded {
		t.Fatal("jobs.embedded should be false from env")
	}
	if cfg.Jobs.Token != "jobs-secret" {
		t.Fatalf("jobs.token = %q, want jobs-secret", cfg.Jobs.Token)
	}
}
