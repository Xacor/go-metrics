package main

import (
	"github.com/Xacor/go-metrics/internal/agent/http"
)

func main() {
	poller := http.NewPoller(2, 10, "http://localhost:8080")
	poller.Run()

}
