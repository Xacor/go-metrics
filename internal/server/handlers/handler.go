package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/Xacor/go-metrics/internal/server/model"
	"github.com/Xacor/go-metrics/internal/server/storage"
	"github.com/go-chi/chi/v5"
)

type API struct {
	Repo storage.MetricRepo
}

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

	data, err := api.Repo.Get(metricID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	valStr := fmt.Sprintf("%v", data.Value)

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(valStr))
}

func (api *API) MetricsHandler(w http.ResponseWriter, r *http.Request) {
	data, _ := api.Repo.All()
	resp, err := json.Marshal(data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	w.Write(resp)
}

func (api *API) UpdateHandler(w http.ResponseWriter, r *http.Request) {
	var metricType, metricID, metricValue string
	if metricType = chi.URLParam(r, "metricType"); metricType == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if metricID = chi.URLParam(r, "metricID"); metricID == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if metricValue = chi.URLParam(r, "metricValue"); metricValue == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var metric model.Metric

	switch metricType {
	case "counter":
		log.Println("type counter")
		v, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		metric = model.Metric{
			ID:    metricID,
			Type:  model.Counter,
			Value: v,
		}

	case "gauge":
		v, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		metric = model.Metric{
			ID:    metricID,
			Type:  model.Gauge,
			Value: v,
		}

	default:
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// create if doesnt exist
	if _, err := api.Repo.Get(metricID); err != nil {
		if _, err = api.Repo.Create(metric); err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	if _, err := api.Repo.Update(metric); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
