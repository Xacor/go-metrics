package http

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"reflect"
	"time"

	"github.com/Xacor/go-metrics/internal/agent/metric"
)

type Poller struct {
	pollInterval   int
	reportInterval int
	address        string
	metrics        *metric.Metrics
	client         *http.Client
}

func NewPoller(pollInterval, reportInterval int, address string, metrics *metric.Metrics, client *http.Client) *Poller {
	return &Poller{
		pollInterval:   pollInterval,
		reportInterval: reportInterval,
		address:        address,
		metrics:        metrics,
		client:         client,
	}
}

func (p *Poller) Run() {
	log.Println("poller started")
	for i := 0; ; i++ {
		time.Sleep(time.Second * 1)
		if i%p.pollInterval == 0 {
			p.metrics.Update()
		}
		if i%p.reportInterval == 0 {
			if err := p.SendRequests(); err != nil {
				log.Println(err)
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
			log.Println("unexpected kind:", value.Kind(), value)
		}

		json, err := json.Marshal(metric)
		if err != nil {
			log.Println(err)
			continue
		}
		resp, err := http.Post(p.address+"/update/", "application/json", bytes.NewReader(json))

		resp.Body.Close()
	}
	return nil
}
