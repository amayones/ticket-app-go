package services

import (
	"context"
	"fmt"
	"net"
	"strings"

	"golang-backend/models"
	"golang-backend/repositories"
	"golang-backend/utils"
)

// AuditService mencatat jejak aksi ke CPAUDITLOG (best-effort dari handler).
type AuditService struct {
	repo repositories.AuditRepositoryInterface
}

func NewAuditService(repo repositories.AuditRepositoryInterface) *AuditService {
	return &AuditService{repo: repo}
}

// Log menyimpan satu baris audit. Gagal tulis tidak menggagalkan request
// (caller mengabaikan error), tapi error dikembalikan untuk testing.
func (s *AuditService) Log(ctx context.Context, actorCode, action, entity, entityCode, detail, ip string) error {
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

func (s *AuditService) List(ctx context.Context, f models.AuditFilter) ([]models.AuditLog, error) {
	logs, err := s.repo.List(ctx, f)
	if err != nil {
		return nil, fmt.Errorf("list audit: %w", err)
	}
	return logs, nil
}

func (s *AuditService) CountSince(ctx context.Context, hours int) (int, error) {
	return s.repo.CountSince(ctx, hours)
}

// ClientIP mengambil host dari RemoteAddr (tanpa port).
func ClientIP(remoteAddr string) string {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return strings.TrimSpace(remoteAddr)
	}
	return host
}
