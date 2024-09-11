package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/melkomukovki/go-musthave-metrics/internal/agent"
	"github.com/melkomukovki/go-musthave-metrics/internal/config"
)

func updateMetrics(agent *agent.Agent) {
	for {
		agent.PollMetrics()
		<-time.After(time.Second * time.Duration(agent.GetPollInterval()))
	}
}

func sendMetrics(c *resty.Client, agent *agent.Agent) {
	for {
		agent.ReportMetrics(c)
		<-time.After(time.Second * time.Duration(agent.GetReportInterval()))
	}
}

func main() {
	cfg, err := config.GetClientConfig()
	if err != nil {
		log.Fatal(err)
	}

	agent := agent.NewAgent(&cfg)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	client := resty.New()

	go updateMetrics(agent)
	go sendMetrics(client, agent)

	<-sigs
}
