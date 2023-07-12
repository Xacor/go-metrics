package handlers

import "net/http"

func (api *API) Ping(w http.ResponseWriter, r *http.Request) {
	if err := api.repo.Ping(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
