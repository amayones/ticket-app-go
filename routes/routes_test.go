package routes

// Wiring test: verifies route↔middleware↔handler integration end-to-end
// with a stub service (no database needed).

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	"golang-backend/handlers"
	"golang-backend/models"
	"golang-backend/services"
	"golang-backend/utils"
)

const auditSecret = "audit-secret-32-chars-minimum-xxxx"

type stubService struct{}

func (s *stubService) ListUsers(ctx context.Context, limit, offset int) ([]models.UserResponse, error) {
	return []models.UserResponse{{Code: "USR-000001", Username: "budi", Email: "budi@x.com", RoleCode: models.RoleUser}}, nil
}
func (s *stubService) GetUserByCode(ctx context.Context, code string) (*models.User, error) {
	return &models.User{Code: code, Username: "budi", Email: "budi@x.com", RoleCode: models.RoleUser}, nil
}
func (s *stubService) CreateUser(ctx context.Context, u, e, p string) (string, error) {
	if u == "" || e == "" || p == "" {
		return "", services.ErrInputRequired
	}
	return "USR-000007", nil
}
func (s *stubService) UpdateUser(ctx context.Context, code string, in models.UpdateUserRequest) error {
	return nil
}
func (s *stubService) DeleteUser(ctx context.Context, code string) error { return nil }
func (s *stubService) Login(ctx context.Context, u, p string) (string, string, error) {
	return "", "", services.ErrInvalidLogin
}
func (s *stubService) RefreshAccessToken(ctx context.Context, t string) (string, string, error) {
	return "", "", services.ErrInvalidRefresh
}
func (s *stubService) Logout(ctx context.Context, t string) error { return nil }
func (s *stubService) LogoutAll(ctx context.Context, code string) error {
	return nil
}
func (s *stubService) ListRoles(ctx context.Context) ([]models.Role, error) {
	return []models.Role{{Code: models.RoleUser, Name: "Pengguna"}}, nil
}
func (s *stubService) CheckPermission(ctx context.Context, userCode, perm string) error {
	if userCode == "USR-ADMIN" {
		return nil
	}
	return services.ErrForbidden
}
func (s *stubService) CreateRole(ctx context.Context, code, name string) (*models.Role, error) {
	return &models.Role{Code: code, Name: name}, nil
}
func (s *stubService) DeleteRole(ctx context.Context, code string) error { return nil }
func (s *stubService) GetRoleDetail(ctx context.Context, code string) (*models.RoleDetail, error) {
	return &models.RoleDetail{Role: models.Role{Code: code, Name: "Test"}, Permissions: []string{}}, nil
}
func (s *stubService) ListPermissions(ctx context.Context) ([]models.Permission, error) {
	return []models.Permission{{Code: models.PermUserRead, Name: "Lihat user", Group: "USER"}}, nil
}
func (s *stubService) SetRolePermissions(ctx context.Context, role string, perms []string) error {
	return nil
}
func (s *stubService) UpdateUserRole(ctx context.Context, userCode, roleCode string) error {
	return nil
}
func (s *stubService) ListSessions(ctx context.Context, userCode string) ([]models.Session, error) {
	return []models.Session{}, nil
}
func (s *stubService) ListAllSessions(ctx context.Context, limit, offset int) ([]models.Session, error) {
	return []models.Session{}, nil
}
func (s *stubService) RevokeSession(ctx context.Context, caller string, id int, all bool) error {
	return nil
}
func (s *stubService) CountUsers(ctx context.Context) (int, error)   { return 1, nil }
func (s *stubService) CountRoles(ctx context.Context) (int, error)   { return 2, nil }
func (s *stubService) CountActiveSessions(ctx context.Context) (int, error) {
	return 0, nil
}
func (s *stubService) CleanupExpiredTokens(ctx context.Context) (int64, error) {
	return 0, nil
}

type stubAudit struct{}

func (s *stubAudit) Log(ctx context.Context, actor, action, entity, entityCode, detail, ip string) error {
	return nil
}
func (s *stubAudit) List(ctx context.Context, f models.AuditFilter) ([]models.AuditLog, error) {
	return []models.AuditLog{}, nil
}
func (s *stubAudit) CountSince(ctx context.Context, hours int) (int, error) { return 0, nil }

type stubSys struct{}

func (s *stubSys) Error(ctx context.Context, source, msg string) error { return nil }
func (s *stubSys) List(ctx context.Context, f models.SyslogFilter) ([]models.SysLog, error) {
	return []models.SysLog{}, nil
}
func (s *stubSys) CountSince(ctx context.Context, hours int, level string) (int, error) {
	return 0, nil
}
func (s *stubSys) Prune(ctx context.Context, days int) (int64, error) { return 0, nil }

type stubNotif struct{}

func (s *stubNotif) ListTemplates(ctx context.Context, active bool) ([]models.NotifTemplate, error) {
	return []models.NotifTemplate{}, nil
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

func auditRouter(t *testing.T) *chi.Mux {
	t.Helper()
	svc := &stubService{}
	h := handlers.NewUserHandler(svc, &stubAudit{}, &stubSys{})
	admin := handlers.NewAdminHandler(svc, &stubAudit{}, &stubSys{}, &stubNotif{})
	cfg := DefaultRouteConfig(auditSecret)
	cfg.LoginLimit = 2
	return SetupRoutesWithConfig(h, admin, cfg)
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
	if rec := doReq(t, r, "POST", "/api/users", `{"username":"x"}`, ""); rec.Code != 400 {
		t.Fatalf("register unknown field must 400, got %d (%s)", rec.Code, rec.Body.String())
	}
	if rec := doReq(t, r, "POST", "/api/users", `{"username":"budi","email":"b@x.com","password":"password123"}`, ""); rec.Code != 201 {
		t.Fatalf("register: got %d (%s)", rec.Code, rec.Body.String())
	}
	if rec := doReq(t, r, "POST", "/api/login", `{"username":"b","password":"p"}`, ""); rec.Code != 401 || !strings.Contains(rec.Body.String(), "error") {
		t.Fatalf("login bad creds must be 401 JSON, got %d (%s)", rec.Code, rec.Body.String())
	}
	if rec := doReq(t, r, "GET", "/api/users", "", ""); rec.Code != 401 {
		t.Fatalf("protected without token must 401, got %d", rec.Code)
	}
	tok, err := utils.GenerateAccessToken(auditSecret, "USR-000001", "budi", models.RoleUser)
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
	// Own sessions: both roles may read.
	for _, tok := range []string{adminTok, userTok} {
		if rec := doReq(t, r, "GET", "/api/admin/sessions", "", tok); rec.Code != 200 {
			t.Fatalf("own sessions must be 200, got %d (%s)", rec.Code, rec.Body.String())
		}
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
		{"POST", "/api/admin/roles", `{"code":"EDITOR","name":"Editor"}`},
		{"PUT", "/api/admin/roles/EDITOR/permissions", `{"permissions":["USER_READ"]}`},
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
