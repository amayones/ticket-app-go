package users

import (
	"context"
	"database/sql"
	"sort"
	"sync"

	"golang-backend/models"
)

type MockUserRepository struct {
	mu        sync.Mutex
	Users     map[string]*models.User // keyed by CODE
	GetAllErr error
	CreateErr error
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{Users: make(map[string]*models.User)}
}

func (m *MockUserRepository) List(ctx context.Context, limit, offset int) ([]models.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.GetAllErr != nil {
		return nil, m.GetAllErr
	}
	codes := make([]string, 0, len(m.Users))
	for code := range m.Users {
		codes = append(codes, code)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(codes)))
	var users []models.User
	for i := offset; i < len(codes) && len(users) < limitOrDefault(limit); i++ {
		u := *m.Users[codes[i]]
		u.Password = ""
		users = append(users, u)
	}
	if users == nil {
		users = []models.User{}
	}
	return users, nil
}

func limitOrDefault(limit int) int {
	if limit <= 0 {
		return 50
	}
	return limit
}

func (m *MockUserRepository) GetByCode(ctx context.Context, code string) (*models.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	user, ok := m.Users[code]
	if !ok {
		return nil, sql.ErrNoRows
	}
	cp := *user
	return &cp, nil
}

func (m *MockUserRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, u := range m.Users {
		if u.Username == username {
			cp := *u
			return &cp, nil
		}
	}
	return nil, sql.ErrNoRows
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, u := range m.Users {
		if u.Email == email {
			cp := *u
			return &cp, nil
		}
	}
	return nil, sql.ErrNoRows
}

func (m *MockUserRepository) Create(ctx context.Context, user *models.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.CreateErr != nil {
		return m.CreateErr
	}
	for _, u := range m.Users {
		if u.Username == user.Username {
			return errUsernameConstraint
		}
		if u.Email == user.Email {
			return errEmailConstraint
		}
		if u.Code == user.Code {
			return errCodeConstraint
		}
	}
	cp := *user
	m.Users[user.Code] = &cp
	return nil
}

func (m *MockUserRepository) Update(ctx context.Context, user *models.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.Users[user.Code]; !ok {
		return sql.ErrNoRows
	}
	cp := *user
	m.Users[user.Code] = &cp
	return nil
}

func (m *MockUserRepository) Delete(ctx context.Context, code string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.Users[code]; !ok {
		return sql.ErrNoRows
	}
	delete(m.Users, code)
	return nil
}

// DeleteByRole meniru cascade hapus role: semua user pada role itu dihapus.
func (m *MockUserRepository) DeleteByRole(ctx context.Context, roleCode string) (int, error) {
	n := 0
	for code, u := range m.Users {
		if u.RoleCode == roleCode {
			delete(m.Users, code)
			n++
		}
	}
	return n, nil
}

func (m *MockUserRepository) UpdateRole(ctx context.Context, code, roleCode string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	u, ok := m.Users[code]
	if !ok {
		return sql.ErrNoRows
	}
	u.RoleCode = roleCode
	return nil
}

func (m *MockUserRepository) Count(ctx context.Context) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.Users), nil
}

func (m *MockUserRepository) CountByRole(ctx context.Context, roleCode string) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	n := 0
	for _, u := range m.Users {
		if u.RoleCode == roleCode {
			n++
		}
	}
	return n, nil
}

type MockRefreshTokenRepository struct {
	mu     sync.Mutex
	Tokens map[string]*models.RefreshToken // keyed by hash
}

func NewMockRefreshTokenRepository() *MockRefreshTokenRepository {
	return &MockRefreshTokenRepository{Tokens: make(map[string]*models.RefreshToken)}
}

func (m *MockRefreshTokenRepository) Create(ctx context.Context, token *models.RefreshToken) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := *token
	m.Tokens[token.Token] = &cp
	return nil
}

func (m *MockRefreshTokenRepository) GetByTokenHash(ctx context.Context, hash string) (*models.RefreshToken, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	rt, ok := m.Tokens[hash]
	if !ok {
		return nil, sql.ErrNoRows
	}
	cp := *rt
	return &cp, nil
}

func (m *MockRefreshTokenRepository) DeleteByTokenHash(ctx context.Context, hash string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.Tokens[hash]; !ok {
		return false, nil
	}
	delete(m.Tokens, hash)
	return true, nil
}

func (m *MockRefreshTokenRepository) DeleteByUserCode(ctx context.Context, userCode string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for k, v := range m.Tokens {
		if v.UserCode == userCode {
			delete(m.Tokens, k)
		}
	}
	return nil
}

func (m *MockRefreshTokenRepository) DeleteExpired(ctx context.Context) (int64, error) {
	return 0, nil
}

func (m *MockRefreshTokenRepository) CountByUserCode(ctx context.Context, userCode string) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	count := 0
	for _, v := range m.Tokens {
		if v.UserCode == userCode {
			count++
		}
	}
	return count, nil
}

func (m *MockRefreshTokenRepository) CountActive(ctx context.Context) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.Tokens), nil
}

func (m *MockRefreshTokenRepository) ListByUserCode(ctx context.Context, userCode string) ([]models.Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]models.Session, 0)
	for _, v := range m.Tokens {
		if v.UserCode == userCode {
			out = append(out, models.Session{ID: v.ID, UserCode: v.UserCode, ExpiresAt: v.ExpiresAt, CreatedAt: v.CreatedAt})
		}
	}
	return out, nil
}

