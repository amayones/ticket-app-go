package main

import (
	"context"
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"

	"golang-backend/config"
	"golang-backend/handlers"
	"golang-backend/repositories"
	"golang-backend/routes"
	"golang-backend/services"
)

//go:embed all:frontend/dist
var embeddedDist embed.FS

func main() {
	log.Println("Starting Go Backend...")
	config.LoadEnv()
	db := config.ConnectDatabase()
	defer db.Close()
	userRepository := repositories.NewUserRepository(db)
	refreshTokenRepository := repositories.NewRefreshTokenRepository(db)
	userService := services.NewUserService(userRepository, refreshTokenRepository)
	userHandler := handlers.NewUserHandler(userService)
	r := routes.SetupRoutes(userHandler)
	if err := attachEmbeddedSPA(r); err != nil {
		log.Printf("WARN: frontend/dist not found (%v)", err)
		log.Println("WARN: jalankan: npm --prefix frontend run build  atau dev: npm --prefix frontend run dev")
		log.Println("WARN: API tetap jalan di /api, tapi \"/\" akan 404 sampai frontend dibuild")
	} else {
		log.Println("Frontend embedded from frontend/dist (single binary mode)")
	}
	port := config.GetEnvDefault("APP_PORT", "8080")
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	go func() {
		log.Printf("Application is running on http://localhost:%s\n", port)
		log.Printf("API available at http://localhost:%s/api\n", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server failed to start:", err)
		}
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}
	log.Println("Server exited gracefully")
}

func attachEmbeddedSPA(r *chi.Mux) error {
	sub, err := fs.Sub(embeddedDist, "frontend/dist")
	if err != nil {
		return err
	}
	if _, err := fs.Stat(sub, "index.html"); err != nil {
		return err
	}
	r.NotFound(spaHandler(sub).ServeHTTP)
	return nil
}

func spaHandler(staticFS fs.FS) http.Handler {
	fileServer := http.FileServer(http.FS(staticFS))
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if strings.HasPrefix(req.URL.Path, "/api") {
			http.NotFound(w, req)
			return
		}
		path := strings.TrimPrefix(req.URL.Path, "/")
		if path == "" {
			serveIndex(staticFS, w, req)
			return
		}
		if f, err := fs.Stat(staticFS, path); err == nil && !f.IsDir() {
			fileServer.ServeHTTP(w, req)
			return
		}
		serveIndex(staticFS, w, req)
	})
}

func serveIndex(staticFS fs.FS, w http.ResponseWriter, r *http.Request) {
	data, err := fs.ReadFile(staticFS, "index.html")
	if err != nil {
		http.Error(w, "index.html not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}
