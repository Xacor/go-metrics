package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Xacor/go-metrics/internal/logger"
	"github.com/Xacor/go-metrics/internal/server/config"
	"github.com/Xacor/go-metrics/internal/server/core"
	"github.com/Xacor/go-metrics/internal/server/core/db"
	"github.com/Xacor/go-metrics/internal/server/handlers/database"
	"github.com/Xacor/go-metrics/internal/server/handlers/metrics"
	"github.com/Xacor/go-metrics/internal/server/interceptors"
	"github.com/Xacor/go-metrics/internal/server/middleware"
	"github.com/Xacor/go-metrics/internal/server/storage"
	"github.com/Xacor/go-metrics/proto"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	_ "google.golang.org/grpc/encoding/gzip" // Install the gzip compressor
)

var (
	buildVersion string = "N/A"
	buildDate    string = "N/A"
	buildCommit  string = "N/A"
)

func printInfo() {
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)
}

func main() {
	printInfo()

	gracefullShutdown := make(chan os.Signal, 2)
	signal.Notify(gracefullShutdown, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	cfg := config.Config{}
	err := cfg.ParseAll()
	if err != nil {
		log.Fatalf("can't parse configuration: %v", err)
	}

	if err = logger.Initialize(cfg.LogLevel); err != nil {
		log.Fatalf("can't initialize zap logger: %v", err)
	}

	l := logger.Get()
	defer l.Sync()

	r := chi.NewRouter()
	_, err = middleware.RegisterMiddlewares(r, &cfg)
	if err != nil {
		l.Error("failed to configure middleware", zap.Error(err))
	}

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

	grpc := startGRPC(cfg, l, repo)

	<-gracefullShutdown

	l.Info("shutting down")
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := srv.Shutdown(timeoutCtx); err != nil {
		l.Error(err.Error())
	}

	grpc.GracefulStop()
}

func startGRPC(cfg config.Config, log *zap.Logger, repo storage.MetricRepo) *grpc.Server {
	listen, err := net.Listen("tcp", cfg.GAddress)
	if err != nil {
		log.Fatal("unable to listen tcp", zap.Error(err))
	}

	opts := make([]grpc.ServerOption, 0)
	if cfg.GRPCConfig.TLSCertFile != "" && cfg.GRPCConfig.TLSKeyFile != "" {
		creds, err := credentials.NewServerTLSFromFile(cfg.TLSCertFile, cfg.GRPCConfig.TLSKeyFile)
		if err != nil {
			log.Fatal("failed to create credentials: %v", zap.Error(err))
		}
		opts = append(opts, grpc.Creds(creds))
	}

	opts = append(opts, interceptors.RegisterUnaryInterceptorChain(cfg))

	s := grpc.NewServer(opts...)
	proto.RegisterMetricsServer(s, core.NewMetricsServer(repo, log))

	go func() {
		if err := s.Serve(listen); err != nil {
			log.Fatal("unable to start grpc server", zap.Error(err))
		}
	}()

	return s
}
