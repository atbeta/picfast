package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"github.com/pbeta/imgapi/internal/config"
	"github.com/pbeta/imgapi/internal/domain"
	"github.com/pbeta/imgapi/internal/handler"
	"github.com/pbeta/imgapi/internal/router"
	"github.com/pbeta/imgapi/internal/sqlc"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	slog.Info("starting imgapi", "port", cfg.Server.Port)

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
	jwtSvc := handler.NewJWTService(&cfg.JWT)

	var spaHandler *handler.SPAHandler
	webDir := cfg.Server.WebDir
	if webDir == "" {
		webDir = "web-dist"
	}
	if info, err := os.Stat(webDir); err == nil && info.IsDir() {
		spaHandler = handler.NewSPAHandler(os.DirFS(webDir))
		slog.Info("serving frontend", "dir", webDir)
	}

	r := router.New(queries, pool, cfg, jwtSvc, spaHandler)

	seedAdmin(queries, cfg)

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

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down server")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("server forced to shutdown", "error", err)
	}

	slog.Info("server stopped")
}

func runMigrations(dbURL string) {
	m, err := migrate.New("file://migrations", dbURL)
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

func seedAdmin(queries *sqlc.Queries, cfg *config.Config) {
	if cfg.App.AdminEmail == "" || cfg.App.AdminPassword == "" {
		return
	}

	ctx := context.Background()
	existing, _ := queries.GetUserByEmail(ctx, cfg.App.AdminEmail)
	if existing.ID != 0 {
		return
	}

	group, err := queries.GetDefaultGroup(ctx)
	if err != nil {
		slog.Warn("no default group for admin seeding")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(cfg.App.AdminPassword), bcrypt.DefaultCost)
	if err != nil {
		slog.Error("failed to hash admin password", "error", err)
		return
	}

	slog.Info("seeding admin user", "email", cfg.App.AdminEmail)
	_, err = queries.CreateAdminUser(ctx, sqlc.CreateAdminUserParams{
		GroupID:       domain.PgInt8(group.ID),
		Email:         cfg.App.AdminEmail,
		Password:      string(hash),
		Name:          "Admin",
		Role:          string(domain.RoleAdmin),
		CapacityBytes: 0,
		Settings:      []byte(`{}`),
		Status:        int16(domain.UserStatusActive),
		EmailVerified: true,
	})
	if err != nil {
		slog.Error("failed to seed admin", "error", err)
	}
}
