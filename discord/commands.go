package discord

import (
	"fmt"
	"strconv"
	"time"

	"github.com/algao1/ichor/store"
	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/session"
	"github.com/diamondburned/arikawa/v3/utils/json/option"
	"github.com/diamondburned/arikawa/v3/utils/sendpart"
	"go.uber.org/zap"
	"gonum.org/v1/gonum/stat"
)

const (
	WarnLevel1 = 3381504  // #339900
	WarnLevel2 = 10079283 // #99cc33
	WarnLevel3 = 16763904 // #ffcc00
	WarnLevel4 = 16750950 // #ff9966
	WarnLevel5 = 13382400 // #cc3300
)

var defaultFooter = discord.EmbedFooter{
	Text: "This bot is still under construction.",
}

var inlineBlankField = discord.EmbedField{
	Name:   "\u200b",
	Value:  "\u200b",
	Inline: true,
}

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
				Min:         option.ZeroInt,
				Required:    true,
			},
		},
	},
	{
		Name:        "carbohydrates",
		Description: "Insert the estimated carbohydrate intake.",
		Options: discord.CommandOptions{
			&discord.IntegerOption{
				OptionName:  "amount",
				Description: "Amount of carbohydrates (grams).",
				Min:         option.ZeroInt,
				Required:    true,
			},
			&discord.IntegerOption{
				OptionName:  "offset",
				Description: "Offset in minutes.",
				Min:         option.ZeroInt,
				Required:    false,
			},
		},
	},
	{
		Name:        "insulin",
		Description: "Insert the amount of insulin taken.",
		Options: discord.CommandOptions{
			&discord.StringOption{
				OptionName:  "type",
				Description: "Type of insulin",
				Choices: []discord.StringChoice{
					{Name: store.RapidActing, Value: store.RapidActing},
					{Name: store.LongActing, Value: store.LongActing},
				},
				Required: true,
			},
			&discord.IntegerOption{
				OptionName:  "units",
				Description: "Units of insulin.",
				Min:         option.ZeroInt,
				Required:    true,
			},
			&discord.IntegerOption{
				OptionName:  "offset",
				Description: "Offset in minutes.",
				Min:         option.ZeroInt,
				Required:    false,
			},
		},
	},
}

type GlucoseReport struct {
	Description string

	// Current and future value.
	Value     float64
	Trend     store.Trend
	Predicted float64

	// 12h overview.
	Mean           float64
	Std            float64
	TimeInRange    float64
	TimeBelowRange float64
	TimeAboveRange float64

	Chart sendpart.File
}

type WeeklyReport struct {
	Description string

	// Weekly overview.
	TimeInRange    float64
	TimeBelowRange float64
	TimeAboveRange float64
	WeeklyChange   float64 // Change in TimeInRange from last week.

	Chart sendpart.File
}

type CarbohydrateResponse struct {
	Value int
	Time  time.Time
}

