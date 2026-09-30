package event

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"golang-backend/internal/export"
	"golang-backend/models"
	"golang-backend/services"
	"golang-backend/utils"
)

// Event service: validation, slug, status transitions, ticket types,
// attendees. Ownership is always the caller's code ( cross-owner rows
// behave as not found).
type ServiceInterface interface {
	MineList(ctx context.Context, owner, status, category, city, from, to, q, sort string, limit, offset int) ([]models.MyEventItem, error)
	CreateEvent(ctx context.Context, owner string, in models.EventInput) (string, error)
	EventDetail(ctx context.Context, owner, code string) (*models.EventDetailOwner, error)
	UpdateEvent(ctx context.Context, owner, code string, in models.EventInput) (bool, error)
	PublishEvent(ctx context.Context, owner, code string) ([]string, error)
	UnpublishEvent(ctx context.Context, owner, code string) error
	CancelEvent(ctx context.Context, owner, code, reason string) (int, error)
	DeleteEvent(ctx context.Context, owner, code string) error
	DuplicateEvent(ctx context.Context, owner, code string) (string, error)
	TypeList(ctx context.Context, owner, event string) ([]models.TicketTypeRow, models.TicketTypeSummary, error)
	CreateType(ctx context.Context, owner, event string, in models.TicketTypeInput) (string, error)
	UpdateType(ctx context.Context, owner, event, typeCode string, in models.TicketTypeInput) error
	SetTypeActive(ctx context.Context, owner, event, typeCode string, active bool) error
	DeleteType(ctx context.Context, owner, event, typeCode string) error
	ReorderTypes(ctx context.Context, owner, event string, codes []string) error
	Attendees(ctx context.Context, owner, event, typeCode, attendance, payment, q string, limit, offset int) ([]models.AttendeeRow, models.AttendeeSummary, error)
	AttendeeExport(ctx context.Context, owner, event, typeCode, attendance, payment, q string) ([]export.Table, error)
}

type Service struct {
	repo RepositoryInterface
}

func NewService(repo RepositoryInterface) *Service {
	return &Service{repo: repo}
}

func nowStr() string { return time.Now().Format("2006-01-02 15:04:05") }

func plusOne(d string) string {
	t, err := time.Parse("2006-01-02", d)
	if err != nil {
		return d
	}
	return t.AddDate(0, 0, 1).Format("2006-01-02")
}

