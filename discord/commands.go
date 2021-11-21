package discord

import (
	"fmt"
	"strconv"
	"time"

	"github.com/algao1/ichor/store"
	"github.com/bwmarrin/discordgo"
	"gonum.org/v1/gonum/stat"
)

var loc, _ = time.LoadLocation("Canada/Eastern")

const (
	GlucoseDataUsage = "!glucose h/d/w/m #"
)

func inlineStr(s string) string {
	return fmt.Sprintf("```\n%s```", s)
}

func cmdGetGlucoseData(dg *discordgo.Session, m *discordgo.MessageCreate, s *store.Store, args []string) error {
	var msg string

	// Note: Probably need to find a more elegant way to return error messages.
	// 		 This is getting a little bit annoying...
	if len(args) != 2 {
		msg = fmt.Sprintf("need %d args but got %d: %s", 2, len(args), GlucoseDataUsage)
		_, err := dg.ChannelMessageSend(m.ChannelID, inlineStr(msg))
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
		_, err := dg.ChannelMessageSend(m.ChannelID, inlineStr(msg))
		return err
	}

	n, err := strconv.Atoi(args[1])
	if err != nil {
		msg = fmt.Sprintf("not a number %s: %s", args[1], GlucoseDataUsage)
		_, err := dg.ChannelMessageSend(m.ChannelID, inlineStr(msg))
		return err
	}

	timeUnit *= time.Duration(n)

	tps, err := s.GetPoints(time.Now().Add(-timeUnit), time.Now(), "glucose")
	if err != nil {
		msg = fmt.Sprintf("unable to get points: %s", err)
		_, err := dg.ChannelMessageSend(m.ChannelID, inlineStr(msg))
		return err
	}

	x := make([]float64, len(tps))

	for i, tp := range tps {
		msg += fmt.Sprintf("%s %5.2f\n", tp.Time.In(loc).Format("2006-01-02 03:04 PM"), tp.Value)
		x[i] = tp.Value
	}

	mean := stat.Mean(x, nil)
	std := stat.StdDev(x, nil)

	msg += "\n"
	msg += fmt.Sprintf("mean: %.2f std: %.2f", mean, std)

	_, err = dg.ChannelMessageSend(m.ChannelID, inlineStr(msg))
	return err
}
