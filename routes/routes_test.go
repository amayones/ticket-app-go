package routes

// Wiring test: verifies route↔middleware↔handler integration end-to-end
// with stub services (no database needed). One stub per menu.

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	faudit "golang-backend/features/audit"
	fnotif "golang-backend/features/notifications"
	froles "golang-backend/features/roles"
	fsecurity "golang-backend/features/security"
	fsessions "golang-backend/features/sessions"
	fsyslog "golang-backend/features/syslog"
	fusers "golang-backend/features/users"
	"golang-backend/models"
	"golang-backend/services"
	"golang-backend/utils"
)

const auditSecret = "audit-secret-32-chars-minimum-xxxx"

// --- Menu: users ---

type stubUsers struct{}

func (s *stubUsers) ListUsers(ctx context.Context, limit, offset int) ([]models.UserResponse, error) {
	return []models.UserResponse{{Code: "USR-000001", Username: "budi", Email: "budi@x.com", RoleCode: models.RoleUser}}, nil
}
func (s *stubUsers) GetUserByCode(ctx context.Context, code string) (*models.User, error) {
	return &models.User{Code: code, Username: "budi", Email: "budi@x.com", RoleCode: models.RoleUser}, nil
}
func (s *stubUsers) CreateUser(ctx context.Context, u, e, p, role string) (string, error) {
	if u == "" || e == "" || p == "" {
		return "", services.ErrInputRequired
	}
	return "USR-000007", nil
}
func (s *stubUsers) UpdateUser(ctx context.Context, code string, in models.UpdateUserRequest) error {
	return nil
}
func (s *stubUsers) DeleteUser(ctx context.Context, code string) error { return nil }
func (s *stubUsers) Login(ctx context.Context, u, p string) (string, string, error) {
	return "", "", services.ErrInvalidLogin
}
func (s *stubUsers) RefreshAccessToken(ctx context.Context, t string) (string, string, error) {
	return "", "", services.ErrInvalidRefresh
}
func (s *stubUsers) Logout(ctx context.Context, t string) error { return nil }
func (s *stubUsers) LogoutAll(ctx context.Context, code string) error {
	return nil
}
func (s *stubUsers) CountUsers(ctx context.Context) (int, error) { return 1, nil }

// --- Menu: roles ---

type stubRoles struct{}

func (s *stubRoles) CheckPermission(ctx context.Context, userCode, perm string) error {
	if userCode == "USR-ADMIN" {
		return nil
	}
	return services.ErrForbidden
}
func (s *stubRoles) ListRoles(ctx context.Context) ([]models.Role, error) {
	return []models.Role{{Code: models.RoleUser, Name: "Pengguna"}}, nil
}
func (s *stubRoles) CreateRole(ctx context.Context, code, name string) (*models.Role, error) {
	return &models.Role{Code: code, Name: name}, nil
}
func (s *stubRoles) DeleteRole(ctx context.Context, code string) (int, error) { return 0, nil }
func (s *stubRoles) GetRoleDetail(ctx context.Context, code string) (*models.RoleDetail, error) {
	return &models.RoleDetail{Role: models.Role{Code: code, Name: "Test"}, Permissions: []string{}}, nil
}
func (s *stubRoles) ListPermissions(ctx context.Context) ([]models.Permission, error) {
	return []models.Permission{{Code: models.MenuUsers, Name: "Akses menu User Account", Group: "SYSTEM"}}, nil
}
func (s *stubRoles) GetRolePermissions(ctx context.Context, roleCode string) ([]string, error) {
	return []string{models.MenuUsers}, nil
}
func (s *stubRoles) SetRolePermissions(ctx context.Context, role string, perms []string) error {
	return nil
}
func (s *stubRoles) UpdateUserRole(ctx context.Context, userCode, roleCode string) error {
	return nil
}
func (s *stubRoles) CountRoles(ctx context.Context) (int, error) { return 2, nil }
func (s *stubRoles) ListMenus(ctx context.Context) ([]models.Menu, error) {
	return []models.Menu{{Code: models.MenuUsers, Module: "SYSTEM", Label: "User Account", Mcontrol: "users", SortOrder: 1}}, nil
}
func (s *stubRoles) GetMenu(ctx context.Context, code string) (*models.Menu, error) {
	return &models.Menu{Code: code, Module: "SYSTEM", Label: "Test"}, nil
}
func (s *stubRoles) CreateMenu(ctx context.Context, in models.MenuInput) (*models.Menu, error) {
	return &models.Menu{Code: "MENU_TEST", Module: "SYSTEM", Label: "Test"}, nil
}
func (s *stubRoles) UpdateMenu(ctx context.Context, code string, in models.MenuUpdateInput) (*models.Menu, error) {
	return &models.Menu{Code: code, Module: in.Module, Label: in.Label}, nil
}
func (s *stubRoles) DeleteMenu(ctx context.Context, code string) error { return nil }
func (s *stubRoles) GetMatrix(ctx context.Context, role string) ([]models.MatrixRow, error) {
	return []models.MatrixRow{}, nil
}
func (s *stubRoles) MyMenus(ctx context.Context, userCode string) ([]models.MenuEntry, error) {
	return []models.MenuEntry{}, nil
}
func (s *stubRoles) ListModules(ctx context.Context) ([]models.Module, error) {
	return []models.Module{{Code: "SYSTEM", Label: "System", SortOrder: 1}}, nil
}
func (s *stubRoles) CreateModule(ctx context.Context, code, label string, sort int) (*models.Module, error) {
	return &models.Module{Code: code, Label: label, SortOrder: sort}, nil
}
func (s *stubRoles) DeleteModule(ctx context.Context, code string) error { return nil }

