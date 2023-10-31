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

	"github.com/Xacor/go-metrics/internal/agent/config"
	"github.com/Xacor/go-metrics/internal/agent/metric"
	"github.com/Xacor/go-metrics/internal/logger"
	"go.uber.org/zap"

	poller "github.com/Xacor/go-metrics/internal/agent/http"
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

	l.Info("agent configuration", zap.Any("cfg", cfg))

	key, err := cfg.GetKey()
	if err != nil {
		l.Error("failed to get key", zap.Error(err))
	}

	monitor, err := metric.NewMonitor(time.Duration(cfg.GetPollInterval()) * time.Second)
	if err != nil {
		l.Error("failed to create monitor", zap.Error(err))
	}
	defer monitor.Close()

	publicKey, err := cfg.GetPublicKey()
	if err != nil {
		l.Error("failed to get public key", zap.Error(err))
	}

	pcfg := poller.PollerConfig{
		ReportInterval: cfg.GetReportInterval(),
		RateLimit:      cfg.GetRateLimit(),
		Address:        cfg.GetURL(),
		Key:            key,
		MetricCh:       monitor.C,
		Client:         &http.Client{},
		Logger:         l,
		PublicKey:      publicKey,
	}

	poller := poller.NewPoller(&pcfg)

	gracefullShutdown := make(chan os.Signal, 2)
	signal.Notify(gracefullShutdown, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	ctx, stopMonitor := context.WithCancel(context.Background())
	go poller.Run(ctx)

	l.Info("signal received, gracefully shutting down", zap.Any("signal", <-gracefullShutdown))
	stopMonitor()
	l.Info("gracefully shutting down")
}
