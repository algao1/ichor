package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/algao1/ichor/discord"
	"github.com/algao1/ichor/glucose/dexcom"
	"github.com/algao1/ichor/glucose/predictor"
	"github.com/algao1/ichor/store"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var (
	uid         string
	token       string
	dexAccount  string
	dexPassword string
	serverAddr  string

	export bool
)

func init() {
	flag.BoolVar(&export, "e", false, "export db as csv")

	flag.StringVar(&token, "t", "", "discord bot token")
	flag.StringVar(&uid, "u", "", "discord user id")
	flag.StringVar(&dexAccount, "a", "", "dexcom account")
	flag.StringVar(&dexPassword, "p", "", "dexcom password")
	flag.StringVar(&serverAddr, "s", "localhost:50051", "inference server address")

	flag.Parse()
}

func main() {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	logger = logger.Named("ichor")

	s, err := store.Create(logger.Named("store"))
	if err != nil {
		logger.Fatal("failed to create store",
			zap.Error(err),
		)
	}
	s.Initialize()

	if export {
		if err := s.Export("data"); err != nil {
			logger.Fatal("failed to export store as csv",
				zap.Error(err),
			)
		}
		return
	}

	// Will temporarily overwrite all previous configs.
	storeConfig := store.Config{
		WarningTimeout: 1 * time.Hour,
		LowThreshold:   3.7,
		HighThreshold:  10.0,
	}

	if err := s.AddObject(store.IndexConfig, storeConfig); err != nil {
		logger.Panic("failed to save default store configuration",
			zap.Error(err),
		)
	}
	logger.Info("saved default store configuration",
		zap.Any("store config", storeConfig),
	)

	dc := dexcom.New(dexAccount, dexPassword, logger.Named("dexcom client"))
	go RunUploader(dc, s, logger)

	alertCh := make(chan discord.Alert)

	puid, err := strconv.ParseFloat(uid, 64)
	if err != nil {
		logger.Fatal("failed to parse Discord uid",
			zap.String("uid", uid),
			zap.Error(err),
		)
	}

	db, err := discord.Create(token, puid, s, alertCh)
	if err != nil {
		logger.Fatal("failed to create Discord bot",
			zap.Error(err),
		)
	}

	conn, err := grpc.Dial(serverAddr, grpc.WithInsecure())
	if err != nil {
		logger.Fatal("failed to create insecure client connection",
			zap.String("address", serverAddr),
			zap.Error(err),
		)
	}

	p := predictor.New(conn)
	go RunPredictor(p, s, alertCh)

	db.Run(context.Background())
	defer db.Stop()

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	logger.Info("shutting down ichor")
}
