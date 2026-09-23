package services

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"golang-backend/models"
	"golang-backend/repositories"
	"golang-backend/utils"
)

var (
	ErrUsernameRequired = errors.New("username, email, and password are required")
	ErrUsernameTooShort = errors.New("username must be at least 3 characters")
	ErrInvalidEmail     = errors.New("invalid email format")
	ErrPasswordTooShort = errors.New("password must be at least 8 characters")
	ErrUsernameTaken    = errors.New("username already taken")
	ErrEmailTaken       = errors.New("email already registered")
	ErrInvalidLogin     = errors.New("invalid username or password")
	ErrUserNotFound     = errors.New("user not found")
	ErrInvalidRefresh   = errors.New("invalid or expired refresh token")
	ErrForbidden        = errors.New("forbidden")
)

const maxRefreshTokensPerUser = 5

type UserService struct {
	Repository        repositories.UserRepositoryInterface
	RefreshRepository repositories.RefreshTokenRepositoryInterface
}

func NewUserService(repository repositories.UserRepositoryInterface, refreshRepository repositories.RefreshTokenRepositoryInterface) *UserService {
	return &UserService{
		Repository:        repository,
		RefreshRepository: refreshRepository,
	}
}

func (s *UserService) validateUserInput(username, email, password string) error {
	if strings.TrimSpace(username) == "" ||
		strings.TrimSpace(email) == "" ||
		strings.TrimSpace(password) == "" {
		return ErrUsernameRequired
	}
	if !utils.IsValidUsername(username) {
		return ErrUsernameTooShort
	}
	if !utils.IsValidEmail(email) {
		return ErrInvalidEmail
	}
	if !utils.IsValidPassword(password) {
		return ErrPasswordTooShort
	}
	return nil
}

func (s *UserService) mapDBError(err error) error {
	if strings.Contains(err.Error(), "UQ_users_username") {
		return ErrUsernameTaken
	}
	if strings.Contains(err.Error(), "UQ_users_email") {
		return ErrEmailTaken
	}
	return err
}

func (s *UserService) GetAllUsers() ([]models.User, error) {
	return s.Repository.GetAll()
}

func (s *UserService) GetUserByID(id int) (*models.User, error) {
	user, err := s.Repository.GetByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

func (s *UserService) CreateUser(username, email, password string) error {
	if err := s.validateUserInput(username, email, password); err != nil {
		return err
	}
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return err
	}
	user := models.User{
		Username: username,
		Email:    email,
		Password: hashedPassword,
	}
	if err := s.Repository.Create(&user); err != nil {
		return s.mapDBError(err)
	}
	return nil
}

func (s *UserService) UpdateUser(id int, username, email, password string) error {
	if err := s.validateUserInput(username, email, password); err != nil {
		return err
	}
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return err
	}
	user := models.User{
		ID:       id,
		Username: username,
		Email:    email,
		Password: hashedPassword,
	}
	if err := s.Repository.Update(&user); err != nil {
		if err == sql.ErrNoRows {
			return ErrUserNotFound
		}
		return s.mapDBError(err)
	}
	return nil
}

func (s *UserService) DeleteUser(id int) error {
	err := s.Repository.Delete(id)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrUserNotFound
		}
		return err
	}
	return nil
}

func (s *UserService) Login(username, password string) (accessToken string, refreshToken string, err error) {
	if strings.TrimSpace(username) == "" || strings.TrimSpace(password) == "" {
		return "", "", ErrInvalidLogin
	}
	user, err := s.Repository.GetByUsername(username)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", "", ErrInvalidLogin
		}
		return "", "", err
	}
	if !utils.CheckPasswordHash(password, user.Password) {
		return "", "", ErrInvalidLogin
	}
	accessToken, err = utils.GenerateAccessToken(user.ID, user.Username)
	if err != nil {
		return "", "", err
	}
	refreshToken, err = utils.GenerateRefreshToken()
	if err != nil {
		return "", "", err
	}
	if count, err := s.RefreshRepository.CountByUserID(user.ID); err == nil {
		for count >= maxRefreshTokensPerUser {
			if err := s.RefreshRepository.DeleteOldestByUserID(user.ID); err != nil {
				break
			}
			count--
		}
	}
	rt := models.RefreshToken{
		UserID:    user.ID,
		Token:     refreshToken,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}
	if err := s.RefreshRepository.Create(&rt); err != nil {
		return "", "", err
	}
	return accessToken, refreshToken, nil
}

func (s *UserService) RefreshAccessToken(refreshToken string) (string, string, error) {
	if strings.TrimSpace(refreshToken) == "" {
		return "", "", ErrInvalidRefresh
	}
	rt, err := s.RefreshRepository.GetByToken(refreshToken)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", "", ErrInvalidRefresh
		}
		return "", "", err
	}
	if time.Now().After(rt.ExpiresAt) {
		_ = s.RefreshRepository.DeleteByToken(refreshToken)
		return "", "", ErrInvalidRefresh
	}
	user, err := s.Repository.GetByID(rt.UserID)
	if err != nil {
		if err == sql.ErrNoRows {
			_ = s.RefreshRepository.DeleteByToken(refreshToken)
			return "", "", ErrInvalidRefresh
		}
		return "", "", err
	}
	newAccessToken, err := utils.GenerateAccessToken(user.ID, user.Username)
	if err != nil {
		return "", "", err
	}
	newRefreshToken, err := utils.GenerateRefreshToken()
	if err != nil {
		return "", "", err
	}
	_ = s.RefreshRepository.DeleteByToken(refreshToken)
	newRT := models.RefreshToken{
		UserID:    user.ID,
		Token:     newRefreshToken,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}
	if err := s.RefreshRepository.Create(&newRT); err != nil {
		return "", "", err
	}
	return newAccessToken, newRefreshToken, nil
}

func (s *UserService) Logout(refreshToken string) error {
	if strings.TrimSpace(refreshToken) == "" {
		return ErrInvalidRefresh
	}
	return s.RefreshRepository.DeleteByToken(refreshToken)
}