// parseDateTime accepts "YYYY-MM-DD HH:MM" (datetime-local) and full ISO.
func parseDateTime(s string) (time.Time, error) {
	for _, layout := range []string{"2006-01-02 15:04", "2006-01-02 15:04:05", time.RFC3339} {
		if t, err := time.ParseInLocation(layout, s, time.Local); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("must be YYYY-MM-DD HH:MM")
}

var slugReplacer = regexp.MustCompile(`[^a-z0-9]+`)

// slugify renders URL-safe slugs; uniqueness is enforced by the caller.
func slugify(title string) string {
	s := strings.ToLower(strings.TrimSpace(title))
	s = slugReplacer.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if len(s) > 160 {
		s = strings.Trim(s[:160], "-")
	}
	if s == "" {
		s = "event"
	}
	return s
}

func (s *Service) uniqueSlug(ctx context.Context, title, exclude string) (string, error) {
	base := slugify(title)
	slug := base
	for i := 2; ; i++ {
		taken, err := s.repo.SlugExists(ctx, slug, exclude)
		if err != nil {
			return "", fmt.Errorf("check slug: %w", err)
		}
		if !taken {
			return slug, nil
		}
		slug = fmt.Sprintf("%s-%d", base, i)
	}
}

// validateInput checks the shared create/update rules.
func (s *Service) validateInput(ctx context.Context, in *models.EventInput, isNew bool) error {
	in.Title = strings.TrimSpace(in.Title)
	if len(in.Title) < 3 || len(in.Title) > 200 {
		return fmt.Errorf("%w: title must be 3-200 characters", services.ErrInvalidFilter)
	}
	in.CategoryCode = strings.TrimSpace(strings.ToUpper(in.CategoryCode))
	if in.CategoryCode == "" {
		return fmt.Errorf("%w: category is required", services.ErrInvalidFilter)
	}
	if ok, err := s.repo.CategoryExists(ctx, in.CategoryCode); err != nil {
		return fmt.Errorf("check category: %w", err)
	} else if !ok {
		return fmt.Errorf("%w: unknown category", services.ErrInvalidFilter)
	}
	in.CityCode = strings.TrimSpace(strings.ToUpper(in.CityCode))
	if in.CityCode == "" {
		return fmt.Errorf("%w: city is required", services.ErrInvalidFilter)
	}
	if ok, err := s.repo.CityExists(ctx, in.CityCode); err != nil {
		return fmt.Errorf("check city: %w", err)
	} else if !ok {
		return fmt.Errorf("%w: unknown city", services.ErrInvalidFilter)
	}
	in.Venue = strings.TrimSpace(in.Venue)
	if in.Venue == "" || len(in.Venue) > 200 {
		return fmt.Errorf("%w: venue is required (max 200 chars)", services.ErrInvalidFilter)
	}
	if len(in.Description) > 8000 {
		return fmt.Errorf("%w: description must be at most 8000 characters", services.ErrInvalidFilter)
	}
	if len(in.Terms) > 8000 {
		return fmt.Errorf("%w: terms must be at most 8000 characters", services.ErrInvalidFilter)
	}
	if len(in.PosterURL) > 500 {
		return fmt.Errorf("%w: poster URL too long", services.ErrInvalidFilter)
	}
	if in.HasLatitude && (in.Latitude < -90 || in.Latitude > 90) {
		return fmt.Errorf("%w: latitude must be -90..90", services.ErrInvalidFilter)
	}
	if in.HasLongitude && (in.Longitude < -180 || in.Longitude > 180) {
		return fmt.Errorf("%w: longitude must be -180..180", services.ErrInvalidFilter)
	}
	start, err := parseDateTime(strings.TrimSpace(in.StartAt))
	if err != nil {
		return fmt.Errorf("%w: invalid start: %v", services.ErrInvalidFilter, err)
	}
	var end time.Time
	hasEnd := strings.TrimSpace(in.EndAt) != ""
	if hasEnd {
		end, err = parseDateTime(strings.TrimSpace(in.EndAt))
		if err != nil {
			return fmt.Errorf("%w: invalid end: %v", services.ErrInvalidFilter, err)
		}
		if !end.After(start) {
			return fmt.Errorf("%w: end must be after start", services.ErrInvalidFilter)
		}
	}
	if isNew && start.Before(time.Now().Add(-5*time.Minute)) {
		return fmt.Errorf("%w: start must not be in the past for new events", services.ErrInvalidFilter)
	}
	in.StartAt = start.Format("2006-01-02 15:04")
	if hasEnd {
		in.EndAt = end.Format("2006-01-02 15:04")
	} else {
		in.EndAt = ""
	}
	in.Description = utils.SanitizeHTML(in.Description)
	in.Terms = utils.SanitizeHTML(in.Terms)
	return nil
}

func (s *Service) MineList(ctx context.Context, owner, status, category, city, from, to, q, sort string, limit, offset int) ([]models.MyEventItem, error) {
	_ = s.repo.MarkEnded(ctx, owner, nowStr()) // opportunistic; list must not fail on it
	status = strings.TrimSpace(strings.ToUpper(status))
	if status != "" {
		switch status {
		case models.EventDraft, models.EventPublished, models.EventCompleted,
			models.EventEnded, models.EventCancelled:
		default:
			return nil, fmt.Errorf("%w: unknown status", services.ErrInvalidFilter)
		}
	}
	sort = strings.ToLower(strings.TrimSpace(sort))
	if sort != "" && sort != "nearest" && sort != "newest" && sort != "popular" {
		return nil, fmt.Errorf("%w: sort must be nearest, newest or popular", services.ErrInvalidFilter)
	}
	if from != "" && !utils.IsValidDate(from) {
		return nil, fmt.Errorf("%w: from must be YYYY-MM-DD", services.ErrInvalidFilter)
	}
	if to != "" && !utils.IsValidDate(to) {
		return nil, fmt.Errorf("%w: to must be YYYY-MM-DD", services.ErrInvalidFilter)
	}
	toPlusOne := ""
	if to != "" {
		toPlusOne = plusOne(to)
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	items, err := s.repo.MineList(ctx, owner, status,
		strings.TrimSpace(strings.ToUpper(category)),
		strings.TrimSpace(strings.ToUpper(city)),
		from, toPlusOne, strings.TrimSpace(q), sort, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list events: %w", err)
	}
	return items, nil
}

func (s *Service) CreateEvent(ctx context.Context, owner string, in models.EventInput) (string, error) {
	if err := s.validateInput(ctx, &in, true); err != nil {
		return "", err
	}
	slug, err := s.uniqueSlug(ctx, in.Title, "")
	if err != nil {
		return "", err
	}
	code, err := utils.GenerateCode(utils.EventCodePrefix)
	if err != nil {
		return "", err
	}
	if err := s.repo.CreateEvent(ctx, owner, code, slug, in); err != nil {
		return "", fmt.Errorf("create event: %w", err)
	}
	return code, nil
}

func (s *Service) loadOwned(ctx context.Context, owner, code string) (*models.EventDetailOwner, error) {
	_ = s.repo.MarkEnded(ctx, owner, nowStr())
	ev, err := s.repo.EventByCode(ctx, owner, code)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, services.ErrEventNotFound
		}
		return nil, fmt.Errorf("get event: %w", err)
	}
	return ev, nil
}

