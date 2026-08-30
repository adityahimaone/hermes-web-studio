package config

import (
	"os"
	"strings"
	"time"
)

type Config struct {
	Host            string
	Port            string
	GatewayBaseURL  string
	GatewayAPIKey   string
	DefaultModel    string
	DefaultProvider string
	ReadTimeout     time.Duration
}

func Load() Config {
	return Config{
		Host:            env("HERMES_WEBUI_HOST", "127.0.0.1"),
		Port:            env("HERMES_WEBUI_PORT", "8787"),
		GatewayBaseURL:  strings.TrimRight(env("HERMES_WEBUI_GATEWAY_BASE_URL", "http://127.0.0.1:8642"), "/"),
		GatewayAPIKey:   firstEnv("HERMES_WEBUI_GATEWAY_API_KEY", "API_SERVER_KEY"),
		DefaultModel:    env("HERMES_WEBUI_DEFAULT_MODEL", "default"),
		DefaultProvider: os.Getenv("HERMES_WEBUI_DEFAULT_PROVIDER"),
		ReadTimeout:     durationEnv("HERMES_WEBUI_GATEWAY_READ_TIMEOUT", 10*time.Minute),
	}
}

func env(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func firstEnv(keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(os.Getenv(key)); value != "" {
			return value
		}
	}
	return ""
}

func durationEnv(key string, fallback time.Duration) time.Duration {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(value)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}
