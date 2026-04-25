package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pbeta/imgapi/internal/config"
	"github.com/pbeta/imgapi/internal/handler"
	"github.com/pbeta/imgapi/internal/handler/middleware"
	"github.com/pbeta/imgapi/internal/sqlc"
)

func New(
	queries *sqlc.Queries,
	pool *pgxpool.Pool,
	cfg *config.Config,
	jwtSvc *handler.JWTService,
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

	authHandler := handler.NewAuthHandler(queries, pool, jwtSvc, cfg)
	userHandler := handler.NewUserHandler(queries)
	adminGroupHandler := handler.NewAdminGroupHandler(queries)
	adminStrategyHandler := handler.NewAdminStrategyHandler(queries)
	adminUserHandler := handler.NewAdminUserHandler(queries)
	adminImageHandler := handler.NewAdminImageHandler(queries)
	adminSettingHandler := handler.NewAdminSettingHandler(cfg, config.NewSetter(cfg))

	loginLimiter := middleware.NewRateLimiter(3, 60*1e9)

	r.Route("/api/v1", func(r chi.Router) {
		// Auth
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", authHandler.Register)
			r.With(middleware.RateLimit(loginLimiter, func(r *http.Request) string { return r.RemoteAddr })).
				Post("/login", authHandler.Login)
			r.Post("/refresh", authHandler.Refresh)
			r.Post("/logout", authHandler.Logout)
		})

		// Authenticated user routes
		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth(jwtSvc))

			r.Get("/users/me", userHandler.GetProfile)
			r.Put("/users/me", userHandler.UpdateProfile)

			// Strategies (available to user's group)
			r.Get("/strategies", adminStrategyHandler.List)
		})

		// Admin routes
		r.Route("/admin", func(r chi.Router) {
			r.Use(middleware.Auth(jwtSvc))
			r.Use(middleware.Admin)

			// Users
			r.Get("/users", adminUserHandler.List)
			r.Get("/users/{id}", adminUserHandler.Get)
			r.Put("/users/{id}", adminUserHandler.Update)
			r.Delete("/users/{id}", adminUserHandler.Delete)

			// Groups
			r.Get("/groups", adminGroupHandler.List)
			r.Get("/groups/{id}", adminGroupHandler.Get)
			r.Post("/groups", adminGroupHandler.Create)
			r.Put("/groups/{id}", adminGroupHandler.Update)
			r.Delete("/groups/{id}", adminGroupHandler.Delete)
			r.Put("/groups/{id}/strategies", adminGroupHandler.SetStrategies)

			// Strategies
			r.Get("/strategies", adminStrategyHandler.List)
			r.Get("/strategies/{id}", adminStrategyHandler.Get)
			r.Post("/strategies", adminStrategyHandler.Create)
			r.Put("/strategies/{id}", adminStrategyHandler.Update)
			r.Delete("/strategies/{id}", adminStrategyHandler.Delete)

			// Images
			r.Get("/images", adminImageHandler.List)
			r.Delete("/images/{id}", adminImageHandler.Delete)

			// Settings
			r.Get("/settings", adminSettingHandler.Get)
			r.Put("/settings", adminSettingHandler.Update)
		})
	})

	return r
}
