package handlers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/Xacor/go-metrics/internal/server/global"
	"github.com/Xacor/go-metrics/internal/server/model"
	"github.com/Xacor/go-metrics/internal/server/storage"
)

type API struct {
	Repo storage.MetricRepo
}

func (api *API) UpdateHandler(w http.ResponseWriter, r *http.Request) {
	mtype := global.ValidType.FindStringSubmatch(r.URL.Path)[1]

	mid := global.ValidID.FindStringSubmatch(r.URL.Path)[1]

	mvalue := global.ValidValue.FindStringSubmatch(r.URL.Path)[1]

	var metric model.Metric

	if mtype == "counter" {
		log.Println("type counter")
		v, err := strconv.ParseInt(mvalue, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		metric = model.Metric{
			ID:    mid,
			Type:  model.Counter,
			Value: v,
		}
	} else {
		v, err := strconv.ParseFloat(mvalue, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		metric = model.Metric{
			ID:    mid,
			Type:  model.Gauge,
			Value: v,
		}
	}
	if _, err := api.Repo.Get(mid); err != nil {
		if _, err = api.Repo.Create(metric); err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	if _, err := api.Repo.Update(metric); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
	}

	log.Println(api.Repo.Get(mid))
}
