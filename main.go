package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/algao1/ichor/dexcom"
	"github.com/algao1/ichor/discord"
	"github.com/algao1/ichor/glucose"
	"github.com/algao1/ichor/store"
	"google.golang.org/grpc"
)

var (
	token       string
	dexAccount  string
	dexPassword string
	serverAddr  string
)

func init() {
	flag.StringVar(&token, "t", "", "bot token")
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

	dc := dexcom.New(dexAccount, dexPassword)
	go dexcom.RunUploader(dc, s)

	conn, err := grpc.Dial(serverAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}

	predictor := glucose.New(conn)
	go glucose.RunPredictor(predictor, s)

	db, err := discord.Create(token, s)
	if err != nil {
		log.Fatal(err)
	}
	db.Run()

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	db.Stop()
}
