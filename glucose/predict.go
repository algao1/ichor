package glucose

import (
	"context"

	"github.com/algao1/ichor/pb"
	"github.com/algao1/ichor/store"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Client struct {
	gc    pb.GlucoseClient
	store *store.Store
}

func New(conn *grpc.ClientConn, store *store.Store) *Client {
	return &Client{
		gc:    pb.NewGlucoseClient(conn),
		store: store,
	}
}

func (c *Client) Predict(ctx context.Context, pts []*store.TimePoint) (*store.TimePoint, error) {
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
	}, nil
}
