package metric

import (
	"time"

	"github.com/Xacor/go-metrics/internal/logger"
	"go.uber.org/zap"
)

// Монитор содержит канал, по которому отрпавляет значения метрик раз в d сек
type Monitor struct {
	C       chan UpdateResult
	t       *time.Ticker
	metrics *Metrics
}

type UpdateResult struct {
	Metrtics Metrics
	Err      error
}

func NewMonitor(d time.Duration) (*Monitor, error) {
	m, err := NewMetrics()
	if err != nil {
		return nil, err
	}
	c := make(chan UpdateResult, 1)
	t := time.NewTicker(d)
	monitor := &Monitor{
		C:       c,
		t:       t,
		metrics: m,
	}
	monitor.run()

	return monitor, nil
}

func (m *Monitor) Close() {
	close(m.C)
}

// вспомогательная функция, чтобы слать метрки в канал без блокировки
func sendResult(c chan UpdateResult, val UpdateResult) {
	select {
	case c <- val:
	default:
	}
}

func (m *Monitor) run() {
	logger.Get().Debug("[monitor] started")
	go func() {
		for t := range m.t.C {
			logger.Get().Debug("[monitor]", zap.Time("tick", t))
			err := update(m.metrics)
			result := UpdateResult{
				Metrtics: *m.metrics,
				Err:      err,
			}
			sendResult(m.C, result)
		}
	}()
}
