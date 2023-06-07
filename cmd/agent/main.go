package main

import (
	"flag"
	"log"

	"github.com/Xacor/go-metrics/internal/agent/http"
	"github.com/caarlos0/env/v6"
)

type Config struct {
	Address        string `env:"ADDRESS"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	PollInterval   int    `env:"POLL_INTERVAL"`
}

func main() {
	var cfg Config
	flag.StringVar(&cfg.Address, "a", "localhost:8080", "server address")
	flag.IntVar(&cfg.ReportInterval, "r", 10, "report interval")
	flag.IntVar(&cfg.PollInterval, "p", 2, "poll interval")
	flag.Parse()

	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	poller := http.NewPoller(cfg.PollInterval, cfg.ReportInterval, cfg.Address)
	poller.Run()
}
