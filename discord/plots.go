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
	"gonum.org/v1/plot/font"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
)

var warnColour, _ = colorful.Hex("#484a47")

var (
	carbColour, _    = colorful.Hex("#21897e")
	insulinColour, _ = colorful.Hex("#8980f5")
)

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

type InvertPyramidGlyph struct{}

func (InvertPyramidGlyph) DrawGlyph(c *draw.Canvas, sty draw.GlyphStyle, pt vg.Point) {
	sinπover6 := font.Length(math.Sin(math.Pi / 6))
	cosπover6 := font.Length(math.Cos(math.Pi / 6))

	r := sty.Radius + (sty.Radius-sty.Radius*sinπover6)/2
	p := make(vg.Path, 0, 4)
	p.Move(vg.Point{X: pt.X, Y: pt.Y - r})
	p.Line(vg.Point{X: pt.X - r*cosπover6, Y: pt.Y + r*sinπover6})
	p.Line(vg.Point{X: pt.X + r*cosπover6, Y: pt.Y + r*sinπover6})
	p.Close()
	c.Fill(p)
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

// TODO: Generalize plotCarbohydrates and plotInsulin better to avoid
//			 redundant code and to support predictions.
// TODO: Also need to account for stacking values (which is rare).

func plotCarbohydrates(yrange float64, xys plotter.XYs, carbs []store.Carbohydrate, p *plot.Plot) error {
	offset := yrange / 20
	carbxys := make(plotter.XYs, 0)
	c := 0

	for _, carb := range carbs {
		carbx := float64(carb.Time.Unix())
		for c < len(xys)-1 && carbx-xys[c].X > 0 {
			c++
		}

		carby := xys[c].Y
		if c < len(xys)-1 {
			carby += (xys[c].Y - xys[c+1].Y) * (carbx - xys[c].X) / 300
		}

		carbxys = append(carbxys, plotter.XY{X: carbx, Y: carby - offset})
	}

	cs, err := plotter.NewScatter(carbxys)
	if err != nil {
		return err
	}
	cs.GlyphStyle.Color = carbColour
	cs.GlyphStyle.Shape = draw.PyramidGlyph{}
	cs.GlyphStyle.Radius = 0.2 * font.Centimeter

	p.Add(cs)
	p.Legend.Add("Carbohydrates", cs)

	return nil
}

func plotInsulin(yrange float64, xys plotter.XYs, insulin []store.Insulin, p *plot.Plot) error {
	offset := yrange / 20
	dosexys := make(plotter.XYs, 0)
	c := 0

	for _, dose := range insulin {
		dosex := float64(dose.Time.Unix())
		for c < len(xys)-1 && dosex-xys[c].X > 0 {
			c++
		}

		dosey := xys[c].Y
		if c < len(xys)-1 {
			dosey += (xys[c].Y - xys[c+1].Y) * (dosex - xys[c].X) / 300
		}

		dosexys = append(dosexys, plotter.XY{X: dosex, Y: xys[c].Y + offset})
	}

	ds, err := plotter.NewScatter(dosexys)
	if err != nil {
		return err
	}
	ds.GlyphStyle.Color = insulinColour
	ds.GlyphStyle.Shape = InvertPyramidGlyph{}
	ds.GlyphStyle.Radius = 0.2 * font.Centimeter

	p.Add(ds)
	p.Legend.Add("Insulin", ds)

	return nil
}

func PlotRecentAndPreds(min, max float64, pts []store.TimePoint, preds []store.TimePoint,
	carbs []store.Carbohydrate, insulin []store.Insulin) (io.Reader, error) {
	if len(pts) == 0 {
		return nil, fmt.Errorf("no points given")
	}

	p := plot.New()
	p.Title.Text = "Current Values"
	p.X.Label.Text = "Hour (EST)"
	p.Y.Label.Text = "Glucose (mmol/l)"
	p.X.Tick.Marker = RecentTicks{}

	p.Y.Min = math.Max(0, min-1)
	p.Y.Max = max + 1

	minSoFar := min
	maxSoFar := max

	xys := make(plotter.XYs, len(pts))
	for i, pt := range pts {
		maxSoFar = math.Max(maxSoFar, pt.Value)
		minSoFar = math.Min(minSoFar, pt.Value)
		xys[i] = plotter.XY{X: float64(pt.Time.Unix()), Y: pt.Value}
	}

	l, err := plotter.NewLine(xys)
	if err != nil {
		return nil, err
	}
	p.Add(l)
	p.Legend.Add("Observed", l)

	// TODO: Plot predictions here...

	if err = plotCarbohydrates(maxSoFar-minSoFar, xys, carbs, p); err != nil {
		return nil, err
	}

	if err = plotInsulin(maxSoFar-minSoFar, xys, insulin, p); err != nil {
		return nil, err
	}

	if err = plotLowHighLines(min, max, p); err != nil {
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
