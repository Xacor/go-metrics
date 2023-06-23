package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/Xacor/go-metrics/internal/server/config"
	"github.com/Xacor/go-metrics/internal/server/handlers"
	"github.com/Xacor/go-metrics/internal/server/logger"
	"github.com/Xacor/go-metrics/internal/server/middleware"
	"github.com/Xacor/go-metrics/internal/server/storage"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

func main() {

	cfg := config.Config{}
	err := cfg.ParseAll()
	if err != nil {
		log.Fatalf("can't parse configuration: %v", err)
	}

	if err := logger.Initialize(cfg.LogLevel); err != nil {
		log.Fatalf("can't initialize zap logger: %v", err)
	}

	r := chi.NewRouter()
	r.Use(middleware.NewLogger(logger.Log))
	r.Use(chimiddleware.Recoverer)

	api := handlers.NewAPI(storage.NewMemStorage(logger.Log), logger.Log)
	api.RegisterRoutes(r)

	logger.Log.Info(fmt.Sprintf("starting serving on %s", cfg.Address))
	err = http.ListenAndServe(cfg.Address, r)
	if err != nil {
		logger.Log.Fatal(fmt.Sprintf("can't start serving: %v", err))
	}
}
