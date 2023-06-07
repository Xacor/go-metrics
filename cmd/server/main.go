package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/Xacor/go-metrics/internal/server/handlers"
	"github.com/Xacor/go-metrics/internal/server/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

var (
	addr = flag.String("a", "localhost:8080", "endpoint server")
)

func main() {
	flag.Parse()
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

	log.Println("started serving on", *addr)
	err := http.ListenAndServe(*addr, r)
	if err != nil {
		log.Fatal(err)
	}
}
