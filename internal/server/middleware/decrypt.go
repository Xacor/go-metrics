package middleware

import (
	"bytes"
	"crypto"
	"crypto/rsa"
	"io"
	"net/http"

	"github.com/Xacor/go-metrics/internal/logger"
	"go.uber.org/zap"
)

type decryptReader struct {
	r  io.ReadCloser
	br *bytes.Reader
}

func newDecryptReader(r io.ReadCloser, b []byte) *decryptReader {
	return &decryptReader{
		r:  r,
		br: bytes.NewReader(b),
	}
}

func (c *decryptReader) Read(p []byte) (n int, err error) {
	return c.br.Read(p)
}

func (c *decryptReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return nil
}

func WithRsaDecrypt(key *rsa.PrivateKey) func(next http.Handler) http.Handler {
	l := logger.Get()
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sendsEncrypted := key != nil
			if sendsEncrypted {
				var body []byte
				if _, err := r.Body.Read(body); err != nil {
					return
				}

				decryptedBytes, err := key.Decrypt(nil, body, &rsa.OAEPOptions{Hash: crypto.SHA256})
				if err != nil {
					l.Error("unable to configure decrypt body", zap.Error(err))
					return
				}

				r.Body = newDecryptReader(r.Body, decryptedBytes)
			}

			next.ServeHTTP(w, r)
		})
	}

}
