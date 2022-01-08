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

// Insulin types.
const (
	RapidActing = "rapid"
	LongActing  = "long"
)

var Fields = []string{
	FieldGlucose,
	FieldGlucosePred,
	FieldCarbohydrate,
	FieldInsulin,
	FieldObject,
}

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
	Time  time.Time `csv:"time"`
	Value float64   `csv:"value"`
	Trend Trend     `csv:"trend"`
}

type Carbohydrate struct {
	Time  time.Time `csv:"time"`
	Value int       `csv:"value"`
}

type Insulin struct {
	Time  time.Time `csv:"time"`
	Type  string    `csv:"type"`
	Value int       `csv:"value"`
}

type Config struct {
	WarningTimeout time.Duration
	LowThreshold   float64
	HighThreshold  float64
}
