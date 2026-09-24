package audit

import (
	"context"
	"fmt"

	"golang-backend/models"
	"golang-backend/utils"
)

// AuditService mencatat jejak aksi ke CPAUDITLOG (best-effort dari handler).
type ServiceInterface interface {
	Log(ctx context.Context, actorCode, action, entity, entityCode, detail, ip string) error
	List(ctx context.Context, f models.AuditFilter) ([]models.AuditLog, error)
	CountSince(ctx context.Context, hours int) (int, error)
}

type Service struct {
	repo RepositoryInterface
}

func NewService(repo RepositoryInterface) *Service {
	return &Service{repo: repo}
}

// Log menyimpan satu baris audit. Gagal tulis tidak menggagalkan request
// (caller mengabaikan error), tapi error dikembalikan untuk testing.
func (s *Service) Log(ctx context.Context, actorCode, action, entity, entityCode, detail, ip string) error {
	code, err := utils.GenerateCode(utils.AuditCodePrefix)
	if err != nil {
		return err
	}
	return s.repo.Create(ctx, &models.AuditLog{
		Code:       code,
		ActorCode:  actorCode,
		Action:     action,
		Entity:     entity,
		EntityCode: entityCode,
		Detail:     detail,
		IPAddress:  ip,
	})
}

func (s *Service) List(ctx context.Context, f models.AuditFilter) ([]models.AuditLog, error) {
	logs, err := s.repo.List(ctx, f)
	if err != nil {
		return nil, fmt.Errorf("list audit: %w", err)
	}
	return logs, nil
}

func (s *Service) CountSince(ctx context.Context, hours int) (int, error) {
	return s.repo.CountSince(ctx, hours)
}
