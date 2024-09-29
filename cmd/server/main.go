package main

import (
	"log"
	"time"

	"github.com/melkomukovki/go-musthave-metrics/internal/server"
	"github.com/melkomukovki/go-musthave-metrics/internal/server/config"
	"github.com/melkomukovki/go-musthave-metrics/internal/storage"
)

func main() {
	cfg, err := config.GetServerConfig()
	if err != nil {
		log.Fatal(err)
	}

	store := storage.NewMemStorage(cfg.StoreInterval, cfg.FileStoragePath)

	if cfg.Restore {
		err := store.RestoreStorage()
		if err != nil {
			log.Fatal(err)
		}
	}

	if !store.SyncStore {
		go func() {
			for {
				time.Sleep(time.Duration(cfg.StoreInterval) * time.Second)
				store.BackupMetrics()
			}
		}()
	}

	engine := server.NewServerRouter(store)

	if err := engine.Run(cfg.Address); err != nil {
		log.Fatal(err)
	}
}
