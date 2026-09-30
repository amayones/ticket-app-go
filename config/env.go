package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Supported database engines (DB_CONNECTION).
const (
	DBSQLServer = "sqlserver"
	DBPostgres  = "postgres"
	DBSQLite    = "sqlite"
)

// Config holds all runtime configuration in one place.
// Load once at startup and inject where needed (no global os.Getenv scattering).
type Config struct {
	AppName string
	AppEnv  string
	AppPort string

	// DBConnection selects the engine: sqlserver (default), postgres, sqlite.
	// sqlite memakai DBDatabase sebagai path file (mis. ./data/go-core.db)
	// dan mengabaikan host/port/username/password.
	DBConnection string
	DBHost       string
	DBPort       string
	DBDatabase   string
	DBUsername   string
	DBPassword   string

	JWTSecret string

	// QRSecret signs ticket QR tokens (TICKETING module). Optional:
	// empty falls back to JWTSecret so existing .env files keep working.
	QRSecret string

	// RefundCutoffHours: refund requests close this many hours before the
	// event starts (default 24). FINANCE owns the payout.
	RefundCutoffHours int

	// Token lifetimes (umur sesi). ACCESS_TOKEN_MINUTES dalam menit
	// (default 15), REFRESH_TOKEN_DAYS dalam hari (default 7).
	// Nilai <= 0 / bukan angka otomatis fallback ke default.
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration

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
		AppName:         getEnvDefault("APP_NAME", "Go Core"),
		AppEnv:          getEnvDefault("APP_ENV", "development"),
		AppPort:         getEnvDefault("APP_PORT", "1067"),
		DBConnection:    normalizeDriver(getEnvDefault("DB_CONNECTION", DBSQLServer)),
		DBHost:          strings.TrimSpace(os.Getenv("DB_HOST")),
		DBPort:          strings.TrimSpace(os.Getenv("DB_PORT")),
		DBDatabase:      strings.TrimSpace(os.Getenv("DB_DATABASE")),
		DBUsername:      strings.TrimSpace(os.Getenv("DB_USERNAME")),
		DBPassword:      os.Getenv("DB_PASSWORD"), // keep as-is; may contain spaces
		JWTSecret:       strings.TrimSpace(os.Getenv("JWT_SECRET")),
		QRSecret:        strings.TrimSpace(os.Getenv("QR_SECRET")),
		AccessTokenTTL:  time.Duration(getEnvIntDefault("ACCESS_TOKEN_MINUTES", 15)) * time.Minute,
		RefreshTokenTTL: time.Duration(getEnvIntDefault("REFRESH_TOKEN_DAYS", 7)) * 24 * time.Hour,
		DBMaxOpenConns:  getEnvIntDefault("DB_MAX_OPEN_CONNS", 25),
		DBMaxIdleConns:  getEnvIntDefault("DB_MAX_IDLE_CONNS", 10),
		RefundCutoffHours: getEnvIntDefault("REFUND_CUTOFF_HOURS", 24),
	}

	var missing []string
	required := [][2]string{
		{"DB_DATABASE", cfg.DBDatabase},
		{"JWT_SECRET", cfg.JWTSecret},
	}
	if cfg.DBConnection != DBSQLite {
		required = append(required,
			[2]string{"DB_HOST", cfg.DBHost},
			[2]string{"DB_PORT", cfg.DBPort},
			[2]string{"DB_USERNAME", cfg.DBUsername},
			[2]string{"DB_PASSWORD", cfg.DBPassword},
		)
	}
	for _, kv := range required {
		if strings.TrimSpace(kv[1]) == "" {
			missing = append(missing, kv[0])
		}
	}
	if len(missing) > 0 {
		return Config{}, fmt.Errorf("missing required env: %s", strings.Join(missing, ", "))
	}
	if cfg.DBConnection != DBSQLite {
		if _, err := strconv.Atoi(cfg.DBPort); err != nil {
			return Config{}, fmt.Errorf("invalid DB_PORT %q: must be numeric", cfg.DBPort)
		}
	}
	if len(cfg.JWTSecret) < 32 {
		return Config{}, fmt.Errorf("JWT_SECRET must be at least 32 characters")
	}
	if cfg.QRSecret == "" {
		cfg.QRSecret = cfg.JWTSecret
	} else if len(cfg.QRSecret) < 32 {
		return Config{}, fmt.Errorf("QR_SECRET must be at least 32 characters")
	}
	return cfg, nil
}

// normalizeDriver menerima alias umum (mssql, postgresql) dan menolak
// engine yang belum didukung dengan fallback aman ke sqlserver.
func normalizeDriver(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case DBPostgres, "postgresql", "pg":
		return DBPostgres
	case DBSQLite, "sqlite3":
		return DBSQLite
	case DBSQLServer, "mssql", "":
		return DBSQLServer
	default:
		return DBSQLServer
	}
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
