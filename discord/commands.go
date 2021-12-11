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

func sendEmbeddedMessage(dg *discordgo.Session, cid, title, desc string, color int, image *discordgo.File) {
	var files []*discordgo.File
	var url string

	if image != nil {
		url = "attachment://" + image.Name
		files = []*discordgo.File{image}
	}

	dg.ChannelMessageSendComplex(cid, &discordgo.MessageSend{
		Embed: &discordgo.MessageEmbed{
			Title:       title,
			Description: desc,
			Image:       &discordgo.MessageEmbedImage{URL: url},
			Color:       color,
		},
		Files: files,
	})
}

func cmdGetGlucoseData(dg *discordgo.Session, m *discordgo.MessageCreate, s *store.Store, args []string) {
	var msg string

	// Note: Probably need to find a more elegant way to return error messages.
	// 			 This is getting a little bit annoying...
	if len(args) != 2 {
		msg = fmt.Sprintf("need %d args but got %d: %s", 2, len(args), GlucoseDataUsage)
		sendEmbeddedMessage(dg, m.ChannelID, "", msg, WarnLevel3, nil)
		return
	}

	var timeUnit time.Duration
	switch args[0] {
	case "h":
		timeUnit = time.Hour
	case "d":
		timeUnit = time.Hour * 24
	case "w":
		timeUnit = time.Hour * 7 * 24
	case "m":
		timeUnit = time.Hour * 4 * 7 * 24
	default:
		msg = fmt.Sprintf("unknown timeframe %s: %s", args[0], GlucoseDataUsage)
		sendEmbeddedMessage(dg, m.ChannelID, "", msg, WarnLevel3, nil)
		return
	}

	n, err := strconv.Atoi(args[1])
	if err != nil {
		msg = fmt.Sprintf("not a number %s: %s", args[1], GlucoseDataUsage)
		sendEmbeddedMessage(dg, m.ChannelID, "", msg, WarnLevel3, nil)
		return
	}

	timeUnit *= time.Duration(n)

	pts, err := s.GetPoints(time.Now().Add(-timeUnit), time.Now(), store.FieldGlucose)
	if err != nil {
		msg = fmt.Sprintf("unable to get points: %s", err)
		sendEmbeddedMessage(dg, m.ChannelID, "", msg, WarnLevel3, nil)
		return
	}

	x := make([]float64, len(pts))

	for i, pt := range pts {
		msg += fmt.Sprintf("%s %5.2f\n", localFormat(pt.Time), pt.Value)
		x[i] = pt.Value
	}

	mean := stat.Mean(x, nil)
	std := stat.StdDev(x, nil)

	msg += "\n"
	msg += fmt.Sprintf("mean: %.2f std: %.2f", mean, std)

	sendEmbeddedMessage(dg, m.ChannelID, "", msg, WarnLevel1, nil)
}

func cmdGetPredictions(dg *discordgo.Session, m *discordgo.MessageCreate, s *store.Store, args []string) {
	var msg string

	pts, err := s.GetLastPoints(store.FieldGlucose, 1)
	if err != nil {
		msg = fmt.Sprintf("unable to get points: %s", err)
		sendEmbeddedMessage(dg, m.ChannelID, "", msg, WarnLevel3, nil)
		return
	}
	pt := pts[0]

	preds, err := s.GetPoints(time.Now(), time.Now().Add(2*time.Hour), store.FieldGlucosePred)
	if err != nil {
		msg = fmt.Sprintf("unable to get predictions: %s", err)
		sendEmbeddedMessage(dg, m.ChannelID, "", msg, WarnLevel3, nil)
		return
	}

	msg += fmt.Sprintf("%s %5.2f (current)\n", localFormat(pt.Time), pt.Value)
	for _, pred := range preds {
		msg += fmt.Sprintf("%s %5.2f\n", localFormat(pred.Time), pred.Value)
	}

	sendEmbeddedMessage(dg, m.ChannelID, "", msg, WarnLevel1, nil)
}

func cmdGetWeeklyReport(dg *discordgo.Session, m *discordgo.MessageCreate, s *store.Store, args []string) {
	var msg string

	if len(args) != 1 {
		msg = fmt.Sprintf("need %d args but got %d: %s", 1, len(args), GlucoseWeeklyUsage)
		sendEmbeddedMessage(dg, m.ChannelID, "", msg, WarnLevel3, nil)
		return
	}

	n, err := strconv.Atoi(args[0])
	if err != nil {
		msg = fmt.Sprintf("not a number %s: %s", args[0], GlucoseWeeklyUsage)
		sendEmbeddedMessage(dg, m.ChannelID, "", msg, WarnLevel3, nil)
		return
	}

	t := time.Now().In(loc).AddDate(0, 0, -7*n)
	ws := weekStart(t)

	pts, err := s.GetPoints(ws, ws.AddDate(0, 0, 7), store.FieldGlucose)
	if err != nil {
		msg = fmt.Sprintf("unable to get points: %s", err)
		sendEmbeddedMessage(dg, m.ChannelID, "", msg, WarnLevel3, nil)
		return
	}

	confObj, err := s.GetObject(store.IndexConfig)
	if err != nil {
		msg = fmt.Sprintf("unable to load config: %s", err)
		sendEmbeddedMessage(dg, m.ChannelID, "", msg, WarnLevel3, nil)
		return
	}

	var conf store.Config
	json.Unmarshal(confObj, &conf)

	r, err := PlotOverlayWeekly(conf.LowThreshold, conf.HighThreshold, pts)
	if err != nil {
		msg = fmt.Sprintf("unable to generate weekly plot: %s", err)
		sendEmbeddedMessage(dg, m.ChannelID, "", msg, WarnLevel3, nil)
		return
	}

	sendEmbeddedMessage(dg, m.ChannelID, "", msg, WarnLevel1, &discordgo.File{Name: "weeklyOverlay.png", Reader: r})
}
