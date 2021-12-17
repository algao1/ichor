package discord

import (
	"bytes"
	"fmt"
	"image/color"
	"io"
	"math"
	"time"

	"github.com/algao1/ichor/store"
	"github.com/lucasb-eyer/go-colorful"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

var warnColour, _ = colorful.Hex("#484a47")

var (
	MondayColour, _    = colorful.Hex("#517AB8")
	TuesdayColour, _   = colorful.Hex("#191970")
	WednesdayColour, _ = colorful.Hex("#ff4500")
	ThursdayColour, _  = colorful.Hex("#e9967a")
	FridayColour, _    = colorful.Hex("#00ff7f")
	SaturdayColour, _  = colorful.Hex("#00bfff")
	SundayColour, _    = colorful.Hex("#ff1493")
)

var weekDayColours = map[string]color.Color{
	"Monday":    MondayColour,
	"Tuesday":   TuesdayColour,
	"Wednesday": WednesdayColour,
	"Thursday":  ThursdayColour,
	"Friday":    FridayColour,
	"Saturday":  SaturdayColour,
	"Sunday":    SundayColour,
}

type HourTicks struct {
	Ticker plot.Ticker
	Time   func(t float64) time.Time
}

func (t HourTicks) Ticks(min, max float64) []plot.Tick {
	ticks := []plot.Tick{}

	for i := 0; i <= 24; i++ {
		var label string
		if i%3 == 0 {
			label = time.Date(0, 0, 0, i, 0, 0, 0, loc).Format(HourFormat)
		}

		ticks = append(ticks, plot.Tick{
			Value: float64(i * 3600),
			Label: label,
		})
	}

	return ticks
}

type RecentTicks struct {
	Ticker plot.Ticker
	Time   func(t float64) time.Time
}

func (t RecentTicks) Ticks(min, max float64) []plot.Tick {
	ticks := []plot.Tick{}

	st := time.Unix(int64(min), 0).In(loc)
	st = time.Date(st.Year(), st.Month(), st.Day(), st.Hour(), 0, 0, 0, loc)

	et := time.Unix(int64(max), 0).In(loc)

	for ; st.Before(et); st = st.Add(15 * time.Minute) {
		tick := plot.Tick{Value: float64(st.Unix())}
		if st.Minute() == 0 {
			tick.Label = st.Format(HourFormat)
		}
		ticks = append(ticks, tick)
	}

	return ticks
}

func plotLowHighLines(min, max float64, p *plot.Plot) error {
	tl, err := plotter.NewLine(plotter.XYs{plotter.XY{X: p.X.Min, Y: max}, plotter.XY{X: p.X.Max, Y: max}})
	if err != nil {
		return err
	}
	tl.LineStyle.Color = warnColour
	tl.LineStyle.Width = vg.Points(1)
	tl.LineStyle.Dashes = []vg.Length{vg.Points(5), vg.Points(10)}

	bl, err := plotter.NewLine(plotter.XYs{plotter.XY{X: p.X.Min, Y: min}, plotter.XY{X: p.X.Max, Y: min}})
	if err != nil {
		return err
	}
	bl.LineStyle.Color = warnColour
	bl.LineStyle.Width = vg.Points(1)
	bl.LineStyle.Dashes = []vg.Length{vg.Points(5), vg.Points(10)}

	p.Add(tl, bl)

	return nil
}

func PlotRecentAndPreds(min, max float64, pts []store.TimePoint, preds []store.TimePoint) (io.Reader, error) {
	if len(pts) == 0 {
		return nil, fmt.Errorf("no points given")
	}

	p := plot.New()
	p.Title.Text = "Current Values"
	p.X.Label.Text = "Hour (EST)"
	p.Y.Label.Text = "Glucose (mmol/l)"
	p.X.Tick.Marker = RecentTicks{}

	p.Y.Min = math.Max(0, min-1)

	xys := make(plotter.XYs, len(pts))
	for i, pt := range pts {
		xys[i] = plotter.XY{X: float64(pt.Time.Unix()), Y: pt.Value}
	}

	l, err := plotter.NewLine(xys)
	if err != nil {
		return nil, err
	}
	p.Add(l)
	p.Legend.Add("Observed", l)

	// Plot predictions here...

	err = plotLowHighLines(min, max, p)
	if err != nil {
		return nil, err
	}

	wt, err := p.WriterTo(18*vg.Inch, 6*vg.Inch, "png")
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	_, err = wt.WriteTo(buf)
	if err != nil {
		return nil, err
	}

	return buf, nil
}

func PlotOverlayWeekly(min, max float64, pts []store.TimePoint) (io.Reader, error) {
	if len(pts) == 0 {
		return nil, fmt.Errorf("no points given")
	}

	start := pts[0].Time
	end := pts[len(pts)-1].Time
	if end.Sub(start).Hours() > 7*24 {
		return nil, fmt.Errorf("points must be within a week")
	}

	dotted := true
	if end.In(loc).Month() == time.Now().In(loc).Month() &&
		end.In(loc).Day() == time.Now().In(loc).Day() {
		dotted = false
	}

	p := plot.New()
	p.Title.Text = "Weekly Overlay"
	p.X.Label.Text = "Hour (EST)"
	p.Y.Label.Text = "Glucose (mmol/l)"
	p.X.Tick.Marker = HourTicks{}

	p.X.Min = 0
	p.X.Max = 24 * 3600

	p.Y.Min = math.Max(0, min-1)

	dayPts := make(map[string]plotter.XYs)

	for _, pt := range pts {
		wd := pt.Time.In(loc).Weekday().String()
		if _, ok := dayPts[wd]; !ok {
			dayPts[wd] = make(plotter.XYs, 0)
		}

		dayPts[wd] = append(dayPts[wd],
			plotter.XY{
				X: float64(daySeconds(pt.Time)),
				Y: pt.Value,
			},
		)
	}

	for day, pts := range dayPts {
		l, err := plotter.NewLine(pts)
		if err != nil {
			return nil, err
		}

		// Graph formatting.
		if day != time.Now().In(loc).Weekday().String() || dotted {
			l.LineStyle.Width = vg.Points(1)
			l.LineStyle.Dashes = []vg.Length{vg.Points(3), vg.Points(3)}
		}
		l.LineStyle.Color = weekDayColours[day]
		p.Add(l)
		p.Legend.Add(day, l)
	}

	err := plotLowHighLines(min, max, p)
	if err != nil {
		return nil, err
	}

	wt, err := p.WriterTo(18*vg.Inch, 6*vg.Inch, "png")
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	_, err = wt.WriteTo(buf)
	if err != nil {
		return nil, err
	}

	return buf, nil
}
