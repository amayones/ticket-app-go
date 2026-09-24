package routes

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"golang-backend/handlers"
	appmw "golang-backend/middleware"
	"golang-backend/models"
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

func SetupRoutesWithConfig(userHandler *handlers.UserHandler, adminHandler *handlers.AdminHandler, cfg RouteConfig) *chi.Mux {
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
		r.With(auth).Get("/roles", userHandler.ListRoles)
		r.Route("/users", func(r chi.Router) {
			r.With(auth).Get("/", userHandler.GetUsers)
			r.With(registerLimiter.Middleware).Post("/", userHandler.CreateUser)
			r.With(auth).Get("/{code}", userHandler.GetUserByCode)
			r.With(auth).Put("/{code}", userHandler.UpdateUser)
			r.With(auth).Delete("/{code}", userHandler.DeleteUser)
			r.With(auth).Post("/{code}/logout-all", userHandler.LogoutAll)
			r.With(auth, appmw.RequirePermission(userHandler.Service, models.PermUserRoleAssign)).
				Put("/{code}/role", adminHandler.UpdateUserRole)
		})

		// --- Menu admin (RBAC) ---
		r.Route("/admin", func(r chi.Router) {
			// Role & Permission
			r.With(auth, appmw.RequirePermission(userHandler.Service, models.PermRoleManage)).Post("/roles", adminHandler.CreateRole)
			r.With(auth, appmw.RequirePermission(userHandler.Service, models.PermRoleRead)).Get("/roles/{code}", adminHandler.GetRoleDetail)
			r.With(auth, appmw.RequirePermission(userHandler.Service, models.PermRoleManage)).Delete("/roles/{code}", adminHandler.DeleteRole)
			r.With(auth, appmw.RequirePermission(userHandler.Service, models.PermRoleRead)).Get("/permissions", adminHandler.ListPermissions)
			r.With(auth, appmw.RequirePermission(userHandler.Service, models.PermPermissionAssign)).Put("/roles/{code}/permissions", adminHandler.SetRolePermissions)
			// Authentication & Session Management
			r.With(auth).Get("/sessions", adminHandler.ListMySessions)
			r.With(auth, appmw.RequirePermission(userHandler.Service, models.PermSessionManage)).Get("/sessions/all", adminHandler.ListAllSessions)
			r.With(auth).Delete("/sessions/{id}", adminHandler.RevokeSession)
			// Audit Log
			r.With(auth, appmw.RequirePermission(userHandler.Service, models.PermAuditRead)).Get("/audit", adminHandler.ListAudit)
			// Security Center
			r.With(auth, appmw.RequirePermission(userHandler.Service, models.PermSecurityRead)).Get("/security/summary", adminHandler.SecuritySummary)
			// Error / System Log
			r.With(auth, appmw.RequirePermission(userHandler.Service, models.PermSyslogRead)).Get("/syslogs", adminHandler.ListSyslog)
			r.With(auth, appmw.RequirePermission(userHandler.Service, models.PermSyslogManage)).Delete("/syslogs", adminHandler.PruneSyslog)
			// Notification Template & Log
			r.With(auth, appmw.RequirePermission(userHandler.Service, models.PermNotifRead)).Get("/notifications/templates", adminHandler.ListTemplates)
			r.With(auth, appmw.RequirePermission(userHandler.Service, models.PermNotifManage)).Post("/notifications/templates", adminHandler.CreateTemplate)
			r.With(auth, appmw.RequirePermission(userHandler.Service, models.PermNotifManage)).Put("/notifications/templates/{code}", adminHandler.UpdateTemplate)
			r.With(auth, appmw.RequirePermission(userHandler.Service, models.PermNotifManage)).Delete("/notifications/templates/{code}", adminHandler.DeleteTemplate)
			r.With(auth, appmw.RequirePermission(userHandler.Service, models.PermNotifSend)).Post("/notifications/send", adminHandler.SendNotification)
			r.With(auth, appmw.RequirePermission(userHandler.Service, models.PermNotifRead)).Get("/notifications/logs", adminHandler.ListNotifLogs)
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
