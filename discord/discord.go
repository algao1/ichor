package discord

import (
	"fmt"
	"log"
	"strings"

	"github.com/algao1/ichor/store"
	"github.com/diamondburned/arikawa/discord"
	"github.com/diamondburned/arikawa/gateway"
	"github.com/diamondburned/arikawa/session"
)

type Bot struct {
	ses    *session.Session
	sto    *store.Store
	alerts <-chan Alert

	uid  discord.UserID
	chid discord.ChannelID
}

func Create(token string, uid float64, sto *store.Store, alertCh <-chan Alert) (*Bot, error) {
	ses, err := session.New("Bot " + token)
	if err != nil {
		return nil, err
	}

	// Verify that we can create a private channel.
	duid := discord.UserID(uid)
	uch, err := ses.Client.CreatePrivateChannel(duid)
	if err != nil {
		return nil, err
	}

	b := &Bot{
		ses:    ses,
		sto:    sto,
		alerts: alertCh,
		uid:    duid,
		chid:   uch.ID,
	}

	// Add handlers.
	ses.Gateway.AddIntent(gateway.IntentDirectMessages)
	ses.AddHandler(b.makeMessageCreate())

	if alertCh != nil {
		go b.handleAlerts()
	}

	return b, nil
}

func (b *Bot) Run() error {
	return b.ses.Open()
}

func (b *Bot) Stop() error {
	return b.ses.Close()
}

func (b *Bot) handleAlerts() {
	for {
		var msg string
		alert := <-b.alerts

		obs, err := b.sto.GetLastPoints(store.FieldGlucose, 1)
		if err != nil {
			msg = fmt.Sprintf("unable to get points: %s", err)
			sendWarnMessage(b.ses, b.chid, msg)
			continue
		}
		ob := obs[0]

		preds, err := b.sto.GetLastPoints(store.FieldGlucosePred, 1)
		if err != nil {
			msg = fmt.Sprintf("unable to get points: %s", err)
			sendWarnMessage(b.ses, b.chid, msg)
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

		b.ses.SendEmbed(b.chid, discord.Embed{
			Description: msg,
			Color:       discord.Color(WarnLevel5),
		})
	}
}

func (b *Bot) makeMessageCreate() func(c *gateway.MessageCreateEvent) {
	return func(c *gateway.MessageCreateEvent) {
		m, err := b.ses.Message(b.chid, c.ID)
		if err != nil {
			log.Println("Message not found:", c.ID)
		}

		if m.Author.ID != b.uid {
			return
		}

		parsed := strings.Fields(m.Content)
		cmd := parsed[0]

		switch cmd {
		case "!glucose":
			b.cmdSendGlucoseData(parsed[1:])
		case "!weekly":
			b.cmdSendWeeklyReport(parsed[1:])
		default:
		}
	}
}
