package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Xacor/go-metrics/internal/server/model"
	"github.com/go-chi/chi/v5"
)

func (api *API) UpdateHandler(w http.ResponseWriter, r *http.Request) {
	var metricType, metricID, metricValue string
	if metricType = chi.URLParam(r, "metricType"); metricType == "" {
		api.logger.Error("empty metric type")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if metricID = chi.URLParam(r, "metricID"); metricID == "" {
		api.logger.Error("empty metric id")
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if metricValue = chi.URLParam(r, "metricValue"); metricValue == "" {
		api.logger.Error("empty metric value")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var metric model.Metrics

	switch metricType {
	case model.TypeCounter:
		v, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			api.logger.Error("can't parse int value")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		metric = model.Metrics{
			ID:    metricID,
			MType: model.TypeCounter,
			Delta: &v,
		}

	case model.TypeGauge:
		v, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		metric = model.Metrics{
			ID:    metricID,
			MType: model.TypeGauge,
			Value: &v,
		}

	default:
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// create if doesnt exist
	if _, err := api.repo.Get(metricID); err != nil {
		if _, err = api.repo.Create(metric); err != nil {
			api.logger.Error(err.Error())
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	if _, err := api.repo.Update(metric); err != nil {
		api.logger.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (api *API) UpdateJSON(w http.ResponseWriter, r *http.Request) {
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

	// проверка на существование метрики с таким ID
	var result model.Metrics
	if _, err := api.repo.Get(metric.ID); err != nil {
		// если нет, то создать
		result, err = api.repo.Create(metric)
		if err != nil {
			api.logger.Error(err.Error())
			w.WriteHeader(http.StatusInternalServerError)
		}
	} else {
		result, err = api.repo.Update(metric)
		if err != nil {
			api.logger.Error(err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	json, err := json.Marshal(result)
	if err != nil {
		api.logger.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(json)
}
