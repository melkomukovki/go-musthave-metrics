package config

import (
	"flag"

	"github.com/caarlos0/env/v6"
)

var DefaultAddress = "localhost:8080"

type ServerConfig struct {
	Address string `env:"ADDRESS"`
}

type ClientConfig struct {
	Address        string `env:"ADDRESS"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	PollInterval   int    `env:"POLL_INTERVAL"`
}

func GetServerConfig() ServerConfig {
	var cfg ServerConfig

	if err := env.Parse(&cfg); err != nil {
		panic(err)
	}

	if cfg.Address != "" {
		return cfg
	}

	fullAddres := flag.String("a", DefaultAddress, "Server address and port")
	flag.Parse()

	cfg.Address = *fullAddres
	return cfg
}

func GetClientConfig() ClientConfig {
	var cfg ClientConfig

	if err := env.Parse(&cfg); err != nil {
		panic(err)
	}

	fullAddress := flag.String("a", "localhost:8080", "Server address and port")
	reportInterval := flag.Int("r", 10, "Report interval (sec)")
	pollInterval := flag.Int("p", 2, "Poll metric interval (sec)")
	flag.Parse()

	if cfg.Address == "" {
		cfg.Address = *fullAddress
	}
	if cfg.ReportInterval == 0 {
		cfg.ReportInterval = *reportInterval
	}
	if cfg.PollInterval == 0 {
		cfg.PollInterval = *pollInterval
	}

	return cfg
}
