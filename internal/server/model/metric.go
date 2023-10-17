// Модуль описывает используемые модели данных.
package model

// Возможные типы метрик.
const (
	TypeCounter = "counter"
	TypeGauge   = "gauge"
)

// Модель метрики
type Metrics struct {
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
	Name  string   `json:"id"`
	MType string   `json:"type"`
}
