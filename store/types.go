package store

import "time"

const (
	FieldGlucose      = "glucose"
	FieldGlucosePred  = "glucose-pred"
	FieldCarbohydrate = "carbohydrate"
	FieldInsulin      = "insulin"
	FieldObject       = "obj"

	IndexConfig        = "config"
	IndexTimeoutExpire = "timeout-expire"
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

type Carbohydrate struct {
	Time  time.Time
	Value int
}

type Insulin struct {
	Time  time.Time
	Type  string
	Value int
}

type Config struct {
	WarningTimeout time.Duration
	LowThreshold   float64
	HighThreshold  float64
}
