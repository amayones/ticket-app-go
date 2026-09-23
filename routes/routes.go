package routes

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"golang-backend/handlers"
	appmw "golang-backend/middleware"
)

// RouteConfig injects secrets and limits (no magic inside router).
type RouteConfig struct {
	JWTSecret      string
	LoginLimit     int
	LoginWindow    time.Duration
	RegisterLimit  int
	RegisterWindow time.Duration
	RefreshLimit   int
	RefreshWindow  time.Duration
	TrustProxy     bool
	RequestTimeout time.Duration
}

func DefaultRouteConfig(jwtSecret string) RouteConfig {
	return RouteConfig{
		JWTSecret:      jwtSecret,
		LoginLimit:     5,
		LoginWindow:    time.Minute,
		RegisterLimit:  10,
		RegisterWindow: time.Minute,
		RefreshLimit:   30,
		RefreshWindow:  time.Minute,
		RequestTimeout: 10 * time.Second,
	}
}

func SetupRoutesWithConfig(userHandler *handlers.UserHandler, cfg RouteConfig) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(cfg.RequestTimeout))
	r.Use(secureHeaders)

	r.MethodNotAllowed(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
	})
	r.NotFound(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Not found"})
	})

	r.Get("/healthz", userHandler.Health)

	loginLimiter := appmw.NewRateLimiterWithOptions(cfg.LoginLimit, cfg.LoginWindow, cfg.TrustProxy)
	registerLimiter := appmw.NewRateLimiterWithOptions(cfg.RegisterLimit, cfg.RegisterWindow, cfg.TrustProxy)
	refreshLimiter := appmw.NewRateLimiterWithOptions(cfg.RefreshLimit, cfg.RefreshWindow, cfg.TrustProxy)
	auth := appmw.NewAuth(cfg.JWTSecret)

	r.Route("/api", func(r chi.Router) {
		r.With(loginLimiter.Middleware).Post("/login", userHandler.Login)
		r.With(refreshLimiter.Middleware).Post("/refresh", userHandler.RefreshToken)
		r.With(refreshLimiter.Middleware).Post("/logout", userHandler.Logout)
		r.Route("/users", func(r chi.Router) {
			r.With(auth).Get("/", userHandler.GetUsers)
			r.With(registerLimiter.Middleware).Post("/", userHandler.CreateUser)
			r.With(auth).Get("/{id}", userHandler.GetUserByID)
			r.With(auth).Put("/{id}", userHandler.UpdateUser)
			r.With(auth).Delete("/{id}", userHandler.DeleteUser)
			r.With(auth).Post("/{id}/logout-all", userHandler.LogoutAll)
		})
	})
	return r
}

func secureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		next.ServeHTTP(w, r)
	})
}
