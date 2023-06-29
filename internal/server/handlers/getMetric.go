package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

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
	var valStr string
	switch data.MType {
	case model.TypeCounter:
		valStr = fmt.Sprintf("%v", *data.Delta)
	case model.TypeGauge:
		valStr = fmt.Sprintf("%v", *data.Value)
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(valStr))
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

func (api *API) MetricJSON(w http.ResponseWriter, r *http.Request) {
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
	api.logger.Info(fmt.Sprintf("requested metric %+v", metric))
	result, err := api.repo.Get(metric.ID)

	// если на найдена такая метрика, то формируется ответ с нулевым значением
	// возможно стоит сразу создавать такую метрику в бд
	if err != nil {
		api.logger.Info("metric not found")
		result.ID = metric.ID
		result.MType = metric.MType

		if result.MType == model.TypeCounter {
			var delta int64 = 0
			result.Delta = &delta
		} else {
			var value float64 = 0
			result.Value = &value
		}
	}

	api.logger.Info(fmt.Sprintf("responsed metric %+v", result))

	json, err := json.Marshal(result)
	if err != nil {
		api.logger.Error(err.Error())
		w.Write([]byte(err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(json)
}
