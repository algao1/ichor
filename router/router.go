package router

import (
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/algao1/ichor/store"
	"github.com/gin-gonic/gin"
)

type Router struct {
	*gin.Engine
}

func Create(tss *store.TimeSeriesStore) *Router {
	r := gin.Default()

	r.POST("/upload/glucose", uploadGlucose(tss))

	return &Router{r}
}

func uploadGlucose(tss *store.TimeSeriesStore) func(*gin.Context) {
	return func(c *gin.Context) {
		var hd HealthData

		if c.Bind(&hd) == nil {
			// Apple Health weirdness with getting health samples.
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
	}
}
