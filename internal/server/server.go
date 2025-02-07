// Package server реализует логику связывания необходимых компонентов для работы сервера
package server

import (
	"context"
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
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
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
	controllers.NewHandler(router, appService, cfg.HashKey, certKey, cfg.TrustedSubnet)

	srv := &http.Server{
		Addr:    cfg.Address,
		Handler: router,
	}

	go func() {
		if err = srv.ListenAndServe(); err != nil {
			log.Fatal().Err(err).Msg("error while running server")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	<-quit
	log.Info().Msg("shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("error while shutting down server")
	}

	log.Info().Msg("server gracefully stopped")
}
