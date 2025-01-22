// Package config define client config structure
package config

import (
	"encoding/json"
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
	DefaultCryptoKey      = ""
	DefaultConfigPath     = ""
)

// ClientConfig structure define
type ClientConfig struct {
	Address        string `json:"address" env:"ADDRESS"`                 // Server address with port
	ReportInterval int    `json:"report_interval" env:"REPORT_INTERVAL"` // Metrics poll interval in seconds
	PollInterval   int    `json:"poll_interval" env:"POLL_INTERVAL"`     // Metrics report interval in seconds
	HashKey        string `json:"key" env:"KEY"`                         // Secret key using for hashing
	RateLimit      int    `json:"rate_limit" env:"RATE_LIMIT"`           // Maximum concurrent connections to server
	CryptoKey      string `json:"crypto_key" env:"CRYPTO_KEY"`           // Path to file with public crypto key
	ConfigPath     string `env:"CONFIG"`                                 // Path to JSON file with configuration
}

// GetClientConfig allow to get ClientConfig
func GetClientConfig() (ClientConfig, error) {
	var cfg ClientConfig

	flag.StringVar(&cfg.Address, "a", DefaultAddress, "Server address and port")
	flag.IntVar(&cfg.ReportInterval, "r", DefaultReportInterval, "Report interval (sec)")
	flag.IntVar(&cfg.PollInterval, "p", DefaultPollInterval, "Poll metric interval (sec)")
	flag.StringVar(&cfg.HashKey, "k", DefaultHashKey, "Hash key for calculation HashSHA256 header")
	flag.IntVar(&cfg.RateLimit, "l", DefaultRateLimit, "Limit for outgoing requests")
	flag.StringVar(&cfg.CryptoKey, "crypto-key", DefaultCryptoKey, "Path to public crypto key")
	flag.StringVar(&cfg.ConfigPath, "c", DefaultConfigPath, "Path to config file")
	flag.Parse()

	envConfigPath := os.Getenv("CONFIG")
	if envConfigPath != "" {
		cfg.ConfigPath = envConfigPath
	}

	if cfg.ConfigPath != "" {
		fileCfg, err := loadConfigFromFile(cfg.ConfigPath)
		if err != nil {
			return ClientConfig{}, err
		}
		mergeConfig(&cfg, fileCfg)
	}

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

	if envCryptoKey := os.Getenv("CRYPTO_KEY"); envCryptoKey != "" {
		cfg.CryptoKey = envCryptoKey
	}

	if cfg.PollInterval <= 0 || cfg.PollInterval > 100 {
		return ClientConfig{}, fmt.Errorf("wrong value PollInterval: %d. Must be: 0 < PollInterval <= 100", cfg.PollInterval)
	}

	if cfg.ReportInterval <= 0 || cfg.ReportInterval > 100 {
		return ClientConfig{}, fmt.Errorf("wrong value ReportInterval: %d. Must be: 0 < ReportInterval <= 100", cfg.ReportInterval)
	}

	// Validate crypto key path and return err if not exists
	if cfg.CryptoKey != "" {
		_, err := os.Stat(cfg.CryptoKey)
		if err != nil {
			return ClientConfig{}, err
		}
	}

	return cfg, nil
}

func loadConfigFromFile(path string) (ClientConfig, error) {
	file, err := os.Open(path)
	if err != nil {
		return ClientConfig{}, err
	}
	defer func() {
		err := file.Close()
		if err != nil {
			fmt.Printf("Error closing file: %v", err)
		}
	}()

	var cfg ClientConfig
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&cfg); err != nil {
		return ClientConfig{}, err
	}

	return cfg, nil
}

func mergeConfig(cfg *ClientConfig, fileCfg ClientConfig) {
	if cfg.Address == DefaultAddress && fileCfg.Address != "" {
		cfg.Address = fileCfg.Address
	}
	if cfg.ReportInterval == DefaultReportInterval && fileCfg.ReportInterval != 0 {
		cfg.ReportInterval = fileCfg.ReportInterval
	}
	if cfg.PollInterval == DefaultPollInterval && fileCfg.PollInterval != 0 {
		cfg.PollInterval = fileCfg.PollInterval
	}
	if cfg.CryptoKey == DefaultCryptoKey && fileCfg.CryptoKey != "" {
		cfg.CryptoKey = fileCfg.CryptoKey
	}
	if cfg.RateLimit == DefaultRateLimit && fileCfg.RateLimit != 0 {
		cfg.RateLimit = fileCfg.RateLimit
	}
	if cfg.HashKey == DefaultHashKey && fileCfg.HashKey != "" {
		cfg.HashKey = fileCfg.HashKey
	}
}
