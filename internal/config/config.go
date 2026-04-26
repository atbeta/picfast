package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	Storage  StorageConfig  `mapstructure:"storage"`
	App      AppConfig      `mapstructure:"app"`
}

type ServerConfig struct {
	Port      int    `mapstructure:"port"`
	BaseURL   string `mapstructure:"base_url"`
	WebDir    string `mapstructure:"web_dir"`
}

type DatabaseConfig struct {
	URL string `mapstructure:"url"`
}

type JWTConfig struct {
	Secret     string        `mapstructure:"secret"`
	AccessTTL  time.Duration `mapstructure:"access_ttl"`
	RefreshTTL time.Duration `mapstructure:"refresh_ttl"`
}

type StorageConfig struct {
	LocalRoot    string `mapstructure:"local_root"`
	ThumbnailDir string `mapstructure:"thumbnail_dir"`
}

type AppConfig struct {
	Name                string `mapstructure:"name"`
	AllowGuestUpload    bool   `mapstructure:"allow_guest_upload"`
	AllowRegistration   bool   `mapstructure:"allow_registration"`
	UserInitialCapacity int64  `mapstructure:"user_initial_capacity"`
	AdminEmail          string `mapstructure:"admin_email"`
	AdminPassword       string `mapstructure:"admin_password"`
	ModerationMode      string `mapstructure:"moderation_mode"` // disabled, manual, auto
}

type Setter struct {
	cfg *Config
}

func NewSetter(cfg *Config) *Setter {
	return &Setter{cfg: cfg}
}

func (s *Setter) SetAppName(name string)         { s.cfg.App.Name = name }
func (s *Setter) SetAllowGuestUpload(v bool)     { s.cfg.App.AllowGuestUpload = v }
func (s *Setter) SetAllowRegistration(v bool)    { s.cfg.App.AllowRegistration = v }
func (s *Setter) SetUserInitialCapacity(v int64) { s.cfg.App.UserInitialCapacity = v }
func (s *Setter) SetModerationMode(mode string)  { s.cfg.App.ModerationMode = mode }

func Load() (*Config, error) {
	v := viper.New()

	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("/etc/imgapi")

	v.SetEnvPrefix("IMGAPI")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	setDefaults(v)

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("read config: %w", err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	return &cfg, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.base_url", "http://localhost:8080")
	v.SetDefault("server.web_dir", "")

	v.SetDefault("database.url", "postgres://imgapi:imgapi@localhost:5432/imgapi?sslmode=disable")

	v.SetDefault("jwt.secret", "change-me-in-production")
	v.SetDefault("jwt.access_ttl", 15*time.Minute)
	v.SetDefault("jwt.refresh_ttl", 168*time.Hour)

	v.SetDefault("storage.local_root", "./data/uploads")
	v.SetDefault("storage.thumbnail_dir", "./data/thumbnails")

	v.SetDefault("app.name", "ImageAPI")
	v.SetDefault("app.allow_guest_upload", true)
	v.SetDefault("app.allow_registration", true)
	v.SetDefault("app.user_initial_capacity", int64(524288000))
	v.SetDefault("app.moderation_mode", "disabled")
}
