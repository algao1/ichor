package store

import "time"

const (
	FieldGlucose     = "glucose"
	FieldGlucosePred = "glucose-pred"
)

type Trend int

const (
	DoubleUp Trend = iota
	SingleUp
	HalfUp
	Flat
	HalfDown
	SingleDown
	DoubleDown
	Missing
)

type TimePoint struct {
	Time  time.Time
	Value float64
	Trend Trend
}
