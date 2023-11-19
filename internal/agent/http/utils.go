package http

import (
	"bytes"
	"compress/gzip"
	"crypto/hmac"
	"crypto/sha256"
	"errors"
	"net"
)

func Sign(data []byte, key string) ([]byte, error) {
	h := hmac.New(sha256.New, []byte(key))
	if _, err := h.Write(data); err != nil {
		return nil, err
	}

	return h.Sum(nil), nil
}

func Compress(data []byte) ([]byte, error) {
	var b bytes.Buffer

	w := gzip.NewWriter(&b)
	if _, err := w.Write(data); err != nil {
		return nil, err
	}

	if err := w.Close(); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

type Semaphore struct {
	semaCh chan struct{}
}

func NewSemaphore(maxReq int) *Semaphore {
	return &Semaphore{
		semaCh: make(chan struct{}, maxReq),
	}
}

func (s *Semaphore) Acquire() {
	s.semaCh <- struct{}{}
}

func (s *Semaphore) Release() {
	<-s.semaCh
}

func GetLocalIP() (string, error) {
	ips, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	for i := range ips {
		if ips[i].String() != "127.0.0.1/8" {
			ip, _, err := net.ParseCIDR(ips[i].String())
			if err != nil {
				return "", err
			}

			return ip.String(), nil
		}
	}

	return "", errors.New("no IP available")
}
