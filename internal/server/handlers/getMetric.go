package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Xacor/go-metrics/internal/server/model"
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
	switch data.MType {
	case model.TypeCounter:
		w.Write([]byte([]byte(strconv.FormatInt(*data.Delta, 10))))
	case model.TypeGauge:
		w.Write([]byte([]byte(strconv.FormatFloat(*data.Value, 'f', -1, 64))))
	}

	w.Header().Set("Content-Type", "text/plain")
}

func (api *API) MetricsHandler(w http.ResponseWriter, r *http.Request) {
	data, err := api.repo.All()
	if err != nil {
		api.logger.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal(data)
	if err != nil {
		api.logger.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	w.Write(resp)
}

func (api *API) MetricJson(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "application/json" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var metric model.Metrics
	var buf bytes.Buffer

	if _, err := buf.ReadFrom(r.Body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := json.Unmarshal(buf.Bytes(), &metric); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	metric, err := api.repo.Get(metric.ID)
	if err != nil {
		api.logger.Info(metric.ID)
		w.Write([]byte(err.Error()))
		w.WriteHeader(http.StatusNotFound)
		return
	}

	json, err := json.Marshal(metric)
	if err != nil {
		api.logger.Error(err.Error())
		w.Write([]byte(err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	api.logger.Info(string(json))
	w.Header().Set("Content-Type", "application/json")
	w.Write(json)
}
