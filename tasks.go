package main

import (
	"context"
	"fmt"
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
	DefaultLookBack = -4 * time.Hour
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
				logger.Debug("failed to save glucose reading",
					zap.Any("reading", tr),
					zap.Error(err),
				)
			}
		}
	}
}

func RunPredictor(client *predictor.Client, s *store.Store, logger *zap.Logger, alertCh chan<- discord.Alert) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for ; true; <-ticker.C {
		// Panic early, if no configuration could be found.
		var conf store.Config
		err := s.GetObject(store.IndexConfig, &conf)
		if err != nil {
			panic(fmt.Errorf("failed to load config: %w", err))
		}

		var expire time.Time
		s.GetObject(store.IndexTimeoutExpire, &expire)

		var pastPoints []store.TimePoint
		err = s.GetLastPoints(store.FieldGlucose, 4*12, &pastPoints)
		if err != nil {
			logger.Info("failed to get past points",
				zap.Error(err),
			)
			continue
		}

		var pastInsulin []store.Insulin
		err = s.GetPoints(time.Now().Add(DefaultLookBack), time.Now(), store.FieldInsulin, &pastInsulin)
		if err != nil {
			logger.Info("failed to get past insulin values",
				zap.Error(err),
			)
			continue
		}

		var pastCarbs []store.Carbohydrate
		err = s.GetPoints(time.Now().Add(DefaultLookBack), time.Now(), store.FieldCarbohydrate, &pastCarbs)
		if err != nil {
			logger.Info("failed to get past carbohydrate values",
				zap.Error(err),
			)
			continue
		}

		fpts, err := client.Predict(context.Background(), pastPoints, pastInsulin, pastCarbs)
		if err != nil {
			logger.Info("failed to make a prediction",
				zap.Error(err),
			)
			continue
		}

		for _, fpt := range fpts {
			s.AddPoint(store.FieldGlucosePred, fpt.Time, &store.TimePoint{
				Time:  fpt.Time,
				Value: fpt.Value,
				Trend: fpt.Trend,
			})
		}

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
