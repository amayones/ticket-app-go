package routes

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	faudit "golang-backend/features/audit"
	fnotif "golang-backend/features/notifications"
	froles "golang-backend/features/roles"
	fsecurity "golang-backend/features/security"
	fsessions "golang-backend/features/sessions"
	fsyslog "golang-backend/features/syslog"
	fusers "golang-backend/features/users"
	appmw "golang-backend/middleware"
	"golang-backend/models"
)

// RouteConfig injects secrets and limits (no magic inside router).
type RouteConfig struct {
	JWTSecret      string
	LoginLimit     int
	LoginWindow    time.Duration
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
		RefreshLimit:   30,
		RefreshWindow:  time.Minute,
		RequestTimeout: 10 * time.Second,
	}
}

// Deps carries one handler per menu (features/<menu>).
// Tambah menu backend = tambah 1 field + 1 blok route di bawah.
type Deps struct {
	Users         *fusers.Handler
	Roles         *froles.Handler
	Sessions      *fsessions.Handler
	Audit         *faudit.Handler
	Security      *fsecurity.Handler
	Syslog        *fsyslog.Handler
	Notifications *fnotif.Handler
}

func SetupRoutesWithConfig(d Deps, cfg RouteConfig) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(cfg.RequestTimeout))
	r.Use(middleware.StripSlashes)
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

	r.Get("/healthz", d.Users.Health)

	loginLimiter := appmw.NewRateLimiterWithOptions(cfg.LoginLimit, cfg.LoginWindow, cfg.TrustProxy)
	refreshLimiter := appmw.NewRateLimiterWithOptions(cfg.RefreshLimit, cfg.RefreshWindow, cfg.TrustProxy)
	auth := appmw.NewAuth(cfg.JWTSecret)
	need := func(perm string) func(http.Handler) http.Handler {
		return appmw.RequirePermission(d.Roles.Service, perm)
	}
	r.Route("/api", func(r chi.Router) {
		// Menu: users (auth inti + CRUD akun).
		r.With(loginLimiter.Middleware).Post("/login", d.Users.Login)
		r.With(refreshLimiter.Middleware).Post("/refresh", d.Users.RefreshToken)
		r.With(refreshLimiter.Middleware).Post("/logout", d.Users.Logout)
		r.With(auth).Get("/roles", d.Roles.ListRoles)
		// Menu milik user login (sidebar + placeholder 404). Auth saja.
		r.With(auth).Get("/menus/mine", d.Roles.MyMenus)
		r.Route("/users", func(r chi.Router) {
			r.With(auth).Get("/me", d.Users.GetMe)
			r.With(auth, need(models.MenuUsers)).Post("/", d.Users.CreateUser)
			r.With(auth, need(models.MenuUsers)).Get("/", d.Users.GetUsers)
			r.With(auth, need(models.MenuUsers)).Get("/{code}", d.Users.GetUserByCode)
			r.With(auth).Put("/{code}", d.Users.UpdateUser)
			r.With(auth).Delete("/{code}", d.Users.DeleteUser)
			r.With(auth).Post("/{code}/logout-all", d.Users.LogoutAll)
			r.With(auth, need(models.MenuUsers)).Put("/{code}/role", d.Roles.UpdateUserRole)
		})

		r.Route("/admin", func(r chi.Router) {
			// Module SYSTEM: setiap endpoint memakai permission menu yang sama.
			r.With(auth, need(models.MenuRoles)).Post("/roles", d.Roles.CreateRole)
			r.With(auth, need(models.MenuRoles)).Get("/roles/{code}", d.Roles.GetRoleDetail)
			r.With(auth, need(models.MenuRoles)).Delete("/roles/{code}", d.Roles.DeleteRole)
			r.With(auth, need(models.MenuRoles)).Get("/permissions", d.Roles.ListPermissions)
			r.With(auth, need(models.MenuRoles)).Put("/roles/{code}/permissions", d.Roles.SetRolePermissions)
			// Registry menu (CPMENU + CPMATRIX). Buat/hapus tanpa auto-grant role.
			r.With(auth, need(models.MenuRoles)).Get("/menus", d.Roles.ListMenus)
			r.With(auth, need(models.MenuRoles)).Post("/menus", d.Roles.CreateMenu)
			r.With(auth, need(models.MenuRoles)).Delete("/menus/{code}", d.Roles.DeleteMenu)
			r.With(auth, need(models.MenuRoles)).Get("/matrix", d.Roles.GetMatrix)
			// Master modul (CPMATRIX tabel). Urut: modul -> menu -> permission.
			r.With(auth, need(models.MenuRoles)).Get("/modules", d.Roles.ListModules)
			r.With(auth, need(models.MenuRoles)).Post("/modules", d.Roles.CreateModule)
			r.With(auth, need(models.MenuRoles)).Delete("/modules/{code}", d.Roles.DeleteModule)
			// Menu: sessions. Lihat semua sesi butuh permission; revoke sesi
			// lain ditangani RevokeSession (manageAll dari permission caller).
			r.With(auth, need(models.MenuSessions)).Get("/sessions/all", d.Sessions.ListAllSessions)
			// Menu: audit.
			r.With(auth, need(models.MenuAudit)).Get("/audit", d.Audit.ListAudit)
			// Menu: security.
			r.With(auth, need(models.MenuSecurity)).Get("/security/summary", d.Security.SecuritySummary)
			// Menu: syslog.
			r.With(auth, need(models.MenuSyslog)).Get("/syslogs", d.Syslog.ListSyslog)
			r.With(auth, need(models.MenuSyslog)).Delete("/syslogs", d.Syslog.PruneSyslog)
			// Menu: notifications.
			r.With(auth, need(models.MenuNotifications)).Get("/notifications/templates", d.Notifications.ListTemplates)
			r.With(auth, need(models.MenuNotifications)).Post("/notifications/templates", d.Notifications.CreateTemplate)
			r.With(auth, need(models.MenuNotifications)).Put("/notifications/templates/{code}", d.Notifications.UpdateTemplate)
			r.With(auth, need(models.MenuNotifications)).Delete("/notifications/templates/{code}", d.Notifications.DeleteTemplate)
			r.With(auth, need(models.MenuNotifications)).Post("/notifications/send", d.Notifications.SendNotification)
			r.With(auth, need(models.MenuNotifications)).Get("/notifications/logs", d.Notifications.ListNotifLogs)
		})
		// Menu: sessions (self-service). Auth saja, tanpa permission menu,
		// supaya user biasa bisa lihat & cabut sesinya sendiri.
		r.With(auth).Get("/sessions/mine", d.Sessions.ListMySessions)
		r.With(auth).Delete("/sessions/{id}", d.Sessions.RevokeSession)
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
