package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/Xacor/go-metrics/internal/server/config"
	"github.com/Xacor/go-metrics/internal/server/handlers"
	"github.com/Xacor/go-metrics/internal/server/logger"
	"github.com/Xacor/go-metrics/internal/server/middleware"
	"github.com/Xacor/go-metrics/internal/server/storage"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	cfg := config.Config{}
	err := cfg.ParseAll()
	if err != nil {
		log.Fatalf("can't parse configuration: %v", err)
	}

	if err := logger.Initialize(cfg.LogLevel); err != nil {
		log.Fatalf("can't initialize zap logger: %v", err)
	}

	r := chi.NewRouter()
	r.Use(middleware.WithLogging)
	r.Use(middleware.WithCompression)
	r.Use(chimiddleware.Recoverer)

	s := storage.NewMemStorage()

	if cfg.Restore {
		if err := storage.Load(cfg.FileStoragePath, s); err != nil {
			logger.Log.Error(err.Error())
		}
	}

	api := handlers.NewAPI(s, logger.Log)
	api.RegisterRoutes(r)

	srv := http.Server{
		Addr:    cfg.Address,
		Handler: r,
	}

	logger.Log.Info(fmt.Sprintf("starting serving on %s", cfg.Address))
	go srv.ListenAndServe()

	go func() {
		t := time.NewTicker(time.Duration(cfg.StoreInterval) * time.Second)
		for range t.C {
			logger.Log.Info("saving current state")
			err = storage.Save(cfg.FileStoragePath, s)
			if err != nil {
				logger.Log.Error(err.Error())
			}
		}
	}()

	<-ctx.Done()

	stop()

	logger.Log.Info("shutting down")
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	err = storage.Save(cfg.FileStoragePath, s)
	if err != nil {
		logger.Log.Error(err.Error())
	}

	if err := srv.Shutdown(timeoutCtx); err != nil {
		logger.Log.Error(err.Error())
	}

	defer logger.Log.Sync()
}
