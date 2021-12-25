package predictor

import (
	"context"
	"fmt"
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

func (c *Client) Predict(ctx context.Context, pts []store.TimePoint) ([]store.TimePoint, error) {
	if len(pts) == 0 {
		return nil, fmt.Errorf("no points given")
	}

	values := make([]float64, len(pts))
	for i, pt := range pts {
		values[i] = pt.Value
	}

	res, err := c.gc.Predict(ctx, &pb.Features{
		Values: values,
		Time:   timestamppb.New(pts[len(pts)-1].Time),
	})
	if err != nil {
		return nil, err
	}

	rpts := make([]store.TimePoint, len(res.Labels))
	for i, label := range res.Labels {
		rpts[i] = store.TimePoint{
			Value: label.GetValue(),
			Time:  label.GetTime().AsTime().Round(5 * time.Minute),
			Trend: store.Missing,
		}
	}

	return rpts, nil
}
