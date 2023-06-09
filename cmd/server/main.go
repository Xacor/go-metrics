package main

import (
	"log"
	"net/http"

	"github.com/Xacor/go-metrics/internal/server/config"
	"github.com/Xacor/go-metrics/internal/server/handlers"
	"github.com/Xacor/go-metrics/internal/server/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	cfg := config.Config{}
	err := cfg.ParseAll()
	if err != nil {
		log.Fatal(err)
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	api := handlers.NewAPI(storage.NewMemStorage())
	api.RegisterRoutes(r)

	log.Println("started serving on", cfg.Address)
	err = http.ListenAndServe(cfg.Address, r)
	if err != nil {
		log.Fatal(err)
	}
}
