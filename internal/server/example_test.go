package server

import (
	"github.com/gin-gonic/gin"
	"github.com/melkomukovki/go-musthave-metrics/internal/config"
	"github.com/melkomukovki/go-musthave-metrics/internal/controllers"
	"github.com/melkomukovki/go-musthave-metrics/internal/infra/memstorage"
	"github.com/melkomukovki/go-musthave-metrics/internal/services"
	"log"
)

func Example_run() {
	// Get server config
	cfg, _ := config.GetServerConfig()

	// Create ServiceRepository
	serviceRepository := memstorage.NewClient(cfg.StoreInterval, cfg.FileStoragePath, cfg.Restore)

	// Wire repository and logic
	appService := &services.Service{
		ServiceRepo: serviceRepository,
	}

	// Create gin engine with routes
	router := gin.Default()
	controllers.NewHandler(router, appService, cfg.HashKey)

	// Run server
	if err := router.Run(cfg.Address); err != nil {
		log.Fatal("error while running server")
	}
}
