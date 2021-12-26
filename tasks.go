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
	"go.uber.org/zap"
)

const (
	DefaultMinutes  = 1440
	DefaultMaxCount = 288
)

func RunUploader(client *dexcom.Client, s *store.Store, logger *zap.Logger) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for time.Now(); true; <-ticker.C {
		trs, err := client.GetReadings(DefaultMinutes, DefaultMaxCount)
		if err != nil {
			logger.Info("failed to fetch readings",
				zap.Error(err),
			)
			continue
		}

		for _, tr := range trs {
			err := s.AddPoint(store.FieldGlucose, tr.Time, &store.TimePoint{
				Time:  tr.Time,
				Value: tr.Mmol,
				Trend: tr.Trend,
			})
			if err != nil {
				logger.Info("failed to save glucose reading",
					zap.Any("reading", tr),
					zap.Error(err),
				)
			}
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

		fpts, err := client.Predict(context.Background(), pastPoints)
		if err != nil {
			log.Printf("Failed to make a prediction: %s\n", err)
			continue
		}

		for _, fpt := range fpts {
			s.AddPoint(store.FieldGlucosePred, fpt.Time, &store.TimePoint{
				Time:  fpt.Time,
				Value: fpt.Value,
				Trend: fpt.Trend,
			})
		}

		// Everything below here is very not production ready.
		// TODO: Fix everything.

		var conf store.Config
		if err = s.GetObject(store.IndexConfig, &conf); err != nil {
			panic(fmt.Errorf("failed to load config: %w", err))
		}

		var expire time.Time
		s.GetObject(store.IndexTimeoutExpire, &expire)

		if expire.After(time.Now()) {
			continue
		}

		// if ftp.Value <= conf.LowThreshold {
		// 	alertCh <- discord.Low
		// 	s.AddObject(store.IndexTimeoutExpire, time.Now().Add(conf.WarningTimeout))
		// } else if ftp.Value >= conf.HighThreshold {
		// 	alertCh <- discord.High
		// 	s.AddObject(store.IndexTimeoutExpire, time.Now().Add(conf.WarningTimeout))
		// }
	}
}
