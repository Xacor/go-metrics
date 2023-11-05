package middleware

import (
	"net"

	"github.com/Xacor/go-metrics/internal/server/config"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

func RegisterMiddlewares(r *chi.Mux, cfg *config.Config) (chi.Middlewares, error) {
	var signKey string
	if cfg.KeyFile != "" {
		key, err := cfg.GetKey()
		if err != nil {
			return nil, err
		}
		signKey = key
	}

	r.Use(WithLogging)

	if cfg.TrustedSubnet != "" {
		_, trustedNet, err := net.ParseCIDR(cfg.TrustedSubnet)
		if err != nil {
			return nil, err
		}
		r.Use(WithCheckSubnet(trustedNet))
	}

	r.Use(WithCheckSignature(signKey))
	r.Use(WithCompressRead)

	if cfg.CryptoKeyPrivateFile != "" {
		pkey, err := cfg.GetPrivateKey()
		if err != nil {
			return nil, err
		}
		r.Use(WithRsaDecrypt(pkey))
	}

	r.Use(WithCompressWrite)
	r.Use(WithSignature(signKey))
	r.Use(chimiddleware.Recoverer)

	r.Mount("/debug", chimiddleware.Profiler())

	return r.Middlewares(), nil
}
