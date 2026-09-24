package security

import (
	"context"
	"fmt"

	"golang-backend/models"
)

// Security Center: agregasi hitung lintas menu untuk ringkasan.
// Reader dipenuhi service fitur lain (tanpa import antar-fitur).
type usersCounter interface {
	CountUsers(ctx context.Context) (int, error)
}

type rolesCounter interface {
	CountRoles(ctx context.Context) (int, error)
}

type sessionsCounter interface {
	CountActiveSessions(ctx context.Context) (int, error)
}

type auditCounter interface {
	CountSince(ctx context.Context, hours int) (int, error)
}

type syslogCounter interface {
	CountSince(ctx context.Context, hours int, level string) (int, error)
}

type notifCounter interface {
	CountSentSince(ctx context.Context, hours int) (int, error)
}

type ServiceInterface interface {
	Summary(ctx context.Context) (models.SecuritySummary, error)
}

type Service struct {
	users    usersCounter
	roles    rolesCounter
	sessions sessionsCounter
	audit    auditCounter
	sys      syslogCounter
	notif    notifCounter
}

func NewService(
	users usersCounter,
	roles rolesCounter,
	sessions sessionsCounter,
	audit auditCounter,
	sys syslogCounter,
	notif notifCounter,
) *Service {
	return &Service{users: users, roles: roles, sessions: sessions, audit: audit, sys: sys, notif: notif}
}

func (s *Service) Summary(ctx context.Context) (models.SecuritySummary, error) {
	var out models.SecuritySummary
	var err error
	if out.TotalUsers, err = s.users.CountUsers(ctx); err != nil {
		return out, fmt.Errorf("count users: %w", err)
	}
	if out.TotalRoles, err = s.roles.CountRoles(ctx); err != nil {
		return out, fmt.Errorf("count roles: %w", err)
	}
	if out.ActiveSessions, err = s.sessions.CountActiveSessions(ctx); err != nil {
		return out, fmt.Errorf("count sessions: %w", err)
	}
	if out.AuditLast24h, err = s.audit.CountSince(ctx, 24); err != nil {
		return out, fmt.Errorf("count audit: %w", err)
	}
	if out.ErrorsLast24h, err = s.sys.CountSince(ctx, 24, models.SyslogError); err != nil {
		return out, fmt.Errorf("count errors: %w", err)
	}
	if out.NotifSentLast24h, err = s.notif.CountSentSince(ctx, 24); err != nil {
		return out, fmt.Errorf("count notifications: %w", err)
	}
	return out, nil
}
