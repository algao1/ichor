package main

import (
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/algao1/ichor/store"
	"github.com/gin-gonic/gin"
)

func sRouter(tss *store.TimeSeriesStore) *gin.Engine {
	r := gin.Default()

	r.GET("/download", func(c *gin.Context) {
		params := c.Request.URL.Query()

		startStr := params.Get("start")
		if startStr == "" {
			startStr = "1970-01-01"
		}

		endStr := params.Get("end")
		if endStr == "" {
			endStr = "9999-12-31"
		}

		start, _ := time.Parse("2006-01-02", startStr)
		end, _ := time.Parse("2006-01-02", endStr)

		pts, _ := tss.GetPoints(start, end, "glucose")

		c.JSON(http.StatusOK, gin.H{"points": pts})
	})

	r.POST("/upload", func(c *gin.Context) {
		var hd HealthData

		if c.Bind(&hd) == nil {
			glucoseValues := strings.Split(hd.Glucose.Values, "\n")
			dateStrs := strings.Split(hd.Glucose.Dates, "\n")

			for i := range glucoseValues {
				date, err := time.Parse("2006-01-02T15:04:05-07:00", dateStrs[i])
				if err != nil {
					log.Fatal(err)
				}

				val, err := strconv.ParseFloat(glucoseValues[i], 64)
				if err != nil {
					log.Fatal(err)
				}

				tss.AddPoint("glucose", store.TimePoint{Time: date, Value: val})
			}

			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		}
	})

	return r
}

func main() {
	tss, err := store.Create()
	if err != nil {
		panic(err)
	}

	tss.Initialize()

	r := sRouter(tss)
	r.Run(":8080")
}
