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

func Create(s *store.Store) *Router {
	r := gin.Default()

	r.POST("/upload/glucose", uploadGlucose(s))

	return &Router{r}
}

func uploadGlucose(s *store.Store) func(*gin.Context) {
	return func(c *gin.Context) {
		var hd HealthData

		if c.Bind(&hd) == nil {
			// Apple Health weirdness with getting health samples.
			glucoseValues := strings.Split(hd.Data.Values, "\n")
			dateStrs := strings.Split(hd.Data.Dates, "\n")

			for i := range glucoseValues {
				date, err := time.Parse("2006-01-02T15:04:05-07:00", dateStrs[i])
				if err != nil {
					log.Fatal(err)
				}

				val, err := strconv.ParseFloat(glucoseValues[i], 64)
				if err != nil {
					log.Fatal(err)
				}

				s.AddPoint(store.FieldGlucose, &store.TimePoint{Time: date, Value: val})
			}

			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		}
	}
}
