package main

import (
	"log"

	"github.com/melkomukovki/go-musthave-metrics/internal/config"
	"github.com/melkomukovki/go-musthave-metrics/internal/server"
	"github.com/melkomukovki/go-musthave-metrics/internal/storage"
)

func main() {
	cfg, err := config.GetServerConfig()
	if err != nil {
		log.Fatal(err)
	}

	store := storage.NewMemStorage(cfg.StoreInterval, cfg.FileStoragePath)

	if cfg.Restore {
		store.RestoreStorage()
	}

	if !store.SyncStore {
		go store.BackupMetrics()
	}

	engine := server.NewServerRouter(store)

	if err := engine.Run(cfg.Address); err != nil {
		log.Fatal(err)
	}
}
