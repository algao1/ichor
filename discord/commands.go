package discord

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/algao1/ichor/store"
	"github.com/bwmarrin/discordgo"
	"gonum.org/v1/gonum/stat"
)

const (
	WarnLevel1 = 3381504  // #339900
	WarnLevel2 = 10079283 // #99cc33
	WarnLevel3 = 16763904 // #ffcc00
	WarnLevel4 = 16750950 // #ff9966
	WarnLevel5 = 13382400 // #cc3300
)

func sendWarnMessage(dg *discordgo.Session, cid, desc string) {
	dg.ChannelMessageSendComplex(cid, &discordgo.MessageSend{
		Embed: &discordgo.MessageEmbed{
			Title:       "Warning",
			Description: desc,
			Color:       int(WarnLevel3),
		},
	})
}

func cmdGetGlucoseData(dg *discordgo.Session, m *discordgo.MessageCreate, s *store.Store, args []string) {
	var msg string

	pts, err := s.GetPoints(time.Now().Add(-12*time.Hour), time.Now(), store.FieldGlucose)
	if err != nil {
		msg = fmt.Sprintf("unable to get points: %s", err)
		sendWarnMessage(dg, m.ChannelID, msg)
		return
	}

	x := make([]float64, len(pts))
	for i, pt := range pts {
		x[i] = pt.Value
	}

	confObj, err := s.GetObject(store.IndexConfig)
	if err != nil {
		msg = fmt.Sprintf("unable to load config: %s", err)
		sendWarnMessage(dg, m.ChannelID, msg)
		return
	}

	var conf store.Config
	json.Unmarshal(confObj, &conf)

	r, err := PlotRecentAndPreds(conf.LowThreshold, conf.HighThreshold, pts, nil)
	if err != nil {
		msg = fmt.Sprintf("unable to generate graph: %s", err)
		sendWarnMessage(dg, m.ChannelID, msg)
		return
	}

	curPt := pts[len(pts)-1]

	mean := stat.Mean(x, nil)
	std := stat.StdDev(x, nil)

	img := &discordgo.File{Name: "recentAndPreds.png", Reader: r}

	dg.ChannelMessageSendComplex(m.ChannelID, &discordgo.MessageSend{
		Embed: &discordgo.MessageEmbed{
			Title: "Recent Glucose & Predictions",
			Image: &discordgo.MessageEmbedImage{URL: "attachment://" + img.Name},
			Fields: []*discordgo.MessageEmbedField{
				{Name: "Current", Value: floatToString(curPt.Value)},
				{Name: "Trend", Value: trendToString(curPt.Trend)},
				{Name: "Mean", Value: floatToString(mean)},
				{Name: "Standard Deviation", Value: floatToString(std)},
			},
			Color: int(WarnLevel1),
		},
		Files: []*discordgo.File{img},
	})
}

func cmdGetWeeklyReport(dg *discordgo.Session, m *discordgo.MessageCreate, s *store.Store, args []string) {
	var msg string

	if len(args) != 1 {
		msg = fmt.Sprintf("need %d args but got %d: %s", 1, len(args), GlucoseWeeklyUsage)
		sendWarnMessage(dg, m.ChannelID, msg)
		return
	}

	n, err := strconv.Atoi(args[0])
	if err != nil {
		msg = fmt.Sprintf("not a number %s: %s", args[0], GlucoseWeeklyUsage)
		sendWarnMessage(dg, m.ChannelID, msg)
		return
	}

	t := time.Now().In(loc).AddDate(0, 0, -7*n)
	ws := weekStart(t)

	pts, err := s.GetPoints(ws, ws.AddDate(0, 0, 7), store.FieldGlucose)
	if err != nil {
		msg = fmt.Sprintf("unable to get points: %s", err)
		sendWarnMessage(dg, m.ChannelID, msg)
		return
	}

	confObj, err := s.GetObject(store.IndexConfig)
	if err != nil {
		msg = fmt.Sprintf("unable to load config: %s", err)
		sendWarnMessage(dg, m.ChannelID, msg)
		return
	}

	var conf store.Config
	json.Unmarshal(confObj, &conf)

	r, err := PlotOverlayWeekly(conf.LowThreshold, conf.HighThreshold, pts)
	if err != nil {
		msg = fmt.Sprintf("unable to generate weekly plot: %s", err)
		sendWarnMessage(dg, m.ChannelID, msg)
		return
	}

	img := &discordgo.File{Name: "weeklyOverlay.png", Reader: r}

	dg.ChannelMessageSendComplex(m.ChannelID, &discordgo.MessageSend{
		Embed: &discordgo.MessageEmbed{
			Title: "Weekly Overlay",
			Image: &discordgo.MessageEmbedImage{URL: "attachment://" + img.Name},
			Color: int(WarnLevel1),
		},
		Files: []*discordgo.File{img},
	})
}
