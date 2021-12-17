package discord

import (
	"context"
	"fmt"

	"github.com/algao1/ichor/store"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/session"
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
	ses.AddIntents(gateway.IntentDirectMessages)
	ses.AddHandler(interactionCreate(b.ses, b.sto))

	if alertCh != nil {
		go b.handleAlerts()
	}

	return b, nil
}

func (b *Bot) Run(ctx context.Context) error {
	err := b.ses.Open(ctx)
	if err != nil {
		return err
	}

	app, err := b.ses.CurrentApplication()
	if err != nil {
		return err
	}
	appID := app.ID

	commands, err := b.ses.Commands(appID)
	if err != nil {
		return err
	}

	// Delete old commands.
	for _, command := range commands {
		b.ses.DeleteCommand(appID, command.ID)
	}

	// Add registered commands.
	for _, command := range registeredCommands {
		_, err := b.ses.CreateCommand(appID, command)
		if err != nil {
			return err
		}
	}

	return nil
}

func (b *Bot) Stop() error {
	return b.ses.Close()
}

func (b *Bot) handleAlerts() {
	for {
		var msg string
		alert := <-b.alerts

		var obs []store.TimePoint
		if err := b.sto.GetLastPoints(store.FieldGlucose, 1, &obs); err != nil {
			msg = fmt.Sprintf("unable to get points: %s", err)
			sendWarnMessage(b.ses, b.chid, msg)
			continue
		}
		ob := obs[0]

		var preds []store.TimePoint
		if err := b.sto.GetLastPoints(store.FieldGlucosePred, 1, &preds); err != nil {
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

		b.ses.SendEmbeds(b.chid, discord.Embed{
			Description: msg,
			Color:       discord.Color(WarnLevel5),
		})
	}
}
