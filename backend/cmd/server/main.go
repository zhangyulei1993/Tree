package main

import (
	"log"

	"tree/backend/internal/config"
	"tree/backend/internal/database"
	"tree/backend/internal/router"
)

func main() {
	cfg := config.Load()

	database.Init(cfg)

	r := router.New(cfg)

	if err := r.Run(cfg.Addr()); err != nil {
		log.Fatal(err)
	}
}
