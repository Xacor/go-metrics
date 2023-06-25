package handlers

import (
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
}
