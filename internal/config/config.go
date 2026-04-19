package config

import (
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	HTTPAddr              string
	DatabaseURL           string
	ServiceAccessPassword string
	RoomTTL               time.Duration
	JitsiDomain           string
	JitsiAppID            string
	JitsiAppSecret        string
	JitsiAudience         string
	JitsiSubject          string
	JitsiTokenTTL         time.Duration
	APITokenSecret        string
	APITokenTTL           time.Duration
	CleanupInterval       time.Duration
	AllowedOrigins        []string
	LogLevelValue         string
	AppBuildID            string
}

func Load() Config {
	allowedOrigins := envCSV("ALLOWED_ORIGINS", "http://localhost:8080,http://localhost:5173")
	return Config{
		HTTPAddr:              env("HTTP_ADDR", ":8080"),
		DatabaseURL:           env("DATABASE_URL", "postgres://fistream:fistream@localhost:5432/fistream?sslmode=disable"),
		ServiceAccessPassword: env("SERVICE_ACCESS_PASSWORD", ""),
		RoomTTL:               envDuration("ROOM_TTL", 2*time.Hour),
		JitsiDomain:           env("JITSI_DOMAIN", "localhost:8000"),
		JitsiAppID:            env("JITSI_APP_ID", "fistream"),
		JitsiAppSecret:        env("JITSI_APP_SECRET", "dev-jitsi-secret"),
		JitsiAudience:         env("JITSI_AUDIENCE", "fistream"),
		JitsiSubject:          env("JITSI_SUBJECT", "meet.jitsi"),
		JitsiTokenTTL:         envDuration("JITSI_TOKEN_TTL", 2*time.Hour),
		APITokenSecret:        env("API_TOKEN_SECRET", "dev-api-token-secret"),
		APITokenTTL:           envDuration("API_TOKEN_TTL", 12*time.Hour),
		CleanupInterval:       envDuration("CLEANUP_INTERVAL", time.Minute),
		AllowedOrigins:        allowedOrigins,
		LogLevelValue:         env("LOG_LEVEL", "info"),
		AppBuildID:            env("APP_BUILD_ID", "dev"),
	}
}

func (c Config) LogLevel() slog.Level {
	switch strings.ToLower(c.LogLevelValue) {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func env(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func envDuration(key string, fallback time.Duration) time.Duration {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func envCSV(key, fallback string) []string {
	value := env(key, fallback)
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func envInt(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