func (s *Service) EventDetail(ctx context.Context, owner, code string) (*models.EventDetailOwner, error) {
	ev, err := s.loadOwned(ctx, owner, code)
	if err != nil {
		return nil, err
	}
	missing, err := s.checklist(ctx, owner, ev)
	if err != nil {
		return nil, err
	}
	ev.Missing = missing
	return ev, nil
}

// checklist returns unmet publish requirements (empty = ready).
func (s *Service) checklist(ctx context.Context, owner string, ev *models.EventDetailOwner) ([]string, error) {
	missing := []string{}
	types, err := s.repo.TypeList(ctx, owner, ev.Code)
	if err != nil {
		return nil, fmt.Errorf("list types: %w", err)
	}
	active := 0
	for _, t := range types {
		if t.IsActive {
			active++
		}
	}
	if active == 0 {
		missing = append(missing, "Minimal 1 tipe tiket aktif")
	}
	start, err := parseDateTime(ev.StartAt)
	if err != nil || (!evEndValid(ev)) {
		missing = append(missing, "Jadwal mulai/selesai tidak valid")
	} else if start.Before(time.Now()) {
		missing = append(missing, "Jadwal mulai sudah lewat")
	}
	if strings.TrimSpace(ev.PosterURL) == "" {
		missing = append(missing, "Poster/banner belum diisi")
	}
	return missing, nil
}

func evEndValid(ev *models.EventDetailOwner) bool {
	if strings.TrimSpace(ev.EndAt) == "" {
		return true
	}
	return ev.EndAt > ev.StartAt
}

func (s *Service) UpdateEvent(ctx context.Context, owner, code string, in models.EventInput) (bool, error) {
	ev, err := s.loadOwned(ctx, owner, code)
	if err != nil {
		return false, err
	}
	if ev.Status != models.EventDraft && ev.Status != models.EventPublished {
		return false, fmt.Errorf("%w: only draft or published events can be edited", services.ErrEventState)
	}
	if err := s.validateInput(ctx, &in, false); err != nil {
		return false, err
	}
	notify := false
	if ev.Status == models.EventPublished && ev.Sold > 0 {
		if in.StartAt != ev.StartAt || in.Venue != ev.Venue {
			notify = true
		}
	}
	if err := s.repo.UpdateEvent(ctx, owner, code, in); err != nil {
		return false, fmt.Errorf("update event: %w", err)
	}
	return notify, nil
}

func (s *Service) PublishEvent(ctx context.Context, owner, code string) ([]string, error) {
	ev, err := s.loadOwned(ctx, owner, code)
	if err != nil {
		return nil, err
	}
	if ev.Status != models.EventDraft {
		return nil, fmt.Errorf("%w: only draft events can be published", services.ErrEventState)
	}
	missing, err := s.checklist(ctx, owner, ev)
	if err != nil {
		return nil, err
	}
	if len(missing) > 0 {
		return missing, services.ErrEventNotReady
	}
	if err := s.repo.SetStatus(ctx, owner, code, models.EventPublished, ""); err != nil {
		return nil, fmt.Errorf("publish event: %w", err)
	}
	return nil, nil
}

