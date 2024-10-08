package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"
)

const (
	DefaultAddress        = "localhost:8080"
	DefaultPollInterval   = 2
	DefaultReportInterval = 10
)

type ClientConfig struct {
	Address        string `env:"ADDRESS"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	PollInterval   int    `env:"POLL_INTERVAL"`
}

func GetClientConfig() (ClientConfig, error) {
	var cfg ClientConfig

	flag.StringVar(&cfg.Address, "a", DefaultAddress, "Server address and port")
	flag.IntVar(&cfg.ReportInterval, "r", DefaultReportInterval, "Report interval (sec)")
	flag.IntVar(&cfg.PollInterval, "p", DefaultPollInterval, "Poll metric interval (sec)")
	flag.Parse()

	if envAddress := os.Getenv("ADDRESS"); envAddress != "" {
		cfg.Address = envAddress
	}
	if envReportInterval := os.Getenv("REPORT_INTERVAL"); envReportInterval != "" {
		iReportInterval, err := strconv.Atoi(envReportInterval)
		if err != nil {
			return ClientConfig{}, fmt.Errorf("invalid value for env variable `REPORT_INTERVAL`")
		}
		cfg.ReportInterval = iReportInterval
	}
	if envPollInterval := os.Getenv("POLL_INTERVAL"); envPollInterval != "" {
		iPollInterval, err := strconv.Atoi(envPollInterval)
		if err != nil {
			return ClientConfig{}, fmt.Errorf("invalid value for env variavle `POLL_INTERVAL`")
		}
		cfg.PollInterval = iPollInterval
	}

	if cfg.PollInterval <= 0 || cfg.PollInterval > 100 {
		return ClientConfig{}, fmt.Errorf("wrong value PollInterval: %d. Must be: 0 < PollInterval <= 100", cfg.PollInterval)
	}

	if cfg.ReportInterval <= 0 || cfg.ReportInterval > 100 {
		return ClientConfig{}, fmt.Errorf("wrong value ReportInterval: %d. Must be: 0 < ReportInterval <= 100", cfg.ReportInterval)
	}

	return cfg, nil
}
