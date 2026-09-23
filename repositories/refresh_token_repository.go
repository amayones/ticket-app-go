package repositories

import (
	"database/sql"

	"golang-backend/models"
)

type RefreshTokenRepositoryInterface interface {
	Create(token *models.RefreshToken) error
	GetByToken(token string) (*models.RefreshToken, error)
	DeleteByToken(token string) error
	DeleteByUserID(userID int) error
	CountByUserID(userID int) (int, error)
	DeleteOldestByUserID(userID int) error
}

type RefreshTokenRepository struct {
	DB *sql.DB
}

func NewRefreshTokenRepository(db *sql.DB) *RefreshTokenRepository {
	return &RefreshTokenRepository{
		DB: db,
	}
}

func (r *RefreshTokenRepository) Create(token *models.RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (user_id, token, expires_at)
		VALUES (@p1, @p2, @p3)
	`
	_, err := r.DB.Exec(query, token.UserID, token.Token, token.ExpiresAt)
	return err
}

func (r *RefreshTokenRepository) GetByToken(token string) (*models.RefreshToken, error) {
	query := `
		SELECT id, user_id, token, expires_at, created_at
		FROM refresh_tokens
		WHERE token = @p1
	`
	var rt models.RefreshToken
	err := r.DB.QueryRow(query, token).Scan(
		&rt.ID,
		&rt.UserID,
		&rt.Token,
		&rt.ExpiresAt,
		&rt.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &rt, nil
}

func (r *RefreshTokenRepository) DeleteByToken(token string) error {
	query := `DELETE FROM refresh_tokens WHERE token = @p1`
	_, err := r.DB.Exec(query, token)
	return err
}

func (r *RefreshTokenRepository) DeleteByUserID(userID int) error {
	query := `DELETE FROM refresh_tokens WHERE user_id = @p1`
	_, err := r.DB.Exec(query, userID)
	return err
}

func (r *RefreshTokenRepository) CountByUserID(userID int) (int, error) {
	query := `SELECT COUNT(*) FROM refresh_tokens WHERE user_id = @p1`
	var count int
	err := r.DB.QueryRow(query, userID).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *RefreshTokenRepository) DeleteOldestByUserID(userID int) error {
	query := `
		DELETE FROM refresh_tokens
		WHERE id = (
			SELECT TOP 1 id FROM refresh_tokens WHERE user_id = @p1 ORDER BY created_at ASC
		)
	`
	_, err := r.DB.Exec(query, userID)
	return err
}
