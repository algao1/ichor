package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/algao1/ichor/discord"
	"github.com/algao1/ichor/router"
	"github.com/algao1/ichor/store"
)

var (
	token string
)

func init() {
	flag.StringVar(&token, "t", "", "bot token")
	flag.Parse()
}

func main() {
	s, err := store.Create()
	if err != nil {
		panic(err)
	}

	s.Initialize()

	r := router.Create(s)
	go r.Run(":8080")

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
