package server

import (
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"github.com/melkomukovki/go-musthave-metrics/internal/config"
	"github.com/melkomukovki/go-musthave-metrics/internal/controllers"
	"github.com/melkomukovki/go-musthave-metrics/internal/infra/memstorage"
	"github.com/melkomukovki/go-musthave-metrics/internal/infra/postgres"
	"github.com/melkomukovki/go-musthave-metrics/internal/services"
)

func Run() {
	cfg, err := config.GetServerConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("config error. can't initialize config")
	}

	var serviceRepository services.ServiceRepository
	if cfg.DataSourceName != "" {
		store, err := postgres.NewClient(cfg.DataSourceName)
		if err != nil {
			log.Fatal().Err(err).Msg("can't initialize postgresql storage")
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
	controllers.NewHandler(router, appService, cfg.HashKey)

	if err = router.Run(cfg.Address); err != nil {
		log.Fatal().Err(err).Msg("error while running server")
	}
}
