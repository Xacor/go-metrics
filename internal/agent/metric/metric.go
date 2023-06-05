package metric

import (
	"log"
	"math/rand"
	"reflect"
	"runtime"
)

type Gauge float64

type Counter int64

type Metrics struct {
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
	PollCount     Counter
	RandomValue   Gauge
}

// type Runtime struct {
// 	Alloc         Gauge
// 	BuckHashSys   Gauge
// 	Frees         Gauge
// 	GCCPUFraction Gauge
// 	GCSys         Gauge
// 	HeapAlloc     Gauge
// 	HeapIdle      Gauge
// 	HeapInuse     Gauge
// 	HeapObjects   Gauge
// 	HeapReleased  Gauge
// 	HeapSys       Gauge
// 	LastGC        Gauge
// 	Lookups       Gauge
// 	MCacheInuse   Gauge
// 	MCacheSys     Gauge
// 	MSpanInuse    Gauge
// 	MSpanSys      Gauge
// 	Mallocs       Gauge
// 	NextGC        Gauge
// 	NumForcedGC   Gauge
// 	NumGC         Gauge
// 	OtherSys      Gauge
// 	PauseTotalNs  Gauge
// 	StackInuse    Gauge
// 	StackSys      Gauge
// 	Sys           Gauge
// 	TotalAlloc    Gauge
// }

// Returns up to date Metrics.
func NewMetrics() *Metrics {
	var m Metrics
	m.Update()
	return &m
}

// Updates Metrics with up to date values.
func (m *Metrics) Update() {
	log.Println("updating metrics")
	var runTime runtime.MemStats
	runtime.ReadMemStats(&runTime)

	values := reflect.ValueOf(m).Elem()
	types := values.Type()
	runTimeValues := reflect.ValueOf(runTime)

	// parse MemStats to Metrics
	for i := 0; i < values.NumField(); i++ {
		field := types.Field(i)
		runTimeField := runTimeValues.FieldByName(field.Name)

		var v float64
		switch runTimeField.Kind() { //nolint:exhaustive // default case present
		case reflect.Float64:
			v = runTimeField.Float()

		case reflect.Uint64, reflect.Uint32:
			v = float64(runTimeField.Uint())

		default:
			continue
		}

		values.Field(i).SetFloat(v)
	}

	m.PollCount++
	m.RandomValue = Gauge(rand.Float64()) //nolint:gosec // its ok to use weak random algorythm here
}
