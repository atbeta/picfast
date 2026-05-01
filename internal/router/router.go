package router

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"time"

	"github.com/atbeta/picfast/internal/config"
	"github.com/atbeta/picfast/internal/handler"
	"github.com/atbeta/picfast/internal/handler/middleware"
	"github.com/atbeta/picfast/internal/service"
	mailservice "github.com/atbeta/picfast/internal/service/mail"
	"github.com/atbeta/picfast/internal/service/moderation"
	"github.com/atbeta/picfast/internal/sqlc"
	"github.com/atbeta/picfast/internal/version"
	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/modelcontextprotocol/go-sdk/auth"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func New(
	queries *sqlc.Queries,
	pool *pgxpool.Pool,
	cfg *config.Config,
	jwtSvc *handler.JWTService,
	spaHandler *handler.SPAHandler,
	mailSender mailservice.Sender,
) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.StructuredLogger)
	r.Use(chimw.Recoverer)
	r.Use(chimw.RealIP)
	r.Use(chimw.RequestID)
	r.Use(chimw.Timeout(60 * time.Second))
	r.Use(middleware.Metrics)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		status := map[string]interface{}{
			"status":    "ok",
			"database":  "ok",
			"storage":   "ok",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		}
		code := http.StatusOK

		if err := pool.Ping(r.Context()); err != nil {
			status["status"] = "degraded"
			status["database"] = err.Error()
			code = http.StatusServiceUnavailable
		}

		for _, dir := range []string{cfg.Storage.LocalRoot, cfg.Storage.ThumbnailDir} {
			if dir == "" {
				continue
			}
			if _, err := os.Stat(dir); err != nil {
				if os.IsNotExist(err) {
					if mkErr := os.MkdirAll(dir, 0755); mkErr != nil {
						status["status"] = "degraded"
						status["storage"] = "cannot create " + dir + ": " + mkErr.Error()
						code = http.StatusServiceUnavailable
					}
				} else {
					status["status"] = "degraded"
					status["storage"] = "cannot access " + dir + ": " + err.Error()
					code = http.StatusServiceUnavailable
				}
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)
		if err := json.NewEncoder(w).Encode(status); err != nil {
			slog.Warn("failed to write health response", "error", err)
		}
	})

	r.Get("/metrics", promhttp.Handler().ServeHTTP)
	r.Get("/openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		spec, err := os.ReadFile("api/openapi.yaml")
		if err != nil {
			http.Error(w, "openapi spec not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/yaml; charset=utf-8")
		if _, err := w.Write(spec); err != nil {
			slog.Warn("failed to write openapi spec", "error", err)
		}
	})
	r.Get("/docs", func(w http.ResponseWriter, r *http.Request) {
		server := cfg.ServerSnapshot()
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if _, err := w.Write([]byte(fmt.Sprintf(`<!doctype html>
<html>
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>PicFast API Docs</title>
  </head>
  <body>
    <script id="api-reference" data-url="%s"></script>
    <script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
  </body>
</html>`, server.BaseURL+"/openapi.yaml"))); err != nil {
			slog.Warn("failed to write docs page", "error", err)
		}
	})

	// Services
	uploadSvc := service.NewUploadService(queries, pool, cfg)
	deleteSvc := service.NewDeleteService(queries, pool, cfg.Storage.ThumbnailDir)

	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			deleted, err := deleteSvc.CleanExpiredImages(context.Background(), 100)
			if err != nil {
				slog.Warn("failed to clean expired images", "error", err)
			} else if deleted > 0 {
				slog.Info("cleaned expired images", "deleted", deleted)
			}
		}
	}()

	// Handlers
	authHandler := handler.NewAuthHandler(queries, pool, jwtSvc, cfg, mailSender)
	userHandler := handler.NewUserHandler(queries)
	imageHandler := handler.NewImageHandler(queries, uploadSvc, deleteSvc, cfg.Server.BaseURL, cfg.App.AuditUploadLogs)
	albumHandler := handler.NewAlbumHandler(queries, pool)
	fileHandler := handler.NewFileHandler(queries, cfg.Server.BaseURL, cfg.Storage.ThumbnailDir)
	adminGroupHandler := handler.NewAdminGroupHandler(queries, pool)
	adminStrategyHandler := handler.NewAdminStrategyHandler(queries)
	adminUserHandler := handler.NewAdminUserHandler(queries)
	adminImageHandler := handler.NewAdminImageHandler(queries, deleteSvc, cfg.Server.BaseURL)
	adminSettingHandler := handler.NewAdminSettingHandler(cfg, config.NewSetter(cfg), queries, mailSender != nil && mailSender.Ready())
	adminAuditHandler := handler.NewAdminAuditHandler(queries)
	adminObservabilityHandler := handler.NewAdminObservabilityHandler(queries, pool, cfg, mailSender != nil && mailSender.Ready())

	// MCP Server
	mcpFactory := handler.NewMCPServerFactory(queries, pool, cfg)
	mcpServer := mcpFactory.CreateServer()
	mcpAuth := handler.NewMCPAuth(queries)
	mcpHandler := mcp.NewStreamableHTTPHandler(func(r *http.Request) *mcp.Server {
		return mcpServer
	}, nil)

	// Content Moderation
	app := cfg.AppSnapshot()
	modMode, err := moderation.ParseMode(app.ModerationMode)
	if err != nil {
		slog.Warn("invalid moderation mode, disabling moderation", "mode", app.ModerationMode, "error", err)
	}
	if modMode == "" {
		modMode = moderation.ModeDisabled
	}
	moderator, err := moderation.New(modMode, queries)
	if err != nil {
		slog.Warn("failed to create moderator", "mode", modMode, "error", err)
		moderator = moderation.NewNoopModerator()
	}
	modMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r.WithContext(moderation.WithModerator(r.Context(), moderator)))
		})
	}

	loginLimiter := middleware.NewRateLimiter(5, 60*1e9)

	// Image file serving: /i/{key}.{ext} — with OptionalAuth so private images can be accessed by owner
	r.With(middleware.OptionalAuth(middleware.NewJWTAuthenticator(jwtSvc))).Get("/i/{key}.{ext}", fileHandler.ServeImage)

	// Thumbnail serving: /t/{hash}.png
	r.With(middleware.OptionalAuth(middleware.NewJWTAuthenticator(jwtSvc))).Get("/t/{hash}.png", fileHandler.ServeThumbnail)

	// Ensure thumbnail directory exists
	if err := os.MkdirAll(cfg.Storage.ThumbnailDir, 0755); err != nil {
		slog.Warn("failed to ensure thumbnail directory", "dir", cfg.Storage.ThumbnailDir, "error", err)
	}

	r.Route("/api/v1", func(r chi.Router) {
		// Public site config
		r.Get("/config", func(w http.ResponseWriter, r *http.Request) {
			server, app := cfg.RuntimeSnapshot()
			handler.Success(w, map[string]interface{}{
				"app_name":                   app.Name,
				"site_description":           app.SiteDescription,
				"allow_guest_upload":         app.AllowGuestUpload,
				"allow_registration":         app.AllowRegistration,
				"require_email_verification": app.RequireEmailVerification && mailSender != nil && mailSender.Ready(),
				"base_url":                   server.BaseURL,
				"icp_number":                 app.ICPNumber,
				"icp_link":                   app.ICPLink,
				"psb_number":                 app.PSBNumber,
				"psb_link":                   app.PSBLink,
				"analytics_provider":         app.AnalyticsProvider,
				"analytics_config":           normalizeJSON(app.AnalyticsConfig),
				"github_url":                 version.DefaultGitHubURL(),
			})
		})
		r.Get("/version", func(w http.ResponseWriter, r *http.Request) {
			handler.Success(w, version.Info())
		})

		// Auth
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", authHandler.Register)
			r.Post("/verify-email", authHandler.VerifyEmail)
			r.Post("/resend-verification", authHandler.ResendVerification)
			r.With(middleware.RateLimit(loginLimiter, func(r *http.Request) string {
				host, _, err := net.SplitHostPort(r.RemoteAddr)
				if err != nil || host == "" {
					return r.RemoteAddr
				}
				return host
			})).
				Post("/login", authHandler.Login)
			r.Post("/refresh", authHandler.Refresh)
		})

		// Authenticated user routes
		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth(middleware.NewJWTAuthenticator(jwtSvc)))

			r.Post("/auth/logout", authHandler.Logout)

			r.Get("/users/me", userHandler.GetProfile)
			r.Put("/users/me", userHandler.UpdateProfile)

			// Images — upload injects moderator into context
			r.With(modMiddleware).Post("/images", imageHandler.Upload)
			r.Get("/images", imageHandler.List)
			r.Get("/images/{key}", imageHandler.Get)
			r.Delete("/images/{key}", imageHandler.Delete)
			r.Patch("/images/{key}", imageHandler.Update)

			// Albums
			r.Get("/albums", albumHandler.List)
			r.Post("/albums", albumHandler.Create)
			r.Put("/albums/{id}", albumHandler.Update)
			r.Delete("/albums/{id}", albumHandler.Delete)

			// Strategies (available to user's group)
			r.Get("/strategies", adminStrategyHandler.List)

			// API Tokens (for MCP / AI integration)
			apiTokenHandler := handler.NewAPITokenHandler(queries)
			r.Post("/api-tokens", apiTokenHandler.Create)
			r.Get("/api-tokens", apiTokenHandler.List)
			r.Delete("/api-tokens/{id}", apiTokenHandler.Delete)
		})

		// Optional auth for guest upload
		r.With(middleware.OptionalAuth(middleware.NewJWTAuthenticator(jwtSvc))).
			Post("/upload", imageHandler.Upload)

		// ShareX endpoints
		sharexHandler := handler.NewShareXHandler(uploadSvc, cfg.Server.BaseURL)
		r.With(middleware.OptionalAuth(middleware.NewJWTAuthenticator(jwtSvc)), modMiddleware).Post("/sharex/upload", sharexHandler.Upload)
		r.Get("/sharex/config", sharexHandler.Config)

		// Admin routes
		r.Route("/admin", func(r chi.Router) {
			r.Use(middleware.Auth(middleware.NewJWTAuthenticator(jwtSvc)))

			r.Use(middleware.Admin)

			r.Get("/users", adminUserHandler.List)
			r.Get("/users/{id}", adminUserHandler.Get)
			r.Put("/users/{id}", adminUserHandler.Update)
			r.Delete("/users/{id}", adminUserHandler.Delete)

			r.Get("/groups", adminGroupHandler.List)
			r.Get("/groups/{id}", adminGroupHandler.Get)
			r.Post("/groups", adminGroupHandler.Create)
			r.Put("/groups/{id}", adminGroupHandler.Update)
			r.Delete("/groups/{id}", adminGroupHandler.Delete)
			r.Put("/groups/{id}/strategies", adminGroupHandler.SetStrategies)

			r.Get("/strategies", adminStrategyHandler.List)
			r.Get("/strategies/{id}", adminStrategyHandler.Get)
			r.Post("/strategies", adminStrategyHandler.Create)
			r.Put("/strategies/{id}", adminStrategyHandler.Update)
			r.Delete("/strategies/{id}", adminStrategyHandler.Delete)

			// Strategy health check
			r.Get("/strategies/health", func(w http.ResponseWriter, r *http.Request) {
				strategies, err := queries.ListStrategies(r.Context())
				if err != nil {
					slog.Warn("failed to list strategies for health check", "error", err)
					handler.Fail(w, http.StatusInternalServerError, "failed to list strategies")
					return
				}
				results := make([]map[string]interface{}, 0, len(strategies))
				for _, st := range strategies {
					store, err := service.GetStorageForStrategy(st)
					item := map[string]interface{}{
						"id":      st.ID,
						"name":    st.Name,
						"type":    st.StrategyType,
						"healthy": false,
					}
					if err != nil {
						item["error"] = "failed to init: " + err.Error()
					} else {
						defer store.Close()
						health := store.HealthCheck(r.Context())
						item["healthy"] = health.Healthy
						if health.Error != "" {
							item["error"] = health.Error
						}
					}
					results = append(results, item)
				}
				handler.Success(w, results)
			})

			r.Get("/images", adminImageHandler.List)
			r.Delete("/images/{id}", adminImageHandler.Delete)

			// Content moderation (admin only)
			modHandler := handler.NewModerationHandler(queries)
			r.Get("/moderation/pending", modHandler.ListPending)
			r.Post("/moderation/{id}/approve", modHandler.Approve)
			r.Post("/moderation/{id}/reject", modHandler.Reject)

			r.Get("/settings", adminSettingHandler.Get)
			r.Put("/settings", adminSettingHandler.Update)
			r.Get("/audit-logs", adminAuditHandler.List)
			r.Get("/observability/summary", adminObservabilityHandler.Summary)

			if cfg.ServerSnapshot().EnablePprof {
				// Debug / pprof (admin only)
				r.Route("/debug", func(r chi.Router) {
					r.Get("/pprof/", func(w http.ResponseWriter, req *http.Request) {
						pprof.Index(w, req)
					})
					r.Get("/pprof/cmdline", func(w http.ResponseWriter, req *http.Request) {
						pprof.Cmdline(w, req)
					})
					r.Get("/pprof/profile", func(w http.ResponseWriter, req *http.Request) {
						pprof.Profile(w, req)
					})
					r.Get("/pprof/symbol", func(w http.ResponseWriter, req *http.Request) {
						pprof.Symbol(w, req)
					})
					r.Get("/pprof/trace", func(w http.ResponseWriter, req *http.Request) {
						pprof.Trace(w, req)
					})
					r.Get("/pprof/goroutine", pprof.Handler("goroutine").ServeHTTP)
					r.Get("/pprof/heap", pprof.Handler("heap").ServeHTTP)
					r.Get("/pprof/allocs", pprof.Handler("allocs").ServeHTTP)
					r.Get("/pprof/threadcreate", pprof.Handler("threadcreate").ServeHTTP)
					r.Get("/pprof/block", pprof.Handler("block").ServeHTTP)
					r.Get("/pprof/mutex", pprof.Handler("mutex").ServeHTTP)
				})
			}
		})
	})

	// MCP endpoint (AI integration) — protected by API token auth
	mcpVerifier := auth.TokenVerifier(mcpAuth.VerifyToken)
	r.Mount("/mcp", auth.RequireBearerToken(mcpVerifier, nil)(mcpHandler))

	// SPA frontend — must be last so API routes take priority
	if spaHandler != nil {
		r.NotFound(spaHandler.Serve)
	}

	return r
}

func normalizeJSON(raw json.RawMessage) json.RawMessage {
	if len(raw) == 0 || string(raw) == "null" {
		return json.RawMessage(`{}`)
	}
	return raw
}
