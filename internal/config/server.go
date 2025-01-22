// Package config define server config structure
package config

import (
	"encoding/json"
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
	DefaultCryptoKey       = ""               // Path to file with private key
	DefaultConfigPath      = ""               // Path to json config file
)

// ServerConfig server config structure
type ServerConfig struct {
	Address         string `json:"address" env:"ADDRESS"`
	StoreInterval   int    `json:"store_interval" env:"STORE_INTERVAL"`
	FileStoragePath string `json:"file_storage_path" env:"FILE_STORAGE_PATH"`
	Restore         bool   `json:"restore" env:"RESTORE"`
	DataSourceName  string `json:"database_dsn" env:"DATABASE_DSN"`
	HashKey         string `json:"key" env:"KEY"`
	CryptoKey       string `json:"crypto_key" env:"CRYPTO_KEY"`
	ConfigPath      string `env:"CONFIG"`
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
	flag.StringVar(&cfg.CryptoKey, "crypto-key", DefaultCryptoKey, "Path to private crypto key")
	flag.StringVar(&cfg.ConfigPath, "c", DefaultConfigPath, "Configuration file path")
	flag.Parse()

	envConfigPath := os.Getenv("CONFIG")
	if envConfigPath != "" {
		cfg.ConfigPath = envConfigPath
	}

	if cfg.ConfigPath != "" {
		fileCfg, err := loadConfigFromFile(cfg.ConfigPath)
		if err != nil {
			return ServerConfig{}, err
		}
		mergeConfig(&cfg, fileCfg)
	}

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

	if envCryptoKey := os.Getenv("CRYPTO_KEY"); envCryptoKey != "" {
		cfg.CryptoKey = envCryptoKey
	}

	// Validate file path and create if not exists
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

	// Validate crypto key path and return err if not exists
	if cfg.CryptoKey != "" {
		_, err := os.Stat(cfg.CryptoKey)
		if err != nil {
			return ServerConfig{}, err
		}
	}

	return cfg, nil
}

func loadConfigFromFile(path string) (ServerConfig, error) {
	file, err := os.Open(path)
	if err != nil {
		return ServerConfig{}, err
	}
	defer func() {
		err := file.Close()
		if err != nil {
			fmt.Printf("Error closing file: %v", err)
		}
	}()

	var cfg ServerConfig
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&cfg); err != nil {
		return ServerConfig{}, err
	}

	return cfg, nil
}

func mergeConfig(cfg *ServerConfig, fileCfg ServerConfig) {
	if cfg.Address == DefaultAddress && fileCfg.Address != "" {
		cfg.Address = fileCfg.Address
	}
	if cfg.StoreInterval == DefaultStoreInterval && fileCfg.StoreInterval != 0 {
		cfg.StoreInterval = fileCfg.StoreInterval
	}
	if cfg.FileStoragePath == DefaultFileStoragePath && fileCfg.FileStoragePath != "" {
		cfg.FileStoragePath = fileCfg.FileStoragePath
	}
	if cfg.Restore == DefaultRestore && fileCfg.Restore != false {
		cfg.Restore = fileCfg.Restore
	}
	if cfg.DataSourceName == DefaultDSN && fileCfg.DataSourceName != "" {
		cfg.DataSourceName = fileCfg.DataSourceName
	}
	if cfg.HashKey == DefaultHashKey && fileCfg.HashKey != "" {
		cfg.HashKey = fileCfg.HashKey
	}
	if cfg.CryptoKey == DefaultCryptoKey && fileCfg.CryptoKey != "" {
		cfg.CryptoKey = fileCfg.CryptoKey
	}
}
