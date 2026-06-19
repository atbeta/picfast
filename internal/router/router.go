package router

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"net/http/pprof"
	"os"
	"strings"
	"time"

	"github.com/atbeta/picfast/internal/clientip"
	"github.com/atbeta/picfast/internal/config"
	"github.com/atbeta/picfast/internal/docsui"
	"github.com/atbeta/picfast/internal/domain"
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
	"go.yaml.in/yaml/v3"
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
	r.Use(middleware.SecurityHeaders(cfg))
	r.Use(chimw.Recoverer)
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), domain.ContextKeyOriginalAddr, r.RemoteAddr)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	})
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if clientip.IsTrustedProxy(r.RemoteAddr) {
				chimw.RealIP(next).ServeHTTP(w, r)
			} else {
				next.ServeHTTP(w, r)
			}
		})
	})
	r.Use(chimw.RequestID)
	if cfg.Server.ReadTimeout > 0 {
		r.Use(chimw.Timeout(cfg.Server.ReadTimeout))
	}
	r.Use(middleware.Metrics)
	r.Use(middleware.RequireSetupCompleteForWrites(queries))

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

	serveOpenAPISpec := func(format string) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, HEAD, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			raw, err := os.ReadFile("api/openapi.yaml")
			if err != nil {
				http.Error(w, "openapi spec not found", http.StatusNotFound)
				return
			}
			server := cfg.ServerSnapshot()
			baseURL := strings.TrimRight(server.BaseURL, "/")
			if baseURL != "" {
				raw = []byte(strings.Replace(string(raw), "http://localhost:8080/api/v1", baseURL+"/api/v1", 1))
			}
			if v := strings.TrimSpace(version.Version); v != "" {
				raw = []byte(strings.ReplaceAll(string(raw), `version: "1.0"`, fmt.Sprintf(`version: "%s"`, v)))
			}

			if format == "json" {
				var spec any
				if err := yaml.Unmarshal(raw, &spec); err != nil {
					http.Error(w, "failed to parse openapi spec", http.StatusInternalServerError)
					return
				}
				w.Header().Set("Content-Type", "application/json; charset=utf-8")
				enc := json.NewEncoder(w)
				enc.SetIndent("", "  ")
				if err := enc.Encode(spec); err != nil {
					slog.Warn("failed to write openapi json spec", "error", err)
				}
				return
			}

			w.Header().Set("Content-Type", "application/yaml; charset=utf-8")
			if _, err := w.Write(raw); err != nil {
				slog.Warn("failed to write openapi spec", "error", err)
			}
		}
	}
	r.Get("/openapi.yaml", serveOpenAPISpec("yaml"))
	r.Head("/openapi.yaml", serveOpenAPISpec("yaml"))
	r.Options("/openapi.yaml", serveOpenAPISpec("yaml"))
	r.Get("/openapi.json", serveOpenAPISpec("json"))
	r.Head("/openapi.json", serveOpenAPISpec("json"))
	r.Options("/openapi.json", serveOpenAPISpec("json"))
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
    <script src="/docs/assets/scalar-api-reference.js"></script>
  </body>
