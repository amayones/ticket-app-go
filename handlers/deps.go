package handlers

import (
	"context"

	"golang-backend/models"
)

// Shared dependency interfaces (diimplementasikan services, dimock di test).

type AuditLogger interface {
	Log(ctx context.Context, actorCode, action, entity, entityCode, detail, ip string) error
}

type AuditReader interface {
	AuditLogger
	List(ctx context.Context, f models.AuditFilter) ([]models.AuditLog, error)
	CountSince(ctx context.Context, hours int) (int, error)
}

type SysLogger interface {
	Error(ctx context.Context, source, message string) error
}

type SyslogAdmin interface {
	SysLogger
	List(ctx context.Context, f models.SyslogFilter) ([]models.SysLog, error)
	CountSince(ctx context.Context, hours int, level string) (int, error)
	Prune(ctx context.Context, days int) (int64, error)
}

type Notifier interface {
	ListTemplates(ctx context.Context, activeOnly bool) ([]models.NotifTemplate, error)
	CreateTemplate(ctx context.Context, name, channel, subject, body string, active bool) (*models.NotifTemplate, error)
	UpdateTemplate(ctx context.Context, code, name, channel, subject, body string, active bool) error
	DeleteTemplate(ctx context.Context, code string) error
	Send(ctx context.Context, req models.NotifSendRequest) (*models.NotifLog, error)
	ListLogs(ctx context.Context, limit, offset int) ([]models.NotifLog, error)
}
