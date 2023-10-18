package metric

import (
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

type Process struct {
	CPUutilization []Gauge
	TotalMemory    Gauge
	FreeMemory     Gauge
}

func NewProcess() (*Process, error) {
	cores, err := cpu.Counts(true)
	if err != nil {
		return nil, err
	}
	return &Process{
		TotalMemory:    0,
		FreeMemory:     0,
		CPUutilization: make([]Gauge, cores),
	}, nil
}

func (p *Process) Update() error {
	v, err := mem.VirtualMemory()
	if err != nil {
		return err
	}
	p.TotalMemory = Gauge(v.Total)
	p.FreeMemory = Gauge(v.Free)

	cpu, err := cpu.Percent(0, true)
	if err != nil {
		return err
	}
	for i, v := range cpu {
		p.CPUutilization[i] = Gauge(v)
	}

	return nil
}
