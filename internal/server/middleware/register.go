package middleware

import (
	"github.com/Xacor/go-metrics/internal/logger"
	"github.com/Xacor/go-metrics/internal/server/config"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

func RegisterMiddlewares(r *chi.Mux, cfg *config.Config) chi.Middlewares {
	key, err := cfg.GetKey()
	if err != nil {
		logger.Get().Error("unable to get signature key", zap.Error(err))
	}

	// r.Use(chimiddleware.Timeout(time.Second))
	r.Use(WithLogging)
	r.Use(WithCompressRead)
	r.Use(WithCheckSignature(key))
	r.Use(WithCompressWrite)
	r.Use(WithSignature(key))
	r.Use(chimiddleware.Recoverer)

	r.Mount("/debug", chimiddleware.Profiler())

	return r.Middlewares()
}