</html>`, server.BaseURL+"/openapi.yaml"))); err != nil {
			slog.Warn("failed to write docs page", "error", err)
		}
	})
	docsAssets, err := fs.Sub(docsui.Assets, "static")
	if err != nil {
		slog.Error("failed to load embedded docs assets", "error", err)
	} else {
		docsFS := http.StripPrefix("/docs/assets/", http.FileServerFS(docsAssets))
		r.Get("/docs/assets/*", func(w http.ResponseWriter, r *http.Request) {
			docsFS.ServeHTTP(w, r)
		})
	}

	// Services
	uploadSvc := service.NewUploadService(queries, pool, cfg)
	deleteSvc := service.NewDeleteService(queries, pool, cfg.Storage.ThumbnailDir)

	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			deleted, err := deleteSvc.CleanExpiredImages(context.Background(), int32(cfg.Server.ExpiredCleanupBatchSize))
			if err != nil {
				slog.Warn("failed to clean expired images", "error", err)
			} else if deleted > 0 {
				slog.Info("cleaned expired images", "deleted", deleted)
			}
		}
	}()

	// Handlers
	authHandler := handler.NewAuthHandler(queries, pool, jwtSvc, cfg, mailSender)
	oauthHandler := handler.NewOAuthHandler(queries, pool, jwtSvc, cfg)
	setupHandler := handler.NewSetupHandler(queries, pool, jwtSvc, cfg)
	userHandler := handler.NewUserHandler(queries)
	imageHandler := handler.NewImageHandler(queries, uploadSvc, deleteSvc, cfg.Server.BaseURL, cfg.App.AuditUploadLogs, cfg.App.MaxUploadBytes)
	albumHandler := handler.NewAlbumHandler(queries, pool)
	fileHandler := handler.NewFileHandler(queries, cfg.Server.BaseURL, cfg.Storage.ThumbnailDir)
	adminGroupHandler := handler.NewAdminGroupHandler(queries, pool)
	adminStrategyHandler := handler.NewAdminStrategyHandler(queries)
	adminUserHandler := handler.NewAdminUserHandler(queries)
	adminImageHandler := handler.NewAdminImageHandler(queries, deleteSvc, cfg.Server.BaseURL)
	adminSettingHandler := handler.NewAdminSettingHandler(cfg, config.NewSetter(cfg), queries, mailSender != nil && mailSender.Ready())
	adminAuditHandler := handler.NewAdminAuditHandler(queries)
	adminObservabilityHandler := handler.NewAdminObservabilityHandler(queries, pool, cfg, mailSender != nil && mailSender.Ready())

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

	loginLimiter := middleware.NewRateLimiter(5, time.Minute)
	mailActionLimiter := middleware.NewRateLimiter(3, 10*time.Minute)
	oauthStartLimiter := middleware.NewRateLimiter(10, time.Minute)
	oauthCallbackLimiter := middleware.NewRateLimiter(5, time.Minute)
	oauthLinkLimiter := middleware.NewRateLimiter(10, time.Minute)
	clientIPKey := func(r *http.Request) string {
		return clientip.FromRequest(r)
	}

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
			setupRequired := false
			if count, err := queries.CountUsers(r.Context()); err != nil {
				handler.Fail(w, http.StatusServiceUnavailable, "failed to read setup status")
				return
			} else {
				setupRequired = count == 0
			}
			providerList := cfg.OAuthProviderList()
			handler.Success(w, map[string]interface{}{
				"app_name":                    app.Name,
				"site_description":            app.SiteDescription,
				"favicon_url":                 app.FaviconURL,
				"allow_guest_upload":          app.AllowGuestUpload,
				"guest_capacity_bytes":        app.GuestCapacityBytes,
				"allow_registration":          app.AllowRegistration,
				"allow_oauth_registration":    app.AllowOauthRegistration,
				"allow_user_image_processing": app.AllowUserImageProcessing,
				"skip_image_processing":       app.SkipImageProcessing,
				"require_email_verification":  app.RequireEmailVerification && mailSender != nil && mailSender.Ready(),
				"base_url":                    server.BaseURL,
				"default_image_ttl":           app.DefaultImageTTL.String(),
				"guest_image_ttl":             app.GuestImageTTL.String(),
				"footer_text_1":               app.FooterText1,
				"footer_link_1":               app.FooterLink1,
				"footer_text_2":               app.FooterText2,
				"footer_link_2":               app.FooterLink2,
				"analytics_provider":          app.AnalyticsProvider,
				"analytics_config":            normalizeJSON(app.AnalyticsConfig),
				"theme_config":                normalizeJSON(app.ThemeConfig),
				"default_copy_format":           app.DefaultCopyFormat,
				"copy_template":                 app.CopyTemplate,
					"github_url":                  version.DefaultGitHubURL(),
				"setup_required":              setupRequired,
				"oauth_providers":             providerList,
			})
		})
		r.Get("/version", func(w http.ResponseWriter, r *http.Request) {
			handler.Success(w, version.Info())
		})

		// First-run setup
		r.Route("/setup", func(r chi.Router) {
			r.Get("/status", setupHandler.Status)
			r.Post("/admin", setupHandler.CreateAdmin)
		})

		// Auth
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", authHandler.Register)
			r.Post("/verify-email", authHandler.VerifyEmail)
			r.With(middleware.RateLimit(mailActionLimiter, clientIPKey)).Post("/resend-verification", authHandler.ResendVerification)
			r.With(middleware.RateLimit(mailActionLimiter, clientIPKey)).Post("/forgot-password", authHandler.ForgotPassword)
			r.Post("/reset-password", authHandler.ResetPassword)
			r.With(middleware.RateLimit(loginLimiter, clientIPKey)).
				Post("/login", authHandler.Login)
			r.Post("/refresh", authHandler.Refresh)

			// OAuth
			r.Route("/oauth", func(r chi.Router) {
				r.Get("/providers", oauthHandler.Providers)
				r.With(middleware.RateLimit(oauthStartLimiter, clientIPKey)).
					Get("/{provider}", oauthHandler.Start)
				r.With(middleware.RateLimit(oauthCallbackLimiter, clientIPKey)).
					Get("/{provider}/callback", oauthHandler.Callback)
			})
		})

		// Authenticated user routes: MCP-accessible routes accept both JWT and API Token;
		// other user routes only accept JWT (e.g. logout, profile update, tokens CRUD).
		// Write operations within the dual-auth group require "write" scope when
		// authenticated via API token; JWT sessions have no scope restriction.
		dualAuth := middleware.DualAuth(middleware.NewJWTAuthenticator(jwtSvc), queries)

		r.Group(func(r chi.Router) {
			r.Use(dualAuth)

			r.Get("/users/me", userHandler.GetProfile)

			// Images — write operations require "write" scope for API tokens
			r.With(modMiddleware, middleware.RequireScope("write")).Post("/images", imageHandler.Upload)
			r.Get("/images", imageHandler.List)
			r.Get("/images/{key}", imageHandler.Get)
			r.With(middleware.RequireScope("write")).Delete("/images/{key}", imageHandler.Delete)
			r.With(middleware.RequireScope("write")).Post("/images/batch-delete", imageHandler.BatchDelete)
			r.With(middleware.RequireScope("write")).Patch("/images/{key}", imageHandler.Update)

			// Albums — write operations require "write" scope for API tokens
			r.Get("/albums", albumHandler.List)
			r.With(middleware.RequireScope("write")).Post("/albums", albumHandler.Create)
			r.With(middleware.RequireScope("write")).Put("/albums/{id}", albumHandler.Update)
			r.With(middleware.RequireScope("write")).Delete("/albums/{id}", albumHandler.Delete)

			// Strategies (available to user's group)
			r.Get("/strategies", adminStrategyHandler.List)
		})

		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth(middleware.NewJWTAuthenticator(jwtSvc)))

			r.Post("/auth/logout", authHandler.Logout)
			r.Put("/users/me", userHandler.UpdateProfile)

			// OAuth link/unlink
			r.With(middleware.RateLimit(oauthLinkLimiter, clientIPKey)).
				Get("/auth/oauth/{provider}/link", oauthHandler.Link)
			r.With(middleware.RateLimit(oauthLinkLimiter, clientIPKey)).
				Delete("/auth/oauth/{provider}", oauthHandler.Unlink)
			r.Get("/auth/oauth/identities", oauthHandler.Identities)

			// API Tokens (for API / MCP integration)
			apiTokenHandler := handler.NewAPITokenHandler(queries)
			r.Post("/api-tokens", apiTokenHandler.Create)
			r.Get("/api-tokens", apiTokenHandler.List)
			r.Delete("/api-tokens/{id}", apiTokenHandler.Delete)
		})

		// Optional auth for guest upload (also accepts API tokens)
		r.With(middleware.OptionalDualAuth(middleware.NewJWTAuthenticator(jwtSvc), queries)).
			Post("/upload", imageHandler.Upload)

			// ShareX endpoints
		sharexHandler := handler.NewShareXHandler(uploadSvc, cfg.Server.BaseURL, cfg.App.MaxUploadBytes)
		r.With(middleware.OptionalDualAuth(middleware.NewJWTAuthenticator(jwtSvc), queries), modMiddleware).Post("/sharex/upload", sharexHandler.Upload)
		r.Get("/sharex/config", sharexHandler.Config)

		// Flat upload endpoints (PicGo, uPic, Dropshare, etc.)
		flatHandler := handler.NewFlatUploadHandler(uploadSvc, cfg.Server.BaseURL, cfg.App.MaxUploadBytes)
		r.With(middleware.OptionalDualAuth(middleware.NewJWTAuthenticator(jwtSvc), queries), modMiddleware).Post("/flat/upload", flatHandler.Upload)

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
			modHandler := handler.NewModerationHandler(queries, cfg.ServerSnapshot().BaseURL)
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
