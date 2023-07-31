package metric

import (
	"errors"
	"reflect"
	runt "runtime"
)

type Runtime struct {
	Alloc         Gauge
	BuckHashSys   Gauge
	Frees         Gauge
	GCCPUFraction Gauge
	GCSys         Gauge
	HeapAlloc     Gauge
	HeapIdle      Gauge
	HeapInuse     Gauge
	HeapObjects   Gauge
	HeapReleased  Gauge
	HeapSys       Gauge
	LastGC        Gauge
	Lookups       Gauge
	MCacheInuse   Gauge
	MCacheSys     Gauge
	MSpanInuse    Gauge
	MSpanSys      Gauge
	Mallocs       Gauge
	NextGC        Gauge
	NumForcedGC   Gauge
	NumGC         Gauge
	OtherSys      Gauge
	PauseTotalNs  Gauge
	StackInuse    Gauge
	StackSys      Gauge
	Sys           Gauge
	TotalAlloc    Gauge
}

func NewRuntime() *Runtime {
	return &Runtime{}
}

func (r *Runtime) Update() error {
	var runTime runt.MemStats
	runt.ReadMemStats(&runTime)

	values := reflect.ValueOf(r).Elem()
	types := values.Type()
	runTimeValues := reflect.ValueOf(runTime)

	for i := 0; i < values.NumField(); i++ {
		field := types.Field(i)
		runTimeField := runTimeValues.FieldByName(field.Name)

		var v float64
		switch runTimeField.Kind() {
		case reflect.Float64:
			v = runTimeField.Float()

		case reflect.Uint64, reflect.Uint32:
			v = float64(runTimeField.Uint())

		default:
			return errors.New("invalid metric type")
		}

		values.Field(i).SetFloat(v)
	}
	return nil
}
