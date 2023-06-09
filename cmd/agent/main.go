package main

import (
	"log"
	"net/http"

	"github.com/Xacor/go-metrics/internal/agent/config"
	"github.com/Xacor/go-metrics/internal/agent/metric"

	poller "github.com/Xacor/go-metrics/internal/agent/http"
)

func main() {
	cfg := config.Config{}
	err := cfg.ParseAll()
	if err != nil {
		log.Fatal(err)
	}

	poller := poller.NewPoller(
		cfg.GetPollInterval(),
		cfg.GetReportInterval(),
		cfg.GetURL(),
		metric.NewMetrics(),
		&http.Client{},
	)
	poller.Run()
}
