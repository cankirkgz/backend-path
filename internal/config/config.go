package config

import (
	"os"
	"strconv"
)

type Config struct {
	AppEnv        string
	Port          string
	DbDsn         string
	LogLevel      string
	RedisAddr     string
	RedisPassword string
	RedisDB       int
}

func MustLoad() Config {
	return Config{
		AppEnv:        getEnv("APP_ENV", "local"),
		Port:          getEnv("PORT", "8080"),
		DbDsn:         getEnv("DB_DSN", ""),
		LogLevel:      getEnv("LOG_LEVEL", "info"),
		RedisAddr:     getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       getEnvAsInt("REDIS_DB", 0),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}

	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsedValue, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return parsedValue
}
