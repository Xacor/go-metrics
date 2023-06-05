package http

import (
	"log"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"time"

	"github.com/Xacor/go-metrics/internal/agent/metric"
)

type Poller struct {
	pollInterval   uint
	reportInterval uint
	address        string
	metrics        *metric.Metrics
	client         *http.Client
}

func NewPoller(pollInterval, reportInterval uint, address string) *Poller {
	return &Poller{
		pollInterval:   pollInterval,
		reportInterval: reportInterval,
		metrics:        metric.NewMetrics(),
		address:        address,
		client:         &http.Client{},
	}
}

func (p *Poller) Run() {
	log.Println("poller started")
	for i := uint(0); ; i++ {
		log.Println("poller tick:", i)
		time.Sleep(time.Second * 1)
		if i%p.pollInterval == 0 {
			p.metrics.Update()
		}
		if i%p.reportInterval == 0 {
			p.SendRequests()
		}
	}
}

// http://<АДРЕС_СЕРВЕРА>/update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>

func (p *Poller) SendRequests() error {
	values := reflect.ValueOf(p.metrics).Elem()
	types := values.Type()
	for i := 0; i < values.NumField(); i++ {
		field := types.Field(i)
		value := values.Field(i)
		var strVal, strType string

		switch value.Kind() {
		case reflect.Uint64:
			strVal = strconv.FormatUint(value.Uint(), 10)
			strType = "counter"

		case reflect.Int64:
			strVal = strconv.FormatInt(value.Int(), 10)
			strType = "counter"

		case reflect.Float64:
			strVal = strconv.FormatFloat(value.Float(), 'f', -1, 64)
			strType = "gauge"

		default:
			log.Println("unexpected kind:", value.Kind(), value)
		}
		url, err := url.JoinPath(p.address, "update", strType, field.Name, strVal)
		log.Println(url)
		if err != nil {
			return err
		}

		resp, err := http.Post(url, "text/plain", nil)
		if err != nil {
			log.Println(err)
		}
		log.Println(resp.StatusCode)
		resp.Body.Close()
	}
	return nil
}