func (s *Service) UnpublishEvent(ctx context.Context, owner, code string) error {
	ev, err := s.loadOwned(ctx, owner, code)
	if err != nil {
		return err
	}
	if ev.Status != models.EventPublished {
		return fmt.Errorf("%w: only published events can be unpublished", services.ErrEventState)
	}
	if ev.Sold > 0 {
		return fmt.Errorf("%w: cannot unpublish with tickets sold", services.ErrEventState)
	}
	if err := s.repo.SetStatus(ctx, owner, code, models.EventDraft, ""); err != nil {
		return fmt.Errorf("unpublish event: %w", err)
	}
	return nil
}

func (s *Service) CancelEvent(ctx context.Context, owner, code, reason string) (int, error) {
	ev, err := s.loadOwned(ctx, owner, code)
	if err != nil {
		return 0, err
	}
	if ev.Status != models.EventDraft && ev.Status != models.EventPublished {
		return 0, fmt.Errorf("%w: only draft or published events can be cancelled", services.ErrEventState)
	}
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return 0, fmt.Errorf("%w: cancel reason is required", services.ErrInvalidFilter)
	}
	affected, err := s.repo.AffectedBuyers(ctx, code)
	if err != nil {
		return 0, fmt.Errorf("count buyers: %w", err)
	}
	if err := s.repo.SetStatus(ctx, owner, code, models.EventCancelled, reason); err != nil {
		return 0, fmt.Errorf("cancel event: %w", err)
	}
	return affected, nil
}

func (s *Service) DeleteEvent(ctx context.Context, owner, code string) error {
	ev, err := s.loadOwned(ctx, owner, code)
	if err != nil {
		return err
	}
	if ev.Status != models.EventDraft {
		return fmt.Errorf("%w: only draft events can be deleted", services.ErrEventState)
	}
	published, err := s.repo.WasPublished(ctx, owner, code)
	if err != nil {
		return fmt.Errorf("check publish history: %w", err)
	}
	if published {
		return fmt.Errorf("%w: already published events cannot be deleted", services.ErrEventState)
	}
	if err := s.repo.SoftDelete(ctx, owner, code); err != nil {
		return fmt.Errorf("delete event: %w", err)
	}
	return nil
}

func (s *Service) UpdateType(ctx context.Context, owner, event, typeCode string, in models.TicketTypeInput) error {
	if _, err := s.validateType(ctx, owner, event, &in, typeCode); err != nil {
		return err
	}
	sold, err := s.repo.TypeSold(ctx, typeCode)
	if err != nil {
		return fmt.Errorf("type sold: %w", err)
	}
	if in.Quota < sold {
		return fmt.Errorf("%w: quota must not go below sold tickets (%d)", services.ErrInvalidFilter, sold)
	}
	if err := s.repo.UpdateType(ctx, owner, event, typeCode, in); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return services.ErrEventNotFound
		}
		return fmt.Errorf("update type: %w", err)
	}
	return nil
}

func (s *Service) DuplicateEvent(ctx context.Context, owner, code string) (string, error) {
	ev, err := s.loadOwned(ctx, owner, code)
	if err != nil {
		return "", err
	}
	newCode, err := utils.GenerateCode(utils.EventCodePrefix)
	if err != nil {
		return "", err
	}
	newSlug, err := s.uniqueSlug(ctx, ev.Title+" copy", "")
	if err != nil {
		return "", err
	}
	if err := s.repo.Duplicate(ctx, owner, code, newCode, newSlug); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", services.ErrEventNotFound
		}
		return "", fmt.Errorf("duplicate event: %w", err)
	}
	return newCode, nil
}

func (s *Service) SetTypeActive(ctx context.Context, owner, event, typeCode string, active bool) error {
	if _, err := s.loadOwned(ctx, owner, event); err != nil {
		return err
	}
	if err := s.repo.SetTypeActive(ctx, owner, event, typeCode, active); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return services.ErrEventNotFound
		}
		return fmt.Errorf("set type active: %w", err)
	}
	return nil
}

