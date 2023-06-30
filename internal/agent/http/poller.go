package http

import (
	"bytes"
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
	var metric Metrics

	for i := 0; i < values.NumField(); i++ {
		field := types.Field(i)
		value := values.Field(i)

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

		json, err := json.Marshal(metric)
		if err != nil {
			p.logger.Error(err.Error())
			continue
		}
		reader := bytes.NewReader(json)
		resp, err := p.client.Post(p.address+"/update/", "application/json", reader)
		if err != nil {
			p.logger.Error(err.Error())
			continue
		}
		err = resp.Body.Close()
		if err != nil {
			p.logger.Error(err.Error())
		}
	}
	return nil
}
