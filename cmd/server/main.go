package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/Xacor/go-metrics/internal/server/handlers"
	"github.com/Xacor/go-metrics/internal/server/storage"
	"github.com/caarlos0/env/v6"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Config struct {
	Address string `env:"ADDRESS"`
}

func main() {
	var cfg Config
	flag.StringVar(&cfg.Address, "a", "localhost:8080", "server address")
	flag.Parse()
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	api := handlers.API{
		Repo: storage.NewMemStorage(),
	}

	r.Get("/", api.MetricsHandler)
	r.Route("/value", func(r chi.Router) {
		r.Get("/{metricType}/{metricID}", api.MetricHandler)
	})
	r.Route("/update", func(r chi.Router) {
		r.Post("/{metricType}/{metricID}/{metricValue}", api.UpdateHandler)
	})

	log.Println("started serving on", cfg.Address)
	err = http.ListenAndServe(cfg.Address, r)
	if err != nil {
		log.Fatal(err)
	}
}
