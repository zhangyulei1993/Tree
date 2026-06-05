package main

import (
	"log"

	"tree/backend/internal/app"
	"tree/backend/internal/common/config"
	"tree/backend/internal/common/logger"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	logg, err := logger.New(cfg.Log.Level, cfg.App.Env)
	if err != nil {
		log.Fatalf("init logger: %v", err)
	}
	defer func() {
		_ = logg.Sync()
	}()

	server := app.NewServer(cfg, logg)
	if err := server.Run(); err != nil {
		logg.Fatal("server stopped", logger.Error(err))
	}
}
