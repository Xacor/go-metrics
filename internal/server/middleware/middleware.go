package middleware

import (
	"net/http"

	"github.com/Xacor/go-metrics/internal/server/global"
)

type Middleware func(http.Handler) http.Handler

func Conveyor(h http.Handler, middlewares ...Middleware) http.Handler {
	for _, middleware := range middlewares {
		h = middleware(h)
	}
	return h
}

func Post(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if http.MethodPost != r.Method {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func TextPlain(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "text/plain" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func ValidateParams(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mtype := global.ValidType.FindStringSubmatch(r.URL.Path)
		if mtype == nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		mid := global.ValidID.FindStringSubmatch(r.URL.Path)
		if mid == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		mvalue := global.ValidValue.FindStringSubmatch(r.URL.Path)
		if mvalue == nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// по хорошему писать извлеченные параметры в контекст, чтоб второй раз не делать тоже самое

		next.ServeHTTP(w, r)
	})
}
