package services

import (
	"context"
	"fmt"

	"golang-backend/models"
	"golang-backend/repositories"
	"golang-backend/utils"
)

// SyslogService mencatat error/sistem ke CPSYSLOG.
type SyslogService struct {
	repo repositories.SyslogRepositoryInterface
}

func NewSyslogService(repo repositories.SyslogRepositoryInterface) *SyslogService {
	return &SyslogService{repo: repo}
}

func (s *SyslogService) write(ctx context.Context, level, source, message string) error {
	code, err := utils.GenerateCode(utils.SyslogCodePrefix)
	if err != nil {
		return err
	}
	if len(message) > 1000 {
		message = message[:1000]
	}
	return s.repo.Create(ctx, &models.SysLog{Code: code, Level: level, Source: source, Message: message})
}

func (s *SyslogService) Error(ctx context.Context, source, message string) error {
	return s.write(ctx, models.SyslogError, source, message)
}

func (s *SyslogService) Warn(ctx context.Context, source, message string) error {
	return s.write(ctx, models.SyslogWarn, source, message)
}

func (s *SyslogService) Info(ctx context.Context, source string, message string) error {
	return s.write(ctx, models.SyslogInfo, source, message)
}

func (s *SyslogService) List(ctx context.Context, f models.SyslogFilter) ([]models.SysLog, error) {
	logs, err := s.repo.List(ctx, f)
	if err != nil {
		return nil, fmt.Errorf("list syslog: %w", err)
	}
	return logs, nil
}

func (s *SyslogService) CountSince(ctx context.Context, hours int, level string) (int, error) {
	return s.repo.CountSince(ctx, hours, level)
}

func (s *SyslogService) Prune(ctx context.Context, days int) (int64, error) {
	if days < 1 {
		days = 30
	}
	return s.repo.PruneBefore(ctx, days)
}
