package http

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
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
			if err := p.SendRequests(); err != nil {
				p.logger.Error(err.Error())
			}
		}
	}
}

func (p *Poller) SendRequests() error {
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
			p.logger.Info(fmt.Sprintf("unexpected kind: %v, value: %v", value.Kind(), value))
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

	request, err := http.NewRequest(http.MethodPost, p.address+"/update/", reader)
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
	_, err := w.Write(data)
	if err != nil {
		return nil, fmt.Errorf("failed write data to compress temporary buffer: %v", err)
	}

	err = w.Close()
	if err != nil {
		return nil, fmt.Errorf("failed compress data: %v", err)
	}

	return b.Bytes(), nil
}
