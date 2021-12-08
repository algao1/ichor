package discord

import (
	"time"
)

type Alert int

const (
	Low Alert = iota
	High
)

var loc, _ = time.LoadLocation("Canada/Eastern")

const (
	GlucoseDataUsage    = "!glucose h/d/w/m #"
	GlucosePredictUsage = "!predict"

	TimeFormat = "2006-01-02 03:04 PM"
)

func localFormat(t time.Time) string {
	return t.In(loc).Format(TimeFormat)
}
