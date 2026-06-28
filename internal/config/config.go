package config

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/spf13/viper"
	"github.com/subosito/gotenv"
)

type Config struct {
	mu          sync.RWMutex
	Server      ServerConfig   `mapstructure:"server"`
	Database    DatabaseConfig `mapstructure:"database"`
	JWT         JWTConfig      `mapstructure:"jwt"`
	Storage     StorageConfig  `mapstructure:"storage"`
	Mail        MailConfig     `mapstructure:"mail"`
	App         AppConfig      `mapstructure:"app"`
	OAuth       OAuthConfig    `mapstructure:"oauth"`
	Webhook     WebhookConfig  `mapstructure:"webhook"`
	SecretKey   string         `mapstructure:"secret_key"`
	secretBytes []byte
}

type ServerConfig struct {
	Port                    int           `mapstructure:"port"`
	MetricsAddr             string        `mapstructure:"metrics_addr"`
	MetricsPort             int           `mapstructure:"metrics_port"` // legacy fallback for metrics_addr
	BaseURL                 string        `mapstructure:"base_url"`
	WebDir                  string        `mapstructure:"web_dir"`
	EnablePprof             bool          `mapstructure:"pprof_enabled"`
	ReadTimeout             time.Duration `mapstructure:"read_timeout"`
	ExpiredCleanupBatchSize int           `mapstructure:"expired_cleanup_batch_size"`
	TrustedProxies          []string      `mapstructure:"trusted_proxies"`
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

const DefaultJWTSecret = "change-me-in-production"

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

type OAuthConfig struct {
	Providers []OAuthProviderConfig `mapstructure:"providers"`
}

type OAuthProviderConfig struct {
	ID           string   `mapstructure:"id"`
	DisplayName  string   `mapstructure:"display_name"`
	Type         string   `mapstructure:"type"` // "oidc" (default) or "github"
	ClientID     string   `mapstructure:"client_id"`
	ClientSecret string   `mapstructure:"client_secret"`
	Issuer       string   `mapstructure:"issuer"`
	AuthURL      string   `mapstructure:"auth_url"`
	TokenURL     string   `mapstructure:"token_url"`
	UserInfoURL  string   `mapstructure:"userinfo_url"` // optional; when set, overrides the userinfo endpoint in both manual and discovery modes
	EmailURL     string   `mapstructure:"email_url"`
	Scopes       []string `mapstructure:"scopes"`
	JWKSURL      string   `mapstructure:"jwks_url"`
	Enabled      bool     `mapstructure:"enabled"`
}

type AppConfig struct {
	Name                     string          `mapstructure:"name"`
	WebBaseURL               string          `mapstructure:"web_base_url"`
	SiteDescription          string          `mapstructure:"site_description"`
	FaviconURL               string          `mapstructure:"favicon_url"`
	AllowGuestUpload         bool            `mapstructure:"allow_guest_upload"`
	GuestCapacityBytes       int64           `mapstructure:"guest_capacity_bytes"`
	AllowRegistration        bool            `mapstructure:"allow_registration"`
	AllowUserImageProcessing bool            `mapstructure:"allow_user_image_processing"`
	SkipImageProcessing      bool            `mapstructure:"skip_image_processing"`
	RequireEmailVerification bool            `mapstructure:"require_email_verification"`
	AuditUploadLogs          bool            `mapstructure:"audit_upload_logs"`
	MaxUploadBytes           int64           `mapstructure:"max_upload_bytes"`
	UserInitialCapacity      int64           `mapstructure:"user_initial_capacity"`
	DefaultImageTTL          time.Duration   `mapstructure:"default_image_ttl"`
	GuestImageTTL            time.Duration   `mapstructure:"guest_image_ttl"`
	AdminEmail               string          `mapstructure:"admin_email"`
	AdminPassword            string          `mapstructure:"admin_password"`
	ModerationMode           string          `mapstructure:"moderation_mode"`
	FooterText1              string          `mapstructure:"footer_text_1"`
	FooterLink1              string          `mapstructure:"footer_link_1"`
	FooterText2              string          `mapstructure:"footer_text_2"`
	FooterLink2              string          `mapstructure:"footer_link_2"`
	AllowOauthRegistration   bool            `mapstructure:"allow_oauth_registration"`
	AnalyticsProvider        string          `mapstructure:"analytics_provider"`
	AnalyticsConfig          json.RawMessage `mapstructure:"analytics_config"`
	ThemeConfig              json.RawMessage `mapstructure:"theme_config"`
	InstanceID               string          `mapstructure:"instance_id"`
}

type WebhookConfig struct {
	Enabled          bool          `mapstructure:"enabled"`
	MaxPerUser       int           `mapstructure:"max_per_user"`
	AllowHTTP        bool          `mapstructure:"allow_http"`
	AllowPrivateURLs bool          `mapstructure:"allow_private_urls"`
	WorkerInterval   time.Duration `mapstructure:"worker_interval"`
	DeliveryTimeout  time.Duration `mapstructure:"delivery_timeout"`
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

func (s *Setter) SetAllowOauthRegistration(v bool) {
	s.cfg.mu.Lock()
	defer s.cfg.mu.Unlock()
	s.cfg.App.AllowOauthRegistration = v
}

func (s *Setter) SetAllowUserImageProcessing(v bool) {
	s.cfg.mu.Lock()
	defer s.cfg.mu.Unlock()
	s.cfg.App.AllowUserImageProcessing = v
}

func (s *Setter) SetSkipImageProcessing(v bool) {
	s.cfg.mu.Lock()
	defer s.cfg.mu.Unlock()
	s.cfg.App.SkipImageProcessing = v
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

func (s *Setter) SetGuestImageTTL(v time.Duration) {
	s.cfg.mu.Lock()
	defer s.cfg.mu.Unlock()
	s.cfg.App.GuestImageTTL = v
}

func (s *Setter) SetModerationMode(mode string) {
	s.cfg.mu.Lock()
	defer s.cfg.mu.Unlock()
	s.cfg.App.ModerationMode = mode
}

func (s *Setter) SetFooterItems(text1, link1, text2, link2 string) {
	s.cfg.mu.Lock()
	defer s.cfg.mu.Unlock()
	s.cfg.App.FooterText1 = text1
	s.cfg.App.FooterLink1 = link1
	s.cfg.App.FooterText2 = text2
	s.cfg.App.FooterLink2 = link2
}

func (s *Setter) SetAnalytics(provider string, cfg json.RawMessage) {
	s.cfg.mu.Lock()
	defer s.cfg.mu.Unlock()
	s.cfg.App.AnalyticsProvider = provider
	s.cfg.App.AnalyticsConfig = cfg
}

func (s *Setter) SetThemeConfig(cfg json.RawMessage) {
	s.cfg.mu.Lock()
	defer s.cfg.mu.Unlock()
	s.cfg.App.ThemeConfig = cfg
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

func (c *Config) UsesDefaultJWTSecret() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.JWT.Secret == DefaultJWTSecret
}

func (c *Config) SecretKeyBytes() []byte {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if len(c.secretBytes) == 32 {
		return c.secretBytes
	}
	key := sha256.Sum256([]byte(c.JWT.Secret))
	return key[:]
}

func (c *Config) SetSecretKeyBytes(key []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.secretBytes = key
}

func (c *Config) EnabledOAuthProviders() []OAuthProviderConfig {
	c.mu.RLock()
	defer c.mu.RUnlock()
	var out []OAuthProviderConfig
	for _, p := range c.OAuth.Providers {
		if p.Enabled {
			out = append(out, p)
		}
	}
	return out
}

func (c *Config) OAuthProviderByID(id string) (OAuthProviderConfig, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for _, p := range c.OAuth.Providers {
		if p.ID == id && p.Enabled {
			return p, true
		}
	}
	return OAuthProviderConfig{}, false
}

func (c *Config) OAuthProviderList() []map[string]string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make([]map[string]string, 0, len(c.OAuth.Providers))
	for _, p := range c.OAuth.Providers {
		if !p.Enabled {
			continue
		}
		out = append(out, map[string]string{
			"id":           p.ID,
			"display_name": p.DisplayName,
		})
	}
	return out
}

func (c MailConfig) IsConfigured() bool {
	return strings.TrimSpace(c.Host) != "" &&
		c.Port > 0 &&
		strings.TrimSpace(c.FromEmail) != ""
}

func Load() (*Config, error) {
	v := viper.New()

	// Load optional local .env for `go run ./cmd/picfast`.
	// Missing file is fine; process env vars still take precedence.
	_ = gotenv.Load(".env")

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
	if v.IsSet("app.theme_config") {
		if raw := strings.TrimSpace(v.GetString("app.theme_config")); raw != "" {
			cfg.App.ThemeConfig = json.RawMessage(raw)
		}
	}
	if len(cfg.App.ThemeConfig) == 0 {
		cfg.App.ThemeConfig = json.RawMessage(`{}`)
	}
	if strings.TrimSpace(cfg.Server.MetricsAddr) == "" {
		cfg.Server.MetricsAddr = fmt.Sprintf("127.0.0.1:%d", cfg.Server.MetricsPort)
	}

	return &cfg, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.metrics_addr", "")
	v.SetDefault("server.metrics_port", 9090)
	v.SetDefault("server.base_url", "http://localhost:8080")
	v.SetDefault("server.web_dir", "")
	v.SetDefault("server.pprof_enabled", false)
	v.SetDefault("server.read_timeout", 60*time.Second)
	v.SetDefault("server.expired_cleanup_batch_size", 100)

	v.SetDefault("database.url", "postgres://picfast:picfast@localhost:5432/picfast?sslmode=disable")

	v.SetDefault("jwt.secret", DefaultJWTSecret)
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
	v.SetDefault("app.web_base_url", "")
	v.SetDefault("app.site_description", "Modern self-hosted image hosting.")
	v.SetDefault("app.favicon_url", "")
	v.SetDefault("app.allow_guest_upload", false)
	v.SetDefault("app.guest_capacity_bytes", int64(10737418240))
	v.SetDefault("app.allow_registration", false)
	v.SetDefault("app.allow_oauth_registration", false)
	v.SetDefault("app.allow_user_image_processing", true)
	v.SetDefault("app.skip_image_processing", false)
	v.SetDefault("app.require_email_verification", false)
	v.SetDefault("app.audit_upload_logs", false)
	v.SetDefault("app.max_upload_bytes", int64(50<<20))
	v.SetDefault("app.user_initial_capacity", int64(524288000))
	v.SetDefault("app.default_image_ttl", time.Duration(0))
	v.SetDefault("app.guest_image_ttl", time.Duration(0))
	v.SetDefault("app.admin_email", "")
	v.SetDefault("app.admin_password", "")
	v.SetDefault("app.moderation_mode", "")
	v.SetDefault("app.footer_text_1", "")
	v.SetDefault("app.footer_link_1", "")
	v.SetDefault("app.footer_text_2", "")
	v.SetDefault("app.footer_link_2", "")
	v.SetDefault("app.analytics_provider", "")
	v.SetDefault("app.instance_id", "")

	v.SetDefault("webhook.enabled", true)
	v.SetDefault("webhook.max_per_user", 10)
	v.SetDefault("webhook.allow_http", false)
	v.SetDefault("webhook.allow_private_urls", false)
	v.SetDefault("webhook.worker_interval", 5*time.Second)
	v.SetDefault("webhook.delivery_timeout", 10*time.Second)

	v.SetDefault("secret_key", "")
}
