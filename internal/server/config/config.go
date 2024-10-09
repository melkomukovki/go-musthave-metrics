package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

const (
	DefaultAddress         = "localhost:8080"
	DefaultStoreInterval   = 300
	DefaultFileStoragePath = "storefile.json"
	DefaultRestore         = true
	DefaultDSN             = ""
)

type ServerConfig struct {
	Address         string `env:"ADDRESS"`
	StoreInterval   int    `env:"STORE_INTERVAL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         bool   `env:"RESTORE"`
	DataSourceName  string `env:"DATABASE_DSN"`
}

func GetServerConfig() (ServerConfig, error) {
	var cfg ServerConfig

	flag.StringVar(&cfg.Address, "a", DefaultAddress, "Server address and port")
	flag.IntVar(&cfg.StoreInterval, "i", DefaultStoreInterval, "Store interval")
	flag.StringVar(&cfg.FileStoragePath, "f", DefaultFileStoragePath, "File with metrics")
	flag.BoolVar(&cfg.Restore, "r", DefaultRestore, "Restore from file (bool)")
	flag.StringVar(&cfg.DataSourceName, "d", DefaultDSN, "Database DSN")
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
	if envDatabaseDSN := os.Getenv("DATABASE_DSN"); envDatabaseDSN != "" {
		cfg.DataSourceName = envDatabaseDSN
	}

	return cfg, nil
}
