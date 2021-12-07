package main

import (
	"context"
	"encoding/json"
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

	for {
		t := <-ticker.C
		trs, err := client.GetReadings(DefaultMinutes, DefaultMaxCount)
		if err != nil {
			log.Println("Failed to get readings: " + t.Format(time.RFC3339))
			continue
		}

		for _, tr := range trs {
			s.AddPoint(store.FieldGlucose, &store.TimePoint{
				Time:  tr.Time,
				Value: tr.Mmol,
			})
		}
	}
}

func RunPredictor(client *predictor.Client, s *store.Store, alertCh chan<- discord.Alert) {
	ticker := time.NewTicker(1 * time.Minute)

	for {
		<-ticker.C

		pastPoints, err := s.GetLastPoints(store.FieldGlucose, 24)
		if err != nil {
			log.Printf("Failed to get past points: %s\n", err)
			continue
		}

		ftp, err := client.Predict(context.Background(), pastPoints)
		if err != nil {
			log.Printf("Failed to make a prediction: %s\n", err)
			continue
		}

		fmt.Println(ftp.Time, ftp.Value)

		s.AddPoint(store.FieldGlucosePred, &store.TimePoint{
			Time:  ftp.Time.Add(24 * 5 * time.Minute),
			Value: ftp.Value,
			Trend: ftp.Trend,
		})

		// Everything below here is very not production ready.
		// TODO: Fix everything.

		confObj, err := s.GetObject(store.IndexConfig)
		if err != nil {
			log.Printf("Could not access config settings")
			continue
		}

		var conf store.Config
		err = json.Unmarshal(confObj, &conf)
		if err != nil {
			panic("Invalid configuration")
		}

		var expire time.Time
		expireObj, err := s.GetObject(store.IndexTimeoutExpore)
		if err == nil {
			json.Unmarshal(expireObj, &expire)
		}

		if expire.After(time.Now()) {
			continue
		}

		if ftp.Value <= conf.LowThreshold {
			alertCh <- discord.Low
			s.AddObject(store.IndexTimeoutExpore, time.Now().Add(conf.WarningTimeout))
		} else if ftp.Value >= conf.HighThreshold {
			alertCh <- discord.High
			s.AddObject(store.IndexTimeoutExpore, time.Now().Add(conf.WarningTimeout))
		}
	}
}