type InsulinResponse struct {
	Type  string
	Units int
	Time  time.Time
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

func getAllOptions(opts []discord.CommandInteractionOption) map[string]string {
	optsMap := make(map[string]string)
	for _, opt := range opts {
		optsMap[opt.Name] = opt.Value.String()
	}
	return optsMap
}

func interactionCreate(ses *session.Session, sto *store.Store, logger *zap.Logger) func(e *gateway.InteractionCreateEvent) {
	return func(e *gateway.InteractionCreateEvent) {
		var resp api.InteractionResponse

		// TODO: Really need to move each case to its own function.

		logger.Info("got InteractionCreateEvent",
			zap.Any("event", e),
		)

		switch data := e.Data.(type) {
		case *discord.CommandInteraction:
			switch data.Name {
			case "glucose":
				gr, err := glucoseReport(sto)
				if err != nil {
					logger.Info("failed to get glucose report",
						zap.Error(err),
					)
					resp = interactionWarnResponse(err.Error())
					break
				}

				resp = api.InteractionResponse{
					Type: api.MessageInteractionWithSource,
					Data: &api.InteractionResponseData{
						Embeds: &[]discord.Embed{
							{
								Title:       "Recent Glucose & Predictions",
								Description: gr.Description,
								Image:       &discord.EmbedImage{URL: "attachment://" + gr.Chart.Name},
								Fields: []discord.EmbedField{
									// Line 1.
									{Name: "Current", Value: floatToString(gr.Value), Inline: true},
									{Name: "Trend", Value: "\\" + trendToString(gr.Trend), Inline: true},
									{Name: "Predicted", Value: floatToString(gr.Predicted), Inline: true},
									// Line 2.
									{Name: "Mean", Value: floatToString(gr.Mean), Inline: true},
									{Name: "Std Dev", Value: floatToString(gr.Std), Inline: true},
									inlineBlankField,
									// Line 3.
									{Name: "In Range", Value: floatToString(gr.TimeInRange), Inline: true},
									{Name: "Below Range", Value: floatToString(gr.TimeBelowRange), Inline: true},
									{Name: "Above Range", Value: floatToString(gr.TimeAboveRange), Inline: true},
								},
								Footer: &defaultFooter,
								Color:  discord.Color(WarnLevel1),
							},
						},
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
					logger.Info("failed to get weekly report",
						zap.Error(err),
					)
					resp = interactionWarnResponse(err.Error())
					break
				}

				resp = api.InteractionResponse{
					Type: api.MessageInteractionWithSource,
					Data: &api.InteractionResponseData{
						Embeds: &[]discord.Embed{{
							Title:       "Weekly Overlay",
							Description: wr.Description,
							Image:       &discord.EmbedImage{URL: "attachment://" + wr.Chart.Name},
							Fields: []discord.EmbedField{
								// Line 1.
								{Name: "In Range", Value: floatToString(wr.TimeInRange), Inline: true},
								{Name: "Below Range", Value: floatToString(wr.TimeBelowRange), Inline: true},
								{Name: "Above Range", Value: floatToString(wr.TimeAboveRange), Inline: true},
								// Line 2.
								{Name: "Weekly Change", Value: signedFloatString(wr.WeeklyChange), Inline: true},
							},
							Footer: &defaultFooter,
							Color:  discord.Color(WarnLevel1),
						}},
						Files: []sendpart.File{wr.Chart},
					},
				}
			case "carbohydrates":
				var val, offset int

				optsMap := getAllOptions(data.Options)
				val, err := strconv.Atoi(optsMap["amount"])
				if err != nil {
					resp = interactionWarnResponse(err.Error())
					break
				}

				if off, ok := optsMap["offset"]; ok {
					offset, err = strconv.Atoi(off)
					if err != nil {
						resp = interactionWarnResponse(err.Error())
						break
					}
				}

				cr, err := addCarbohydrate(val, offset, sto)
				if err != nil {
					logger.Info("failed to add carbohydrate intake",
						zap.Error(err),
					)
					resp = interactionWarnResponse(err.Error())
					break
				}

				resp = api.InteractionResponse{
					Type: api.MessageInteractionWithSource,
					Data: &api.InteractionResponseData{
						Embeds: &[]discord.Embed{
							{
								Fields: []discord.EmbedField{
									{Name: "Amount", Value: strconv.Itoa(val) + " grams", Inline: true},
									{Name: "Time", Value: cr.Time.In(loc).Format("Jan 02 15:04:05"), Inline: true},
								},
								Footer: &defaultFooter,
								Color:  discord.Color(WarnLevel1),
							},
						},
					},
				}
			case "insulin":
				var insulin string
				var units, offset int

				optsMap := getAllOptions(data.Options)

				units, err := strconv.Atoi(optsMap["units"])
				if err != nil {
					resp = interactionWarnResponse(err.Error())
					break
				}

				if off, ok := optsMap["offset"]; ok {
					offset, err = strconv.Atoi(off)
					if err != nil {
						resp = interactionWarnResponse(err.Error())
						break
					}
				}

				insulin = optsMap["type"]

				ir, err := addInsulin(insulin, units, offset, sto)
				if err != nil {
					logger.Info("failed to add insulin intake",
						zap.Error(err),
					)
					resp = interactionWarnResponse(err.Error())
					break
				}

				resp = api.InteractionResponse{
					Type: api.MessageInteractionWithSource,
					Data: &api.InteractionResponseData{
						Embeds: &[]discord.Embed{
							{
								Fields: []discord.EmbedField{
									{Name: "Units", Value: strconv.Itoa(ir.Units) + " units", Inline: true},
									{Name: "Time", Value: ir.Time.In(loc).Format("Jan 02 15:04:05"), Inline: true},
								},
								Footer: &defaultFooter,
								Color:  discord.Color(WarnLevel1),
							},
						},
					},
				}
			}
		}

		if err := ses.RespondInteraction(e.ID, e.Token, resp); err != nil {
			logger.Info("failed to send interaction callback",
				zap.Error(err),
			)
		}
	}
}

func glucoseReport(sto *store.Store) (*GlucoseReport, error) {
	start := time.Now().Add(-12 * time.Hour)
	end := time.Now()

	// Get glucose values.
	var pts []store.TimePoint
	err := sto.GetPoints(start, end, store.FieldGlucose, &pts)
	if err != nil {
		return nil, fmt.Errorf("unable to get points: %w", err)
	}

	// Get future glucose predictions.
	var preds []store.TimePoint
	err = sto.GetPoints(end, end.Add(6*time.Hour), store.FieldGlucosePred, &preds)
	if err != nil {
		return nil, fmt.Errorf("unable to get predictions: %w", err)
	}

	// Get carbohydrate intakes.
	var carbs []store.Carbohydrate
	err = sto.GetPoints(start, end, store.FieldCarbohydrate, &carbs)
	if err != nil {
		return nil, fmt.Errorf("unable to get carbohydrates: %w", err)
	}

	// Get insulin doses.
	var insulin []store.Insulin
	err = sto.GetPoints(start, end, store.FieldInsulin, &insulin)
	if err != nil {
		return nil, fmt.Errorf("unable to get insulin doses: %w", err)
	}

	var conf store.Config
	if err = sto.GetObject(store.IndexConfig, &conf); err != nil {
		return nil, fmt.Errorf("unable to load config: %w", err)
	}

	r, err := PlotRecentAndPreds(conf.LowThreshold, conf.HighThreshold, pts, preds, carbs, insulin)
	if err != nil {
		return nil, fmt.Errorf("unable to generate daily graph: %w", err)
	}

	curPt := pts[len(pts)-1]
	predPt := store.TimePoint{Value: -1}
	if len(preds) != 0 {
		predPt = preds[len(preds)-1]
	}

	total := float64(len(pts))
	var within, below, above float64

	x := make([]float64, len(pts))
	for i, pt := range pts {
		x[i] = pt.Value

		if pt.Value < conf.LowThreshold {
			below++
		} else if pt.Value > conf.HighThreshold {
			above++
		} else {
			within++
		}
	}

	mean := stat.Mean(x, nil)
	std := stat.StdDev(x, nil)

	return &GlucoseReport{
		Description: fmt.Sprintf("%s - %s",
			start.In(loc).Format("Jan 02 15:04:05"),
			end.In(loc).Format("Jan 02 15:04:05"),
		),
		Value:          curPt.Value,
		Trend:          curPt.Trend,
		Predicted:      predPt.Value,
		Mean:           mean,
		Std:            std,
		TimeInRange:    within / total,
		TimeBelowRange: below / total,
		TimeAboveRange: above / total,
		Chart:          sendpart.File{Name: "glucoseChart.png", Reader: r},
	}, nil
}

func weeklyReport(offset int, sto *store.Store) (*WeeklyReport, error) {
	t := time.Now().In(loc).AddDate(0, 0, -7*offset)
	ws := weekStart(t)
	we := ws.AddDate(0, 0, 7)

	var pts []store.TimePoint
	err := sto.GetPoints(ws, we, store.FieldGlucose, &pts)
	if err != nil {
		return nil, fmt.Errorf("unable to get points: %w", err)
	}

	// Get last week's points.
	var lwPts []store.TimePoint
	err = sto.GetPoints(ws.AddDate(0, 0, -7), ws, store.FieldGlucose, &lwPts)
	if err != nil {
		return nil, fmt.Errorf("unable to get last week's points: %w", err)
	}

	var conf store.Config
	if err = sto.GetObject(store.IndexConfig, &conf); err != nil {
		return nil, fmt.Errorf("unable to load config: %w", err)
	}

	r, err := PlotOverlayWeekly(conf.LowThreshold, conf.HighThreshold, pts)
	if err != nil {
		return nil, fmt.Errorf("unable to generate weekly plot: %w", err)
	}

	total := float64(len(pts))
	var within, below, above float64

	x := make([]float64, len(pts))
	for i, pt := range pts {
		x[i] = pt.Value

		if pt.Value < conf.LowThreshold {
			below++
		} else if pt.Value > conf.HighThreshold {
			above++
		} else {
			within++
		}
	}

	lwTotal := float64(len(lwPts))
	var lwWithin float64
	for _, pt := range lwPts {
		if pt.Value >= conf.LowThreshold && pt.Value <= conf.HighThreshold {
			lwWithin++
		}
	}

	return &WeeklyReport{
		Description: fmt.Sprintf("%s - %s",
			ws.In(loc).Format("Mon, 02 Jan 2006"),
			ws.AddDate(0, 0, 6).In(loc).Format("Mon, 02 Jan 2006"),
		),
		TimeInRange:    within / total,
		TimeBelowRange: below / total,
		TimeAboveRange: above / total,
		WeeklyChange:   within/total - lwWithin/lwTotal,
		Chart:          sendpart.File{Name: "weeklyOverlay.png", Reader: r},
	}, nil
}

func addCarbohydrate(val, offset int, sto *store.Store) (*CarbohydrateResponse, error) {
	when := time.Now().In(loc).Add(-time.Duration(offset) * time.Minute)
	err := sto.AddPoint(store.FieldCarbohydrate, when, store.Carbohydrate{
		Time:  when,
		Value: val,
	})
	if err != nil {
		return nil, err
	}

	return &CarbohydrateResponse{
		Value: val,
		Time:  when,
	}, nil
}

func addInsulin(insulin string, units, offset int, sto *store.Store) (*InsulinResponse, error) {
	when := time.Now().In(loc).Add(-time.Duration(offset) * time.Minute)
	err := sto.AddPoint(store.FieldInsulin, when, store.Insulin{
		Time:  when,
		Type:  insulin,
		Value: units,
	})
	if err != nil {
		return nil, err
	}

	return &InsulinResponse{
		Type:  insulin,
		Time:  when,
		Units: units,
	}, nil
}
