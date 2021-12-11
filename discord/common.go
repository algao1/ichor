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
	GlucoseWeeklyUsage  = "!weekly [+/-]#"
	GlucosePredictUsage = "!predict"

	TimeFormat = "2006-01-02 03:04 PM"
	HourFormat = "3 PM"
)

func weekStart(t time.Time) time.Time {
	rounded := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, loc)
	if wd := rounded.Weekday(); wd == time.Sunday {
		rounded = rounded.AddDate(0, 0, -6)
	} else {
		rounded = rounded.AddDate(0, 0, -int(wd)+1)
	}
	return rounded
}

func daySeconds(t time.Time) int {
	year, month, day := t.Date()
	t2 := time.Date(year, month, day, 0, 0, 0, 0, loc)
	return int(t.Sub(t2).Seconds())
}

func localFormat(t time.Time) string {
	return t.In(loc).Format(TimeFormat)
}
