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
	"github.com/Xacor/go-metrics/internal/server/handlers"
	"github.com/Xacor/go-metrics/internal/server/middleware"
	"github.com/Xacor/go-metrics/internal/server/storage"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
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

	r := chi.NewRouter()
	r.Use(middleware.WithLogging)
	r.Use(middleware.WithCompressRead)
	r.Use(middleware.WithCompressWrite)
	r.Use(chimiddleware.Recoverer)

	ms, err := storage.NewPostgreStorage(cfg.DatabaseDSN)
	if err != nil {
		l.Fatal(err.Error())
	}
	defer ms.Close()

	fs, err := storage.NewFileStorage(cfg.FileStoragePath)
	if err != nil {
		l.Fatal(err.Error())
	}
	if cfg.Restore {
		if err := fs.Load(ms); err != nil {
			l.Error(err.Error())
		}
	}

	api := handlers.NewAPI(ms, l)
	api.RegisterRoutes(r)

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
			err = fs.Save(ms)
			if err != nil {
				l.Error(err.Error())
			}
		}
	}()

	<-gracefullShutdown

	l.Info("shutting down")
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	err = fs.Save(ms)
	if err != nil {
		l.Error(err.Error())
	}

	if err := srv.Shutdown(timeoutCtx); err != nil {
		l.Error(err.Error())
	}

	defer l.Sync()

}
