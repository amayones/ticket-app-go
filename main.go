package main

import (
	"context"
	"embed"
	"flag"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"

	"golang-backend/config"
	"golang-backend/handlers"
	"golang-backend/internal/pidfile"
	"golang-backend/repositories"
	"golang-backend/routes"
	"golang-backend/services"
)

//go:embed all:frontend/dist
var embeddedDist embed.FS

var version = "dev"

func main() {
	hideFlag := flag.Bool("hide", false, "hide console window (Windows tray mode)")
	trayFlag := flag.Bool("tray", false, "alias for --hide")
	flag.Parse()
	if *hideFlag || *trayFlag {
		hideConsole()
		defer showConsole()
	}

	slog.Info("starting go-core", "version", version)

	cfg, err := config.Load()
	if err != nil {
		fail("invalid configuration", err)
	}

	if _, err := os.Stat(pidfile.Path()); err == nil {
		if oldPID, rerr := pidfile.Read(); rerr == nil && oldPID != os.Getpid() {
			slog.Warn("stale PID file exists, overwriting", "old_pid", oldPID)
		}
	}
	if err := pidfile.Write(); err != nil {
		slog.Warn("could not write pid file", "err", err)
	}
	defer pidfile.Remove()

	db, err := config.ConnectDatabase(cfg)
	if err != nil {
		fail("database connection failed", err)
	}
	defer db.Close()

	dialect := repositories.ParseDialect(cfg.DBConnection)
	userRepository := repositories.NewUserRepository(db, dialect)
	refreshTokenRepository := repositories.NewRefreshTokenRepository(db, dialect)
	roleRepository := repositories.NewRoleRepository(db, dialect)
	auditRepository := repositories.NewAuditRepository(db, dialect)
	syslogRepository := repositories.NewSyslogRepository(db, dialect)
	notifRepository := repositories.NewNotificationRepository(db, dialect)
	userService, err := services.NewUserService(userRepository, refreshTokenRepository, roleRepository, cfg.JWTSecret)
	if err != nil {
		fail("service init failed", err)
	}
	auditService := services.NewAuditService(auditRepository)
	syslogService := services.NewSyslogService(syslogRepository)
	notifService := services.NewNotificationService(notifRepository)
	userHandler := handlers.NewUserHandler(userService, auditService, syslogService)
	adminHandler := handlers.NewAdminHandler(userService, auditService, syslogService, notifService)

	routeCfg := routes.DefaultRouteConfig(cfg.JWTSecret)
	r := routes.SetupRoutesWithConfig(userHandler, adminHandler, routeCfg)
	if err := attachEmbeddedSPA(r); err != nil {
		slog.Warn("frontend/dist missing; API only", "err", err)
	}

	// Background cleanup of expired refresh tokens every hour.
	go func() {
		t := time.NewTicker(time.Hour)
		defer t.Stop()
		for range t.C {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			if n, err := userService.CleanupExpiredTokens(ctx); err != nil {
				slog.Warn("cleanup expired tokens failed", "err", err)
			} else if n > 0 {
				slog.Info("cleaned expired tokens", "count", n)
			}
			cancel()
		}
	}()

	server := &http.Server{
		Addr:         ":" + cfg.AppPort,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	go func() {
		slog.Info("listening", "port", cfg.AppPort, "api", "/api", "health", "/healthz")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server failed", "err", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		slog.Error("forced shutdown", "err", err)
		os.Exit(1)
	}
	slog.Info("server exited gracefully")
}

func attachEmbeddedSPA(r *chi.Mux) error {
	sub, err := fs.Sub(embeddedDist, "frontend/dist")
	if err != nil {
		return err
	}
	if _, err := fs.Stat(sub, "index.html"); err != nil {
		return err
	}
	// Cache index.html at startup instead of reading per request.
	indexData, err := fs.ReadFile(sub, "index.html")
	if err != nil {
		return err
	}
	r.NotFound(spaHandler(sub, indexData).ServeHTTP)
	return nil
}

func spaHandler(staticFS fs.FS, indexData []byte) http.Handler {
	fileServer := http.FileServer(http.FS(staticFS))
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if strings.HasPrefix(req.URL.Path, "/api") || strings.HasPrefix(req.URL.Path, "/healthz") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"error":"Not found"}`))
			return
		}
		path := strings.TrimPrefix(req.URL.Path, "/")
		if path == "" {
			serveIndexBytes(w, indexData)
			return
		}
		// Prevent path traversal escaping embed FS.
		if strings.Contains(path, "..") {
			serveIndexBytes(w, indexData)
			return
		}
		if f, err := fs.Stat(staticFS, path); err == nil && !f.IsDir() {
			fileServer.ServeHTTP(w, req)
			return
		}
		serveIndexBytes(w, indexData)
	})
}

func serveIndexBytes(w http.ResponseWriter, data []byte) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

// fail logs a fatal startup error and removes the PID file written earlier
// (os.Exit skips defers, so cleanup must happen explicitly here).
func fail(msg string, err error) {
	slog.Error(msg, "err", err)
	pidfile.Remove()
	os.Exit(1)
}
