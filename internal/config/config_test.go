package config

import "testing"

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
