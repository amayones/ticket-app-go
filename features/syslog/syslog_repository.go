package repositories

import (
	"context"
	"database/sql"
	"time"

	"golang-backend/models"
)

// SyslogRepositoryInterface persists and reads CPSYSLOG.
type SyslogRepositoryInterface interface {
	Create(ctx context.Context, log *models.SysLog) error
	List(ctx context.Context, f models.SyslogFilter) ([]models.SysLog, error)
	CountSince(ctx context.Context, hours int, level string) (int, error)
	PruneBefore(ctx context.Context, days int) (int64, error)
}

type SyslogRepository struct {
	db      *sql.DB
	dialect Dialect
}

func NewSyslogRepository(db *sql.DB, dialect Dialect) SyslogRepositoryInterface {
	return &SyslogRepository{db: db, dialect: dialect}
}

func (r *SyslogRepository) table() string { return r.dialect.Table("CPSYSLOG") }

func (r *SyslogRepository) Create(ctx context.Context, log *models.SysLog) error {
	ctx, cancel := withTimeout(ctx)
	defer cancel()
	_, err := r.db.ExecContext(ctx, r.dialect.Bind(`
		INSERT INTO `+r.table()+` (CODE, LEVEL, SOURCE, MESSAGE)
		VALUES (?, ?, ?, ?)`),
		log.Code, log.Level, log.Source, log.Message)
	return err
}

func (r *SyslogRepository) List(ctx context.Context, f models.SyslogFilter) ([]models.SysLog, error) {
	limit := f.Limit
	if limit <= 0 {
		limit = DefaultListLimit
	}
	if limit > MaxListLimit {
		limit = MaxListLimit
	}
	if f.Offset < 0 {
		f.Offset = 0
	}
	ctx, cancel := withTimeout(ctx)
	defer cancel()
	query := `
		SELECT CODE, LEVEL, SOURCE, MESSAGE, CREATED_AT
		FROM ` + r.table() + `
		WHERE (? = '' OR LEVEL = ?)
		ORDER BY ID DESC `
	var args []any
	if r.dialect == DialectMSSQL {
		query += pageMSSQL()
		args = []any{f.Level, f.Level, f.Offset, limit}
	} else {
		query += pageStd()
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

func (r *SyslogRepository) CountSince(ctx context.Context, hours int, level string) (int, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()
	var n int
	err := r.db.QueryRowContext(ctx, r.dialect.Bind(`
		SELECT COUNT(*) FROM `+r.table()+`
		WHERE CREATED_AT >= ? AND (? = '' OR LEVEL = ?)`),
		time.Now().Add(-time.Duration(hours)*time.Hour), level, level).Scan(&n)
	return n, err
}

func (r *SyslogRepository) PruneBefore(ctx context.Context, days int) (int64, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()
	res, err := r.db.ExecContext(ctx, r.dialect.Bind(
		`DELETE FROM `+r.table()+` WHERE CREATED_AT < ?`),
		time.Now().AddDate(0, 0, -days))
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}
