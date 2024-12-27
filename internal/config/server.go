// Package config define server config structure
package config

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Default server config settings
const (
	DefaultAddress         = "localhost:8080" // Server address
	DefaultStoreInterval   = 300              // Store interval in seconds
	DefaultFileStoragePath = "metrics.json"   // Path to storage file
	DefaultRestore         = true             // Restore metrics from file
	DefaultDSN             = ""               // DSN connection string
	DefaultHashKey         = ""               // Secret string for hashing messages
)

// ServerConfig server config structure
type ServerConfig struct {
	Address         string `env:"ADDRESS"`
	StoreInterval   int    `env:"STORE_INTERVAL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         bool   `env:"RESTORE"`
	DataSourceName  string `env:"DATABASE_DSN"`
	HashKey         string `env:"KEY"`
}

// GetServerConfig allows to get instance of ServerConfig
func GetServerConfig() (ServerConfig, error) {
	var cfg ServerConfig

	flag.StringVar(&cfg.Address, "a", DefaultAddress, "Server address and port")
	flag.IntVar(&cfg.StoreInterval, "i", DefaultStoreInterval, "Store interval")
	flag.StringVar(&cfg.FileStoragePath, "f", DefaultFileStoragePath, "File with metrics")
	flag.BoolVar(&cfg.Restore, "r", DefaultRestore, "Restore from file (bool)")
	flag.StringVar(&cfg.DataSourceName, "d", DefaultDSN, "Database DSN")
	flag.StringVar(&cfg.HashKey, "k", DefaultHashKey, "Hash key for calculation HashSHA256 header")
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

	if envHashKey := os.Getenv("KEY"); envHashKey != "" {
		cfg.HashKey = envHashKey
	}

	// Validate file path
	if cfg.DataSourceName != "" {
		_, err := os.Stat(cfg.FileStoragePath)

		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				f, err := os.Create(cfg.FileStoragePath)
				if err != nil {
					return ServerConfig{}, err
				}
				_ = f.Close()
			} else {
				return ServerConfig{}, nil
			}
		}
	}

	return cfg, nil
}
