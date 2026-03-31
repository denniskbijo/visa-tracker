package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port            string
	DBPath          string
	MigrationsDir   string
	DataDir         string
	StaticDir       string
	TemplatesDir    string
	SponsorCSVURL   string
	RefreshInterval int
}

const defaultSponsorCSV = "https://assets.publishing.service.gov.uk/media/" +
	"69c270b7bb0dfe55b83e4c53/2026-03-24_-_Worker_and_Temporary_Worker.csv"

func Load() *Config {
	return &Config{
		Port:            envOr("PORT", "8080"),
		DBPath:          envOr("DB_PATH", "visa-tracker.db"),
		MigrationsDir:   envOr("MIGRATIONS_DIR", "data/migrations"),
		DataDir:         envOr("DATA_DIR", "data"),
		StaticDir:       envOr("STATIC_DIR", "static"),
		TemplatesDir:    envOr("TEMPLATES_DIR", "internal/templates"),
		SponsorCSVURL:   envOr("SPONSOR_CSV_URL", defaultSponsorCSV),
		RefreshInterval: envOrInt("REFRESH_INTERVAL_HOURS", 24),
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envOrInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}
