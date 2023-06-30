package main

import (
	"log"
	"net/http"

	"github.com/Xacor/go-metrics/internal/agent/config"
	"github.com/Xacor/go-metrics/internal/agent/logger"
	"github.com/Xacor/go-metrics/internal/agent/metric"

	poller "github.com/Xacor/go-metrics/internal/agent/http"
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

	pcfg := poller.PollerConfig{
		PollInterval:   cfg.GetPollInterval(),
		ReportInterval: cfg.GetReportInterval(),
		Address:        cfg.GetURL(),
		Metrics:        metric.NewMetrics(),
		Client:         &http.Client{},
		Logger:         logger.Log,
	}

	poller := poller.NewPoller(&pcfg)
	poller.Run()

	defer logger.Log.Sync()
}
