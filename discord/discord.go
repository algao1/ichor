package discord

import (
	"fmt"
	"strings"

	"github.com/algao1/ichor/store"
	"github.com/bwmarrin/discordgo"
)

type DiscordBot struct {
	dg *discordgo.Session
	s  *store.Store

	uid    string
	cid    string
	alerts <-chan Alert
}

func Create(uid, token string, s *store.Store, alertCh <-chan Alert) (*DiscordBot, error) {
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, err
	}

	// Verify that we can create a private channel.
	uch, err := dg.UserChannelCreate(uid)
	if err != nil {
		return nil, err
	}

	return &DiscordBot{dg, s, uid, uch.ID, alertCh}, nil
}

func (b *DiscordBot) Run() error {
	b.addHandlers()
	return b.dg.Open()
}

func (b *DiscordBot) Stop() error {
	return b.dg.Close()
}

func (b *DiscordBot) addHandlers() {
	b.dg.Identify.Intents = discordgo.IntentsDirectMessages
	b.dg.AddHandler(makeMessageCreate(b.uid, b.s))

	if b.alerts != nil {
		go b.handleAlerts()
	}
}

func (b *DiscordBot) handleAlerts() {
	for {
		var msg string
		alert := <-b.alerts

		obs, err := b.s.GetLastPoints(store.FieldGlucose, 1)
		if err != nil {
			msg = fmt.Sprintf("unable to get points: %s", err)
			b.dg.ChannelMessageSend(b.cid, msg)
			continue
		}
		ob := obs[0]

		preds, err := b.s.GetLastPoints(store.FieldGlucosePred, 1)
		if err != nil {
			msg = fmt.Sprintf("unable to get points: %s", err)
			b.dg.ChannelMessageSend(b.cid, msg)
			continue
		}
		pr := preds[0]

		if alert == Low {
			msg = fmt.Sprintf(
				"🔻 incoming low blood sugar\n%s %5.2f\n%s %5.2f",
				localFormat(ob.Time), ob.Value,
				localFormat(pr.Time), pr.Value,
			)
		} else {
			msg = fmt.Sprintf(
				"🔺 incoming high blood sugar\n%s %5.2f\n%s %5.2f",
				localFormat(ob.Time), ob.Value,
				localFormat(pr.Time), pr.Value,
			)
		}

		b.dg.ChannelMessageSend(b.cid, msg)
	}
}

func makeMessageCreate(uid string, s *store.Store) func(*discordgo.Session, *discordgo.MessageCreate) {
	return func(dg *discordgo.Session, m *discordgo.MessageCreate) {
		// If the author is the bot, don't do anything.
		if m.Author.ID != uid {
			return
		}

		// Definitely need to make command handling/parsing a little
		// more robust.

		parsed := strings.Fields(m.Content)
		cmd := parsed[0]

		switch cmd {
		case "!glucose":
			cmdGetGlucoseData(dg, m, s, parsed[1:])
		case "!weekly":
			cmdGetWeeklyReport(dg, m, s, parsed[1:])
		default:
		}
	}
}
