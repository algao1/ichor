package discord

import (
	"strconv"
	"time"

	"github.com/algao1/ichor/store"
)

type Alert int

const (
	Low Alert = iota
	High
)

var loc, _ = time.LoadLocation("Canada/Eastern")

const (
	GlucoseDataUsage   = "!glucose"
	GlucoseWeeklyUsage = "!weekly [+/-]#"

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

func floatToString(v float64) string {
	return strconv.FormatFloat(v, 'f', 2, 64)
}

func trendToString(t store.Trend) string {
	switch t {
	case store.DoubleUp:
		return "↟"
	case store.SingleUp:
		return "↑"
	case store.HalfUp:
		return "↗"
	case store.Flat:
		return "→"
	case store.HalfDown:
		return "↘"
	case store.SingleDown:
		return "↓"
	case store.DoubleDown:
		return "↡"
	default:
		return "-"
	}
}
