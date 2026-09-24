package repositories

import (
	"context"
	"database/sql"

	"golang-backend/models"
)

// AuditRepositoryInterface persists and reads CPAUDITLOG.
type AuditRepositoryInterface interface {
	Create(ctx context.Context, log *models.AuditLog) error
	List(ctx context.Context, f models.AuditFilter) ([]models.AuditLog, error)
	CountSince(ctx context.Context, hours int) (int, error)
}

type AuditRepository struct {
	db *sql.DB
}

func NewAuditRepository(db *sql.DB) AuditRepositoryInterface {
	return &AuditRepository{db: db}
}

func (r *AuditRepository) Create(ctx context.Context, log *models.AuditLog) error {
	ctx, cancel := withTimeout(ctx)
	defer cancel()
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO dbo.CPAUDITLOG (CODE, ACTOR_CODE, ACTION, ENTITY, ENTITY_CODE, DETAIL, IP_ADDRESS)
		VALUES (@p1, @p2, @p3, @p4, @p5, @p6, @p7)`,
		log.Code, nullStr(log.ActorCode), log.Action, log.Entity,
		nullStr(log.EntityCode), nullStr(log.Detail), nullStr(log.IPAddress))
	return err
}

func nullStr(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func (r *AuditRepository) List(ctx context.Context, f models.AuditFilter) ([]models.AuditLog, error) {
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
		SELECT CODE, ACTOR_CODE, ACTION, ENTITY, ENTITY_CODE, DETAIL, IP_ADDRESS, CREATED_AT
		FROM dbo.CPAUDITLOG
		WHERE (@p1 = '' OR ACTION = @p1)
		  AND (@p2 = '' OR ENTITY = @p2)
		  AND (@p3 = '' OR ACTOR_CODE = @p3)
		ORDER BY ID DESC
		OFFSET @p4 ROWS FETCH NEXT @p5 ROWS ONLY`
	rows, err := r.db.QueryContext(ctx, query, f.Action, f.Entity, f.Actor, f.Offset, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.AuditLog, 0)
	for rows.Next() {
		var l models.AuditLog
		var actor, entityCode, detail, ip sql.NullString
		if err := rows.Scan(&l.Code, &actor, &l.Action, &l.Entity, &entityCode, &detail, &ip, &l.CreatedAt); err != nil {
			return nil, err
		}
		l.ActorCode = actor.String
		l.EntityCode = entityCode.String
		l.Detail = detail.String
		l.IPAddress = ip.String
		out = append(out, l)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *AuditRepository) CountSince(ctx context.Context, hours int) (int, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()
	var n int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM dbo.CPAUDITLOG WHERE CREATED_AT >= DATEADD(HOUR, -@p1, GETDATE())`, hours).Scan(&n)
	return n, err
}
