package main

import (
	"errors"
	"os"
	"time"

	"github.com/melkomukovki/go-musthave-metrics/internal/server"
	"github.com/melkomukovki/go-musthave-metrics/internal/server/config"
	"github.com/melkomukovki/go-musthave-metrics/internal/storage"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const timeFormat = "02/Jan/2006 15:04:05 -0700"

func init() {
	zerolog.TimeFieldFormat = timeFormat
	log.Logger = zerolog.New(os.Stdout).With().Timestamp().Logger()
}

func main() {

	cfg, err := config.GetServerConfig()
	if err != nil {
		log.Fatal().Err(err)
	}
	log.Info().Msg("Server config succesfully loaded")

	var store storage.Storage
	if cfg.DataSourceName != "" {
		store, err = storage.NewPgStorage(cfg.DataSourceName)
		if err != nil {
			log.Fatal().Err(err)
		}
		log.Info().Msg("Initialized postgresql storage")
	} else {
		store = storage.NewMemStorage(cfg.StoreInterval, cfg.FileStoragePath)
		log.Info().Msg("Initialized memstorage")
	}

	if cfg.Restore {
		if _, err := os.Stat(cfg.FileStoragePath); errors.Is(err, os.ErrNotExist) {
			os.Create(cfg.FileStoragePath)
		}
		err := store.RestoreStorage()
		if err != nil {
			log.Fatal().Err(err)
		}
	}

	if !store.SyncStorage() {
		go func() {
			for {
				time.Sleep(time.Duration(cfg.StoreInterval) * time.Second)
				store.BackupMetrics()
			}
		}()
	}

	engine := server.NewServerRouter(store)

	log.Info().Msg("Starting server...")
	if err := engine.Run(cfg.Address); err != nil {
		log.Fatal().Err(err)
	}
}
