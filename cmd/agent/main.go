package main

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/go-resty/resty/v2"

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

	client := resty.New()
	client.SetRetryCount(3).
		SetRetryWaitTime(time.Second).
		SetRetryMaxWaitTime(5 * time.Second).
		SetRetryAfter(func(client *resty.Client, resp *resty.Response) (time.Duration, error) {
			return 0, errors.New("quota exceeded")
		})
	app.Run(client)
}

func printInfo() {
	fmt.Println("Build Version:", buildVersion)
	fmt.Println("Build Date:", buildDate)
	fmt.Println("Build Commit:", buildCommit)
}
