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
	Users     map[int]*models.User
	NextID    int
	GetAllErr error
	CreateErr error
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{Users: make(map[int]*models.User), NextID: 1}
}

func (m *MockUserRepository) List(ctx context.Context, limit, offset int) ([]models.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.GetAllErr != nil {
		return nil, m.GetAllErr
	}
	ids := make([]int, 0, len(m.Users))
	for id := range m.Users {
		ids = append(ids, id)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(ids)))
	var users []models.User
	for i := offset; i < len(ids) && len(users) < limitOrDefault(limit); i++ {
		u := *m.Users[ids[i]]
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

func (m *MockUserRepository) GetByID(ctx context.Context, id int) (*models.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	user, ok := m.Users[id]
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

func (m *MockUserRepository) Create(ctx context.Context, user *models.User) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.CreateErr != nil {
		return 0, m.CreateErr
	}
	for _, u := range m.Users {
		if u.Username == user.Username {
			return 0, errUsernameConstraint
		}
		if u.Email == user.Email {
			return 0, errEmailConstraint
		}
	}
	user.ID = m.NextID
	cp := *user
	m.Users[user.ID] = &cp
	m.NextID++
	return user.ID, nil
}

func (m *MockUserRepository) Update(ctx context.Context, user *models.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.Users[user.ID]; !ok {
		return sql.ErrNoRows
	}
	cp := *user
	m.Users[user.ID] = &cp
	return nil
}

func (m *MockUserRepository) Delete(ctx context.Context, id int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.Users[id]; !ok {
		return sql.ErrNoRows
	}
	delete(m.Users, id)
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

func (m *MockRefreshTokenRepository) DeleteByUserID(ctx context.Context, userID int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for k, v := range m.Tokens {
		if v.UserID == userID {
			delete(m.Tokens, k)
		}
	}
	return nil
}

func (m *MockRefreshTokenRepository) DeleteExpired(ctx context.Context) (int64, error) {
	return 0, nil
}

func (m *MockRefreshTokenRepository) CountByUserID(ctx context.Context, userID int) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	count := 0
	for _, v := range m.Tokens {
		if v.UserID == userID {
			count++
		}
	}
	return count, nil
}

func (m *MockRefreshTokenRepository) DeleteOldestByUserID(ctx context.Context, userID int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	var oldest string
	oldestTime := int64(1<<63 - 1)
	for k, v := range m.Tokens {
		if v.UserID == userID {
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
