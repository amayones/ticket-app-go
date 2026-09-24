package config

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"time"

	_ "github.com/lib/pq"
	_ "github.com/microsoft/go-mssqldb"
	_ "modernc.org/sqlite"
)

// Connection pool defaults (overridable via Config).
const (
	DefaultConnMaxLifetime = 5 * time.Minute
	DefaultConnMaxIdleTime = 2 * time.Minute
	ConnectTimeout         = 10 * time.Second
)

// ConnectDatabase opens the pool for the configured engine with timeouts.
// sqlite membuat file + folder otomatis; pool dibatasi 1 koneksi tulis
// (menghindari SQLITE_BUSY). Driver lain tetap memakai pool penuh.
func ConnectDatabase(cfg Config) (*sql.DB, error) {
	var (
		db         *sql.DB
		err        error
		driverName string
	)
	switch cfg.DBConnection {
	case DBPostgres:
		driverName = "postgres"
		db, err = sql.Open("postgres", postgresDSN(cfg))
	case DBSQLite:
		driverName = "sqlite"
		db, err = sql.Open("sqlite", sqlitePath(cfg))
		cfg.DBMaxOpenConns = 1
		cfg.DBMaxIdleConns = 1
	default:
		driverName = "sqlserver"
		db, err = sql.Open("sqlserver", sqlserverDSN(cfg))
	}
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
	slog.Info("database connected", "driver", driverName)
	return db, nil
}

// sqlserverDSN escapes @ : / ? # in credentials correctly
// (plain Sprintf breaks when passwords contain them).
func sqlserverDSN(cfg Config) string {
	u := &url.URL{
		Scheme:   "sqlserver",
		User:     url.UserPassword(cfg.DBUsername, cfg.DBPassword),
		Host:     net.JoinHostPort(cfg.DBHost, cfg.DBPort),
		RawQuery: "database=" + url.QueryEscape(cfg.DBDatabase),
	}
	return u.String()
}

func postgresDSN(cfg Config) string {
	u := &url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(cfg.DBUsername, cfg.DBPassword),
		Host:     net.JoinHostPort(cfg.DBHost, cfg.DBPort),
		Path:     "/" + cfg.DBDatabase,
		RawQuery: "sslmode=disable&connect_timeout=10",
	}
	return u.String()
}

// sqlitePath resolves the DB file path (relative to workdir) and ensures
// its parent directory exists so first connect never fails with ENOENT.
func sqlitePath(cfg Config) string {
	path := cfg.DBDatabase
	if path == "" {
		path = "./data/go-core.db"
	}
	if dir := filepath.Dir(path); dir != "." && dir != "" {
		_ = os.MkdirAll(dir, 0755)
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return path
	}
	return abs
}
