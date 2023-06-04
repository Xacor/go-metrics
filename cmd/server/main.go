package main

import (
	"log"
	"net/http"

	"github.com/Xacor/go-metrics/internal/server/handlers"
	"github.com/Xacor/go-metrics/internal/server/middleware"
	"github.com/Xacor/go-metrics/internal/server/storage"
)

const (
	addr = ":8080"
)

func main() {
	mux := http.NewServeMux()

	api := handlers.API{
		Repo: storage.NewMemStorage(),
	}
	mux.Handle(`/update/`, middleware.Conveyor(handlers.MakeHandler(api.UpdateHandler), middleware.Post)) // а как это укоротить????

	log.Println("started serving on", addr)
	err := http.ListenAndServe(addr, mux)
	if err != nil {
		log.Fatal(err)
	}
}
