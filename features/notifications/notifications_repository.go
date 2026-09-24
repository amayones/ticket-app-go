package repositories

import (
	"context"
	"database/sql"
	"time"

	"golang-backend/models"
)

// NotificationRepositoryInterface persists CPNOTIFTEMPLATE + CPNOTIFLOG.
type NotificationRepositoryInterface interface {
	ListTemplates(ctx context.Context, activeOnly bool) ([]models.NotifTemplate, error)
	GetTemplate(ctx context.Context, code string) (*models.NotifTemplate, error)
	CreateTemplate(ctx context.Context, t *models.NotifTemplate) error
	UpdateTemplate(ctx context.Context, t *models.NotifTemplate) error
	DeleteTemplate(ctx context.Context, code string) error
	CreateLog(ctx context.Context, l *models.NotifLog) error
	ListLogs(ctx context.Context, limit, offset int) ([]models.NotifLog, error)
	CountSentSince(ctx context.Context, hours int) (int, error)
}

type NotificationRepository struct {
	db      *sql.DB
	dialect Dialect
}

func NewNotificationRepository(db *sql.DB, dialect Dialect) NotificationRepositoryInterface {
	return &NotificationRepository{db: db, dialect: dialect}
}

func (r *NotificationRepository) tmplTable() string { return r.dialect.Table("CPNOTIFTEMPLATE") }
func (r *NotificationRepository) logTable() string  { return r.dialect.Table("CPNOTIFLOG") }

func (r *NotificationRepository) ListTemplates(ctx context.Context, activeOnly bool) ([]models.NotifTemplate, error) {
	ctx, cancel := withTimeout(ctx)
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

func (r *NotificationRepository) GetTemplate(ctx context.Context, code string) (*models.NotifTemplate, error) {
	ctx, cancel := withTimeout(ctx)
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

func (r *NotificationRepository) CreateTemplate(ctx context.Context, t *models.NotifTemplate) error {
	ctx, cancel := withTimeout(ctx)
	defer cancel()
	_, err := r.db.ExecContext(ctx, r.dialect.Bind(`
		INSERT INTO `+r.tmplTable()+` (CODE, NAME, CHANNEL, SUBJECT, BODY, IS_ACTIVE)
		VALUES (?, ?, ?, ?, ?, ?)`),
		t.Code, t.Name, t.Channel, nullStr(t.Subject), t.Body, t.IsActive)
	return err
}

func (r *NotificationRepository) UpdateTemplate(ctx context.Context, t *models.NotifTemplate) error {
	ctx, cancel := withTimeout(ctx)
	defer cancel()
	res, err := r.db.ExecContext(ctx, r.dialect.Bind(`
		UPDATE `+r.tmplTable()+`
		SET NAME = ?, CHANNEL = ?, SUBJECT = ?, BODY = ?, IS_ACTIVE = ?, UPDATED_AT = `+r.dialect.Now()+`
		WHERE CODE = ?`),
		t.Name, t.Channel, nullStr(t.Subject), t.Body, t.IsActive, t.Code)
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

func (r *NotificationRepository) DeleteTemplate(ctx context.Context, code string) error {
	ctx, cancel := withTimeout(ctx)
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

func (r *NotificationRepository) CreateLog(ctx context.Context, l *models.NotifLog) error {
	ctx, cancel := withTimeout(ctx)
	defer cancel()
	_, err := r.db.ExecContext(ctx, r.dialect.Bind(`
		INSERT INTO `+r.logTable()+` (CODE, TEMPLATE_CODE, CHANNEL, RECIPIENT, SUBJECT, BODY, STATUS, ERROR)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`),
		l.Code, nullStr(l.TemplateCode), l.Channel, l.Recipient,
		nullStr(l.Subject), l.Body, l.Status, nullStr(l.Error))
	return err
}

func (r *NotificationRepository) ListLogs(ctx context.Context, limit, offset int) ([]models.NotifLog, error) {
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
	query := `
		SELECT CODE, TEMPLATE_CODE, CHANNEL, RECIPIENT, SUBJECT, BODY, STATUS, ERROR, CREATED_AT
		FROM ` + r.logTable() + `
		ORDER BY ID DESC `
	var args []any
	if r.dialect == DialectMSSQL {
		query += pageMSSQL()
		args = []any{offset, limit}
	} else {
		query += pageStd()
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

func (r *NotificationRepository) CountSentSince(ctx context.Context, hours int) (int, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()
	var n int
	err := r.db.QueryRowContext(ctx, r.dialect.Bind(`
		SELECT COUNT(*) FROM `+r.logTable()+`
		WHERE STATUS = 'SENT' AND CREATED_AT >= ?`),
		time.Now().Add(-time.Duration(hours)*time.Hour)).Scan(&n)
	return n, err
}