func (s *Service) DeleteType(ctx context.Context, owner, event, typeCode string) error {
	if _, err := s.loadOwned(ctx, owner, event); err != nil {
		return err
	}
	sold, err := s.repo.TypeSold(ctx, typeCode)
	if err != nil {
		return fmt.Errorf("type sold: %w", err)
	}
	if sold > 0 {
		return fmt.Errorf("%w: types with sales cannot be deleted, deactivate instead", services.ErrInvalidFilter)
	}
	if err := s.repo.DeleteType(ctx, owner, event, typeCode); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return services.ErrEventNotFound
		}
		return fmt.Errorf("delete type: %w", err)
	}
	return nil
}

func (s *Service) ReorderTypes(ctx context.Context, owner, event string, codes []string) error {
	if _, err := s.loadOwned(ctx, owner, event); err != nil {
		return err
	}
	types, err := s.repo.TypeList(ctx, owner, event)
	if err != nil {
		return fmt.Errorf("list types: %w", err)
	}
	known := make(map[string]bool, len(types))
	for _, t := range types {
		known[t.Code] = true
	}
	seen := make(map[string]bool, len(codes))
	for _, c := range codes {
		if c == "" || !known[c] || seen[c] {
			return fmt.Errorf("%w: reorder must list each type code exactly once", services.ErrInvalidFilter)
		}
		seen[c] = true
	}
	if len(codes) != len(types) {
		return fmt.Errorf("%w: reorder must list each type code exactly once", services.ErrInvalidFilter)
	}
	if err := s.repo.ReorderTypes(ctx, owner, event, codes); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return services.ErrEventNotFound
		}
		return fmt.Errorf("reorder types: %w", err)
	}
	return nil
}

func (s *Service) Attendees(ctx context.Context, owner, event, typeCode, attendance, payment, q string, limit, offset int) ([]models.AttendeeRow, models.AttendeeSummary, error) {
	if _, err := s.loadOwned(ctx, owner, event); err != nil {
		return nil, models.AttendeeSummary{}, err
	}
	attendance = strings.TrimSpace(strings.ToUpper(attendance))
	if attendance != "" && attendance != "HADIR" && attendance != "BELUM" {
		return nil, models.AttendeeSummary{}, fmt.Errorf("%w: attendance must be HADIR or BELUM", services.ErrInvalidFilter)
	}
	payment = strings.TrimSpace(strings.ToUpper(payment))
	if payment != "" {
		switch payment {
		case "PENDING", "PAID", "CANCELLED", "REFUNDED":
		default:
			return nil, models.AttendeeSummary{}, fmt.Errorf("%w: unknown payment status", services.ErrInvalidFilter)
		}
	}
	if limit <= 0 || limit > 500 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	rows, err := s.repo.Attendees(ctx, owner, event, strings.TrimSpace(typeCode), attendance, payment, strings.TrimSpace(q), limit, offset)
	if err != nil {
		return nil, models.AttendeeSummary{}, fmt.Errorf("list attendees: %w", err)
	}
	sum, err := s.repo.AttendeeSummary(ctx, owner, event)
	if err != nil {
		return nil, models.AttendeeSummary{}, fmt.Errorf("attendee summary: %w", err)
	}
	return rows, sum, nil
}

// AttendeeExport flattens the filtered attendee list for CSV.
func (s *Service) AttendeeExport(ctx context.Context, owner, event, typeCode, attendance, payment, q string) ([]export.Table, error) {
	rows, _, err := s.Attendees(ctx, owner, event, typeCode, attendance, payment, q, 5000, 0)
	if err != nil {
		return nil, err
	}
	lines := make([][]string, 0, len(rows))
	for _, a := range rows {
		lines = append(lines, []string{a.TicketCode, a.BuyerName, a.BuyerUsername, a.Email, a.Phone,
			a.TicketType, a.OrderCode, a.PaymentStatus, a.Attendance, a.CheckedInAt, a.PurchaseDate})
	}
	return []export.Table{{
		Title:   fmt.Sprintf("Attendees %s", event),
		Headers: []string{"Ticket", "Name", "Username", "Email", "Phone", "Type", "Order", "Payment", "Attendance", "CheckedIn", "Purchased"},
		Rows:    lines,
	}}, nil
}

