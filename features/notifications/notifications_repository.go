package notifications

import (
	"context"
	"database/sql"
	"time"

	"golang-backend/models"
	"golang-backend/repositories"
)

// RepositoryInterface persists CPNOTIFTEMPLATE + CPNOTIFLOG.
type RepositoryInterface interface {
	ListTemplates(ctx context.Context, activeOnly bool) ([]models.NotifTemplate, error)
	GetTemplate(ctx context.Context, code string) (*models.NotifTemplate, error)
	CreateTemplate(ctx context.Context, t *models.NotifTemplate) error
	UpdateTemplate(ctx context.Context, t *models.NotifTemplate) error
	DeleteTemplate(ctx context.Context, code string) error
	CreateLog(ctx context.Context, l *models.NotifLog) error
	ListLogs(ctx context.Context, limit, offset int) ([]models.NotifLog, error)
	CountSentSince(ctx context.Context, hours int) (int, error)
}

type Repository struct {
	db      *sql.DB
	dialect repositories.Dialect
}

func NewRepository(db *sql.DB, dialect repositories.Dialect) RepositoryInterface {
	return &Repository{db: db, dialect: dialect}
}

func (r *Repository) tmplTable() string { return r.dialect.Table("CPNOTIFTEMPLATE") }
func (r *Repository) logTable() string  { return r.dialect.Table("CPNOTIFLOG") }

func (r *Repository) ListTemplates(ctx context.Context, activeOnly bool) ([]models.NotifTemplate, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	query := `
		SELECT CODE, NAME, CHANNEL, SUBJECT, BODY, IS_ACTIVE, CREATED_AT, UPDATED_AT
		FROM ` + r.tmplTable()
	if activeOnly {
		query += ` WHERE IS_ACTIVE = ` + r.dialect.IsTrue()
	}
	query += ` ORDER BY NAME ASC`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.NotifTemplate, 0)
	for rows.Next() {
		var t models.NotifTemplate
		var subject sql.NullString
		if err := rows.Scan(&t.Code, &t.Name, &t.Channel, &subject, &t.Body, &t.IsActive, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		t.Subject = subject.String
		out = append(out, t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *Repository) GetTemplate(ctx context.Context, code string) (*models.NotifTemplate, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	var t models.NotifTemplate
	var subject sql.NullString
	err := r.db.QueryRowContext(ctx, r.dialect.Bind(`
		SELECT CODE, NAME, CHANNEL, SUBJECT, BODY, IS_ACTIVE, CREATED_AT, UPDATED_AT
		FROM `+r.tmplTable()+` WHERE CODE = ?`), code,
	).Scan(&t.Code, &t.Name, &t.Channel, &subject, &t.Body, &t.IsActive, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, err
	}
	t.Subject = subject.String
	return &t, nil
}

func (r *Repository) CreateTemplate(ctx context.Context, t *models.NotifTemplate) error {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	_, err := r.db.ExecContext(ctx, r.dialect.Bind(`
		INSERT INTO `+r.tmplTable()+` (CODE, NAME, CHANNEL, SUBJECT, BODY, IS_ACTIVE)
		VALUES (?, ?, ?, ?, ?, ?)`),
		t.Code, t.Name, t.Channel, repositories.NullStr(t.Subject), t.Body, t.IsActive)
	return err
}

func (r *Repository) UpdateTemplate(ctx context.Context, t *models.NotifTemplate) error {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	res, err := r.db.ExecContext(ctx, r.dialect.Bind(`
		UPDATE `+r.tmplTable()+`
		SET NAME = ?, CHANNEL = ?, SUBJECT = ?, BODY = ?, IS_ACTIVE = ?, UPDATED_AT = `+r.dialect.Now()+`
		WHERE CODE = ?`),
		t.Name, t.Channel, repositories.NullStr(t.Subject), t.Body, t.IsActive, t.Code)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *Repository) DeleteTemplate(ctx context.Context, code string) error {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	res, err := r.db.ExecContext(ctx,
		r.dialect.Bind(`DELETE FROM `+r.tmplTable()+` WHERE CODE = ?`), code)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *Repository) CreateLog(ctx context.Context, l *models.NotifLog) error {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	_, err := r.db.ExecContext(ctx, r.dialect.Bind(`
		INSERT INTO `+r.logTable()+` (CODE, TEMPLATE_CODE, CHANNEL, RECIPIENT, SUBJECT, BODY, STATUS, ERROR)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`),
		l.Code, repositories.NullStr(l.TemplateCode), l.Channel, l.Recipient,
		repositories.NullStr(l.Subject), l.Body, l.Status, repositories.NullStr(l.Error))
	return err
}

func (r *Repository) ListLogs(ctx context.Context, limit, offset int) ([]models.NotifLog, error) {
	if limit <= 0 {
		limit = repositories.DefaultListLimit
	}
	if limit > repositories.MaxListLimit {
		limit = repositories.MaxListLimit
	}
	if offset < 0 {
		offset = 0
	}
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	query := `
		SELECT CODE, TEMPLATE_CODE, CHANNEL, RECIPIENT, SUBJECT, BODY, STATUS, ERROR, CREATED_AT
		FROM ` + r.logTable() + `
		ORDER BY ID DESC `
	var args []any
	if r.dialect == repositories.DialectMSSQL {
		query += repositories.PageMSSQL()
		args = []any{offset, limit}
	} else {
		query += repositories.PageStd()
		args = []any{limit, offset}
	}
	rows, err := r.db.QueryContext(ctx, r.dialect.Bind(query), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.NotifLog, 0)
	for rows.Next() {
		var l models.NotifLog
		var tmpl, subject, errMsg sql.NullString
		if err := rows.Scan(&l.Code, &tmpl, &l.Channel, &l.Recipient, &subject, &l.Body, &l.Status, &errMsg, &l.CreatedAt); err != nil {
			return nil, err
		}
		l.TemplateCode, l.Subject, l.Error = tmpl.String, subject.String, errMsg.String
		out = append(out, l)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *Repository) CountSentSince(ctx context.Context, hours int) (int, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	var n int
	err := r.db.QueryRowContext(ctx, r.dialect.Bind(`
		SELECT COUNT(*) FROM `+r.logTable()+`
		WHERE STATUS = 'SENT' AND CREATED_AT >= ?`),
		time.Now().Add(-time.Duration(hours)*time.Hour)).Scan(&n)
	return n, err
}
