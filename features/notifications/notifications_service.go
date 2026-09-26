package notifications

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"golang-backend/models"
	"golang-backend/services"
	"golang-backend/utils"
)

// NotificationService mengelola template + riwayat pengiriman.
// "Pengiriman" saat ini dicatat ke CPNOTIFLOG (provider email/SMS nyata
// bisa disuntik di sini tanpa mengubah handler).
type ServiceInterface interface {
	ListTemplates(ctx context.Context, activeOnly bool) ([]models.NotifTemplate, error)
	GetTemplate(ctx context.Context, code string) (*models.NotifTemplate, error)
	CreateTemplate(ctx context.Context, name, channel, subject, body string, active bool) (*models.NotifTemplate, error)
	UpdateTemplate(ctx context.Context, code, name, channel, subject, body string, active bool) error
	DeleteTemplate(ctx context.Context, code string) error
	Send(ctx context.Context, req models.NotifSendRequest) (*models.NotifLog, error)
	ListLogs(ctx context.Context, limit, offset int) ([]models.NotifLog, error)
	CountSentSince(ctx context.Context, hours int) (int, error)
}

type Service struct {
	repo RepositoryInterface
}

func NewService(repo RepositoryInterface) *Service {
	return &Service{repo: repo}
}

func validChannel(c string) error {
	switch strings.ToUpper(strings.TrimSpace(c)) {
	case models.ChannelEmail, models.ChannelPush, models.ChannelInApp:
		return nil
	default:
		return services.ErrInvalidChannel
	}
}

func (s *Service) ListTemplates(ctx context.Context, activeOnly bool) ([]models.NotifTemplate, error) {
	t, err := s.repo.ListTemplates(ctx, activeOnly)
	if err != nil {
		return nil, fmt.Errorf("list templates: %w", err)
	}
	return t, nil
}

func (s *Service) CreateTemplate(ctx context.Context, name, channel, subject, body string, active bool) (*models.NotifTemplate, error) {
	name = strings.TrimSpace(name)
	body = strings.TrimSpace(body)
	if name == "" || body == "" {
		return nil, fmt.Errorf("%w: template name and body are required", services.ErrInvalidTemplate)
	}
	channel = strings.ToUpper(strings.TrimSpace(channel))
	if err := validChannel(channel); err != nil {
		return nil, err
	}
	code, err := utils.GenerateCode(utils.NotifTemplatePrefix)
	if err != nil {
		return nil, err
	}
	t := &models.NotifTemplate{
		Code: code, Name: name, Channel: channel,
		Subject: strings.TrimSpace(subject), Body: body, IsActive: active,
	}
	if err := s.repo.CreateTemplate(ctx, t); err != nil {
		return nil, fmt.Errorf("create template: %w", err)
	}
	return t, nil
}

func (s *Service) UpdateTemplate(ctx context.Context, code, name, channel, subject, body string, active bool) error {
	name = strings.TrimSpace(name)
	body = strings.TrimSpace(body)
	if name == "" || body == "" {
		return fmt.Errorf("%w: template name and body are required", services.ErrInvalidTemplate)
	}
	channel = strings.ToUpper(strings.TrimSpace(channel))
	if err := validChannel(channel); err != nil {
		return err
	}
	err := s.repo.UpdateTemplate(ctx, &models.NotifTemplate{
		Code: code, Name: name, Channel: channel,
		Subject: strings.TrimSpace(subject), Body: body, IsActive: active,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return services.ErrTemplateNotFound
		}
		return fmt.Errorf("update template: %w", err)
	}
	return nil
}

func (s *Service) DeleteTemplate(ctx context.Context, code string) error {
	if err := s.repo.DeleteTemplate(ctx, code); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return services.ErrTemplateNotFound
		}
		return fmt.Errorf("delete template: %w", err)
	}
	return nil
}

func (s *Service) GetTemplate(ctx context.Context, code string) (*models.NotifTemplate, error) {
	tmpl, err := s.repo.GetTemplate(ctx, strings.TrimSpace(code))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, services.ErrTemplateNotFound
		}
		return nil, fmt.Errorf("get template: %w", err)
	}
	return tmpl, nil
}

// render mengganti {{var}} dengan nilai variables.
func render(input string, vars map[string]string) string {
	out := input
	for k, v := range vars {
		out = strings.ReplaceAll(out, "{{"+strings.TrimSpace(k)+"}}", v)
	}
	return out
}

// Send merender template + mencatat ke CPNOTIFLOG (status SENT).
func (s *Service) Send(ctx context.Context, req models.NotifSendRequest) (*models.NotifLog, error) {
	recipient := strings.TrimSpace(req.Recipient)
	if recipient == "" {
		return nil, fmt.Errorf("%w: recipient is required", services.ErrInvalidTemplate)
	}
	tmpl, err := s.repo.GetTemplate(ctx, strings.TrimSpace(req.TemplateCode))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, services.ErrTemplateNotFound
		}
		return nil, fmt.Errorf("get template: %w", err)
	}
	if !tmpl.IsActive {
		return nil, fmt.Errorf("%w: template is inactive", services.ErrInvalidTemplate)
	}
	vars := req.Variables
	if vars == nil {
		vars = map[string]string{}
	}
	code, err := utils.GenerateCode(utils.NotifLogPrefix)
	if err != nil {
		return nil, err
	}
	entry := &models.NotifLog{
		Code: code, TemplateCode: tmpl.Code, Channel: tmpl.Channel,
		Recipient: recipient, Subject: render(tmpl.Subject, vars),
		Body: render(tmpl.Body, vars), Status: models.NotifSent,
	}
	if err := s.repo.CreateLog(ctx, entry); err != nil {
		return nil, fmt.Errorf("log notification: %w", err)
	}
	return entry, nil
}

func (s *Service) ListLogs(ctx context.Context, limit, offset int) ([]models.NotifLog, error) {
	logs, err := s.repo.ListLogs(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list notification logs: %w", err)
	}
	return logs, nil
}

func (s *Service) CountSentSince(ctx context.Context, hours int) (int, error) {
	return s.repo.CountSentSince(ctx, hours)
}
