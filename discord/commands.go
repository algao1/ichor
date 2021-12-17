package discord

import (
	"fmt"
	"log"
	"time"

	"github.com/algao1/ichor/store"
	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/session"
	"github.com/diamondburned/arikawa/v3/utils/sendpart"
	"gonum.org/v1/gonum/stat"
)

const (
	WarnLevel1 = 3381504  // #339900
	WarnLevel2 = 10079283 // #99cc33
	WarnLevel3 = 16763904 // #ffcc00
	WarnLevel4 = 16750950 // #ff9966
	WarnLevel5 = 13382400 // #cc3300
)

var registeredCommands = []api.CreateCommandData{
	{
		Name:        "glucose",
		Description: "Get the current glucose value.",
	},
	{
		Name:        "weekly",
		Description: "Get the weekly overlay of glucose values.",
		Options: discord.CommandOptions{
			&discord.IntegerOption{
				OptionName:  "offset",
				Description: "Weekly offset.",
				Required:    true,
			},
		},
	},
}

type GlucoseReport struct {
	Value float64
	Trend store.Trend
	Mean  float64
	Std   float64
	Chart sendpart.File
}

type WeeklyReport struct {
	Chart sendpart.File
}

func sendWarnMessage(ses *session.Session, cid discord.ChannelID, desc string) {
	ses.SendMessageComplex(cid, api.SendMessageData{
		Embeds: []discord.Embed{{
			Title:       "Warning",
			Description: desc,
			Color:       discord.Color(WarnLevel3),
		}},
	})
}

func interactionWarnResponse(desc string) api.InteractionResponse {
	return api.InteractionResponse{
		Type: api.MessageInteractionWithSource,
		Data: &api.InteractionResponseData{
			Embeds: &[]discord.Embed{{
				Title:       "Warning",
				Description: desc,
				Color:       discord.Color(WarnLevel3),
			}},
		},
	}
}

func interactionCreate(ses *session.Session, sto *store.Store) func(e *gateway.InteractionCreateEvent) {
	return func(e *gateway.InteractionCreateEvent) {
		var resp api.InteractionResponse

		// This is slightly ugly.

		switch data := e.Data.(type) {
		case *discord.CommandInteraction:
			switch data.Name {
			case "glucose":
				gr, err := glucoseReport(sto)
				if err != nil {
					resp = interactionWarnResponse(err.Error())
					break
				}

				resp = api.InteractionResponse{
					Type: api.MessageInteractionWithSource,
					Data: &api.InteractionResponseData{
						Embeds: &[]discord.Embed{{
							Title: "Recent Glucose & Predictions",
							Image: &discord.EmbedImage{URL: "attachment://" + gr.Chart.Name},
							Fields: []discord.EmbedField{
								{Name: "Current", Value: floatToString(gr.Value)},
								{Name: "Trend", Value: "\\" + trendToString(gr.Trend)},
								{Name: "Mean", Value: floatToString(gr.Mean)},
								{Name: "Std", Value: floatToString(gr.Std)},
							},
							Color: discord.Color(WarnLevel1),
						}},
						Files: []sendpart.File{gr.Chart},
					},
				}
			case "weekly":
				n, err := data.Options[0].IntValue()
				if err != nil {
					resp = interactionWarnResponse(err.Error())
					break
				}

				wr, err := weeklyReport(int(n), sto)
				if err != nil {
					resp = interactionWarnResponse(err.Error())
					break
				}

				resp = api.InteractionResponse{
					Type: api.MessageInteractionWithSource,
					Data: &api.InteractionResponseData{
						Embeds: &[]discord.Embed{{
							Title: "Weekly Overlay",
							Image: &discord.EmbedImage{URL: "attachment://" + wr.Chart.Name},
							Color: discord.Color(WarnLevel1),
						}},
						Files: []sendpart.File{wr.Chart},
					},
				}
			}
		}

		if err := ses.RespondInteraction(e.ID, e.Token, resp); err != nil {
			log.Println("failed to send interaction callback: ", err)
		}
	}
}

func glucoseReport(sto *store.Store) (*GlucoseReport, error) {
	var pts []store.TimePoint
	err := sto.GetPoints(time.Now().Add(-12*time.Hour), time.Now(), store.FieldGlucose, &pts)
	if err != nil {
		return nil, fmt.Errorf("unable to get points: %w", err)
	}

	x := make([]float64, len(pts))
	for i, pt := range pts {
		x[i] = pt.Value
	}

	var conf store.Config
	if err = sto.GetObject(store.IndexConfig, &conf); err != nil {
		return nil, fmt.Errorf("unable to load config: %w", err)
	}

	r, err := PlotRecentAndPreds(conf.LowThreshold, conf.HighThreshold, pts, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to generate daily graph: %w", err)
	}

	curPt := pts[len(pts)-1]
	mean := stat.Mean(x, nil)
	std := stat.StdDev(x, nil)

	return &GlucoseReport{
		Value: curPt.Value,
		Trend: curPt.Trend,
		Mean:  mean,
		Std:   std,
		Chart: sendpart.File{Name: "glucoseChart.png", Reader: r},
	}, nil
}

func weeklyReport(offset int, sto *store.Store) (*WeeklyReport, error) {
	t := time.Now().In(loc).AddDate(0, 0, -7*offset)
	ws := weekStart(t)

	var pts []store.TimePoint
	err := sto.GetPoints(ws, ws.AddDate(0, 0, 7), store.FieldGlucose, &pts)
	if err != nil {
		return nil, fmt.Errorf("unable to get points: %w", err)
	}

	var conf store.Config
	if err = sto.GetObject(store.IndexConfig, &conf); err != nil {
		return nil, fmt.Errorf("unable to load config: %w", err)
	}

	r, err := PlotOverlayWeekly(conf.LowThreshold, conf.HighThreshold, pts)
	if err != nil {
		return nil, fmt.Errorf("unable to generate weekly plot: %w", err)
	}

	return &WeeklyReport{
		Chart: sendpart.File{Name: "weeklyOverlay.png", Reader: r},
	}, nil
}
