package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type fileConfig struct {
	URL   string `json:"url,omitempty"`
	Token string `json:"token,omitempty"`
}

func configDir() (string, error) {
	if override := strings.TrimSpace(os.Getenv("PICFAST_CONFIG_DIR")); override != "" {
		return override, nil
	}
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "picfast"), nil
}

func configPath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

func loadConfig() (fileConfig, error) {
	path, err := configPath()
	if err != nil {
		return fileConfig{}, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fileConfig{}, nil
		}
		return fileConfig{}, err
	}
	var cfg fileConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return fileConfig{}, fmt.Errorf("parse config: %w", err)
	}
	cfg.URL = strings.TrimRight(strings.TrimSpace(cfg.URL), "/")
	cfg.Token = strings.TrimSpace(cfg.Token)
	return cfg, nil
}

func saveConfig(cfg fileConfig) error {
	dir, err := configDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	path, err := configPath()
	if err != nil {
		return err
	}
	cfg.URL = strings.TrimRight(strings.TrimSpace(cfg.URL), "/")
	cfg.Token = strings.TrimSpace(cfg.Token)
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	// Best-effort on platforms where chmod matters (no-op-ish on Windows ACLs).
	_ = os.Chmod(path, 0o600)
	return nil
}

func unsetConfigKey(key string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}
	switch key {
	case "url":
		cfg.URL = ""
	case "token":
		cfg.Token = ""
	default:
		return fmt.Errorf("unknown config key %q (want url or token)", key)
	}
	return saveConfig(cfg)
}

func resolveCredentials() (baseURL, token string, err error) {
	cfg, err := loadConfig()
	if err != nil {
		return "", "", err
	}
	baseURL = strings.TrimRight(strings.TrimSpace(os.Getenv("PICFAST_URL")), "/")
	if baseURL == "" {
		baseURL = cfg.URL
	}
	token = strings.TrimSpace(os.Getenv("PICFAST_TOKEN"))
	if token == "" {
		token = cfg.Token
	}
	return baseURL, token, nil
}

func maskToken(token string) string {
	if token == "" {
		return "(not set)"
	}
	if len(token) < 8 {
		return "*****"
	}
	return token[:2] + strings.Repeat("*", len(token)-4) + token[len(token)-2:]
}
