package routes

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	faudit "golang-backend/features/audit"
	fdiscovery "golang-backend/features/discovery"
	fevent "golang-backend/features/event"
	fnotif "golang-backend/features/notifications"
	forganizer "golang-backend/features/organizer"
	froles "golang-backend/features/roles"
	fsecurity "golang-backend/features/security"
	fseller "golang-backend/features/seller"
	fsessions "golang-backend/features/sessions"
	fsyslog "golang-backend/features/syslog"
	fticketing "golang-backend/features/ticketing"
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
	Organizer     *forganizer.Handler
	Seller        *fseller.Handler
	Discovery     *fdiscovery.Handler
	Event         *fevent.Handler
	Ticketing     *fticketing.Handler
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
			r.With(auth, need(models.MenuRoles)).Put("/menus/{code}", d.Roles.UpdateMenu)
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
		// Menu TICKETING: buyer-scoped (tiket/refund milik sendiri) dan
		// officer-scoped (scan hanya event yang ditugaskan). Tanpa permission
		// matriks (aturan kepemilikan/penugasan, bukan hak menu).
		r.Route("/ticketing", func(r chi.Router) {
			r.With(auth).Post("/checkout", d.Ticketing.Checkout)
			r.With(auth).Post("/orders/{code}/pay", d.Ticketing.PayOrder)
			r.With(auth).Post("/orders/{code}/cancel", d.Ticketing.CancelOrder)
			r.With(auth).Get("/tickets", d.Ticketing.MyTickets)
			r.With(auth).Get("/tickets/{code}", d.Ticketing.TicketDetail)
			// QR tanpa middleware auth: <img> tak bisa pasang header, handler
			// menerima Authorization atau ?access_token= (cek di handler).
			r.Get("/tickets/{code}/qr", d.Ticketing.TicketQR)
			r.With(auth).Post("/refunds", d.Ticketing.RequestRefund)
			r.With(auth).Get("/refunds", d.Ticketing.BuyerRefunds)
			r.With(auth).Post("/scan", d.Ticketing.ScanTicket)
			r.With(auth).Get("/scan/events", d.Ticketing.ScanEvents)
			r.With(auth).Get("/scan/summary", d.Ticketing.ScanSummary)
			r.With(auth).Get("/scan/history", d.Ticketing.ScanHistory)
			r.With(auth).Get("/scan/export", d.Ticketing.ScanExport)
		})
		// Menu EVENT: owner-scoped (organizer hanya garap event miliknya).
		// Tanpa permission matriks (aturan kepemilikan, bukan hak menu).
		r.Route("/events", func(r chi.Router) {
			r.With(auth).Get("/mine", d.Event.MineList)
			r.With(auth).Post("/", d.Event.CreateEvent)
			r.With(auth).Post("/upload-poster", d.Event.UploadPoster)
			r.With(auth).Get("/{code}", d.Event.EventDetail)
			r.With(auth).Put("/{code}", d.Event.UpdateEvent)
			r.With(auth).Delete("/{code}", d.Event.DeleteEvent)
			r.With(auth).Post("/{code}/duplicate", d.Event.DuplicateEvent)
			r.With(auth).Post("/{code}/publish", d.Event.PublishEvent)
			r.With(auth).Post("/{code}/unpublish", d.Event.UnpublishEvent)
			r.With(auth).Post("/{code}/cancel", d.Event.CancelEvent)
			r.With(auth).Get("/{code}/ticket-types", d.Event.TypeList)
			r.With(auth).Post("/{code}/ticket-types", d.Event.CreateType)
			r.With(auth).Put("/{code}/ticket-types/order", d.Event.ReorderTypes)
			r.With(auth).Put("/{code}/ticket-types/{tcode}", d.Event.UpdateType)
			r.With(auth).Delete("/{code}/ticket-types/{tcode}", d.Event.DeleteType)
			r.With(auth).Post("/{code}/ticket-types/{tcode}/activate", func(w http.ResponseWriter, req *http.Request) {
				d.Event.SetTypeActive(w, req, true)
			})
			r.With(auth).Post("/{code}/ticket-types/{tcode}/deactivate", func(w http.ResponseWriter, req *http.Request) {
				d.Event.SetTypeActive(w, req, false)
			})
			r.With(auth).Get("/{code}/attendees", d.Event.Attendees)
			r.With(auth).Get("/{code}/attendees/export", d.Event.AttendeeExport)
		})
		// Menu DISCOVERY: publik tanpa auth (home, berita, event, promo).
		// Hanya baris layak-publik yang keluar (publishable filter di service).
		r.Route("/discovery", func(r chi.Router) {
			r.Get("/home", d.Discovery.Home)
			r.Get("/categories", d.Discovery.Categories)
			r.Get("/cities", d.Discovery.Cities)
			r.Get("/news/categories", d.Discovery.NewsCategories)
			r.Get("/news", d.Discovery.News)
			r.Get("/news/{slug}", d.Discovery.NewsDetail)
			r.Get("/events", d.Discovery.Events)
			r.Get("/events/{code}", d.Discovery.EventDetail)
			r.Get("/promotions", d.Discovery.Promotions)
			r.Post("/ads/{code}/impression", d.Discovery.AdImpression)
			r.Post("/ads/{code}/click", d.Discovery.AdClick)
		})
		// Menu dashboard (modul DASHBOARD): permission matriks standar
		// (MENU_ORGANIZER / MENU_SELLER), sama seperti menu lain. Admin
		// memberi akses lewat Role & Permission; tanpa grant = 403.
		org := need(models.MenuOrganizer)
		seller := need(models.MenuSeller)
		r.Route("/dashboard", func(r chi.Router) {
			r.Route("/organizer", func(r chi.Router) {
				r.With(auth, org).Get("/summary", d.Organizer.Summary)
				r.With(auth, org).Get("/orders", d.Organizer.Orders)
				r.With(auth, org).Get("/refunds", d.Organizer.Refunds)
				r.With(auth, org).Post("/refunds/{code}/approve", func(w http.ResponseWriter, req *http.Request) {
					d.Organizer.DecideRefund(w, req, true)
				})
				r.With(auth, org).Post("/refunds/{code}/reject", func(w http.ResponseWriter, req *http.Request) {
					d.Organizer.DecideRefund(w, req, false)
				})
				r.With(auth, org).Get("/checkin-live", d.Organizer.CheckinLive)
				r.With(auth, org).Get("/export", d.Organizer.Export)
			})
			r.Route("/seller", func(r chi.Router) {
				r.With(auth, seller).Get("/summary", d.Seller.Summary)
				r.With(auth, seller).Get("/orders", d.Seller.Orders)
				r.With(auth, seller).Get("/orders/actionable", d.Seller.Actionable)
				r.With(auth, seller).Get("/products/low-stock", d.Seller.LowStock)
				r.With(auth, seller).Get("/complaints", d.Seller.Complaints)
				r.With(auth, seller).Get("/export", d.Seller.Export)
			})
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
