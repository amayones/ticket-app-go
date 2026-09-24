package repositories

import (
	"context"
	"database/sql"
	"time"

	"golang-backend/models"
)

// AuditRepositoryInterface persists and reads CPAUDITLOG.
type AuditRepositoryInterface interface {
	Create(ctx context.Context, log *models.AuditLog) error
	List(ctx context.Context, f models.AuditFilter) ([]models.AuditLog, error)
	CountSince(ctx context.Context, hours int) (int, error)
}

type AuditRepository struct {
	db      *sql.DB
	dialect Dialect
}

func NewAuditRepository(db *sql.DB, dialect Dialect) AuditRepositoryInterface {
	return &AuditRepository{db: db, dialect: dialect}
}

func (r *AuditRepository) table() string { return r.dialect.Table("CPAUDITLOG") }

func (r *AuditRepository) Create(ctx context.Context, log *models.AuditLog) error {
	ctx, cancel := withTimeout(ctx)
	defer cancel()
	_, err := r.db.ExecContext(ctx, r.dialect.Bind(`
		INSERT INTO `+r.table()+` (CODE, ACTOR_CODE, ACTION, ENTITY, ENTITY_CODE, DETAIL, IP_ADDRESS)
		VALUES (?, ?, ?, ?, ?, ?, ?)`),
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
		FROM ` + r.table() + `
		WHERE (? = '' OR ACTION = ?)
		  AND (? = '' OR ENTITY = ?)
		  AND (? = '' OR ACTOR_CODE = ?)
		ORDER BY ID DESC `
	var args []any
	if r.dialect == DialectMSSQL {
		query += pageMSSQL()
		args = []any{f.Action, f.Action, f.Entity, f.Entity, f.Actor, f.Actor, f.Offset, limit}
	} else {
		query += pageStd()
		args = []any{f.Action, f.Action, f.Entity, f.Entity, f.Actor, f.Actor, limit, f.Offset}
	}
	rows, err := r.db.QueryContext(ctx, r.dialect.Bind(query), args...)
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
	err := r.db.QueryRowContext(ctx, r.dialect.Bind(
		`SELECT COUNT(*) FROM `+r.table()+` WHERE CREATED_AT >= ?`),
		time.Now().Add(-time.Duration(hours)*time.Hour)).Scan(&n)
	return n, err
}
