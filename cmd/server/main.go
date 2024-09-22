package main

import (
	"log"

	"github.com/melkomukovki/go-musthave-metrics/internal/config"
	"github.com/melkomukovki/go-musthave-metrics/internal/server"
	"github.com/melkomukovki/go-musthave-metrics/internal/storage"
)

var store = storage.NewMemStorage()

func main() {
	cfg, err := config.GetServerConfig()
	if err != nil {
		log.Fatal(err)
	}

	r := server.NewServerRouter(store)

	r.Run(cfg.Address)
}
