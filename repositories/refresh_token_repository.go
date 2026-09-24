package repositories

import (
	"context"
	"database/sql"

	"golang-backend/models"
)

// RefreshTokenRepositoryInterface works with SHA-256 hashes
// (plaintext tokens never touch the DB). Owner link is USER_CODE.
type RefreshTokenRepositoryInterface interface {
	Create(ctx context.Context, token *models.RefreshToken) error
	GetByTokenHash(ctx context.Context, tokenHash string) (*models.RefreshToken, error)
	// DeleteByTokenHash returns true when a row was actually removed.
	DeleteByTokenHash(ctx context.Context, tokenHash string) (bool, error)
	DeleteByUserCode(ctx context.Context, userCode string) error
	DeleteExpired(ctx context.Context) (int64, error)
	CountByUserCode(ctx context.Context, userCode string) (int, error)
	DeleteOldestByUserCode(ctx context.Context, userCode string) error
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
		INSERT INTO dbo.CPREFRESHTOKEN (USER_CODE, TOKEN, EXPIRES_AT)
		VALUES (@p1, @p2, @p3)
	`
	_, err := r.db.ExecContext(ctx, query, token.UserCode, token.Token, token.ExpiresAt)
	return err
}

func (r *RefreshTokenRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*models.RefreshToken, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()
	query := `
		SELECT ID, USER_CODE, TOKEN, EXPIRES_AT, CREATED_AT
		FROM dbo.CPREFRESHTOKEN
		WHERE TOKEN = @p1
	`
	var rt models.RefreshToken
	err := r.db.QueryRowContext(ctx, query, tokenHash).Scan(
		&rt.ID,
		&rt.UserCode,
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
	res, err := r.db.ExecContext(ctx, `DELETE FROM dbo.CPREFRESHTOKEN WHERE TOKEN = @p1`, tokenHash)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func (r *RefreshTokenRepository) DeleteByUserCode(ctx context.Context, userCode string) error {
	ctx, cancel := withTimeout(ctx)
	defer cancel()
	_, err := r.db.ExecContext(ctx, `DELETE FROM dbo.CPREFRESHTOKEN WHERE USER_CODE = @p1`, userCode)
	return err
}

// DeleteExpired removes already-expired rows. Call from a cron/scheduler
// to keep the table bounded.
func (r *RefreshTokenRepository) DeleteExpired(ctx context.Context) (int64, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()
	res, err := r.db.ExecContext(ctx, `DELETE FROM dbo.CPREFRESHTOKEN WHERE EXPIRES_AT < GETDATE()`)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (r *RefreshTokenRepository) CountByUserCode(ctx context.Context, userCode string) (int, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()
	var count int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM dbo.CPREFRESHTOKEN WHERE USER_CODE = @p1`, userCode).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// DeleteOldestByUserCode evicts one row. NOTE: Count+DeleteOldest+Create is
// still non-atomic under concurrency; DB-level cap (trigger/proc) is the
// full fix. This keeps sessions bounded in the common case.
func (r *RefreshTokenRepository) DeleteOldestByUserCode(ctx context.Context, userCode string) error {
	ctx, cancel := withTimeout(ctx)
	defer cancel()
	query := `
		DELETE FROM dbo.CPREFRESHTOKEN
		WHERE ID = (
			SELECT TOP 1 ID FROM dbo.CPREFRESHTOKEN WHERE USER_CODE = @p1 ORDER BY CREATED_AT ASC, ID ASC
		)
	`
	_, err := r.db.ExecContext(ctx, query, userCode)
	return err
}
