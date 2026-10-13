package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source"
	"github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/atbeta/picfast/internal/clientip"
	"github.com/atbeta/picfast/internal/config"
	"github.com/atbeta/picfast/internal/handler"
	"github.com/atbeta/picfast/internal/router"
	"github.com/atbeta/picfast/internal/service"
	mailservice "github.com/atbeta/picfast/internal/service/mail"
	"github.com/atbeta/picfast/internal/sitesettings"
	"github.com/atbeta/picfast/internal/sqlc"
	"github.com/atbeta/picfast/migrations"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "maintenance":
			os.Exit(runMaintenanceCommand(context.Background(), os.Args[2:], os.Stdout, os.Stderr))
		case "migrate":
			os.Exit(runMigrateCommand(context.Background(), os.Args[2:], os.Stdout, os.Stderr))
		case "jobs":
			os.Exit(runJobsCommand(context.Background(), os.Args[2:], os.Stdout, os.Stderr))
		}
	}

	service.StartupImaging()
	defer service.ShutdownImaging()
	slog.Info("image backend", "backend", service.Capabilities().Backend)

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

	if cfg.Database.AutoMigrate {
		runMigrations(cfg.Database.URL)
	} else {
		slog.Info("auto migrate disabled; run `picfast migrate` before starting")
	}

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
	settingsUpdatedAt, err := sitesettings.Load(context.Background(), queries, cfg)
	if err != nil {
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

	r, runner := router.New(queries, pool, cfg, jwtSvc, spaHandler, mailSender,
		sitesettings.NewRefresher(queries, cfg, settingsUpdatedAt))

	if cfg.Jobs.Embedded {
		runner.Start()
	} else {
		slog.Info("embedded background jobs disabled; trigger POST /internal/jobs/run or `picfast jobs run` on a schedule")
	}

	var metricsSrv *http.Server
	if cfg.Server.MetricsEnabled {
		metricsMux := http.NewServeMux()
		metricsMux.Handle("/metrics", promhttp.Handler())

		metricsSrv = &http.Server{
			Addr:         cfg.Server.MetricsAddr,
			Handler:      metricsMux,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  60 * time.Second,
		}
	} else {
		slog.Info("metrics server disabled", "addr", cfg.Server.MetricsAddr)
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

	if metricsSrv != nil {
		go func() {
			slog.Info("metrics server listening", "addr", metricsSrv.Addr)
			if err := metricsSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				slog.Error("metrics server error", "error", err)
			}
		}()
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down server")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	runner.Stop()

	if metricsSrv != nil {
		if err := metricsSrv.Shutdown(ctx); err != nil {
			slog.Error("metrics server forced to shutdown", "error", err)
		}
	}
	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("server forced to shutdown", "error", err)
	}

	slog.Info("server stopped")
}

// migrateDatabaseURL 把 postgres:// 形式的 DSN 改写成 golang-migrate pgx/v5
// 驱动注册的 pgx5:// scheme，避免引入 lib/pq 驱动（其有未修复的 CVE）。
func migrateDatabaseURL(dbURL string) string {
	for _, scheme := range []string{"postgres://", "postgresql://"} {
		if strings.HasPrefix(dbURL, scheme) {
			return "pgx5://" + strings.TrimPrefix(dbURL, scheme)
		}
	}
	return dbURL
}

func runMigrations(dbURL string) {
	m, err := newMigrator(dbURL)
	if err != nil {
		slog.Error("migration source error", "error", err)
		os.Exit(1)
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		slog.Error("migration failed", "error", err)
		os.Exit(1)
	}
	slog.Info("migrations applied")
}

// newMigrator returns a configured migrator; shared by server startup and the
// `picfast migrate` subcommand.
func newMigrator(dbURL string) (*migrate.Migrate, error) {
	name, src, err := newMigrationSource()
	if err != nil {
		return nil, err
	}
	return migrate.NewWithSourceInstance(name, src, migrateDatabaseURL(dbURL))
}

// newMigrationSource resolves the migration source in order:
// PICFAST_MIGRATIONS_DIR (when it points at a directory) → well-known disk
// directories → the migration files embedded in the binary.
func newMigrationSource() (string, source.Driver, error) {
	if dir, ok := resolveMigrationsDir(); ok {
		drv, err := (&file.File{}).Open("file://" + dir)
		if err != nil {
			return "", nil, err
		}
		return "file", drv, nil
	}
	drv, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return "", nil, err
	}
	return "iofs", drv, nil
}

func resolveMigrationsDir() (string, bool) {
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
			continue
		}
		return abs, true
	}

	return "", false
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
