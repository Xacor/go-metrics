package database

import (
	"net/http"
)

func (h *DBHandler) Ping(w http.ResponseWriter, r *http.Request) {
	if err := h.db.Ping(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
