package glucose

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/algao1/ichor/pb"
	"github.com/algao1/ichor/store"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Client struct {
	gc pb.GlucoseClient
}

func New(conn *grpc.ClientConn) *Client {
	return &Client{
		gc: pb.NewGlucoseClient(conn),
	}
}

func (c *Client) Predict(ctx context.Context, pts []*store.TimePoint) (*store.TimePoint, error) {
	if len(pts) == 0 {
		return nil, fmt.Errorf("no points given")
	}

	values := make([]float64, len(pts))
	for i, pt := range pts {
		values[i] = pt.Value
	}

	pred, err := c.gc.Predict(ctx, &pb.Features{
		Values: values,
		Time:   timestamppb.New(pts[len(pts)-1].Time),
	})
	if err != nil {
		return nil, err
	}

	return &store.TimePoint{
		Value: pred.Value,
		Time:  pred.GetTime().AsTime(),
		Trend: store.Missing,
	}, nil
}

func RunPredictor(client *Client, s *store.Store) {
	ticker := time.NewTicker(1 * time.Minute)

	for {
		<-ticker.C

		pastPoints, err := s.GetLastPoints("glucose", 24)
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

		s.AddPoint("glucose-pred", &store.TimePoint{
			Time:  ftp.Time.Add(24 * 5 * time.Minute),
			Value: ftp.Value,
			Trend: ftp.Trend,
		})
	}
}
