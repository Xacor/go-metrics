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
	"github.com/Xacor/go-metrics/internal/server/handlers/database"
	"github.com/Xacor/go-metrics/internal/server/handlers/metrics"
	"github.com/Xacor/go-metrics/internal/server/middleware"
	"github.com/Xacor/go-metrics/internal/server/storage"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
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
	r.Use(middleware.WithLogging)
	r.Use(middleware.WithCompressRead)
	r.Use(middleware.WithCompressWrite)
	r.Use(chimiddleware.Recoverer)

	var repo storage.Storage

	if cfg.DatabaseDSN != "" {
		ctx, cancelfunc := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancelfunc()
		postgre, err := storage.NewPostgreStorage(ctx, cfg.DatabaseDSN, l)
		if err != nil {
			l.Fatal("can't init db connection", zap.Error(err))
		}
		defer postgre.Close()
		repo = postgre

	} else if cfg.Restore {
		repo = storage.NewMemStorage()
		fs, err := storage.NewFileStorage(cfg.FileStoragePath)
		if err != nil {
			l.Error("cant'init file storage", zap.Error(err))
		}
		if err := fs.Load(repo); err != nil {
			l.Error("can't restore data from file", zap.Error(err))
		}

	} else {
		repo = storage.NewMemStorage()
	}

	api := metrics.NewAPI(repo, l)
	api.RegisterRoutes(r)

	fs, err := storage.NewFileStorage(cfg.FileStoragePath)
	if err != nil {
		l.Fatal(err.Error())
	}
	if cfg.Restore {

	}

	databaseApi := database.NewDBHandler(repo)
	databaseApi.RegisterRoutes(r)

	srv := http.Server{
		Addr:    cfg.Address,
		Handler: r,
	}

	l.Info(fmt.Sprintf("starting serving on %s", cfg.Address))
	go func() {
		srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			l.Fatal(err.Error())
		}

	}()

	go func() {
		t := time.NewTicker(time.Duration(cfg.StoreInterval) * time.Second)
		for range t.C {
			l.Debug("saving current state")
			err = fs.Save(repo)
			if err != nil {
				l.Error(err.Error())
			}
		}
	}()

	<-gracefullShutdown

	l.Info("shutting down")
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	if err := fs.Save(repo); err != nil {
		l.Error(err.Error())
	}

	if err := srv.Shutdown(timeoutCtx); err != nil {
		l.Error(err.Error())
	}
}
