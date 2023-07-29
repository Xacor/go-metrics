package http

import (
	"bytes"
	"net/http"
	"time"

	"github.com/Xacor/go-metrics/internal/agent/metric"
	"go.uber.org/zap"
)

type PollerConfig struct {
	ReportInterval int
	RateLimit      int
	Address        string
	Key            string
	MetricCh       <-chan metric.UpdateResult
	Client         *http.Client
	Logger         *zap.Logger
}

type Poller struct {
	reportInterval int
	rateLimit      int
	address        string
	key            string
	metricCh       <-chan metric.UpdateResult
	client         *http.Client
	logger         *zap.Logger
}

func NewPoller(cfg *PollerConfig) *Poller {
	return &Poller{
		reportInterval: cfg.ReportInterval,
		address:        cfg.Address,
		metricCh:       cfg.MetricCh,
		client:         cfg.Client,
		logger:         cfg.Logger,
		key:            cfg.Key,
		rateLimit:      cfg.RateLimit,
	}
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
