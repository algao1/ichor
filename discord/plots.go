package discord

import (
	"bytes"
	"fmt"
	"image/color"
	"io"
	"time"

	"github.com/algao1/ichor/store"
	"github.com/lucasb-eyer/go-colorful"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

var warnColour = color.RGBA{R: 191, G: 191, B: 191}

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

func plotLowHighLines(min, max float64, p *plot.Plot) error {
	tl, err := plotter.NewLine(plotter.XYs{plotter.XY{X: 0, Y: max}, plotter.XY{X: 24 * 3600, Y: max}})
	if err != nil {
		return err
	}
	tl.LineStyle.Color = warnColour

	bl, err := plotter.NewLine(plotter.XYs{plotter.XY{X: 0, Y: min}, plotter.XY{X: 24 * 3600, Y: min}})
	if err != nil {
		return err
	}
	bl.LineStyle.Color = warnColour

	p.Add(tl)
	p.Add(bl)

	return nil
}

func PlotOverlayWeekly(min, max float64, pts []*store.TimePoint) (io.Reader, error) {
	if len(pts) == 0 {
		return nil, fmt.Errorf("no points given")
	}

	start := pts[0].Time
	end := pts[len(pts)-1].Time
	if end.Sub(start).Hours() > 7*24 {
		return nil, fmt.Errorf("points must be within a week")
	}

	p := plot.New()
	p.Title.Text = "Weekly Overlay"
	p.X.Label.Text = "Hour (EST)"
	p.Y.Label.Text = "Glucose (mmol/l)"
	p.X.Tick.Marker = HourTicks{}

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
		if day != time.Now().In(loc).Weekday().String() {
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
