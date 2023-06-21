package handlers

import (
	"github.com/Xacor/go-metrics/internal/server/storage"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type API struct {
	repo   storage.MetricRepo
	logger *zap.Logger
}

func NewAPI(repo storage.MetricRepo, logger *zap.Logger) *API {

	return &API{repo: repo, logger: logger}
}

func (api *API) RegisterRoutes(router *chi.Mux) {
	router.Get("/", api.MetricsHandler)

	router.Route("/value", func(r chi.Router) {
		r.Get("/{metricType}/{metricID}", api.MetricHandler)
	})

	router.Route("/update", func(r chi.Router) {
		r.Post("/{metricType}/{metricID}/{metricValue}", api.UpdateHandler)
	})
}
