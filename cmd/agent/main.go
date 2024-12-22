package main

import (
	"errors"
	"log"
	"time"

	"github.com/go-resty/resty/v2"

	"github.com/melkomukovki/go-musthave-metrics/internal/agent"
	"github.com/melkomukovki/go-musthave-metrics/internal/agent/config"
)

func main() {
	cfg, err := config.GetClientConfig()
	if err != nil {
		log.Fatal(err)
	}

	app := agent.NewAgent(&cfg)

	client := resty.New()
	client.SetRetryCount(3).
		SetRetryWaitTime(time.Duration(time.Second)).
		SetRetryMaxWaitTime(time.Duration(5 * time.Second)).
		SetRetryAfter(func(client *resty.Client, resp *resty.Response) (time.Duration, error) {
			return 0, errors.New("quota exceeded")
		})
	app.Run(client)
}
