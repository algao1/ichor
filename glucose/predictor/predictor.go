package predictor

import (
	"context"
	"fmt"

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

	n := len(pts)
	values := make([]float64, n)
	for i, pt := range pts {
		values[n-i-1] = pt.Value
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
