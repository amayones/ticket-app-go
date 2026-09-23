package config

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net"
	"net/url"
	"time"

	_ "github.com/microsoft/go-mssqldb"
)

// Connection pool defaults (overridable via Config).
const (
	DefaultConnMaxLifetime = 5 * time.Minute
	DefaultConnMaxIdleTime = 2 * time.Minute
	ConnectTimeout         = 10 * time.Second
)

// ConnectDatabase opens the MSSQL pool with timeouts and safe DSN escaping.
// Returns error instead of log.Fatal so main can fail gracefully and
// integration tests can skip when DB is unavailable.
func ConnectDatabase(cfg Config) (*sql.DB, error) {
	// url.UserPassword escapes @ : / ? # in credentials correctly
	// (plain Sprintf breaks when passwords contain them).
	u := &url.URL{
		Scheme:   "sqlserver",
		User:     url.UserPassword(cfg.DBUsername, cfg.DBPassword),
		Host:     net.JoinHostPort(cfg.DBHost, cfg.DBPort),
		RawQuery: "database=" + url.QueryEscape(cfg.DBDatabase),
	}
	db, err := sql.Open("sqlserver", u.String())
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	db.SetMaxOpenConns(cfg.DBMaxOpenConns)
	db.SetMaxIdleConns(cfg.DBMaxIdleConns)
	db.SetConnMaxLifetime(DefaultConnMaxLifetime)
	db.SetConnMaxIdleTime(DefaultConnMaxIdleTime)

	ctx, cancel := context.WithTimeout(context.Background(), ConnectTimeout)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}
	slog.Info("database connected")
	return db, nil
}
