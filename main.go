package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/algao1/ichor/discord"
	"github.com/algao1/ichor/glucose/dexcom"
	"github.com/algao1/ichor/glucose/predictor"
	"github.com/algao1/ichor/store"
	"google.golang.org/grpc"
)

var (
	uid         string
	token       string
	dexAccount  string
	dexPassword string
	serverAddr  string
)

func init() {
	flag.StringVar(&token, "t", "", "discord bot token")
	flag.StringVar(&uid, "u", "", "discord user id")
	flag.StringVar(&dexAccount, "a", "", "dexcom account")
	flag.StringVar(&dexPassword, "p", "", "dexcom password")
	flag.StringVar(&serverAddr, "s", "localhost:50051", "inference server address")

	flag.Parse()
}

func main() {
	s, err := store.Create()
	if err != nil {
		log.Fatal(err)
	}
	s.Initialize()

	// Will temporarily overwrite all previous configs.
	s.AddObject(store.IndexConfig, store.Config{
		WarningTimeout: 1 * time.Hour,
		LowThreshold:   3.7,
		HighThreshold:  10.0,
	})

	dc := dexcom.New(dexAccount, dexPassword)
	go RunUploader(dc, s)

	alertCh := make(chan discord.Alert)

	db, err := discord.Create(uid, token, s, alertCh)
	if err != nil {
		log.Fatal(err)
	}

	conn, err := grpc.Dial(serverAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}

	p := predictor.New(conn)
	go RunPredictor(p, s, alertCh)

	db.Run()
	defer db.Stop()

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}
