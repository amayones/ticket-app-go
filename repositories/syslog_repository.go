package repositories

import (
	"context"
	"database/sql"

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
	db *sql.DB
}

func NewSyslogRepository(db *sql.DB) SyslogRepositoryInterface {
	return &SyslogRepository{db: db}
}

func (r *SyslogRepository) Create(ctx context.Context, log *models.SysLog) error {
	ctx, cancel := withTimeout(ctx)
	defer cancel()
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO dbo.CPSYSLOG (CODE, LEVEL, SOURCE, MESSAGE)
		VALUES (@p1, @p2, @p3, @p4)`,
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
	rows, err := r.db.QueryContext(ctx, `
		SELECT CODE, LEVEL, SOURCE, MESSAGE, CREATED_AT
		FROM dbo.CPSYSLOG
		WHERE (@p1 = '' OR LEVEL = @p1)
		ORDER BY ID DESC
		OFFSET @p2 ROWS FETCH NEXT @p3 ROWS ONLY`, f.Level, f.Offset, limit)
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
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM dbo.CPSYSLOG
		WHERE CREATED_AT >= DATEADD(HOUR, -@p1, GETDATE()) AND (@p2 = '' OR LEVEL = @p2)`,
		hours, level).Scan(&n)
	return n, err
}

func (r *SyslogRepository) PruneBefore(ctx context.Context, days int) (int64, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()
	res, err := r.db.ExecContext(ctx,
		`DELETE FROM dbo.CPSYSLOG WHERE CREATED_AT < DATEADD(DAY, -@p1, GETDATE())`, days)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}
