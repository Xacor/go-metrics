// Модуль описывает используемые модели данных.
package model

// Возможные типы метрик.
const (
	TypeCounter = "counter"
	TypeGauge   = "gauge"
)

// Модель метрики
type Metrics struct {
	Name  string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}