func (m *MockRefreshTokenRepository) GetByID(ctx context.Context, id int) (*models.Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, v := range m.Tokens {
		if v.ID == id {
			cp := models.Session{ID: v.ID, UserCode: v.UserCode, ExpiresAt: v.ExpiresAt, CreatedAt: v.CreatedAt}
			return &cp, nil
		}
	}
	return nil, sql.ErrNoRows
}

func (m *MockRefreshTokenRepository) ListAll(ctx context.Context, limit, offset int) ([]models.Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]models.Session, 0)
	for _, v := range m.Tokens {
		out = append(out, models.Session{ID: v.ID, UserCode: v.UserCode, ExpiresAt: v.ExpiresAt, CreatedAt: v.CreatedAt})
	}
	return out, nil
}

func (m *MockRefreshTokenRepository) DeleteByID(ctx context.Context, id int) (bool, error) {
	return false, nil
}

func (m *MockRefreshTokenRepository) DeleteOldestByUserCode(ctx context.Context, userCode string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	var oldest string
	oldestTime := int64(1<<63 - 1)
	for k, v := range m.Tokens {
		if v.UserCode == userCode {
			t := v.CreatedAt.UnixNano()
			if t == 0 {
				t = v.ExpiresAt.UnixNano()
			}
			if t < oldestTime {
				oldestTime = t
				oldest = k
			}
		}
	}
	if oldest != "" {
		delete(m.Tokens, oldest)
	}
	return nil
}

type MockRoleRepository struct {
	Roles []models.Role
}

func NewMockRoleRepository() *MockRoleRepository {
	return &MockRoleRepository{Roles: []models.Role{
		{Code: models.RoleAdmin, Name: "Administrator"},
		{Code: models.RoleUser, Name: "Pengguna"},
	}}
}

func (m *MockRoleRepository) List(ctx context.Context) ([]models.Role, error) {
	return m.Roles, nil
}

func (m *MockRoleRepository) GetByCode(ctx context.Context, code string) (*models.Role, error) {
	for _, r := range m.Roles {
		if r.Code == code {
			cp := r
			return &cp, nil
		}
	}
	return nil, sql.ErrNoRows
}

func (m *MockRoleRepository) RoleExists(ctx context.Context, code string) (bool, error) {
	for _, r := range m.Roles {
		if r.Code == code {
			return true, nil
		}
	}
	return false, nil
}

func (m *MockRoleRepository) Create(ctx context.Context, role *models.Role) error {
	m.Roles = append(m.Roles, *role)
	return nil
}

func (m *MockRoleRepository) Delete(ctx context.Context, code string) error {
	for i, r := range m.Roles {
		if r.Code == code {
			m.Roles = append(m.Roles[:i], m.Roles[i+1:]...)
			return nil
		}
	}
	return sql.ErrNoRows
}

func (m *MockRoleRepository) Count(ctx context.Context) (int, error) {
	return len(m.Roles), nil
}

func (m *MockRoleRepository) ListPermissions(ctx context.Context) ([]models.Permission, error) {
	return []models.Permission{{Code: models.MenuUsers, Name: "Akses menu User Account", Group: "SYSTEM"}}, nil
}

func (m *MockRoleRepository) GetRolePermissions(ctx context.Context, roleCode string) ([]string, error) {
	if roleCode == models.RoleAdmin {
		return []string{models.MenuUsers}, nil
	}
	return []string{}, nil
}

func (m *MockRoleRepository) SetRolePermissions(ctx context.Context, roleCode string, permCodes []string) error {
	return nil
}

func (m *MockRoleRepository) HasPermission(ctx context.Context, roleCode, permCode string) (bool, error) {
	if roleCode == models.RoleAdmin {
		return true, nil
	}
	return false, nil
}

func (m *MockRoleRepository) ListMenus(ctx context.Context) ([]models.Menu, error) {
	return []models.Menu{}, nil
}

func (m *MockRoleRepository) GetMenu(ctx context.Context, code string) (*models.Menu, error) {
	return nil, sql.ErrNoRows
}

func (m *MockRoleRepository) CreateMenu(ctx context.Context, menu *models.Menu) error {
	return nil
}

func (m *MockRoleRepository) UpdateMenu(ctx context.Context, menu *models.Menu) error {
	return nil
}

func (m *MockRoleRepository) DeleteGrants(ctx context.Context, roleCode string) error {
	return nil
}

func (m *MockRoleRepository) DeleteMenu(ctx context.Context, code string) error {
	return sql.ErrNoRows
}

func (m *MockRoleRepository) CountMenuUsage(ctx context.Context, code string) (int, error) {
	return 0, nil
}

func (m *MockRoleRepository) ListChildren(ctx context.Context, code string) ([]models.Menu, error) {
	return []models.Menu{}, nil
}

func (m *MockRoleRepository) QueryMatrix(ctx context.Context, roleFilter string) ([]models.MatrixRow, error) {
	return []models.MatrixRow{}, nil
}

func (m *MockRoleRepository) MyMenusByRole(ctx context.Context, roleCode string) ([]models.MenuEntry, error) {
	return []models.MenuEntry{}, nil
}

func (m *MockRoleRepository) ListModules(ctx context.Context) ([]models.Module, error) {
	return []models.Module{}, nil
}

func (m *MockRoleRepository) GetModule(ctx context.Context, code string) (*models.Module, error) {
	return nil, sql.ErrNoRows
}

func (m *MockRoleRepository) CreateModule(ctx context.Context, mod *models.Module) error {
	return nil
}

func (m *MockRoleRepository) DeleteModule(ctx context.Context, code string) error {
	return sql.ErrNoRows
}

func (m *MockRoleRepository) CountModuleMenus(ctx context.Context, code string) (int, error) {
	return 0, nil
}
