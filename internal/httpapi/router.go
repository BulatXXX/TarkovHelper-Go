package httpapi

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"

	"tarkovhelper-api/internal/config"
	"tarkovhelper-api/internal/repo"
)

func NewRouter(cfg config.Config, pool *pgxpool.Pool) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(corsMiddleware(CORSOptions{
		AllowedOrigins: []string{"http://localhost:4200"},
	}))
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)

	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(15 * time.Second))
	r.Use(jsonMiddleware)

	userRepo := repo.NewUserRepo(pool)
	trackedRepo := repo.NewTrackedRepo(pool)

	// health
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, map[string]string{"status": "ok"})
	})

	// public auth
	r.Post("/auth/register", handleRegister(cfg, userRepo))
	r.Post("/auth/login", handleLogin(cfg, userRepo))

	// protected
	r.Group(func(pr chi.Router) {
		pr.Use(authMiddleware(cfg.JWTSecret))

		pr.Get("/me", handleMe(userRepo))
		pr.Get("/tracked", handleGetTracked(trackedRepo))
		pr.Put("/tracked", handlePutTracked(trackedRepo))
	})

	return r
}
