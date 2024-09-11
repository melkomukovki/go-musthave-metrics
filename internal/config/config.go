package config

import (
	"flag"
	"log"

	"github.com/caarlos0/env/v6"
)

const (
	DefaultAddress        = "localhost:8080"
	DefaultPollInterval   = 2
	DefaultReportInterval = 10
)

type ServerConfig struct {
	Address string `env:"ADDRESS"`
}

type ClientConfig struct {
	Address        string `env:"ADDRESS"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	PollInterval   int    `env:"POLL_INTERVAL"`
}

func GetServerConfig() (ServerConfig, error) {
	var cfg ServerConfig

	if err := env.Parse(&cfg); err != nil {
		return ServerConfig{}, err
	}

	if cfg.Address != "" {
		return cfg, nil
	}

	fullAddres := flag.String("a", DefaultAddress, "Server address and port")
	flag.Parse()

	cfg.Address = *fullAddres
	return cfg, nil
}

func GetClientConfig() (ClientConfig, error) {
	var cfg ClientConfig

	if err := env.Parse(&cfg); err != nil {
		return ClientConfig{}, err
	}

	fullAddress := flag.String("a", "localhost:8080", "Server address and port")
	reportInterval := flag.Int("r", DefaultReportInterval, "Report interval (sec)")
	pollInterval := flag.Int("p", DefaultPollInterval, "Poll metric interval (sec)")
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

	if cfg.PollInterval <= 0 || cfg.PollInterval > 100 {
		log.Printf("Wrong value PollInterval: %d. Must be: 0 < PollInterval <= 100", cfg.PollInterval)
		cfg.PollInterval = DefaultPollInterval
		log.Printf("Set PollInterval to default value: %d", DefaultPollInterval)
	}

	if cfg.ReportInterval <= 0 || cfg.ReportInterval > 100 {
		log.Printf("Wrong value ReportInterval: %d. Must be: 0 < ReportInterval <= 100", cfg.ReportInterval)
		cfg.ReportInterval = DefaultReportInterval
		log.Printf("Set ReportInterval to default value: %d", DefaultReportInterval)
	}

	return cfg, nil
}
