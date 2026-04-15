package config

import "os"

type Config struct {
	AppEnv   string
	Port     string
	DbDsn    string
	LogLevel string
}

func MustLoad() Config {
	return Config{
		AppEnv:   getEnv("APP_ENV", "local"),
		Port:     getEnv("PORT", "8080"),
		DbDsn:    getEnv("DB_DSN", ""),
		LogLevel: getEnv("LOG_LEVEL", "info"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
