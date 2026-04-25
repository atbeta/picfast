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

	loginLimiter := middleware.NewRateLimiter(3, 60*1e9) // 3/min

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", authHandler.Register)
			r.With(middleware.RateLimit(loginLimiter, func(r *http.Request) string { return r.RemoteAddr })).
				Post("/login", authHandler.Login)
			r.Post("/refresh", authHandler.Refresh)
			r.Post("/logout", authHandler.Logout)
		})

		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth(jwtSvc))

			r.Get("/users/me", userHandler.GetProfile)
			r.Put("/users/me", userHandler.UpdateProfile)
		})
	})

	return r
}
