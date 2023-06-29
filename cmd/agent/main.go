package main

import (
	"log"
	"net/http"

	"github.com/Xacor/go-metrics/internal/agent/config"
	"github.com/Xacor/go-metrics/internal/agent/metric"
	"go.uber.org/zap"

	poller "github.com/Xacor/go-metrics/internal/agent/http"
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("failed to init logger")
	}

	cfg := config.Config{}
	err = cfg.ParseAll()
	if err != nil {
		logger.Fatal(err.Error())
	}

	poller := poller.NewPoller(
		cfg.GetPollInterval(),
		cfg.GetReportInterval(),
		cfg.GetURL(),
		metric.NewMetrics(),
		&http.Client{},
		logger,
	)
	poller.Run()
}
