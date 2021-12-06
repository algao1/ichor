package discord

import (
	"strings"

	"github.com/algao1/ichor/store"
	"github.com/bwmarrin/discordgo"
)

type DiscordBot struct {
	dg *discordgo.Session
}

func Create(token string, s *store.Store) (*DiscordBot, error) {
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, err
	}

	addHandlers(dg, s)

	return &DiscordBot{dg}, nil
}

func (b *DiscordBot) Run() error {
	return b.dg.Open()
}

func (b *DiscordBot) Stop() error {
	return b.dg.Close()
}

func addHandlers(dg *discordgo.Session, s *store.Store) {
	dg.Identify.Intents = discordgo.IntentsDirectMessages

	dg.AddHandler(makeMessageCreate(s))
}

func makeMessageCreate(s *store.Store) func(*discordgo.Session, *discordgo.MessageCreate) {
	return func(dg *discordgo.Session, m *discordgo.MessageCreate) {
		// If the author is the bot, don't do anything.
		if m.Author.ID == dg.State.User.ID {
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
