package main

import (
	"fmt"
	"log"

	"github.com/melkomukovki/go-musthave-metrics/internal/agent"
	"github.com/melkomukovki/go-musthave-metrics/internal/agent/config"
)

var (
	buildVersion = "N/A"
	buildCommit  = "N/A"
	buildDate    = "N/A"
)

func main() {
	printInfo()

	cfg, err := config.GetClientConfig()
	if err != nil {
		log.Fatal(err)
	}

	app, err := agent.NewAgent(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	app.Run()
}

func printInfo() {
	fmt.Println("Build Version:", buildVersion)
	fmt.Println("Build Date:", buildDate)
	fmt.Println("Build Commit:", buildCommit)
}
