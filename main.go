package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/algao1/ichor/dexcom"
	"github.com/algao1/ichor/discord"
	"github.com/algao1/ichor/store"
)

var (
	token       string
	dexAccount  string
	dexPassword string
)

func init() {
	flag.StringVar(&token, "t", "", "bot token")
	flag.StringVar(&dexAccount, "a", "", "dexcom account")
	flag.StringVar(&dexPassword, "p", "", "dexcom password")

	flag.Parse()
}

func main() {
	s, err := store.Create()
	if err != nil {
		panic(err)
	}
	s.Initialize()

	dc := dexcom.New(dexAccount, dexPassword)
	go dexcom.RunUploader(dc, s)

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
