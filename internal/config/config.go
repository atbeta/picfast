package config

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	mu       sync.RWMutex
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	Storage  StorageConfig  `mapstructure:"storage"`
	Mail     MailConfig     `mapstructure:"mail"`
	App      AppConfig      `mapstructure:"app"`
}

type ServerConfig struct {
	Port        int    `mapstructure:"port"`
	MetricsPort int    `mapstructure:"metrics_port"`
	BaseURL     string `mapstructure:"base_url"`
	WebDir      string `mapstructure:"web_dir"`
	EnablePprof bool   `mapstructure:"pprof_enabled"`
}

type DatabaseConfig struct {
	URL string `mapstructure:"url"`
}

type JWTConfig struct {
	Secret        string        `mapstructure:"secret"`
	AccessTTL     time.Duration `mapstructure:"access_ttl"`
	RefreshTTL    time.Duration `mapstructure:"refresh_ttl"`
	SigningMethod string        `mapstructure:"signing_method"` // HS256, HS384, HS512
}

type StorageConfig struct {
	LocalRoot    string `mapstructure:"local_root"`
	ThumbnailDir string `mapstructure:"thumbnail_dir"`
}

type MailConfig struct {
	Host       string `mapstructure:"host"`
	Port       int    `mapstructure:"port"`
	Username   string `mapstructure:"username"`
	Password   string `mapstructure:"password"`
	FromEmail  string `mapstructure:"from_email"`
	FromName   string `mapstructure:"from_name"`
	Encryption string `mapstructure:"encryption"` // starttls, tls, none
}

type AppConfig struct {
	Name                     string          `mapstructure:"name"`
	SiteDescription          string          `mapstructure:"site_description"`
	FaviconURL               string          `mapstructure:"favicon_url"`
	AllowGuestUpload         bool            `mapstructure:"allow_guest_upload"`
	GuestCapacityBytes       int64           `mapstructure:"guest_capacity_bytes"`
	AllowRegistration        bool            `mapstructure:"allow_registration"`
	AllowUserImageProcessing bool            `mapstructure:"allow_user_image_processing"`
	RequireEmailVerification bool            `mapstructure:"require_email_verification"`
	AuditUploadLogs          bool            `mapstructure:"audit_upload_logs"`
	UserInitialCapacity      int64           `mapstructure:"user_initial_capacity"`
	DefaultImageTTL          time.Duration   `mapstructure:"default_image_ttl"`
	AdminEmail               string          `mapstructure:"admin_email"`
	AdminPassword            string          `mapstructure:"admin_password"`
	ModerationMode           string          `mapstructure:"moderation_mode"` // disabled, manual, auto
	ICPNumber                string          `mapstructure:"icp_number"`
	ICPLink                  string          `mapstructure:"icp_link"`
	PSBNumber                string          `mapstructure:"psb_number"`
	PSBLink                  string          `mapstructure:"psb_link"`
	AnalyticsProvider        string          `mapstructure:"analytics_provider"`
	AnalyticsConfig          json.RawMessage `mapstructure:"analytics_config"`
}

type Setter struct {
	cfg *Config
}

func NewSetter(cfg *Config) *Setter {
	return &Setter{cfg: cfg}
}

func (s *Setter) SetAppName(name string) {
	s.cfg.mu.Lock()
	defer s.cfg.mu.Unlock()
	s.cfg.App.Name = name
}

func (s *Setter) SetSiteDescription(description string) {
	s.cfg.mu.Lock()
	defer s.cfg.mu.Unlock()
	s.cfg.App.SiteDescription = description
}

func (s *Setter) SetFaviconURL(url string) {
	s.cfg.mu.Lock()
	defer s.cfg.mu.Unlock()
	s.cfg.App.FaviconURL = strings.TrimSpace(url)
}

func (s *Setter) SetBaseURL(url string) {
	s.cfg.mu.Lock()
	defer s.cfg.mu.Unlock()
	s.cfg.Server.BaseURL = url
}

func (s *Setter) SetAllowGuestUpload(v bool) {
	s.cfg.mu.Lock()
	defer s.cfg.mu.Unlock()
	s.cfg.App.AllowGuestUpload = v
}

func (s *Setter) SetGuestCapacityBytes(v int64) {
	s.cfg.mu.Lock()
	defer s.cfg.mu.Unlock()
	s.cfg.App.GuestCapacityBytes = v
}

func (s *Setter) SetAllowRegistration(v bool) {
	s.cfg.mu.Lock()
	defer s.cfg.mu.Unlock()
	s.cfg.App.AllowRegistration = v
}

func (s *Setter) SetAllowUserImageProcessing(v bool) {
	s.cfg.mu.Lock()
	defer s.cfg.mu.Unlock()
	s.cfg.App.AllowUserImageProcessing = v
}

