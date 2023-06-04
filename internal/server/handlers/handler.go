package handlers

import (
	"log"
	"net/http"
	"regexp"
	"strconv"

	"github.com/Xacor/go-metrics/internal/server/model"
	"github.com/Xacor/go-metrics/internal/server/storage"
)

type API struct {
	Repo storage.MetricRepo
}

func (api *API) UpdateHandler(w http.ResponseWriter, r *http.Request, mtype, id, value string) {
	var metric model.Metric

	if mtype == "counter" {
		log.Println("type counter")
		v, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		metric = model.Metric{
			ID:    id,
			Type:  model.Counter,
			Value: v,
		}
	} else {
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		metric = model.Metric{
			ID:    id,
			Type:  model.Counter,
			Value: v,
		}
	}
	if _, err := api.Repo.Get(id); err != nil {
		if _, err = api.Repo.Create(metric); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	if _, err := api.Repo.Update(metric); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	log.Println(api.Repo.Get(id))
}

var validType = regexp.MustCompile(`/update/(gauge|counter)/`)
var validID = regexp.MustCompile(`/update/.+/([a-zA-Z]*)/.*`)
var validValue = regexp.MustCompile(`/update/.+/.+/([+-]?[0-9]*[.]?[0-9]+)`)

func MakeHandler(fn func(w http.ResponseWriter, r *http.Request, mtype, id, value string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mtype := validType.FindStringSubmatch(r.URL.Path)
		if mtype == nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		mid := validID.FindStringSubmatch(r.URL.Path)
		if mid == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		mvalue := validValue.FindStringSubmatch(r.URL.Path)
		if mvalue == nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		fn(w, r, mtype[1], mid[1], mvalue[1])
	}
}