// --- Menu: sessions ---

type stubSessions struct{}

func (s *stubSessions) ListSessions(ctx context.Context, userCode string) ([]models.Session, error) {
	return []models.Session{}, nil
}
func (s *stubSessions) ListAllSessions(ctx context.Context, limit, offset int) ([]models.Session, error) {
	return []models.Session{}, nil
}
func (s *stubSessions) RevokeSession(ctx context.Context, caller string, id int, all bool) error {
	return nil
}
func (s *stubSessions) CountActiveSessions(ctx context.Context) (int, error) {
	return 0, nil
}
func (s *stubSessions) CleanupExpiredTokens(ctx context.Context) (int64, error) {
	return 0, nil
}

// --- Menu: audit ---

type stubAudit struct{}

func (s *stubAudit) Log(ctx context.Context, actor, action, entity, entityCode, detail, ip string) error {
	return nil
}
func (s *stubAudit) List(ctx context.Context, f models.AuditFilter) ([]models.AuditLog, error) {
	return []models.AuditLog{}, nil
}
func (s *stubAudit) CountSince(ctx context.Context, hours int) (int, error) { return 0, nil }

// --- Menu: security ---

type stubSecurity struct{}

func (s *stubSecurity) Summary(ctx context.Context) (models.SecuritySummary, error) {
	return models.SecuritySummary{}, nil
}

// --- Menu: syslog ---

type stubSyslog struct{}

func (s *stubSyslog) Error(ctx context.Context, source, msg string) error { return nil }
func (s *stubSyslog) List(ctx context.Context, f models.SyslogFilter) ([]models.SysLog, error) {
	return []models.SysLog{}, nil
}
func (s *stubSyslog) CountSince(ctx context.Context, hours int, level string) (int, error) {
	return 0, nil
}
func (s *stubSyslog) Prune(ctx context.Context, days int) (int64, error) { return 0, nil }

// --- Menu: notifications ---

type stubNotif struct{}

func (s *stubNotif) ListTemplates(ctx context.Context, active bool) ([]models.NotifTemplate, error) {
	return []models.NotifTemplate{}, nil
}
func (s *stubNotif) GetTemplate(ctx context.Context, code string) (*models.NotifTemplate, error) {
	return &models.NotifTemplate{Code: code, Name: "Test"}, nil
}
func (s *stubNotif) CreateTemplate(ctx context.Context, name, channel, subject, body string, active bool) (*models.NotifTemplate, error) {
	return &models.NotifTemplate{Code: "NTM-TEST", Name: name}, nil
}
func (s *stubNotif) UpdateTemplate(ctx context.Context, code, name, channel, subject, body string, active bool) error {
	return nil
}
func (s *stubNotif) DeleteTemplate(ctx context.Context, code string) error { return nil }
func (s *stubNotif) Send(ctx context.Context, req models.NotifSendRequest) (*models.NotifLog, error) {
	return &models.NotifLog{Code: "NTF-TEST", Status: models.NotifSent}, nil
}
func (s *stubNotif) ListLogs(ctx context.Context, limit, offset int) ([]models.NotifLog, error) {
	return []models.NotifLog{}, nil
}
func (s *stubNotif) CountSentSince(ctx context.Context, hours int) (int, error) {
	return 0, nil
}

