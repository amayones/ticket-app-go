package repositories

import (
	"context"
	"database/sql"

	"golang-backend/models"
)

// RefreshTokenRepositoryInterface works with SHA-256 hashes
// (plaintext tokens never touch the DB).
type RefreshTokenRepositoryInterface interface {
	Create(ctx context.Context, token *models.RefreshToken) error
	GetByTokenHash(ctx context.Context, tokenHash string) (*models.RefreshToken, error)
	// DeleteByTokenHash returns true when a row was actually removed.
	DeleteByTokenHash(ctx context.Context, tokenHash string) (bool, error)
	DeleteByUserID(ctx context.Context, userID int) error
	DeleteExpired(ctx context.Context) (int64, error)
	CountByUserID(ctx context.Context, userID int) (int, error)
	DeleteOldestByUserID(ctx context.Context, userID int) error
}

type RefreshTokenRepository struct {
	db *sql.DB
}

func NewRefreshTokenRepository(db *sql.DB) RefreshTokenRepositoryInterface {
	return &RefreshTokenRepository{db: db}
}

func (r *RefreshTokenRepository) Create(ctx context.Context, token *models.RefreshToken) error {
	ctx, cancel := withTimeout(ctx)
	defer cancel()
	query := `
		INSERT INTO refresh_tokens (user_id, token, expires_at)
		VALUES (@p1, @p2, @p3)
	`
	_, err := r.db.ExecContext(ctx, query, token.UserID, token.Token, token.ExpiresAt)
	return err
}

func (r *RefreshTokenRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*models.RefreshToken, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()
	query := `
		SELECT id, user_id, token, expires_at, created_at
		FROM refresh_tokens
		WHERE token = @p1
	`
	var rt models.RefreshToken
	err := r.db.QueryRowContext(ctx, query, tokenHash).Scan(
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

func (r *RefreshTokenRepository) DeleteByTokenHash(ctx context.Context, tokenHash string) (bool, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()
	res, err := r.db.ExecContext(ctx, `DELETE FROM refresh_tokens WHERE token = @p1`, tokenHash)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func (r *RefreshTokenRepository) DeleteByUserID(ctx context.Context, userID int) error {
	ctx, cancel := withTimeout(ctx)
	defer cancel()
	_, err := r.db.ExecContext(ctx, `DELETE FROM refresh_tokens WHERE user_id = @p1`, userID)
	return err
}

// DeleteExpired removes already-expired rows. Call from a cron/scheduler
// to keep the table bounded.
func (r *RefreshTokenRepository) DeleteExpired(ctx context.Context) (int64, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()
	res, err := r.db.ExecContext(ctx, `DELETE FROM refresh_tokens WHERE expires_at < GETDATE()`)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (r *RefreshTokenRepository) CountByUserID(ctx context.Context, userID int) (int, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()
	var count int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM refresh_tokens WHERE user_id = @p1`, userID).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// DeleteOldestByUserID evicts one row. NOTE: Count+DeleteOldest+Create is
// still non-atomic under concurrency; DB-level cap (trigger/proc) is the
// full fix. This keeps sessions bounded in the common case.
func (r *RefreshTokenRepository) DeleteOldestByUserID(ctx context.Context, userID int) error {
	ctx, cancel := withTimeout(ctx)
	defer cancel()
	query := `
		DELETE FROM refresh_tokens
		WHERE id = (
			SELECT TOP 1 id FROM refresh_tokens WHERE user_id = @p1 ORDER BY created_at ASC, id ASC
		)
	`
	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}
