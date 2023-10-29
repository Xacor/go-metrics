package middleware

import (
	"github.com/Xacor/go-metrics/internal/server/config"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

func RegisterMiddlewares(r *chi.Mux, cfg *config.Config) (chi.Middlewares, error) {
	key, err := cfg.GetKey()
	if err != nil {
		return nil, err
	}

	r.Use(WithLogging)
	r.Use(WithCheckSignature(key))
	r.Use(WithCompressRead)
	if cfg.CryptoKeyPrivateFile != "" {
		pkey, err := cfg.GetPrivateKey()
		if err != nil {
			return nil, err
		}
		r.Use(WithRsaDecrypt(pkey))
	}

	r.Use(WithCompressWrite)
	r.Use(WithSignature(key))
	r.Use(chimiddleware.Recoverer)

	r.Mount("/debug", chimiddleware.Profiler())

	return r.Middlewares(), nil
}
