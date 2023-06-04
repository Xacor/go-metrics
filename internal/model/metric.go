package model

import (
	"errors"
	"log"
	"reflect"
)

type MetricType int

const (
	Counter MetricType = iota
	Guage
)

type Metric struct {
	ID    string
	Type  MetricType
	Value interface{}
}

func (m *Metric) Set(value interface{}) error {
	log.Println(reflect.TypeOf(value).String())
	log.Println("set")
	v, ok := value.(float64)
	if !ok {
		return errors.New("unexpected type")
	}
	m.Value = v

	return nil
}

func (m *Metric) Add(value interface{}) error {
	v, ok := value.(int64)
	if !ok {
		return errors.New("unexpected type")
	}
	m.Value = m.Value.(int64) + v

	return nil
}
