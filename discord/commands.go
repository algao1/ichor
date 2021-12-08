package discord

import (
	"fmt"
	"strconv"
	"time"

	"github.com/algao1/ichor/store"
	"github.com/bwmarrin/discordgo"
	"gonum.org/v1/gonum/stat"
)

func cmdGetGlucoseData(dg *discordgo.Session, m *discordgo.MessageCreate, s *store.Store, args []string) error {
	var msg string

	// Note: Probably need to find a more elegant way to return error messages.
	// 			 This is getting a little bit annoying...
	if len(args) != 2 {
		msg = fmt.Sprintf("need %d args but got %d: %s", 2, len(args), GlucoseDataUsage)
		_, err := dg.ChannelMessageSend(m.ChannelID, msg)
		return err
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
		_, err := dg.ChannelMessageSend(m.ChannelID, msg)
		return err
	}

	n, err := strconv.Atoi(args[1])
	if err != nil {
		msg = fmt.Sprintf("not a number %s: %s", args[1], GlucoseDataUsage)
		_, err := dg.ChannelMessageSend(m.ChannelID, msg)
		return err
	}

	timeUnit *= time.Duration(n)

	pts, err := s.GetPoints(time.Now().Add(-timeUnit), time.Now(), store.FieldGlucose)
	if err != nil {
		msg = fmt.Sprintf("unable to get points: %s", err)
		_, err := dg.ChannelMessageSend(m.ChannelID, msg)
		return err
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

	_, err = dg.ChannelMessageSend(m.ChannelID, msg)
	return err
}

func cmdGetPredictions(dg *discordgo.Session, m *discordgo.MessageCreate, s *store.Store, args []string) error {
	var msg string

	pts, err := s.GetLastPoints(store.FieldGlucose, 1)
	if err != nil {
		msg = fmt.Sprintf("unable to get points: %s", err)
		_, err := dg.ChannelMessageSend(m.ChannelID, msg)
		return err
	}
	pt := pts[0]

	preds, err := s.GetPoints(time.Now(), time.Now().Add(2*time.Hour), store.FieldGlucosePred)
	if err != nil {
		msg = fmt.Sprintf("unable to get predictions: %s", err)
		_, err := dg.ChannelMessageSend(m.ChannelID, msg)
		return err
	}

	msg += fmt.Sprintf("%s %5.2f (current)\n", localFormat(pt.Time), pt.Value)
	for _, pred := range preds {
		msg += fmt.Sprintf("%s %5.2f\n", localFormat(pred.Time), pred.Value)
	}

	_, err = dg.ChannelMessageSend(m.ChannelID, msg)
	return err
}
