// Package config define client config structure
package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"
)

// Default client config params
const (
	DefaultAddress        = "localhost:8080"
	DefaultPollInterval   = 2
	DefaultReportInterval = 10
	DefaultHashKey        = ""
	DefaultRateLimit      = 10
)

// ClientConfig structure define
type ClientConfig struct {
	Address        string `env:"ADDRESS"`         // Server address with port
	ReportInterval int    `env:"REPORT_INTERVAL"` // Metrics poll interval in seconds
	PollInterval   int    `env:"POLL_INTERVAL"`   // Metrics report interval in seconds
	HashKey        string `env:"KEY"`             // Secret key using for hashing
	RateLimit      int    `env:"RATE_LIMIT"`      // Maximum concurrent connections to server
}

// GetClientConfig allow to get ClientConfig
func GetClientConfig() (ClientConfig, error) {
	var cfg ClientConfig

	flag.StringVar(&cfg.Address, "a", DefaultAddress, "Server address and port")
	flag.IntVar(&cfg.ReportInterval, "r", DefaultReportInterval, "Report interval (sec)")
	flag.IntVar(&cfg.PollInterval, "p", DefaultPollInterval, "Poll metric interval (sec)")
	flag.StringVar(&cfg.HashKey, "k", DefaultHashKey, "Hash key for calculation HashSHA256 header")
	flag.IntVar(&cfg.RateLimit, "l", DefaultRateLimit, "Limit for outgoing requests")
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

	if envHashKey := os.Getenv("KEY"); envHashKey != "" {
		cfg.HashKey = envHashKey
	}

	if envRateLimit := os.Getenv("RATE_LIMIT"); envRateLimit != "" {
		iRateLimit, err := strconv.Atoi(envRateLimit)
		if err != nil {
			return ClientConfig{}, fmt.Errorf("invalid value for env variavle `RATE_LIMIT`")
		}
		cfg.RateLimit = iRateLimit
	}

	if cfg.PollInterval <= 0 || cfg.PollInterval > 100 {
		return ClientConfig{}, fmt.Errorf("wrong value PollInterval: %d. Must be: 0 < PollInterval <= 100", cfg.PollInterval)
	}

	if cfg.ReportInterval <= 0 || cfg.ReportInterval > 100 {
		return ClientConfig{}, fmt.Errorf("wrong value ReportInterval: %d. Must be: 0 < ReportInterval <= 100", cfg.ReportInterval)
	}

	return cfg, nil
}
