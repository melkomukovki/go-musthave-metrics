package config

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/caarlos0/env/v6"
)

const (
	DefaultAddress         = "localhost:8080"
	DefaultPollInterval    = 2
	DefaultReportInterval  = 10
	DefaultStoreInterval   = 300
	DefaultFileStoragePath = "storefile.txt"
	DefaultRestore         = true
)

type ServerConfig struct {
	Address         string `env:"ADDRESS"`
	StoreInterval   int    `env:"STORE_INTERVAL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         bool   `env:"RESTORE"`
}

type ClientConfig struct {
	Address        string `env:"ADDRESS"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	PollInterval   int    `env:"POLL_INTERVAL"`
}

func GetServerConfig() (ServerConfig, error) {
	var cfg ServerConfig

	flag.StringVar(&cfg.Address, "a", DefaultAddress, "Server address and port")
	flag.IntVar(&cfg.StoreInterval, "i", DefaultStoreInterval, "Store interval")
	flag.StringVar(&cfg.FileStoragePath, "f", DefaultFileStoragePath, "File with metrics")
	flag.BoolVar(&cfg.Restore, "r", DefaultRestore, "Restore from file (bool)")
	flag.Parse()

	if envAddress := os.Getenv("ADDRESS"); envAddress != "" {
		cfg.Address = envAddress
	}
	if envStoreInterval := os.Getenv("STORE_INTERVAL"); envStoreInterval != "" {
		iStoreInterval, err := strconv.Atoi(envStoreInterval)
		if err != nil {
			return ServerConfig{}, fmt.Errorf("invalid value for env variable `STORE_INTERVAL`")
		}
		cfg.StoreInterval = iStoreInterval
	}
	if envFileStorePath := os.Getenv("FILE_STORAGE_PATH"); envFileStorePath != "" {
		cfg.FileStoragePath = envFileStorePath
	}
	if envRestore := os.Getenv("RESTORE"); envRestore != "" {
		if strings.ToLower(envRestore) == "true" {
			cfg.Restore = true
		} else if strings.ToLower(envRestore) == "false" {
			cfg.Restore = false
		} else {
			return ServerConfig{}, fmt.Errorf("invalid value for env variable `RESTORE`")
		}
	}

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
