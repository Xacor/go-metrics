package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Xacor/go-metrics/internal/agent/config"
	"github.com/Xacor/go-metrics/internal/agent/metric"
	"github.com/Xacor/go-metrics/internal/logger"
	"go.uber.org/zap"

	poller "github.com/Xacor/go-metrics/internal/agent/http"
)

var (
	Version string = "N/A"
	Date    string = "N/A"
	Commit  string = "N/A"
)

func printInfo() {
	fmt.Printf("Build version: %s\n", Version)
	fmt.Printf("Build date: %s\n", Date)
	fmt.Printf("Build commit: %s\n", Commit)
}

func main() {
	printInfo()

	cfg := config.Config{}
	err := cfg.ParseAll()
	if err != nil {
		log.Fatalf("can't parse configuration: %v", err)
	}

	if err := logger.Initialize(cfg.LogLevel); err != nil {
		log.Fatalf("can't initialize zap logger: %v", err)
	}
	l := logger.Get()
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

	pcfg := poller.PollerConfig{
		ReportInterval: cfg.GetReportInterval(),
		RateLimit:      cfg.GetRateLimit(),
		Address:        cfg.GetURL(),
		Key:            key,
		MetricCh:       monitor.C,
		Client:         &http.Client{},
		Logger:         l,
	}

	poller := poller.NewPoller(&pcfg)
	poller.Run()

	defer l.Sync()
}
