package discord

import (
	"fmt"
	"log"

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
		if m.Author.Bot {
			return
		}

		tps, err := s.GetLastPoints("glucose", 15)
		if err != nil {
			log.Fatal(err)
		}

		var msg string
		for _, tp := range tps {
			msg += fmt.Sprintln(tp.Time, tp.Value)
		}

		_, err = dg.ChannelMessageSend(m.ChannelID, msg)
		if err != nil {
			log.Fatal(err)
		}
	}
}
