package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Xacor/go-metrics/internal/logger"
	"github.com/Xacor/go-metrics/internal/server/config"
	"github.com/Xacor/go-metrics/internal/server/core/db"
	"github.com/Xacor/go-metrics/internal/server/handlers/database"
	"github.com/Xacor/go-metrics/internal/server/handlers/metrics"
	"github.com/Xacor/go-metrics/internal/server/middleware"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func main() {

	gracefullShutdown := make(chan os.Signal, 1)
	signal.Notify(gracefullShutdown, syscall.SIGINT, syscall.SIGTERM)

	cfg := config.Config{}
	err := cfg.ParseAll()
	if err != nil {
		log.Fatalf("can't parse configuration: %v", err)
	}

	if err := logger.Initialize(cfg.LogLevel); err != nil {
		log.Fatalf("can't initialize zap logger: %v", err)
	}

	l := logger.Get()
	defer l.Sync()

	r := chi.NewRouter()
	middleware.RegisterMiddlewares(r, &cfg)

	repo := db.InitDB(&cfg)
	defer repo.Close()

	metricsAPI := metrics.NewAPI(repo, l)
	metricsAPI.RegisterRoutes(r)

	databaseAPI := database.NewHealthService(repo)
	databaseAPI.RegisterRoutes(r)

	srv := http.Server{
		Addr:    cfg.Address,
		Handler: r,
	}

	l.Info(fmt.Sprintf("starting serving on %s", cfg.Address), zap.Any("server configuration", cfg))
	go func() {
		srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			l.Fatal(err.Error())
		}
	}()

	<-gracefullShutdown

	l.Info("shutting down")
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	if err := srv.Shutdown(timeoutCtx); err != nil {
		l.Error(err.Error())
	}
}
