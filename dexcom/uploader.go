package dexcom

import (
	"fmt"
	"log"
	"time"

	"github.com/algao1/ichor/store"
)

const (
	DefaultMinutes  = 1440
	DefaultMaxCount = 288
)

func RunUploader(client *Client, s *store.Store) {
	ticker := time.NewTicker(1 * time.Minute)

	for {
		t := <-ticker.C
		trs, err := client.GetReadings(DefaultMinutes, DefaultMaxCount)
		if err != nil {
			log.Println("Failed to get readings: " + t.Format(time.RFC3339))
			continue
		}

		for _, tr := range trs {
			fmt.Println(tr)
			s.AddPoint("glucose", &store.TimePoint{
				Time:  tr.Time,
				Value: tr.Mmol,
			})
		}
	}
}
