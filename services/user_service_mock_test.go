package services

import (
	"golang-backend/models"
)

type MockUserRepository struct {
	Users     map[int]*models.User
	NextID    int
	GetAllErr error
	CreateErr error
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		Users:  make(map[int]*models.User),
		NextID: 1,
	}
}

func (m *MockUserRepository) GetAll() ([]models.User, error) {
	if m.GetAllErr != nil {
		return nil, m.GetAllErr
	}
	var users []models.User
	for _, u := range m.Users {
		users = append(users, *u)
	}
	return users, nil
}

func (m *MockUserRepository) GetByID(id int) (*models.User, error) {
	user, ok := m.Users[id]
	if !ok {
		return nil, sqlErrNoRows
	}
	return user, nil
}

func (m *MockUserRepository) GetByUsername(username string) (*models.User, error) {
	for _, u := range m.Users {
		if u.Username == username {
			return u, nil
		}
	}
	return nil, sqlErrNoRows
}

func (m *MockUserRepository) Create(user *models.User) error {
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
	}
	user.ID = m.NextID
	m.Users[user.ID] = user
	m.NextID++
	return nil
}

func (m *MockUserRepository) Update(user *models.User) error {
	if _, ok := m.Users[user.ID]; !ok {
		return sqlErrNoRows
	}
	m.Users[user.ID] = user
	return nil
}

func (m *MockUserRepository) Delete(id int) error {
	if _, ok := m.Users[id]; !ok {
		return sqlErrNoRows
	}
	delete(m.Users, id)
	return nil
}

type MockRefreshTokenRepository struct {
	Tokens map[string]*models.RefreshToken
}

func NewMockRefreshTokenRepository() *MockRefreshTokenRepository {
	return &MockRefreshTokenRepository{
		Tokens: make(map[string]*models.RefreshToken),
	}
}

func (m *MockRefreshTokenRepository) Create(token *models.RefreshToken) error {
	m.Tokens[token.Token] = token
	return nil
}

func (m *MockRefreshTokenRepository) GetByToken(token string) (*models.RefreshToken, error) {
	rt, ok := m.Tokens[token]
	if !ok {
		return nil, sqlErrNoRows
	}
	return rt, nil
}

func (m *MockRefreshTokenRepository) DeleteByToken(token string) error {
	delete(m.Tokens, token)
	return nil
}

func (m *MockRefreshTokenRepository) DeleteByUserID(userID int) error {
	for k, v := range m.Tokens {
		if v.UserID == userID {
			delete(m.Tokens, k)
		}
	}
	return nil
}

func (m *MockRefreshTokenRepository) CountByUserID(userID int) (int, error) {
	count := 0
	for _, v := range m.Tokens {
		if v.UserID == userID {
			count++
		}
	}
	return count, nil
}

func (m *MockRefreshTokenRepository) DeleteOldestByUserID(userID int) error {
	var oldest string
	var oldestTime int64 = 1<<63 - 1
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
