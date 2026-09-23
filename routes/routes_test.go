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
	return []models.UserResponse{{ID: 1, Username: "budi", Email: "budi@x.com"}}, nil
}
func (s *stubService) GetUserByID(ctx context.Context, id int) (*models.User, error) {
	return &models.User{ID: id, Username: "budi", Email: "budi@x.com"}, nil
}
func (s *stubService) CreateUser(ctx context.Context, u, e, p string) (int, error) {
	if u == "" || e == "" || p == "" {
		return 0, services.ErrInputRequired
	}
	return 7, nil
}
func (s *stubService) UpdateUser(ctx context.Context, id int, in models.UpdateUserRequest) error {
	return nil
}
func (s *stubService) DeleteUser(ctx context.Context, id int) error { return nil }
func (s *stubService) Login(ctx context.Context, u, p string) (string, string, error) {
	return "", "", services.ErrInvalidLogin
}
func (s *stubService) RefreshAccessToken(ctx context.Context, t string) (string, string, error) {
	return "", "", services.ErrInvalidRefresh
}
func (s *stubService) Logout(ctx context.Context, t string) error  { return nil }
func (s *stubService) LogoutAll(ctx context.Context, id int) error { return nil }
func (s *stubService) CleanupExpiredTokens(ctx context.Context) (int64, error) {
	return 0, nil
}

func auditRouter(t *testing.T) *chi.Mux {
	t.Helper()
	h := handlers.NewUserHandler(&stubService{})
	cfg := DefaultRouteConfig(auditSecret)
	cfg.LoginLimit = 2
	return SetupRoutesWithConfig(h, cfg)
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
	tok, err := utils.GenerateAccessToken(auditSecret, 1, "budi")
	if err != nil {
		t.Fatal(err)
	}
	if rec := doReq(t, r, "GET", "/api/users", "", tok); rec.Code != 200 || !strings.Contains(rec.Body.String(), "budi") {
		t.Fatalf("protected with token must 200, got %d (%s)", rec.Code, rec.Body.String())
	}
	if rec := doReq(t, r, "GET", "/api/users/1", "", tok); rec.Code != 200 {
		t.Fatalf("detail: got %d (%s)", rec.Code, rec.Body.String())
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
