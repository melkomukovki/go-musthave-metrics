package main

import (
	"log"

	"github.com/go-resty/resty/v2"
	"github.com/melkomukovki/go-musthave-metrics/internal/agent"
	"github.com/melkomukovki/go-musthave-metrics/internal/config"
)

func main() {
	cfg, err := config.GetClientConfig()
	if err != nil {
		log.Fatal(err)
	}

	agent := agent.NewAgent(&cfg)
	agent.Run(resty.New())
}
