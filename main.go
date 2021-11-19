package main

import (
	"github.com/algao1/ichor/router"
	"github.com/algao1/ichor/store"
)

func main() {
	tss, err := store.Create()
	if err != nil {
		panic(err)
	}

	tss.Initialize()

	r := router.Create(tss)
	r.Run(":8080")
}
