package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

const (
	StorageMemory   = "memory"
	StoragePostgres = "postgres"
)

type Config struct {
	StorageType string
	DatabaseURL string
	Port        string
	MaxRetries  int
}

func Load() (*Config, error) {
	cfg := &Config{
		StorageType: strings.ToLower(getEnv("STORAGE_TYPE", StorageMemory)),
		DatabaseURL: getEnv("DATABASE_URL", ""),
		Port:        getEnv("PORT", "8080"),
		MaxRetries:  getEnvAsInt("MAX_RETRIES", 3),
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	switch c.StorageType {
	case StorageMemory, StoragePostgres:
	default:
		return fmt.Errorf("unsupported storage type: %s", c.StorageType)
	}

	if c.Port == "" {
		return fmt.Errorf("port must not be empty")
	}

	port, err := strconv.Atoi(c.Port)
	if err != nil || port < 1 || port > 65535 {
		return fmt.Errorf("port must be a valid integer in range 1..65535")
	}

	if c.MaxRetries < 1 {
		return fmt.Errorf("max retries must be greater than zero")
	}

	if c.StorageType == StoragePostgres && c.DatabaseURL == "" {
		return fmt.Errorf("database url is required for postgres storage")
	}

	return nil
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	strValue := getEnv(key, "")
	if value, err := strconv.Atoi(strValue); err == nil {
		return value
	}
	return fallback
}
