package syslog

import (
	"context"
	"database/sql"
	"time"

	"golang-backend/models"
	"golang-backend/repositories"
)

// RepositoryInterface persists and reads CPSYSLOG.
type RepositoryInterface interface {
	Create(ctx context.Context, log *models.SysLog) error
	List(ctx context.Context, f models.SyslogFilter) ([]models.SysLog, error)
	CountSince(ctx context.Context, hours int, level string) (int, error)
	PruneBefore(ctx context.Context, days int) (int64, error)
}

type Repository struct {
	db      *sql.DB
	dialect repositories.Dialect
}

func NewRepository(db *sql.DB, dialect repositories.Dialect) RepositoryInterface {
	return &Repository{db: db, dialect: dialect}
}

func (r *Repository) table() string { return r.dialect.Table("CPSYSLOG") }

func (r *Repository) Create(ctx context.Context, log *models.SysLog) error {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	_, err := r.db.ExecContext(ctx, r.dialect.Bind(`
		INSERT INTO `+r.table()+` (CODE, LEVEL, SOURCE, MESSAGE)
		VALUES (?, ?, ?, ?)`),
		log.Code, log.Level, log.Source, log.Message)
	return err
}

func (r *Repository) List(ctx context.Context, f models.SyslogFilter) ([]models.SysLog, error) {
	limit := f.Limit
	if limit <= 0 {
		limit = repositories.DefaultListLimit
	}
	if limit > repositories.MaxListLimit {
		limit = repositories.MaxListLimit
	}
	if f.Offset < 0 {
		f.Offset = 0
	}
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	query := `
		SELECT CODE, LEVEL, SOURCE, MESSAGE, CREATED_AT
		FROM ` + r.table() + `
		WHERE (? = '' OR LEVEL = ?)
		ORDER BY ID DESC `
	var args []any
	if r.dialect == repositories.DialectMSSQL {
		query += repositories.PageMSSQL()
		args = []any{f.Level, f.Level, f.Offset, limit}
	} else {
		query += repositories.PageStd()
		args = []any{f.Level, f.Level, limit, f.Offset}
	}
	rows, err := r.db.QueryContext(ctx, r.dialect.Bind(query), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.SysLog, 0)
	for rows.Next() {
		var l models.SysLog
		if err := rows.Scan(&l.Code, &l.Level, &l.Source, &l.Message, &l.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *Repository) CountSince(ctx context.Context, hours int, level string) (int, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	var n int
	err := r.db.QueryRowContext(ctx, r.dialect.Bind(`
		SELECT COUNT(*) FROM `+r.table()+`
		WHERE CREATED_AT >= ? AND (? = '' OR LEVEL = ?)`),
		time.Now().Add(-time.Duration(hours)*time.Hour), level, level).Scan(&n)
	return n, err
}

func (r *Repository) PruneBefore(ctx context.Context, days int) (int64, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	res, err := r.db.ExecContext(ctx, r.dialect.Bind(
		`DELETE FROM `+r.table()+` WHERE CREATED_AT < ?`),
		time.Now().AddDate(0, 0, -days))
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}
