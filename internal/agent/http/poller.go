package http

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"net/http"
	"time"

	"github.com/Xacor/go-metrics/internal/agent/metric"
	"go.uber.org/zap"
)

type PollerConfig struct {
	MetricCh       <-chan metric.UpdateResult
	Client         *http.Client
	Logger         *zap.Logger
	PublicKey      *rsa.PublicKey
	Address        string
	Key            string
	ReportInterval int
	RateLimit      int
}

type Poller struct {
	metricCh       <-chan metric.UpdateResult
	client         *http.Client
	logger         *zap.Logger
	publicKey      *rsa.PublicKey
	address        string
	key            string
	reportInterval int
	rateLimit      int
}

func NewPoller(cfg *PollerConfig) *Poller {
	p := &Poller{
		reportInterval: cfg.ReportInterval,
		address:        cfg.Address,
		metricCh:       cfg.MetricCh,
		client:         cfg.Client,
		logger:         cfg.Logger,
		key:            cfg.Key,
		rateLimit:      cfg.RateLimit,
	}

	if cfg.PublicKey != nil {
		p.publicKey = cfg.PublicKey
	}

	return p
}

func (p *Poller) Run() {
	p.logger.Info("poller started")
	semaphore := NewSemaphore(p.rateLimit)

	t := time.NewTicker(time.Second * time.Duration(p.reportInterval))
	for {
		select {
		case <-t.C:
			res := <-p.metricCh
			if res.Err != nil {
				p.logger.Error("failed to read from UpdateResult", zap.Error(res.Err))
				continue
			}

			go func(m metric.Metrics) {
				semaphore.Acquire()
				defer semaphore.Release()

				err := p.Send(m)
				if err != nil {
					p.retry(p.Send, m)
				}
			}(res.Metrtics)

		case <-p.metricCh:
		}
	}
}

func (p *Poller) Send(m metric.Metrics) error {
	json, err := m.MarshalJSON()
	if err != nil {
		return err
	}

	compressed, err := Compress(json)
	if err != nil {
		return err
	}

	reader := bytes.NewReader(compressed)

	if p.publicKey != nil {
		encryptedBytes, err := rsa.EncryptOAEP(
			sha256.New(),
			rand.Reader,
			p.publicKey,
			compressed,
			nil)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(encryptedBytes)
	}

	request, err := http.NewRequest(http.MethodPost, p.address+"/updates/", reader)
	if err != nil {
		return err
	}

	if p.key != "" {
		sign, err := Sign(json, p.key)
		if err != nil {
			return err
		}
		request.Header.Set("HashSHA256", string(sign))
	}

	request.Header.Set("Content-Encoding", "gzip")
	request.Header.Set("Content-Type", "application/json")
	resp, err := p.client.Do(request)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func (p *Poller) retry(fn func(metric.Metrics) error, arg metric.Metrics) {
	attempts := 0
	var err error
	for i := 1; i < 5; i += 2 {
		time.Sleep(time.Second * time.Duration(i))
		if err = fn(arg); err == nil {
			return
		}
		attempts++
		p.logger.Error("attempt failed", zap.Error(err), zap.Int("attempt #", attempts))
	}
}
