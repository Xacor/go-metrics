package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (api *API) MetricHandler(w http.ResponseWriter, r *http.Request) {
	var metricType, metricID string
	if metricID = chi.URLParam(r, "metricID"); metricID == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if metricType = chi.URLParam(r, "metricType"); metricType == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	data, err := api.repo.Get(metricID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	valStr := fmt.Sprintf("%v", data.Value)

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(valStr))
}

func (api *API) MetricsHandler(w http.ResponseWriter, r *http.Request) {
	data, err := api.repo.All()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal(data)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	w.Write(resp)
}
