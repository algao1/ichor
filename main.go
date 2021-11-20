package main

import (
	"github.com/algao1/ichor/router"
	"github.com/algao1/ichor/store"
)

func main() {
	s, err := store.Create()
	if err != nil {
		panic(err)
	}

	s.Initialize()

	r := router.Create(s)
	r.Run(":8080")
}
