package syslog

import (
	"context"
	"fmt"

	"golang-backend/models"
	"golang-backend/utils"
)

// SyslogService mencatat error/sistem ke CPSYSLOG.
type ServiceInterface interface {
	Error(ctx context.Context, source, message string) error
	List(ctx context.Context, f models.SyslogFilter) ([]models.SysLog, error)
	CountSince(ctx context.Context, hours int, level string) (int, error)
	Prune(ctx context.Context, days int) (int64, error)
}

type Service struct {
	repo RepositoryInterface
}

func NewService(repo RepositoryInterface) *Service {
	return &Service{repo: repo}
}

func (s *Service) write(ctx context.Context, level, source, message string) error {
	code, err := utils.GenerateCode(utils.SyslogCodePrefix)
	if err != nil {
		return err
	}
	if len(message) > 1000 {
		message = message[:1000]
	}
	return s.repo.Create(ctx, &models.SysLog{Code: code, Level: level, Source: source, Message: message})
}

func (s *Service) Error(ctx context.Context, source, message string) error {
	return s.write(ctx, models.SyslogError, source, message)
}

func (s *Service) List(ctx context.Context, f models.SyslogFilter) ([]models.SysLog, error) {
	logs, err := s.repo.List(ctx, f)
	if err != nil {
		return nil, fmt.Errorf("list syslog: %w", err)
	}
	return logs, nil
}

func (s *Service) CountSince(ctx context.Context, hours int, level string) (int, error) {
	return s.repo.CountSince(ctx, hours, level)
}

func (s *Service) Prune(ctx context.Context, days int) (int64, error) {
	if days < 1 {
		days = 30
	}
	return s.repo.PruneBefore(ctx, days)
}
