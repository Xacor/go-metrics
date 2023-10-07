package metric

import "math/rand"

type Additional struct {
	PollCount   Counter
	RandomValue Gauge
}

func NewAdditional() *Additional {
	return &Additional{}
}

func (a *Additional) Update() error {
	a.PollCount++
	a.RandomValue = Gauge(rand.Float64())

	return nil
}
