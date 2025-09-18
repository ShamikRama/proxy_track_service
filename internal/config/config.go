package config

import (
	"os"
	"strconv"
	"time"
)

func Load() (*Config, error) {
	config := &Config{
		Server: ServerConfig{
			Port:         getEnv("SERVER_PORT", "8080"),
			ReadTimeout:  getDurationEnv("SERVER_READ_TIMEOUT", 30*time.Second),
			WriteTimeout: getDurationEnv("SERVER_WRITE_TIMEOUT", 30*time.Second),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getIntEnv("REDIS_DB", 0),
			TTL:      getDurationEnv("REDIS_TTL", 1*time.Hour),
		},
		External: ExternalConfig{
			BaseURL:     getEnv("EXTERNAL_API_BASE_URL", "https://track.4px.com"),
			HashPattern: getEnv("EXTERNAL_API_HASH_PATTERN", "/#/result/0/"),
			Timeout:     getDurationEnv("EXTERNAL_API_TIMEOUT", 60*time.Second),
			RetryCount:  getIntEnv("EXTERNAL_API_RETRY_COUNT", 3),
		},
		Batcher: BatcherConfig{
			BatchSize:    getIntEnv("BATCH_SIZE", 50),
			BatchTimeout: getDurationEnv("BATCH_FLUSH_TIMEOUT", 2*time.Second),
			Workers:      getIntEnv("BATCH_WORKERS", 3),
		},
	}

	return config, nil
}

type Config struct {
	Server   ServerConfig   `json:"server"`
	Redis    RedisConfig    `json:"redis"`
	External ExternalConfig `json:"external"`
	Batcher  BatcherConfig  `json:"batcher"`
}

type ServerConfig struct {
	Port         string        `json:"port"`
	ReadTimeout  time.Duration `json:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout"`
}

type RedisConfig struct {
	Host     string        `json:"host"`
	Port     string        `json:"port"`
	Password string        `json:"password"`
	DB       int           `json:"db"`
	TTL      time.Duration `json:"ttl"`
}

type BatcherConfig struct {
	BatchSize    int           `json:"batch_size"`
	BatchTimeout time.Duration `json:"batch_timeout"`
	Workers      int           `json:"workers"`
}

type ExternalConfig struct {
	BaseURL     string        `json:"base_url"`
	HashPattern string        `json:"hash_pattern"`
	Timeout     time.Duration `json:"timeout"`
	RetryCount  int           `json:"retry_count"`
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
