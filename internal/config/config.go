package config

import (
	"log/slog"
	"os"
)

type Config struct {
	AppEnv          string
	ServerPort      string
	DatabaseDSN     string
	JWTSecret       string
	DefaultEmail    string
	DefaultPassword string
}

func Load() Config {
	cfg := Config{
		AppEnv:          getenv("APP_ENV", "development"),
		ServerPort:      getenv("SERVER_PORT", ":8080"),
		DatabaseDSN:     getenv("DATABASE_DSN", "app.db"),
		JWTSecret:       getenv("JWT_SECRET", "change-me-in-production"),
		DefaultEmail:    getenv("DEFAULT_USER_EMAIL", "admin@jarakey.com"),
		DefaultPassword: getenv("DEFAULT_USER_PASSWORD", "Pa$$w0rd!"),
	}

	setupLogger(cfg.AppEnv)
	return cfg
}

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func setupLogger(env string) {
	level := slog.LevelInfo
	if env == "development" {
		level = slog.LevelDebug
	}

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	slog.SetDefault(slog.New(handler))
}