func (s *Setter) SetRequireEmailVerification(v bool) {
	s.cfg.mu.Lock()
	defer s.cfg.mu.Unlock()
	s.cfg.App.RequireEmailVerification = v
}

func (s *Setter) SetUserInitialCapacity(v int64) {
	s.cfg.mu.Lock()
	defer s.cfg.mu.Unlock()
	s.cfg.App.UserInitialCapacity = v
}

func (s *Setter) SetDefaultImageTTL(v time.Duration) {
	s.cfg.mu.Lock()
	defer s.cfg.mu.Unlock()
	s.cfg.App.DefaultImageTTL = v
}

func (s *Setter) SetModerationMode(mode string) {
	s.cfg.mu.Lock()
	defer s.cfg.mu.Unlock()
	s.cfg.App.ModerationMode = mode
}

func (s *Setter) SetFiling(icpNumber, icpLink, psbNumber, psbLink string) {
	s.cfg.mu.Lock()
	defer s.cfg.mu.Unlock()
	s.cfg.App.ICPNumber = icpNumber
	s.cfg.App.ICPLink = icpLink
	s.cfg.App.PSBNumber = psbNumber
	s.cfg.App.PSBLink = psbLink
}

func (s *Setter) SetAnalytics(provider string, cfg json.RawMessage) {
	s.cfg.mu.Lock()
	defer s.cfg.mu.Unlock()
	s.cfg.App.AnalyticsProvider = provider
	s.cfg.App.AnalyticsConfig = cfg
}

func (c *Config) AppSnapshot() AppConfig {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.App
}

func (c *Config) ServerSnapshot() ServerConfig {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Server
}

func (c *Config) RuntimeSnapshot() (ServerConfig, AppConfig) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Server, c.App
}

func (c MailConfig) IsConfigured() bool {
	return strings.TrimSpace(c.Host) != "" &&
		c.Port > 0 &&
		strings.TrimSpace(c.FromEmail) != ""
}

func Load() (*Config, error) {
	v := viper.New()

	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("/etc/picfast")

	v.SetEnvPrefix("PICFAST")
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
	if v.IsSet("app.analytics_config") {
		if raw := strings.TrimSpace(v.GetString("app.analytics_config")); raw != "" {
			cfg.App.AnalyticsConfig = json.RawMessage(raw)
		}
	}

	if cfg.JWT.Secret == "change-me-in-production" {
		return nil, fmt.Errorf("jwt.secret must be changed from default value; set PICFAST_JWT_SECRET environment variable or configure jwt.secret in config.yaml")
	}

	return &cfg, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.metrics_port", 9090)
	v.SetDefault("server.base_url", "http://localhost:8080")
	v.SetDefault("server.web_dir", "")
	v.SetDefault("server.pprof_enabled", false)

	v.SetDefault("database.url", "postgres://picfast:picfast@localhost:5432/picfast?sslmode=disable")

	v.SetDefault("jwt.secret", "change-me-in-production")
	v.SetDefault("jwt.access_ttl", 15*time.Minute)
	v.SetDefault("jwt.refresh_ttl", 168*time.Hour)
	v.SetDefault("jwt.signing_method", "HS256")

	v.SetDefault("storage.local_root", "./data/uploads")
	v.SetDefault("storage.thumbnail_dir", "./data/thumbnails")

	v.SetDefault("mail.host", "")
	v.SetDefault("mail.port", 587)
	v.SetDefault("mail.username", "")
	v.SetDefault("mail.password", "")
	v.SetDefault("mail.from_email", "")
	v.SetDefault("mail.from_name", "PicFast")
	v.SetDefault("mail.encryption", "starttls")

	v.SetDefault("app.name", "PicFast")
	v.SetDefault("app.site_description", "PicFast is a modern self-hosted image hosting service.")
	v.SetDefault("app.favicon_url", "")
	v.SetDefault("app.allow_guest_upload", false)
	v.SetDefault("app.guest_capacity_bytes", int64(10737418240))
	v.SetDefault("app.allow_registration", false)
	v.SetDefault("app.allow_user_image_processing", true)
	v.SetDefault("app.require_email_verification", false)
	v.SetDefault("app.audit_upload_logs", false)
	v.SetDefault("app.user_initial_capacity", int64(524288000))
	v.SetDefault("app.default_image_ttl", time.Duration(0))
	v.SetDefault("app.admin_email", "")
	v.SetDefault("app.admin_password", "")
	v.SetDefault("app.moderation_mode", "")
	v.SetDefault("app.icp_number", "")
	v.SetDefault("app.icp_link", "https://beian.miit.gov.cn/")
	v.SetDefault("app.psb_number", "")
	v.SetDefault("app.psb_link", "")
	v.SetDefault("app.analytics_provider", "")
}
