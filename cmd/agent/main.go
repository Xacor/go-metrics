package main

import (
	"flag"

	"github.com/Xacor/go-metrics/internal/agent/http"
)

var (
	addr           = flag.String("a", "localhost:8080", "endpoint server")
	reportInterval = flag.Uint("r", 10, "report interval")
	pollInterval   = flag.Uint("p", 2, "poll interval")
)

func main() {
	flag.Parse()

	poller := http.NewPoller(*pollInterval, *reportInterval, *addr)
	poller.Run()

}
