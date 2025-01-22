// Package server реализует логику связывания необходимых компонентов для работы сервера
package server

import (
	"crypto/rsa"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/melkomukovki/go-musthave-metrics/internal/config"
	"github.com/melkomukovki/go-musthave-metrics/internal/controllers"
	"github.com/melkomukovki/go-musthave-metrics/internal/infra/memstorage"
	"github.com/melkomukovki/go-musthave-metrics/internal/infra/postgres"
	"github.com/melkomukovki/go-musthave-metrics/internal/services"
	"github.com/melkomukovki/go-musthave-metrics/internal/utils"
	"github.com/rs/zerolog/log"
)

// Run - подготовка необходимых компонентов и запуск сервера
func Run() {
	cfg, err := config.GetServerConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("config error. can't initialize config")
	}

	// get certificate private key if certificate provided
	var certKey *rsa.PrivateKey
	if cfg.CryptoKey != "" {
		certKey, err = utils.GetPrivateKey(cfg.CryptoKey)
		if err != nil {
			log.Fatal().Err(err).Msg("can't initialize crypto key")
		}
	}

	var serviceRepository services.ServiceRepository
	if cfg.DataSourceName != "" {
		store, e := postgres.NewClient(cfg.DataSourceName)
		if e != nil {
			log.Fatal().Err(e).Msg("can't initialize postgresql storage")
		}
		serviceRepository = &postgres.PgRepository{DB: store}
	} else {
		serviceRepository = memstorage.NewClient(cfg.StoreInterval, cfg.FileStoragePath, cfg.Restore)
	}

	appService := &services.Service{
		ServiceRepo: serviceRepository,
	}

	router := gin.Default()
	pprof.Register(router)
	controllers.NewHandler(router, appService, cfg.HashKey, certKey)

	if err = router.Run(cfg.Address); err != nil {
		log.Fatal().Err(err).Msg("error while running server")
	}
}
