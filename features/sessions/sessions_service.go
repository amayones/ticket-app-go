package sessions

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"golang-backend/models"
	"golang-backend/services"
)

// Session management di atas CPREFRESHTOKEN (hash tidak pernah diekspos).
//
// Konvensi menu: query SQL selalu di variabel `query`, lalu di-run,
// lalu hasilnya dipetakan ke response (lihat sessions_repository.go).
type ServiceInterface interface {
	ListSessions(ctx context.Context, userCode string) ([]models.Session, error)
	ListAllSessions(ctx context.Context, limit, offset int) ([]models.Session, error)
	RevokeSession(ctx context.Context, callerCode string, sessionID int, manageAll bool) error
	CountActiveSessions(ctx context.Context) (int, error)
	CleanupExpiredTokens(ctx context.Context) (int64, error)
}

type Service struct {
	refresh RepositoryInterface
}

func NewService(refresh RepositoryInterface) *Service {
	return &Service{refresh: refresh}
}

func (s *Service) ListSessions(ctx context.Context, userCode string) ([]models.Session, error) {
	sessions, err := s.refresh.ListByUserCode(ctx, userCode)
	if err != nil {
		return nil, fmt.Errorf("list sessions: %w", err)
	}
	return sessions, nil
}

func (s *Service) ListAllSessions(ctx context.Context, limit, offset int) ([]models.Session, error) {
	sessions, err := s.refresh.ListAll(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list all sessions: %w", err)
	}
	return sessions, nil
}

// RevokeSession mencabut satu sesi milik user (pemilik atau MENU_SESSIONS).
// Lookup via GetByID (O(1), tidak pecah saat total sesi > MaxListLimit).
func (s *Service) RevokeSession(ctx context.Context, callerCode string, sessionID int, manageAll bool) error {
	sess, err := s.refresh.GetByID(ctx, sessionID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return services.ErrSessionNotFound
		}
		return fmt.Errorf("lookup session: %w", err)
	}
	owner := sess.UserCode
	if owner == "" {
		return services.ErrSessionNotFound
	}
	if owner != callerCode && !manageAll {
		return services.ErrForbidden
	}
	deleted, err := s.refresh.DeleteByID(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("revoke session: %w", err)
	}
	if !deleted {
		return services.ErrSessionNotFound
	}
	return nil
}

func (s *Service) CountActiveSessions(ctx context.Context) (int, error) {
	return s.refresh.CountActive(ctx)
}

func (s *Service) CleanupExpiredTokens(ctx context.Context) (int64, error) {
	return s.refresh.DeleteExpired(ctx)
}
