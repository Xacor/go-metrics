package middleware

import (
	"compress/gzip"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/Xacor/go-metrics/internal/agent/logger"
)

type compressWriter struct {
	w  http.ResponseWriter
	zw *gzip.Writer
}

func newCompressWriter(w http.ResponseWriter) *compressWriter {
	return &compressWriter{
		w:  w,
		zw: gzip.NewWriter(w),
	}
}

func (c *compressWriter) Header() http.Header {
	return c.w.Header()
}

func (c *compressWriter) Write(p []byte) (int, error) {
	return c.zw.Write(p)
}

func (c *compressWriter) WriteHeader(statusCode int) {
	if statusCode < 300 {
		c.w.Header().Set("Content-Encoding", "gzip")
	}
	c.w.WriteHeader(statusCode)
}

func (c *compressWriter) Close() error {
	return c.zw.Close()
}

type compressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

func newCompressReader(r io.ReadCloser) (*compressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &compressReader{
		r:  r,
		zr: zr,
	}, nil
}

func (c compressReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

func (c *compressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.zr.Close()
}

func WithCompression(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		logger.Log.Info("compression middleware")
		ow := w

		// проверяем, что клиент отправил gzip
		contentEncoding := r.Header.Get("Content-Encoding")
		sendsGzip := strings.Contains(contentEncoding, "gzip")
		if sendsGzip {
			cr, err := newCompressReader(r.Body)
			defer cr.Close()

			if err != nil {
				logger.Log.Error(err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			r.Body = cr
		}

		// проверяем, что клиент поддерживает gzip
		supportsGzip := false
		for _, acceptEncoding := range r.Header.Values("Accept-Encoding") {
			supportsGzip = strings.Contains(acceptEncoding, "gzip")
			if supportsGzip {
				log.Println("client supports gzip")
				break
			}
		}

		if supportsGzip {
			log.Println("client supports gzip and valid Content-Type")
			cw := newCompressWriter(w)
			defer cw.Close()

			ow = cw

			w.Header().Set("Content-Encoding", "gzip")
		}

		log.Println("Serving next")
		next.ServeHTTP(ow, r)
	}
	return http.HandlerFunc(fn)
}