func auditRouter(t *testing.T) *chi.Mux {
	t.Helper()
	usersSvc := &stubUsers{}
	rolesSvc := &stubRoles{}
	auditSvc := &stubAudit{}
	syslogSvc := &stubSyslog{}
	deps := Deps{
		Users:         fusers.NewHandler(usersSvc, rolesSvc, auditSvc, syslogSvc),
		Roles:         froles.NewHandler(rolesSvc, auditSvc),
		Sessions:      fsessions.NewHandler(&stubSessions{}, rolesSvc, auditSvc),
		Audit:         faudit.NewHandler(auditSvc),
		Security:      fsecurity.NewHandler(&stubSecurity{}),
		Syslog:        fsyslog.NewHandler(syslogSvc, auditSvc),
		Notifications: fnotif.NewHandler(&stubNotif{}, auditSvc),
	}
	cfg := DefaultRouteConfig(auditSecret)
	cfg.LoginLimit = 2
	return SetupRoutesWithConfig(deps, cfg)
}

func doReq(t *testing.T, r http.Handler, method, path, body, token string) *httptest.ResponseRecorder {
	t.Helper()
	var reader *strings.Reader
	if body == "" {
		reader = strings.NewReader("")
	} else {
		reader = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func TestAuditWiring(t *testing.T) {
	r := auditRouter(t)

	if rec := doReq(t, r, "GET", "/healthz", "", ""); rec.Code != 200 {
		t.Fatalf("healthz: got %d (%s)", rec.Code, rec.Body.String())
	}
	if rec := doReq(t, r, "POST", "/api/users", `{"username":"budi","email":"b@x.com","password":"password123"}`, ""); rec.Code != 401 {
		t.Fatalf("create user without token must 401, got %d (%s)", rec.Code, rec.Body.String())
	}
	if rec := doReq(t, r, "POST", "/api/login", `{"username":"b","password":"p"}`, ""); rec.Code != 401 || !strings.Contains(rec.Body.String(), "error") {
		t.Fatalf("login bad creds must be 401 JSON, got %d (%s)", rec.Code, rec.Body.String())
	}
	if rec := doReq(t, r, "GET", "/api/users", "", ""); rec.Code != 401 {
		t.Fatalf("protected without token must 401, got %d", rec.Code)
	}
	tok, err := utils.GenerateAccessToken(auditSecret, "USR-ADMIN", "budi", models.RoleAdmin)
	if err != nil {
		t.Fatal(err)
	}
	if rec := doReq(t, r, "GET", "/api/users", "", tok); rec.Code != 200 || !strings.Contains(rec.Body.String(), "budi") {
		t.Fatalf("protected with token must 200, got %d (%s)", rec.Code, rec.Body.String())
	}
	if rec := doReq(t, r, "GET", "/api/users/USR-000001", "", tok); rec.Code != 200 {
		t.Fatalf("detail: got %d (%s)", rec.Code, rec.Body.String())
	}
	if rec := doReq(t, r, "GET", "/api/roles", "", tok); rec.Code != 200 {
		t.Fatalf("roles: got %d (%s)", rec.Code, rec.Body.String())
	}
	if rec := doReq(t, r, "GET", "/api/tidak-ada", "", ""); rec.Code != 404 || !strings.Contains(rec.Body.String(), "error") {
		t.Fatalf("unknown api must be 404 JSON, got %d (%s)", rec.Code, rec.Body.String())
	}
	if rec := doReq(t, r, "PUT", "/api/login", "", ""); rec.Code != 405 {
		t.Fatalf("wrong method must 405, got %d", rec.Code)
	}
	// Login limiter is 2/min and 1 POST used above: 2nd passes, 3rd must 429.
	if rec := doReq(t, r, "POST", "/api/login", `{"username":"b","password":"p"}`, ""); rec.Code != 401 {
		t.Fatalf("2nd login must pass limiter (401 from stub), got %d", rec.Code)
	}
	rec := doReq(t, r, "POST", "/api/login", `{"username":"b","password":"p"}`, "")
	if rec.Code != 429 {
		t.Fatalf("rate limit must 429, got %d (%s)", rec.Code, rec.Body.String())
	}
	if rec.Header().Get("Retry-After") == "" {
		t.Fatal("429 must carry Retry-After header")
	}
}

func TestAdminRBAC(t *testing.T) {
	r := auditRouter(t)
	adminTok, err := utils.GenerateAccessToken(auditSecret, "USR-ADMIN", "admin", models.RoleAdmin)
	if err != nil {
		t.Fatal(err)
	}
	userTok, err := utils.GenerateAccessToken(auditSecret, "USR-000001", "budi", models.RoleUser)
	if err != nil {
		t.Fatal(err)
	}
	// Sesi self-service (/api/sessions/*) hanya butuh auth, tanpa permission.
	if rec := doReq(t, r, "GET", "/api/sessions/mine", "", userTok); rec.Code != 200 {
		t.Fatalf("user self sessions must be 200, got %d (%s)", rec.Code, rec.Body.String())
	}
	if rec := doReq(t, r, "GET", "/api/sessions/mine", "", ""); rec.Code != 401 {
		t.Fatalf("self sessions without token must be 401, got %d", rec.Code)
	}
	// Lihat semua sesi tetap butuh permission MENU_SESSIONS.
	if rec := doReq(t, r, "GET", "/api/admin/sessions/all", "", adminTok); rec.Code != 200 {
		t.Fatalf("admin all sessions must be 200, got %d (%s)", rec.Code, rec.Body.String())
	}
	if rec := doReq(t, r, "GET", "/api/admin/sessions/all", "", userTok); rec.Code != 403 {
		t.Fatalf("user all sessions must be 403, got %d", rec.Code)
	}
	adminPaths := []struct{ method, path, body string }{
		{"GET", "/api/admin/permissions", ""},
		{"GET", "/api/admin/roles/USER", ""},
		{"GET", "/api/admin/sessions/all", ""},
		{"GET", "/api/admin/audit", ""},
		{"GET", "/api/admin/security/summary", ""},
		{"GET", "/api/admin/syslogs", ""},
		{"GET", "/api/admin/notifications/templates", ""},
		{"GET", "/api/admin/notifications/logs", ""},
		{"POST", "/api/users", `{"username":"budi","email":"b@x.com","password":"password123"}`},
		{"POST", "/api/admin/roles", `{"code":"EDITOR","name":"Editor"}`},
		{"PUT", "/api/admin/roles/EDITOR/permissions", `{"permissions":["MENU_USERS"]}`},
		{"GET", "/api/admin/menus", ""},
		{"GET", "/api/admin/matrix", ""},
		{"POST", "/api/admin/menus", `{"code":"MENU_TEST","name":"Test","module":"SYSTEM","label":"Test"}`},
		{"PUT", "/api/admin/menus/MENU_TEST", `{"module":"SYSTEM","label":"Test Baru","sort_order":5}`},
		{"DELETE", "/api/admin/menus/MENU_TEST", ""},
		{"GET", "/api/admin/modules", ""},
		{"POST", "/api/admin/modules", `{"code":"REPORT","label":"Report","sort_order":10}`},
		{"DELETE", "/api/admin/modules/REPORT", ""},
		{"POST", "/api/admin/notifications/send", `{"template_code":"NTM-1","recipient":"a@b.c"}`},
	}
	for _, p := range adminPaths {
		if rec := doReq(t, r, p.method, p.path, p.body, adminTok); rec.Code >= 400 {
			t.Fatalf("admin %s %s must succeed, got %d (%s)", p.method, p.path, rec.Code, rec.Body.String())
		}
		if rec := doReq(t, r, p.method, p.path, p.body, userTok); rec.Code != 403 {
			t.Fatalf("user %s %s must be 403, got %d", p.method, p.path, rec.Code)
		}
	}
	// Self-service menus: auth saja, tanpa permission menu.
	if rec := doReq(t, r, "GET", "/api/menus/mine", "", userTok); rec.Code != 200 {
		t.Fatalf("menus/mine for user must be 200, got %d (%s)", rec.Code, rec.Body.String())
	}
	if rec := doReq(t, r, "GET", "/api/menus/mine", "", ""); rec.Code != 401 {
		t.Fatalf("menus/mine without token must be 401, got %d", rec.Code)
	}
	if rec := doReq(t, r, "PUT", "/api/users/USR-000001/role", `{"role_code":"ADMIN"}`, userTok); rec.Code != 403 {
		t.Fatalf("role assign by non-admin must be 403, got %d", rec.Code)
	}
	if rec := doReq(t, r, "PUT", "/api/users/USR-000001/role", `{"role_code":"ADMIN"}`, adminTok); rec.Code != 200 {
		t.Fatalf("role assign by admin must be 200, got %d (%s)", rec.Code, rec.Body.String())
	}
	if rec := doReq(t, r, "GET", "/api/admin/audit", "", ""); rec.Code != 401 {
		t.Fatalf("admin audit without token must be 401, got %d", rec.Code)
	}
}
