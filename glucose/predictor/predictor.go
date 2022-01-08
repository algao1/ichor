package predictor

import (
	"context"
	"fmt"
	"time"

	"github.com/algao1/ichor/pb"
	"github.com/algao1/ichor/store"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Client struct {
	gc     pb.GlucoseClient
	logger *zap.Logger
}

func New(conn *grpc.ClientConn, logger *zap.Logger) *Client {
	return &Client{
		gc:     pb.NewGlucoseClient(conn),
		logger: logger,
	}
}

// TODO: Needs a rewrite, but let's get it working first...

func (c *Client) Predict(ctx context.Context, pts []store.TimePoint, insulin []store.Insulin,
	carbs []store.Carbohydrate) ([]store.TimePoint, error) {
	if len(pts) == 0 {
		return nil, fmt.Errorf("no points given")
	}

	feats := make([]*pb.Feature, len(pts))
	for i, pt := range pts {
		feats[i] = &pb.Feature{
			Time:    timestamppb.New(pt.Time),
			Glucose: pt.Value,
		}
	}

	var pcounter int
	for _, insul := range insulin {
		if insul.Type == store.RapidActing {
			continue
		}

		for pcounter < len(pts)-1 && insul.Time.After(pts[pcounter].Time) {
			pcounter++
		}
		feats[pcounter].Insulin = float64(insul.Value)
	}

	pcounter = 0
	for _, carb := range carbs {
		for pcounter < len(pts)-1 && carb.Time.After(pts[pcounter].Time) {
			pcounter++
		}
		feats[pcounter].Carbs = float64(carb.Value)
	}

	res, err := c.gc.Predict(ctx, &pb.Features{
		Features: feats,
	})
	if err != nil {
		return nil, err
	}

	c.logger.Info("successfully obtained predictions",
		zap.Int("length", len(res.Labels)),
	)

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
