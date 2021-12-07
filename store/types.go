package store

import "time"

const (
	FieldGlucose     = "glucose"
	FieldGlucosePred = "glucose-pred"
	FieldObject      = "obj"

	IndexConfig        = "config"
	IndexTimeoutExpore = "timeout-expire"
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

type Config struct {
	WarningTimeout time.Duration
	LowThreshold   float64
	HighThreshold  float64
}
