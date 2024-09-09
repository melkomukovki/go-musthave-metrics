package main

import (
	"github.com/gin-gonic/gin"
	"github.com/melkomukovki/go-musthave-metrics/internal/config"
	"github.com/melkomukovki/go-musthave-metrics/internal/server"
	"github.com/melkomukovki/go-musthave-metrics/internal/storage"
)

var store storage.Storage = storage.MemStorage{
	GaugeMetrics:   make(map[string]float64),
	CounterMetrics: make(map[string]int64),
}

func main() {
	cfg := config.GetServerConfig()

	gin.ForceConsoleColor()
	r := server.NewServerRouter(store)

	r.Run(cfg.Address)
}
