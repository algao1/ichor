package discord

import (
	"fmt"
	"strings"

	"github.com/algao1/ichor/store"
	"github.com/bwmarrin/discordgo"
)

type DiscordBot struct {
	UserID string

	dg     *discordgo.Session
	s      *store.Store
	alerts <-chan Alert
}

func Create(uid, token string, s *store.Store, alertCh <-chan Alert) (*DiscordBot, error) {
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, err
	}
	return &DiscordBot{uid, dg, s, alertCh}, nil
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
	b.dg.AddHandler(makeMessageCreate(b.UserID, b.s))

	if b.alerts != nil {
		go b.handleAlerts()
	}
}

func (b *DiscordBot) handleAlerts() {
	uch, err := b.dg.UserChannelCreate(b.UserID)
	if err != nil {
		panic(err) // Panic for now, since we cannot do anything if uid is wrong.
	}

	for {
		select {
		case alert := <-b.alerts:
			var msg string

			pts, err := b.s.GetLastPoints(store.FieldGlucosePred, 1)
			if err != nil {
				msg = fmt.Sprintf("unable to get points: %s", err)
				b.dg.ChannelMessageSend(uch.ID, inlineStr(msg))
				continue
			}
			pt := pts[0]

			if alert == Low {
				msg = fmt.Sprintf(
					"🔻 incoming low blood sugar\n%s %5.2f",
					localFormat(pt.Time),
					pt.Value,
				)
				b.dg.ChannelMessageSend(uch.ID, inlineStr(msg))
			} else {
				msg = fmt.Sprintf(
					"🔺 incoming high blood sugar\n%s %5.2f",
					localFormat(pt.Time),
					pt.Value,
				)
				b.dg.ChannelMessageSend(uch.ID, inlineStr(msg))
			}
		}
	}
}

func makeMessageCreate(uid string, s *store.Store) func(*discordgo.Session, *discordgo.MessageCreate) {
	return func(dg *discordgo.Session, m *discordgo.MessageCreate) {
		// If the author is the bot, don't do anything.
		if m.Author.ID != uid {
			return
		}

		parsed := strings.Fields(m.Content)
		cmd := parsed[0]

		switch cmd {
		case "!glucose":
			cmdGetGlucoseData(dg, m, s, parsed[1:])
		case "!predict":
			cmdGetPredictions(dg, m, s, parsed[1:])
		default:
		}
	}
}
