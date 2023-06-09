package handlers

import (
	"github.com/Xacor/go-metrics/internal/server/storage"
	"github.com/go-chi/chi/v5"
)

type API struct {
	repo storage.MetricRepo
}

func NewAPI(repo storage.MetricRepo) *API {
	api := API{repo: repo}

	return &api
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
