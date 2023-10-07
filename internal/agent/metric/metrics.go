package metric

import (
	"golang.org/x/sync/errgroup"
)

type Gauge float64

type Counter int64

type Metrics struct {
	Runtime    *Runtime
	Process    *Process
	Additional *Additional
}

func NewMetrics() (*Metrics, error) {
	proc, err := NewProcess()
	if err != nil {
		return nil, err
	}

	return &Metrics{
		Runtime:    NewRuntime(),
		Process:    proc,
		Additional: NewAdditional(),
	}, nil
}

func update(m *Metrics) error {
	g := new(errgroup.Group)

	g.Go(m.Runtime.Update)
	g.Go(m.Process.Update)
	g.Go(m.Additional.Update)

	return g.Wait()
}
