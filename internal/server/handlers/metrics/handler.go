package metrics

import (
	"github.com/Xacor/go-metrics/internal/server/storage"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type API struct {
	repo   storage.Storage
	logger *zap.Logger
}

func NewAPI(repo storage.Storage, logger *zap.Logger) *API {

	return &API{repo: repo, logger: logger}
}

func (api *API) RegisterRoutes(router *chi.Mux) {
	router.Get("/", api.MetricsHandler)

	router.Route("/value", func(r chi.Router) {
		r.Post("/", api.MetricJSON)
		r.Get("/{metricType}/{metricID}", api.MetricHandler)
	})

	router.Route("/update", func(r chi.Router) {
		r.Post("/", api.UpdateJSON)
		r.Post("/{metricType}/{metricID}/{metricValue}", api.UpdateHandler)
	})

	router.Post("/updates/", api.UpdateMetrics)
}
