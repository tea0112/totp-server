package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	TOTPPeriod time.Duration
	ServerPort string
}

func Load() *Config {
	period := 300
	if p := os.Getenv("TOTP_PERIOD"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			period = parsed
		}
	}

	port := "8080"
	if p := os.Getenv("SERVER_PORT"); p != "" {
		port = p
	}

	return &Config{
		TOTPPeriod: time.Duration(period) * time.Second,
		ServerPort: port,
	}
}