// validateType checks ticket type rules (create + update share them).
func (s *Service) validateType(ctx context.Context, owner, event string, in *models.TicketTypeInput, exclude string) (string, error) {
	ev, err := s.loadOwned(ctx, owner, event)
	if err != nil {
		return "", err
	}
	in.Name = strings.TrimSpace(in.Name)
	if in.Name == "" || len(in.Name) > 100 {
		return "", fmt.Errorf("%w: type name is required (max 100 chars)", services.ErrInvalidFilter)
	}
	if taken, err := s.repo.TypeNameTaken(ctx, event, in.Name, exclude); err != nil {
		return "", fmt.Errorf("check type name: %w", err)
	} else if taken {
		return "", fmt.Errorf("%w: type name already exists in this event", services.ErrInvalidFilter)
	}
	if in.Price < 0 {
		return "", fmt.Errorf("%w: price must not be negative", services.ErrInvalidFilter)
	}
	if in.Quota < 1 {
		return "", fmt.Errorf("%w: quota must be at least 1", services.ErrInvalidFilter)
	}
	if in.MaxPerPerson != nil && *in.MaxPerPerson < 1 {
		return "", fmt.Errorf("%w: max per person must be at least 1", services.ErrInvalidFilter)
	}
	var saleFrom, saleTo time.Time
	if strings.TrimSpace(in.SaleStartAt) != "" {
		saleFrom, err = parseDateTime(strings.TrimSpace(in.SaleStartAt))
		if err != nil {
			return "", fmt.Errorf("%w: invalid sale start: %v", services.ErrInvalidFilter, err)
		}
		in.SaleStartAt = saleFrom.Format("2006-01-02 15:04")
	}
	if strings.TrimSpace(in.SaleEndAt) != "" {
		saleTo, err = parseDateTime(strings.TrimSpace(in.SaleEndAt))
		if err != nil {
			return "", fmt.Errorf("%w: invalid sale end: %v", services.ErrInvalidFilter, err)
		}
		in.SaleEndAt = saleTo.Format("2006-01-02 15:04")
	}
	if !saleFrom.IsZero() && !saleTo.IsZero() && !saleTo.After(saleFrom) {
		return "", fmt.Errorf("%w: sale end must be after sale start", services.ErrInvalidFilter)
	}
	if strings.TrimSpace(ev.EndAt) != "" && !saleTo.IsZero() {
		if evEnd, err := parseDateTime(ev.EndAt); err == nil && saleTo.After(evEnd) {
			return "", fmt.Errorf("%w: sale must end no later than the event", services.ErrInvalidFilter)
		}
	}
	in.Description = utils.SanitizeHTML(strings.TrimSpace(in.Description))
	return ev.Status, nil
}

func (s *Service) TypeList(ctx context.Context, owner, event string) ([]models.TicketTypeRow, models.TicketTypeSummary, error) {
	ev, err := s.loadOwned(ctx, owner, event)
	if err != nil {
		return nil, models.TicketTypeSummary{}, err
	}
	types, err := s.repo.TypeList(ctx, owner, event)
	if err != nil {
		return nil, models.TicketTypeSummary{}, fmt.Errorf("list types: %w", err)
	}
	sum := models.TicketTypeSummary{EventCode: event, EventTitle: ev.Title, EventStatus: ev.Status}
	for _, t := range types {
		sum.TotalQuota += t.Quota
		sum.TotalSold += t.Sold
		if t.IsActive {
			sum.ActiveCount++
			sum.Potential += float64(t.Quota) * t.Price
		}
	}
	missing, err := s.checklist(ctx, owner, ev)
	if err != nil {
		return nil, models.TicketTypeSummary{}, err
	}
	sum.Missing = missing
	sum.PublishReady = ev.Status == models.EventDraft && len(missing) == 0
	return types, sum, nil
}

func (s *Service) CreateType(ctx context.Context, owner, event string, in models.TicketTypeInput) (string, error) {
	if _, err := s.validateType(ctx, owner, event, &in, ""); err != nil {
		return "", err
	}
	sortOrder := 99
	if in.SortOrder != nil && *in.SortOrder > 0 {
		sortOrder = *in.SortOrder
	} else if next, err := s.repo.NextTypeSort(ctx, event); err == nil {
		sortOrder = next
	}
	code, err := utils.GenerateCode(utils.TicketTypePrefix)
	if err != nil {
		return "", err
	}
	if err := s.repo.CreateType(ctx, owner, event, code, in, sortOrder); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", services.ErrEventNotFound
		}
		return "", fmt.Errorf("create type: %w", err)
	}
	return code, nil
}

