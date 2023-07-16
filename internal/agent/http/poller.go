package http

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"net/http"
	"reflect"
	"time"

	"github.com/Xacor/go-metrics/internal/agent/metric"
	"go.uber.org/zap"
)

type PollerConfig struct {
	PollInterval   int
	ReportInterval int
	Address        string
	Metrics        *metric.Metrics
	Client         *http.Client
	Logger         *zap.Logger
}

type Poller struct {
	pollInterval   int
	reportInterval int
	address        string
	metrics        *metric.Metrics
	client         *http.Client
	logger         *zap.Logger
}

func NewPoller(cfg *PollerConfig) *Poller {
	return &Poller{
		pollInterval:   cfg.PollInterval,
		reportInterval: cfg.ReportInterval,
		address:        cfg.Address,
		metrics:        cfg.Metrics,
		client:         cfg.Client,
		logger:         cfg.Logger,
	}
}

func (p *Poller) Run() {
	p.logger.Info("poller started")
	for i := 0; ; i++ {
		time.Sleep(time.Second * 1)
		if i%p.pollInterval == 0 {
			p.logger.Info("updating metric values")
			p.metrics.Update()
		}
		if i%p.reportInterval == 0 {
			if err := p.SendBatch(); err != nil {
				p.logger.Error("failed to report metrics", zap.Error(err))
				p.logger.Info("retrying to report metrics")
				p.Retry(p.SendBatch)
			}
			p.metrics.PollCount = 0
		}
	}
}

func (p *Poller) SendBatch() error {
	values := reflect.ValueOf(p.metrics).Elem()
	types := values.Type()
	var metrics []Metrics

	for i := 0; i < values.NumField(); i++ {
		field := types.Field(i)
		value := values.Field(i)

		var metric Metrics

		switch value.Kind() {
		case reflect.Uint64:
			v := int64(value.Uint())
			metric = Metrics{
				ID:    field.Name,
				MType: TypeCounter,
				Delta: &v,
			}
		case reflect.Int64:
			v := value.Int()
			metric = Metrics{
				ID:    field.Name,
				MType: TypeCounter,
				Delta: &v,
			}
		case reflect.Float64:
			v := value.Float()
			metric = Metrics{
				ID:    field.Name,
				MType: TypeGauge,
				Value: &v,
			}
		default:
			return errors.New("invalid metric kind")
		}

		metrics = append(metrics, metric)
	}

	json, err := json.Marshal(metrics)
	if err != nil {
		return err
	}

	compressed, err := p.Compress(json)
	if err != nil {
		return err
	}

	reader := bytes.NewReader(compressed)

	request, err := http.NewRequest(http.MethodPost, p.address+"/updates/", reader)
	if err != nil {
		return err
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

func (p *Poller) Compress(data []byte) ([]byte, error) {
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

func (p *Poller) Retry(fn func() error) {
	var err error
	attempts := 0
	for i := 1; i < 5; i += 2 {
		time.Sleep(time.Second * time.Duration(i))
		if err = fn(); err == nil {
			return
		}
		attempts++
		p.logger.Error("attempt failed", zap.Error(err), zap.Int("attempt #", attempts))
	}
	return
}
