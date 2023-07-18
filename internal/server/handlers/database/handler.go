package database

import (
	"github.com/Xacor/go-metrics/internal/server/storage"
	"github.com/go-chi/chi/v5"
)

type HealthService struct {
	db storage.Pinger
}

func NewHealthService(db storage.Pinger) *HealthService {
	return &HealthService{db: db}
}

func (h *HealthService) RegisterRoutes(r *chi.Mux) {
	r.Get("/ping", h.Ping)
}
