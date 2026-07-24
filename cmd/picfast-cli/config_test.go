package main

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestSaveLoadConfigRoundTrip(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("PICFAST_CONFIG_DIR", dir)

	want := fileConfig{URL: "https://img.example.com", Token: "secret-token"}
	if err := saveConfig(want); err != nil {
		t.Fatalf("saveConfig: %v", err)
	}

	got, err := loadConfig()
	if err != nil {
		t.Fatalf("loadConfig: %v", err)
	}
	if got.URL != want.URL || got.Token != want.Token {
		t.Fatalf("loadConfig = %+v, want %+v", got, want)
	}

	path := filepath.Join(dir, "config.json")
	st, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat config: %v", err)
	}
	// Unix permission bits are meaningful; Windows ACLs ignore mode bits.
	if runtime.GOOS != "windows" {
		if perm := st.Mode().Perm(); perm&0o077 != 0 {
			t.Fatalf("config permissions too open: %04o", perm)
		}
	}
}

func TestLoadConfigMissingIsEmpty(t *testing.T) {
	t.Setenv("PICFAST_CONFIG_DIR", t.TempDir())

	got, err := loadConfig()
	if err != nil {
		t.Fatalf("loadConfig: %v", err)
	}
	if got.URL != "" || got.Token != "" {
		t.Fatalf("expected empty config, got %+v", got)
	}
}

func TestResolveCredentialsPrefersEnv(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("PICFAST_CONFIG_DIR", dir)
	if err := saveConfig(fileConfig{URL: "https://from-file.example", Token: "file-token"}); err != nil {
		t.Fatalf("saveConfig: %v", err)
	}

	t.Setenv("PICFAST_URL", "https://from-env.example/")
	t.Setenv("PICFAST_TOKEN", "env-token")

	url, token, err := resolveCredentials()
	if err != nil {
		t.Fatalf("resolveCredentials: %v", err)
	}
	if url != "https://from-env.example" {
		t.Fatalf("url = %q, want env URL without trailing slash", url)
	}
	if token != "env-token" {
		t.Fatalf("token = %q, want env-token", token)
	}
}

func TestResolveCredentialsFallsBackToFile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("PICFAST_CONFIG_DIR", dir)
	t.Setenv("PICFAST_URL", "")
	t.Setenv("PICFAST_TOKEN", "")
	if err := saveConfig(fileConfig{URL: "https://from-file.example", Token: "file-token"}); err != nil {
		t.Fatalf("saveConfig: %v", err)
	}

	url, token, err := resolveCredentials()
	if err != nil {
		t.Fatalf("resolveCredentials: %v", err)
	}
	if url != "https://from-file.example" || token != "file-token" {
		t.Fatalf("got url=%q token=%q", url, token)
	}
}

func TestMaskToken(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"", "(not set)"},
		{"short", "*****"},
		{"abcdefghij", "ab******ij"},
		{"pf_1234567890abcdef", "pf***************ef"},
	}
	for _, tc := range cases {
		if got := maskToken(tc.in); got != tc.want {
			t.Fatalf("maskToken(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestUnsetConfigKey(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("PICFAST_CONFIG_DIR", dir)
	if err := saveConfig(fileConfig{URL: "https://x.example", Token: "tok"}); err != nil {
		t.Fatalf("saveConfig: %v", err)
	}
	if err := unsetConfigKey("token"); err != nil {
		t.Fatalf("unsetConfigKey: %v", err)
	}
	got, err := loadConfig()
	if err != nil {
		t.Fatalf("loadConfig: %v", err)
	}
	if got.Token != "" || got.URL != "https://x.example" {
		t.Fatalf("after unset token: %+v", got)
	}
}
