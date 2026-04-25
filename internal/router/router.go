package router

import (
	"net"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pbeta/imgapi/internal/config"
	"github.com/pbeta/imgapi/internal/handler"
	"github.com/pbeta/imgapi/internal/handler/middleware"
	"github.com/pbeta/imgapi/internal/service"
	"github.com/pbeta/imgapi/internal/sqlc"
)

func New(
	queries *sqlc.Queries,
	pool *pgxpool.Pool,
	cfg *config.Config,
	jwtSvc *handler.JWTService,
	spaHandler *handler.SPAHandler,
) http.Handler {
	r := chi.NewRouter()

	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(chimw.RealIP)
	r.Use(chimw.RequestID)
	r.Use(chimw.Timeout(60))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	// Services
	uploadSvc := service.NewUploadService(queries, pool, cfg)
	deleteSvc := service.NewDeleteService(queries, cfg.Storage.ThumbnailDir)

	// Handlers
	authHandler := handler.NewAuthHandler(queries, pool, jwtSvc, cfg)
	userHandler := handler.NewUserHandler(queries)
	imageHandler := handler.NewImageHandler(queries, uploadSvc, deleteSvc, cfg.Server.BaseURL)
	albumHandler := handler.NewAlbumHandler(queries)
	fileHandler := handler.NewFileHandler(queries, cfg.Server.BaseURL, cfg.Storage.ThumbnailDir)
	adminGroupHandler := handler.NewAdminGroupHandler(queries, pool)
	adminStrategyHandler := handler.NewAdminStrategyHandler(queries)
	adminUserHandler := handler.NewAdminUserHandler(queries)
	adminImageHandler := handler.NewAdminImageHandler(queries, deleteSvc)
	adminSettingHandler := handler.NewAdminSettingHandler(cfg, config.NewSetter(cfg))

	loginLimiter := middleware.NewRateLimiter(3, 60*1e9)

	// Image file serving: /i/{key}.{ext} — with OptionalAuth so private images can be accessed by owner
	r.With(middleware.OptionalAuth(jwtSvc)).Get("/i/{key}.{ext}", fileHandler.ServeImage)

	// Thumbnail serving: /t/{hash}.png
	r.Get("/t/{hash}.png", fileHandler.ServeThumbnail)

	// Ensure thumbnail directory exists
	os.MkdirAll(cfg.Storage.ThumbnailDir, 0755)

	r.Route("/api/v1", func(r chi.Router) {
		// Auth
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", authHandler.Register)
			r.With(middleware.RateLimit(loginLimiter, func(r *http.Request) string {
				host, _, _ := net.SplitHostPort(r.RemoteAddr)
				if host == "" {
					return r.RemoteAddr
				}
				return host
			})).
				Post("/login", authHandler.Login)
			r.Post("/refresh", authHandler.Refresh)
			r.Post("/logout", authHandler.Logout)
		})

		// Authenticated user routes
		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth(jwtSvc))

			r.Get("/users/me", userHandler.GetProfile)
			r.Put("/users/me", userHandler.UpdateProfile)

			// Images
			r.Post("/images", imageHandler.Upload)
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
		})

		// Optional auth for guest upload
		r.With(middleware.OptionalAuth(jwtSvc)).
			Post("/upload", imageHandler.Upload)

		// Admin routes
		r.Route("/admin", func(r chi.Router) {
			r.Use(middleware.Auth(jwtSvc))
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

			r.Get("/images", adminImageHandler.List)
			r.Delete("/images/{id}", adminImageHandler.Delete)

			r.Get("/settings", adminSettingHandler.Get)
			r.Put("/settings", adminSettingHandler.Update)
		})
	})

	// SPA frontend — must be last so API routes take priority
	if spaHandler != nil {
		r.NotFound(spaHandler.Serve)
	}

	return r
}
