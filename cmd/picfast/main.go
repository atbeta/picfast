package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/davidbyttow/govips/v2/vips"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/atbeta/picfast/internal/clientip"
	"github.com/atbeta/picfast/internal/config"
	"github.com/atbeta/picfast/internal/handler"
	"github.com/atbeta/picfast/internal/router"
	mailservice "github.com/atbeta/picfast/internal/service/mail"
	"github.com/atbeta/picfast/internal/sqlc"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "maintenance" {
		os.Exit(runMaintenanceCommand(context.Background(), os.Args[2:], os.Stdout, os.Stderr))
	}

	vips.Startup(&vips.Config{
		ConcurrencyLevel: 2,
		MaxCacheFiles:    0,
		MaxCacheMem:      100 * 1024 * 1024,
		MaxCacheSize:     500,
	})
	defer vips.Shutdown()

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}
	if cfg.UsesDefaultJWTSecret() {
		slog.Warn("insecure default JWT secret is in use; set PICFAST_JWT_SECRET or jwt.secret before production use")
	}
	if cfg.App.RequireEmailVerification && !cfg.Mail.IsConfigured() {
		slog.Warn("email verification is enabled but mail delivery is not configured; the requirement will stay inactive")
	}
	if len(cfg.EnabledOAuthProviders()) > 0 && strings.TrimSpace(cfg.App.WebBaseURL) == "" {
		slog.Warn("OAuth is enabled but app.web_base_url is empty; post-login redirects will likely 404 unless backend serves the SPA. Set PICFAST_APP_WEB_BASE_URL to the frontend origin")
	}
	if len(cfg.EnabledOAuthProviders()) > 0 && strings.TrimSpace(cfg.Server.BaseURL) == "" {
		slog.Error("OAuth requires server.base_url to be set; refusing to start")
		os.Exit(1)
	}
	if strings.HasPrefix(strings.ToLower(cfg.Server.BaseURL), "https://") && len(cfg.Server.TrustedProxies) == 0 {
		slog.Warn("server.base_url is https but no trusted_proxies configured; X-Forwarded-Proto from reverse proxy will be ignored, cookies may not be marked Secure")
	}

	slog.Info("starting picfast", "port", cfg.Server.Port)

	runMigrations(cfg.Database.URL)

	pool, err := pgxpool.New(context.Background(), cfg.Database.URL)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := pool.Ping(context.Background()); err != nil {
		slog.Error("failed to ping database", "error", err)
		os.Exit(1)
	}

	queries := sqlc.New(pool)
	if err := applyPersistedSiteSettings(context.Background(), queries, cfg); err != nil {
		slog.Error("failed to load persisted site settings", "error", err)
		os.Exit(1)
	}
	jwtSvc := handler.NewJWTService(&cfg.JWT)
	clientip.SetTrustedProxies(cfg.Server.TrustedProxies)
	mailSender := mailservice.NewSender(&cfg.Mail)

	if err := bootstrapCoreData(context.Background(), queries, cfg); err != nil {
		slog.Error("failed to bootstrap core data", "error", err)
		os.Exit(1)
	}

	var spaHandler *handler.SPAHandler
	if webDir := resolveWebDir(cfg.Server.WebDir); webDir != "" {
		spaHandler = handler.NewSPAHandler(os.DirFS(webDir))
		slog.Info("serving frontend", "dir", webDir)
	}

	r := router.New(queries, pool, cfg, jwtSvc, spaHandler, mailSender)

	metricsMux := http.NewServeMux()
	metricsMux.Handle("/metrics", promhttp.Handler())

	metricsSrv := &http.Server{
		Addr:         cfg.Server.MetricsAddr,
		Handler:      metricsMux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      r,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		slog.Info("server listening", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	go func() {
		slog.Info("metrics server listening", "addr", metricsSrv.Addr)
		if err := metricsSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("metrics server error", "error", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down server")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := metricsSrv.Shutdown(ctx); err != nil {
		slog.Error("metrics server forced to shutdown", "error", err)
	}
	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("server forced to shutdown", "error", err)
	}

	slog.Info("server stopped")
}

func runMigrations(dbURL string) {
	source, err := resolveMigrationSource()
	if err != nil {
		slog.Warn("migration source error", "error", err)
		return
	}

	m, err := migrate.New(source, dbURL)
	if err != nil {
		slog.Warn("migration source error", "error", err)
		return
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		slog.Error("migration failed", "error", err)
		os.Exit(1)
	}
	slog.Info("migrations applied")
}

func resolveMigrationSource() (string, error) {
	candidates := []string{
		os.Getenv("PICFAST_MIGRATIONS_DIR"),
		"migrations",
		"/migrations",
		"/app/migrations",
	}

	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}
		info, err := os.Stat(candidate)
		if err != nil || !info.IsDir() {
			continue
		}
		abs, err := filepath.Abs(candidate)
		if err != nil {
			return "", err
		}
		return "file://" + abs, nil
	}

	return "", fmt.Errorf("no migrations directory found")
}

func resolveWebDir(configured string) string {
	candidates := []string{}
	if configured != "" {
		candidates = append(candidates, configured)
	}
	candidates = append(candidates, "web-dist", "web/dist")

	for _, dir := range candidates {
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			return dir
		}
	}

	return ""
}

func applyPersistedSiteSettings(ctx context.Context, queries *sqlc.Queries, cfg *config.Config) error {
	settings, err := queries.GetSiteSettings(ctx)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		return err
	}

	cfg.App.Name = settings.AppName
	cfg.Server.BaseURL = settings.AppUrl
	cfg.App.SiteDescription = settings.SiteDescription
	cfg.App.FaviconURL = settings.FaviconUrl
	cfg.App.AllowGuestUpload = settings.AllowGuestUpload
	cfg.App.GuestCapacityBytes = settings.GuestCapacityBytes
	cfg.App.AllowRegistration = settings.AllowRegistration
	cfg.App.AllowOauthRegistration = settings.AllowOauthRegistration
	cfg.App.AllowUserImageProcessing = settings.AllowUserImageProcessing
	cfg.App.SkipImageProcessing = settings.SkipImageProcessing
	cfg.App.RequireEmailVerification = settings.RequireEmailVerification
	cfg.App.UserInitialCapacity = settings.UserInitialCapacity
	if settings.DefaultImageTtl == "" {
		cfg.App.DefaultImageTTL = 0
	} else {
		d, err := time.ParseDuration(settings.DefaultImageTtl)
		if err != nil {
			return fmt.Errorf("parse persisted default_image_ttl: %w", err)
		}
		cfg.App.DefaultImageTTL = d
	}
	if settings.GuestImageTtl == "" {
		cfg.App.GuestImageTTL = 0
	} else {
		d, err := time.ParseDuration(settings.GuestImageTtl)
		if err != nil {
			return fmt.Errorf("parse persisted guest_image_ttl: %w", err)
		}
		cfg.App.GuestImageTTL = d
	}
	cfg.App.ModerationMode = settings.ModerationMode
	cfg.App.FooterText1 = settings.FooterText1
	cfg.App.FooterLink1 = settings.FooterLink1
	cfg.App.FooterText2 = settings.FooterText2
	cfg.App.FooterLink2 = settings.FooterLink2
	cfg.App.AnalyticsProvider = settings.AnalyticsProvider
	cfg.App.AnalyticsConfig = settings.AnalyticsConfig
	cfg.App.ThemeConfig = settings.ThemeConfig
	cfg.App.DefaultCopyFormat = settings.DefaultCopyFormat
	cfg.App.CopyTemplate = settings.CopyTemplate
	return nil
}
