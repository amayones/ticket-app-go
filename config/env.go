package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds all runtime configuration in one place.
// Load once at startup and inject where needed (no global os.Getenv scattering).
type Config struct {
	AppName string
	AppEnv  string
	AppPort string

	DBHost     string
	DBPort     string
	DBDatabase string
	DBUsername string
	DBPassword string

	JWTSecret string

	// Pool tuning with sane defaults, overridable via env.
	DBMaxOpenConns int
	DBMaxIdleConns int
}

// Load reads .env best-effort (12-factor friendly: real env vars win)
// and validates required values. Returns error instead of log.Fatal
// so callers can decide how to fail and tests can cover it.
func Load() (Config, error) {
	// Best-effort: missing .env is fine in prod/Docker/CI where env is injected.
	_ = godotenv.Load()

	cfg := Config{
		AppName:        getEnvDefault("APP_NAME", "GoBackend"),
		AppEnv:         getEnvDefault("APP_ENV", "development"),
		AppPort:        getEnvDefault("APP_PORT", "1067"),
		DBHost:         strings.TrimSpace(os.Getenv("DB_HOST")),
		DBPort:         strings.TrimSpace(os.Getenv("DB_PORT")),
		DBDatabase:     strings.TrimSpace(os.Getenv("DB_DATABASE")),
		DBUsername:     strings.TrimSpace(os.Getenv("DB_USERNAME")),
		DBPassword:     os.Getenv("DB_PASSWORD"), // keep as-is; may contain spaces
		JWTSecret:      strings.TrimSpace(os.Getenv("JWT_SECRET")),
		DBMaxOpenConns: getEnvIntDefault("DB_MAX_OPEN_CONNS", 25),
		DBMaxIdleConns: getEnvIntDefault("DB_MAX_IDLE_CONNS", 10),
	}

	var missing []string
	for _, kv := range [][2]string{
		{"DB_HOST", cfg.DBHost},
		{"DB_PORT", cfg.DBPort},
		{"DB_DATABASE", cfg.DBDatabase},
		{"DB_USERNAME", cfg.DBUsername},
		{"DB_PASSWORD", cfg.DBPassword},
		{"JWT_SECRET", cfg.JWTSecret},
	} {
		if strings.TrimSpace(kv[1]) == "" {
			missing = append(missing, kv[0])
		}
	}
	if len(missing) > 0 {
		return Config{}, fmt.Errorf("missing required env: %s", strings.Join(missing, ", "))
	}
	if _, err := strconv.Atoi(cfg.DBPort); err != nil {
		return Config{}, fmt.Errorf("invalid DB_PORT %q: must be numeric", cfg.DBPort)
	}
	if len(cfg.JWTSecret) < 32 {
		return Config{}, fmt.Errorf("JWT_SECRET must be at least 32 characters")
	}
	return cfg, nil
}

func getEnvDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvIntDefault(key string, fallback int) int {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		return fallback
	}
	return n
}
