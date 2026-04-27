package config

import (
	"os"
	"strconv"
	"time"
)

const (
	defaultRequestLimitPerMinute   = 12
	defaultChallengeReloadInterval = 60 * time.Second
)

type Settings struct {
	GiveFixedFiles        bool
	ChallengeDir          string
	AdminPassword         string
	HTTPAddr              string
	LogLevel              string
	RequestLimitPerMinute int
	CatalogReloadInterval time.Duration
}

func Load() Settings {
	return Settings{
		GiveFixedFiles:        os.Getenv("GIVE_FIXED_FILES") == "true",
		ChallengeDir:          valueOrDefaultString(os.Getenv("CHALLENGE_DIR"), "challenges"),
		AdminPassword:         os.Getenv("ADMIN_PASSWORD"),
		HTTPAddr:              valueOrDefaultString(os.Getenv("HTTP_ADDR"), ":8080"),
		LogLevel:              valueOrDefaultString(os.Getenv("LOG_LEVEL"), "info"),
		RequestLimitPerMinute: valueOrDefaultInt(os.Getenv("REQUEST_LIMIT_PER_MINUTE"), defaultRequestLimitPerMinute),
		CatalogReloadInterval: valueOrDefaultDuration(os.Getenv("CATALOG_RELOAD_INTERVAL"), defaultChallengeReloadInterval),
	}
}

func valueOrDefaultString(value string, fallback string) string {
	if value != "" {
		return value
	}

	return fallback
}

func valueOrDefaultInt(value string, fallback int) int {
	if value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}

	return fallback
}

func valueOrDefaultDuration(value string, fallback time.Duration) time.Duration {
	if value != "" {
		if durationValue, err := time.ParseDuration(value); err == nil {
			return durationValue
		}
	}

	return fallback
}
