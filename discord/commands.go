package discord

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/algao1/ichor/store"
	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/session"
	"github.com/diamondburned/arikawa/v3/utils/sendpart"
	"gonum.org/v1/gonum/stat"
)

const (
	WarnLevel1 = 3381504  // #339900
	WarnLevel2 = 10079283 // #99cc33
	WarnLevel3 = 16763904 // #ffcc00
	WarnLevel4 = 16750950 // #ff9966
	WarnLevel5 = 13382400 // #cc3300
)

func sendWarnMessage(ses *session.Session, cid discord.ChannelID, desc string) {
	ses.SendMessageComplex(cid, api.SendMessageData{
		Embeds: []discord.Embed{{
			Title:       "Warning",
			Description: desc,
			Color:       discord.Color(WarnLevel3),
		}},
	})
}

func (b *Bot) cmdSendGlucoseData(args []string) {
	var msg string

	pts, err := b.sto.GetPoints(time.Now().Add(-12*time.Hour), time.Now(), store.FieldGlucose)
	if err != nil {
		msg = fmt.Sprintf("unable to get points: %s", err)
		sendWarnMessage(b.ses, b.chid, msg)
		return
	}

	x := make([]float64, len(pts))
	for i, pt := range pts {
		x[i] = pt.Value
	}

	confObj, err := b.sto.GetObject(store.IndexConfig)
	if err != nil {
		msg = fmt.Sprintf("unable to load config: %s", err)
		sendWarnMessage(b.ses, b.chid, msg)
		return
	}

	var conf store.Config
	json.Unmarshal(confObj, &conf)

	r, err := PlotRecentAndPreds(conf.LowThreshold, conf.HighThreshold, pts, nil)
	if err != nil {
		msg = fmt.Sprintf("unable to generate graph: %s", err)
		sendWarnMessage(b.ses, b.chid, msg)
		return
	}

	curPt := pts[len(pts)-1]

	mean := stat.Mean(x, nil)
	std := stat.StdDev(x, nil)

	img := sendpart.File{Name: "recentAndPreds.png", Reader: r}

	b.ses.SendMessageComplex(b.chid, api.SendMessageData{
		Embeds: []discord.Embed{{
			Title: "Recent Glucose & Predictions",
			Image: &discord.EmbedImage{URL: "attachment://" + img.Name},
			Fields: []discord.EmbedField{
				{Name: "Current", Value: floatToString(curPt.Value)},
				{Name: "Trend", Value: trendToString(curPt.Trend)},
				{Name: "Mean", Value: floatToString(mean)},
				{Name: "Standard Deviation", Value: floatToString(std)},
			},
			Color: discord.Color(WarnLevel1),
		}},
		Files: []sendpart.File{img},
	})
}

func (b *Bot) cmdSendWeeklyReport(args []string) {
	var msg string

	if len(args) != 1 {
		msg = fmt.Sprintf("need %d args but got %d: %s", 1, len(args), GlucoseWeeklyUsage)
		sendWarnMessage(b.ses, b.chid, msg)
		return
	}

	n, err := strconv.Atoi(args[0])
	if err != nil {
		msg = fmt.Sprintf("not a number %s: %s", args[0], GlucoseWeeklyUsage)
		sendWarnMessage(b.ses, b.chid, msg)
		return
	}

	t := time.Now().In(loc).AddDate(0, 0, -7*n)
	ws := weekStart(t)

	pts, err := b.sto.GetPoints(ws, ws.AddDate(0, 0, 7), store.FieldGlucose)
	if err != nil {
		msg = fmt.Sprintf("unable to get points: %s", err)
		sendWarnMessage(b.ses, b.chid, msg)
		return
	}

	confObj, err := b.sto.GetObject(store.IndexConfig)
	if err != nil {
		msg = fmt.Sprintf("unable to load config: %s", err)
		sendWarnMessage(b.ses, b.chid, msg)
		return
	}

	var conf store.Config
	json.Unmarshal(confObj, &conf)

	r, err := PlotOverlayWeekly(conf.LowThreshold, conf.HighThreshold, pts)
	if err != nil {
		msg = fmt.Sprintf("unable to generate weekly plot: %s", err)
		sendWarnMessage(b.ses, b.chid, msg)
		return
	}

	img := sendpart.File{Name: "weeklyOverlay.png", Reader: r}

	b.ses.SendMessageComplex(b.chid, api.SendMessageData{
		Embeds: []discord.Embed{{
			Title: "Weekly Overlay",
			Image: &discord.EmbedImage{URL: "attachment://" + img.Name},
			Color: discord.Color(WarnLevel1),
		}},
		Files: []sendpart.File{img},
	})
}
