package repositories

import (
	"context"
	"database/sql"

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
	db *sql.DB
}

func NewNotificationRepository(db *sql.DB) NotificationRepositoryInterface {
	return &NotificationRepository{db: db}
}

func (r *NotificationRepository) ListTemplates(ctx context.Context, activeOnly bool) ([]models.NotifTemplate, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()
	query := `
		SELECT CODE, NAME, CHANNEL, SUBJECT, BODY, IS_ACTIVE, CREATED_AT, UPDATED_AT
		FROM dbo.CPNOTIFTEMPLATE
		WHERE (@p1 = 0 OR IS_ACTIVE = 1)
		ORDER BY NAME ASC`
	rows, err := r.db.QueryContext(ctx, query, boolToInt(activeOnly))
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

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func (r *NotificationRepository) GetTemplate(ctx context.Context, code string) (*models.NotifTemplate, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()
	var t models.NotifTemplate
	var subject sql.NullString
	err := r.db.QueryRowContext(ctx, `
		SELECT CODE, NAME, CHANNEL, SUBJECT, BODY, IS_ACTIVE, CREATED_AT, UPDATED_AT
		FROM dbo.CPNOTIFTEMPLATE WHERE CODE = @p1`, code,
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
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO dbo.CPNOTIFTEMPLATE (CODE, NAME, CHANNEL, SUBJECT, BODY, IS_ACTIVE)
		VALUES (@p1, @p2, @p3, @p4, @p5, @p6)`,
		t.Code, t.Name, t.Channel, nullStr(t.Subject), t.Body, t.IsActive)
	return err
}

func (r *NotificationRepository) UpdateTemplate(ctx context.Context, t *models.NotifTemplate) error {
	ctx, cancel := withTimeout(ctx)
	defer cancel()
	res, err := r.db.ExecContext(ctx, `
		UPDATE dbo.CPNOTIFTEMPLATE
		SET NAME = @p1, CHANNEL = @p2, SUBJECT = @p3, BODY = @p4, IS_ACTIVE = @p5, UPDATED_AT = GETDATE()
		WHERE CODE = @p6`,
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
	res, err := r.db.ExecContext(ctx, `DELETE FROM dbo.CPNOTIFTEMPLATE WHERE CODE = @p1`, code)
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
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO dbo.CPNOTIFLOG (CODE, TEMPLATE_CODE, CHANNEL, RECIPIENT, SUBJECT, BODY, STATUS, ERROR)
		VALUES (@p1, @p2, @p3, @p4, @p5, @p6, @p7, @p8)`,
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
	rows, err := r.db.QueryContext(ctx, `
		SELECT CODE, TEMPLATE_CODE, CHANNEL, RECIPIENT, SUBJECT, BODY, STATUS, ERROR, CREATED_AT
		FROM dbo.CPNOTIFLOG
		ORDER BY ID DESC
		OFFSET @p1 ROWS FETCH NEXT @p2 ROWS ONLY`, offset, limit)
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
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM dbo.CPNOTIFLOG
		WHERE STATUS = 'SENT' AND CREATED_AT >= DATEADD(HOUR, -@p1, GETDATE())`, hours).Scan(&n)
	return n, err
}
