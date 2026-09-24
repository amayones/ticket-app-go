package sessions

import (
	"context"
	"fmt"

	"golang-backend/models"
	"golang-backend/repositories"
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

// RevokeSession mencabut satu sesi milik user (pemilik atau SESSION_MANAGE).
func (s *Service) RevokeSession(ctx context.Context, callerCode string, sessionID int, manageAll bool) error {
	sessions, err := s.refresh.ListAll(ctx, repositories.MaxListLimit, 0)
	if err != nil {
		return fmt.Errorf("lookup session: %w", err)
	}
	var owner string
	for _, sess := range sessions {
		if sess.ID == sessionID {
			owner = sess.UserCode
			break
		}
	}
	// Fallback: cari di sesi milik sendiri (di luar halaman ListAll).
	if owner == "" {
		mine, err := s.refresh.ListByUserCode(ctx, callerCode)
		if err != nil {
			return fmt.Errorf("lookup session: %w", err)
		}
		for _, sess := range mine {
			if sess.ID == sessionID {
				owner = sess.UserCode
				break
			}
		}
	}
	if owner == "" {
		return services.ErrInvalidRefresh
	}
	if owner != callerCode && !manageAll {
		return services.ErrForbidden
	}
	deleted, err := s.refresh.DeleteByID(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("revoke session: %w", err)
	}
	if !deleted {
		return services.ErrInvalidRefresh
	}
	return nil
}

func (s *Service) CountActiveSessions(ctx context.Context) (int, error) {
	return s.refresh.CountActive(ctx)
}

func (s *Service) CleanupExpiredTokens(ctx context.Context) (int64, error) {
	return s.refresh.DeleteExpired(ctx)
}
