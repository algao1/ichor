package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/algao1/ichor/discord"
	"github.com/algao1/ichor/glucose/dexcom"
	"github.com/algao1/ichor/glucose/predictor"
	"github.com/algao1/ichor/store"
)

const (
	DefaultMinutes  = 1440
	DefaultMaxCount = 288
)

func RunUploader(client *dexcom.Client, s *store.Store) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for t := time.Now(); true; t = <-ticker.C {
		trs, err := client.GetReadings(DefaultMinutes, DefaultMaxCount)
		if err != nil {
			log.Println("Failed to get readings: " + t.Format(time.RFC3339))
			continue
		}

		for _, tr := range trs {
			s.AddPoint(store.FieldGlucose, tr.Time, &store.TimePoint{
				Time:  tr.Time,
				Value: tr.Mmol,
				Trend: tr.Trend,
			})
		}
	}
}

func RunPredictor(client *predictor.Client, s *store.Store, alertCh chan<- discord.Alert) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for ; true; <-ticker.C {
		var pastPoints []store.TimePoint
		if err := s.GetLastPoints(store.FieldGlucose, 24, &pastPoints); err != nil {
			log.Printf("Failed to get past points: %s\n", err)
			continue
		}

		ftp, err := client.Predict(context.Background(), pastPoints)
		if err != nil {
			log.Printf("Failed to make a prediction: %s\n", err)
			continue
		}

		s.AddPoint(store.FieldGlucosePred, ftp.Time, &store.TimePoint{
			Time:  ftp.Time,
			Value: ftp.Value,
			Trend: ftp.Trend,
		})

		// Everything below here is very not production ready.
		// TODO: Fix everything.

		var conf store.Config
		if err = s.GetObject(store.IndexConfig, &conf); err != nil {
			panic(fmt.Errorf("unable to load config: %w", err))
		}

		var expire time.Time
		s.GetObject(store.IndexTimeoutExpire, &expire)

		if expire.After(time.Now()) {
			continue
		}

		if ftp.Value <= conf.LowThreshold {
			alertCh <- discord.Low
			s.AddObject(store.IndexTimeoutExpire, time.Now().Add(conf.WarningTimeout))
		} else if ftp.Value >= conf.HighThreshold {
			alertCh <- discord.High
			s.AddObject(store.IndexTimeoutExpire, time.Now().Add(conf.WarningTimeout))
		}
	}
}
