package services

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
