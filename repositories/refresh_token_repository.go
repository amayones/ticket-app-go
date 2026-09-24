package repositories

import (
	"context"
	"database/sql"
	"time"

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
	CountActive(ctx context.Context) (int, error)
	DeleteOldestByUserCode(ctx context.Context, userCode string) error
	ListByUserCode(ctx context.Context, userCode string) ([]models.Session, error)
	ListAll(ctx context.Context, limit, offset int) ([]models.Session, error)
	DeleteByID(ctx context.Context, id int) (bool, error)
}

type RefreshTokenRepository struct {
	db      *sql.DB
	dialect Dialect
}

func NewRefreshTokenRepository(db *sql.DB, dialect Dialect) RefreshTokenRepositoryInterface {
	return &RefreshTokenRepository{db: db, dialect: dialect}
}

func (r *RefreshTokenRepository) table() string {
	return r.dialect.Table("CPREFRESHTOKEN")
}

func (r *RefreshTokenRepository) Create(ctx context.Context, token *models.RefreshToken) error {
	ctx, cancel := withTimeout(ctx)
	defer cancel()
	query := `INSERT INTO ` + r.table() + ` (USER_CODE, TOKEN, EXPIRES_AT) VALUES (?, ?, ?)`
	_, err := r.db.ExecContext(ctx, r.dialect.Bind(query),
		token.UserCode, token.Token, token.ExpiresAt)
	return err
}

func (r *RefreshTokenRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*models.RefreshToken, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()
	query := `SELECT ID, USER_CODE, TOKEN, EXPIRES_AT, CREATED_AT
		FROM ` + r.table() + ` WHERE TOKEN = ?`
	var rt models.RefreshToken
	err := r.db.QueryRowContext(ctx, r.dialect.Bind(query), tokenHash).Scan(
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
	query := `DELETE FROM ` + r.table() + ` WHERE TOKEN = ?`
	res, err := r.db.ExecContext(ctx, r.dialect.Bind(query), tokenHash)
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
	query := `DELETE FROM ` + r.table() + ` WHERE USER_CODE = ?`
	_, err := r.db.ExecContext(ctx, r.dialect.Bind(query), userCode)
	return err
}

// DeleteExpired removes already-expired rows. Call from a cron/scheduler
// to keep the table bounded.
func (r *RefreshTokenRepository) DeleteExpired(ctx context.Context) (int64, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()
	query := `DELETE FROM ` + r.table() + ` WHERE EXPIRES_AT < ?`
	res, err := r.db.ExecContext(ctx, r.dialect.Bind(query), time.Now())
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (r *RefreshTokenRepository) CountByUserCode(ctx context.Context, userCode string) (int, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()
	query := `SELECT COUNT(*) FROM ` + r.table() + ` WHERE USER_CODE = ?`
	var count int
	err := r.db.QueryRowContext(ctx, r.dialect.Bind(query), userCode).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *RefreshTokenRepository) CountActive(ctx context.Context) (int, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()
	query := `SELECT COUNT(*) FROM ` + r.table() + ` WHERE EXPIRES_AT >= ?`
	var count int
	err := r.db.QueryRowContext(ctx, r.dialect.Bind(query), time.Now()).Scan(&count)
	return count, err
}

// DeleteOldestByUserCode evicts one row. NOTE: Count+DeleteOldest+Create is
// still non-atomic under concurrency; DB-level cap (trigger/proc) is the
// full fix. This keeps sessions bounded in the common case.
func (r *RefreshTokenRepository) DeleteOldestByUserCode(ctx context.Context, userCode string) error {
	ctx, cancel := withTimeout(ctx)
	defer cancel()
	inner := `SELECT ID FROM ` + r.table() + ` WHERE USER_CODE = ? ORDER BY CREATED_AT ASC, ID ASC`
	if r.dialect == DialectMSSQL {
		inner = `SELECT TOP 1 ID FROM ` + r.table() + ` WHERE USER_CODE = ? ORDER BY CREATED_AT ASC, ID ASC`
	} else {
		inner += ` LIMIT 1`
	}
	query := `DELETE FROM ` + r.table() + ` WHERE ID = (` + inner + `)`
	_, err := r.db.ExecContext(ctx, r.dialect.Bind(query), userCode)
	return err
}

// ListByUserCode returns active sessions of one user (hashes never exposed).
func (r *RefreshTokenRepository) ListByUserCode(ctx context.Context, userCode string) ([]models.Session, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()
	return r.querySessions(ctx, `SELECT t.ID, t.USER_CODE, u.USERNAME, t.EXPIRES_AT, t.CREATED_AT
		FROM `+r.table()+` t JOIN `+r.dialect.Table("CPUSER")+` u ON u.CODE = t.USER_CODE
		WHERE t.USER_CODE = ? AND t.EXPIRES_AT >= ?
		ORDER BY t.CREATED_AT DESC`, userCode, time.Now())
}

// ListAll returns active sessions of all users (admin view).
func (r *RefreshTokenRepository) ListAll(ctx context.Context, limit, offset int) ([]models.Session, error) {
	if limit <= 0 {
		limit = DefaultListLimit
	}
	if limit > MaxListLimit {
		limit = MaxListLimit
	}
	if offset < 0 {
		offset = 0
	}
	ctx, cancel := withTimeout(ctx)
	defer cancel()
	query := `SELECT t.ID, t.USER_CODE, u.USERNAME, t.EXPIRES_AT, t.CREATED_AT
		FROM ` + r.table() + ` t JOIN ` + r.dialect.Table("CPUSER") + ` u ON u.CODE = t.USER_CODE
		WHERE t.EXPIRES_AT >= ? ORDER BY t.CREATED_AT DESC `
	var args []any
	if r.dialect == DialectMSSQL {
		query += pageMSSQL()
		args = []any{time.Now(), offset, limit}
	} else {
		query += pageStd()
		args = []any{time.Now(), limit, offset}
	}
	return r.querySessions(ctx, query, args...)
}

func (r *RefreshTokenRepository) querySessions(ctx context.Context, query string, args ...any) ([]models.Session, error) {
	rows, err := r.db.QueryContext(ctx, r.dialect.Bind(query), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.Session, 0)
	for rows.Next() {
		var s models.Session
		if err := rows.Scan(&s.ID, &s.UserCode, &s.Username, &s.ExpiresAt, &s.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// DeleteByID revokes one session by row ID (owner or admin enforced by caller).
func (r *RefreshTokenRepository) DeleteByID(ctx context.Context, id int) (bool, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()
	query := `DELETE FROM ` + r.table() + ` WHERE ID = ?`
	res, err := r.db.ExecContext(ctx, r.dialect.Bind(query), id)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}
