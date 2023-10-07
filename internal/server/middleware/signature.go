package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"net/http"

	"github.com/Xacor/go-metrics/internal/logger"
	"go.uber.org/zap"
)

func WithCheckSignature(key string) func(next http.Handler) http.Handler {
	l := logger.Get()
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if key == "" {
				next.ServeHTTP(w, r)
				return
			}

			data := make([]byte, r.ContentLength)

			_, err := r.Body.Read(data)
			if err != nil {
				l.Error("WithCheckSignature", zap.Error(err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			h := hmac.New(sha256.New, []byte(key))
			_, err = h.Write(data)
			if err != nil {
				l.Error("WithCheckSignature", zap.Error(err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			csign := []byte(r.Header.Get("HashSHA256"))
			ssign := h.Sum(nil)

			if !hmac.Equal(csign, ssign) {
				l.Warn("signatures are't equal", zap.ByteString("client sign", csign))
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

type signWriter struct {
	w   http.ResponseWriter
	key []byte
}

func newSignWriter(w http.ResponseWriter) *signWriter {
	return &signWriter{w: w}
}

func (s *signWriter) Header() http.Header {
	return s.w.Header()
}

func (s *signWriter) Write(body []byte) (int, error) {
	h := hmac.New(sha256.New, s.key)
	n, err := h.Write(body)
	if err != nil {
		return 0, err
	}

	sign := h.Sum(nil)
	s.w.Header().Set("HashSHA256", string(sign))

	return n, nil
}

func (s *signWriter) WriteHeader(statusCode int) {
	s.w.WriteHeader(statusCode)
}

func WithSignature(key string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ow := w
			if key != "" {
				sw := newSignWriter(w)
				ow = sw
			}

			next.ServeHTTP(ow, r)
		})
	}
}
