package audit

import (
	"context"
	"database/sql"
	"time"

	"golang-backend/models"
	"golang-backend/repositories"
)

// RepositoryInterface persists and reads CPAUDITLOG.
type RepositoryInterface interface {
	Create(ctx context.Context, log *models.AuditLog) error
	List(ctx context.Context, f models.AuditFilter) ([]models.AuditLog, error)
	CountSince(ctx context.Context, hours int) (int, error)
}

type Repository struct {
	db      *sql.DB
	dialect repositories.Dialect
}

func NewRepository(db *sql.DB, dialect repositories.Dialect) RepositoryInterface {
	return &Repository{db: db, dialect: dialect}
}

func (r *Repository) table() string { return r.dialect.Table("CPAUDITLOG") }

func (r *Repository) Create(ctx context.Context, log *models.AuditLog) error {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	_, err := r.db.ExecContext(ctx, r.dialect.Bind(`
		INSERT INTO `+r.table()+` (CODE, ACTOR_CODE, ACTION, ENTITY, ENTITY_CODE, DETAIL, IP_ADDRESS)
		VALUES (?, ?, ?, ?, ?, ?, ?)`),
		log.Code, repositories.NullStr(log.ActorCode), log.Action, log.Entity,
		repositories.NullStr(log.EntityCode), repositories.NullStr(log.Detail), repositories.NullStr(log.IPAddress))
	return err
}

func (r *Repository) List(ctx context.Context, f models.AuditFilter) ([]models.AuditLog, error) {
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
		SELECT CODE, ACTOR_CODE, ACTION, ENTITY, ENTITY_CODE, DETAIL, IP_ADDRESS, CREATED_AT
		FROM ` + r.table() + `
		WHERE (? = '' OR ACTION = ?)
		  AND (? = '' OR ENTITY = ?)
		  AND (? = '' OR ACTOR_CODE = ?)
		ORDER BY ID DESC `
	var args []any
	if r.dialect == repositories.DialectMSSQL {
		query += repositories.PageMSSQL()
		args = []any{f.Action, f.Action, f.Entity, f.Entity, f.Actor, f.Actor, f.Offset, limit}
	} else {
		query += repositories.PageStd()
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

func (r *Repository) CountSince(ctx context.Context, hours int) (int, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	var n int
	err := r.db.QueryRowContext(ctx, r.dialect.Bind(
		`SELECT COUNT(*) FROM `+r.table()+` WHERE CREATED_AT >= ?`),
		time.Now().Add(-time.Duration(hours)*time.Hour)).Scan(&n)
	return n, err
}
