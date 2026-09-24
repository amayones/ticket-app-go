package services

import (
	"context"
	"fmt"

	"golang-backend/models"
	"golang-backend/repositories"
)

// Session management di atas CPREFRESHTOKEN (hash tidak pernah diekspos).

func (s *UserService) ListSessions(ctx context.Context, userCode string) ([]models.Session, error) {
	sessions, err := s.refresh.ListByUserCode(ctx, userCode)
	if err != nil {
		return nil, fmt.Errorf("list sessions: %w", err)
	}
	return sessions, nil
}

func (s *UserService) ListAllSessions(ctx context.Context, limit, offset int) ([]models.Session, error) {
	sessions, err := s.refresh.ListAll(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list all sessions: %w", err)
	}
	return sessions, nil
}

// RevokeSession mencabut satu sesi milik user (pemilik atau SESSION_MANAGE).
func (s *UserService) RevokeSession(ctx context.Context, callerCode string, sessionID int, manageAll bool) error {
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
		return ErrInvalidRefresh
	}
	if owner != callerCode && !manageAll {
		return ErrForbidden
	}
	deleted, err := s.refresh.DeleteByID(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("revoke session: %w", err)
	}
	if !deleted {
		return ErrInvalidRefresh
	}
	return nil
}
