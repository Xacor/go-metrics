package database

import (
	"github.com/Xacor/go-metrics/internal/server/storage"
	"github.com/go-chi/chi/v5"
)

type DBHandler struct {
	db storage.Pinger
}

func NewDBHandler(db storage.Pinger) *DBHandler {
	return &DBHandler{db: db}
}

func (h *DBHandler) RegisterRoutes(r *chi.Mux) {
	r.Get("/ping", h.Ping)
}
