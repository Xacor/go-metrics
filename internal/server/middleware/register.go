package middleware

import (
	"time"

	"github.com/Xacor/go-metrics/internal/server/config"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

func RegisterMiddlewares(r *chi.Mux, cfg *config.Config) chi.Middlewares {
	r.Use(chimiddleware.Timeout(time.Second))
	r.Use(WithLogging)
	r.Use(WithCompressRead)
	r.Use(WithCheckSignature(cfg.Key))
	r.Use(WithCompressWrite)
	r.Use(WithSignature(cfg.Key))
	r.Use(chimiddleware.Recoverer)

	return r.Middlewares()
}
