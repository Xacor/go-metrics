package metric

import (
	"encoding/json"
	"fmt"
	"reflect"
)

const (
	typeCounter = "counter"
	typeGauge   = "gauge"
)

type jsonMetric struct {
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
	ID    string   `json:"id"`
	MType string   `json:"type"`
}

func readStruct(st interface{}) ([]jsonMetric, error) {
	val := reflect.ValueOf(st)
	if val.Kind() == reflect.Pointer {
		val = val.Elem()
	}
	result := make([]jsonMetric, 0)
	for i := 0; i < val.NumField(); i++ {
		f := val.Field(i)

		switch f.Kind() {
		case reflect.Struct:
			r, err := readStruct(f)
			if err != nil {
				return nil, err
			}
			result = append(result, r...)
		case reflect.Pointer:
			if f.Elem().CanInterface() {
				r, err := readStruct(f.Elem().Interface())
				if err != nil {
					return nil, err
				}
				result = append(result, r...)
			}
		case reflect.Slice:
			for j := 0; j < f.Len(); j++ {
				v := f.Index(j).Float()
				result = append(result, jsonMetric{
					ID:    fmt.Sprintf("%v%v", val.Type().Field(i).Name, j+1),
					MType: typeGauge,
					Value: &v,
				})
			}
		case reflect.Uint64:
			v := int64(f.Uint())
			result = append(result, jsonMetric{
				ID:    val.Type().Field(i).Name,
				MType: typeCounter,
				Delta: &v,
			})
		case reflect.Int64:
			v := f.Int()
			result = append(result, jsonMetric{
				ID:    val.Type().Field(i).Name,
				MType: typeCounter,
				Delta: &v,
			})
		case reflect.Float64:
			v := f.Float()
			result = append(result, jsonMetric{
				ID:    val.Type().Field(i).Name,
				MType: typeGauge,
				Value: &v,
			})
		case reflect.Chan:
			continue
		default:
			return nil, fmt.Errorf("ivalid metric Kind: %v %v", f.Kind(), val.Type().Field(i).Name)
		}
	}

	return result, nil
}

func (m *Metrics) MarshalJSON() ([]byte, error) {
	metrics, err := readStruct(m)
	if err != nil {
		return nil, err
	}

	json, err := json.Marshal(metrics)
	if err != nil {
		return nil, err
	}
	return json, nil
}
