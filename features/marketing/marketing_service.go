package marketing

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"time"

	"golang-backend/models"
	"golang-backend/services"
	"golang-backend/utils"
)

type ServiceInterface interface {
	ListPromos(ctx context.Context, organizer, status, q string, limit, offset int) ([]models.PromoRow, error)
	PromoDetail(ctx context.Context, organizer, code string) (*models.PromoRow, []models.PromoUsageRow, error)
	CreatePromo(ctx context.Context, organizer string, in models.PromoInput) (string, error)
	UpdatePromo(ctx context.Context, organizer, code string, in models.PromoInput) error
	DeletePromo(ctx context.Context, organizer, code string) error
	SetPromoActive(ctx context.Context, organizer, code string, active bool) error
	GenerateCode() string
	Positions(ctx context.Context) ([]models.AdPosition, error)
	ListAds(ctx context.Context, organizer, status, q string, limit, offset int) ([]models.AdCampaignRow, error)
	AdDetail(ctx context.Context, organizer, code string) (*models.AdCampaignRow, []models.AdDailyPoint, error)
	CreateAd(ctx context.Context, organizer string, in models.AdCampaignInput) (string, error)
	UpdateAd(ctx context.Context, organizer, code string, in models.AdCampaignInput) error
	SetAdStatus(ctx context.Context, organizer, code, action, note string) error
	SimulatePay(ctx context.Context, organizer, code string) error
	ListNews(ctx context.Context, status, category, q string, limit, offset int) ([]models.NewsRow, error)
	NewsDetail(ctx context.Context, code string) (*models.NewsRow, error)
	CreateNews(ctx context.Context, author string, in models.NewsInput) (string, error)
	UpdateNews(ctx context.Context, code string, in models.NewsInput) error
	DeleteNews(ctx context.Context, code string) error
	SetNewsStatus(ctx context.Context, code, status string) error
	NewsCategories(ctx context.Context) ([]models.NewsCategory, error)
	EventCategories(ctx context.Context) ([]models.EventCategory, error)
	OwnedEvents(ctx context.Context, organizer string) ([]OwnedEvent, error)
}

type Service struct {
	repo   RepositoryInterface
	appEnv string
}

func NewService(repo RepositoryInterface, appEnv string) *Service {
	return &Service{repo: repo, appEnv: appEnv}
}

func parseDateTime(s string) (time.Time, error) {
	for _, layout := range []string{"2006-01-02 15:04", "2006-01-02 15:04:05", time.RFC3339, "2006-01-02"} {
		if t, err := time.ParseInLocation(layout, strings.TrimSpace(s), time.Local); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("must be YYYY-MM-DD HH:MM")
}

var slugRe = regexp.MustCompile(`[^a-z0-9]+`)

func slugify(title string) string {
	s := strings.ToLower(strings.TrimSpace(title))
	s = slugRe.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if len(s) > 160 {
		s = strings.Trim(s[:160], "-")
	}
	if s == "" {
		s = "news"
	}
	return s
}

func (s *Service) uniqueSlug(ctx context.Context, title, exclude string) (string, error) {
	base := slugify(title)
	slug := base
	for i := 2; ; i++ {
		taken, err := s.repo.NewsSlugExists(ctx, slug, exclude)
		if err != nil {
			return "", err
		}
		if !taken {
			return slug, nil
		}
		slug = fmt.Sprintf("%s-%d", base, i)
	}
}

var promoCodeRe = regexp.MustCompile(`^[A-Z0-9_-]+$`)

func (s *Service) GenerateCode() string {
	letters := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	b := make([]rune, 8)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func (s *Service) validatePromo(ctx context.Context, in *models.PromoInput, exclude string) error {
	in.PromoCode = strings.ToUpper(strings.TrimSpace(in.PromoCode))
	if len(in.PromoCode) < 3 || len(in.PromoCode) > 40 {
		return fmt.Errorf("%w: promo code must be 3-40 chars", services.ErrInvalidFilter)
	}
	if !promoCodeRe.MatchString(in.PromoCode) {
		return fmt.Errorf("%w: promo code must be uppercase letters/digits/_/-", services.ErrInvalidFilter)
	}
	if taken, err := s.repo.PromoExists(ctx, in.PromoCode, exclude); err != nil {
		return err
	} else if taken {
		return fmt.Errorf("%w: promo code already exists", services.ErrInvalidFilter)
	}
	in.Description = strings.TrimSpace(in.Description)
	if len(in.Description) > 500 {
		return fmt.Errorf("%w: description max 500 chars", services.ErrInvalidFilter)
	}
	hasPct := in.DiscountPercent > 0
	hasAmt := in.DiscountAmount > 0
	if hasPct == hasAmt {
		return fmt.Errorf("%w: choose percent or amount (exactly one)", services.ErrInvalidFilter)
	}
	if hasPct {
		if in.DiscountPercent < 1 || in.DiscountPercent > 100 {
			return fmt.Errorf("%w: percent must be 1-100", services.ErrInvalidFilter)
		}
		if in.MaxDiscount < 0 {
			return fmt.Errorf("%w: max discount must not be negative", services.ErrInvalidFilter)
		}
	} else {
		if in.DiscountAmount <= 0 {
			return fmt.Errorf("%w: discount amount must be positive", services.ErrInvalidFilter)
		}
		in.MaxDiscount = 0
	}
	if in.MinPurchase < 0 {
		return fmt.Errorf("%w: min purchase must not be negative", services.ErrInvalidFilter)
	}
	if in.MaxUses != nil && *in.MaxUses < 1 {
		return fmt.Errorf("%w: quota must be at least 1", services.ErrInvalidFilter)
	}
	if in.MaxPerUser != nil && *in.MaxPerUser < 1 {
		return fmt.Errorf("%w: max per user must be at least 1", services.ErrInvalidFilter)
	}
	start, err := parseDateTime(in.ValidFrom)
	if err != nil {
		return fmt.Errorf("%w: invalid valid_from: %v", services.ErrInvalidFilter, err)
	}
	end, err := parseDateTime(in.ValidTo)
	if err != nil {
		return fmt.Errorf("%w: invalid valid_to: %v", services.ErrInvalidFilter, err)
	}
	if !end.After(start) {
		return fmt.Errorf("%w: valid_to must be after valid_from", services.ErrInvalidFilter)
	}
	in.ValidFrom = start.Format("2006-01-02 15:04:05")
	in.ValidTo = end.Format("2006-01-02 15:04:05")
	in.ScopeType = strings.ToUpper(strings.TrimSpace(in.ScopeType))
	if in.ScopeType == "" {
		in.ScopeType = models.PromoScopeAll
	}
	switch in.ScopeType {
	case models.PromoScopeAll:
		in.CategoryCodes = nil
		in.EventCodes = nil
	case models.PromoScopeCategory:
		if len(in.CategoryCodes) == 0 {
			return fmt.Errorf("%w: category scope requires at least one category", services.ErrInvalidFilter)
		}
		cats, err := s.repo.EventCategories(ctx)
		if err != nil {
			return err
		}
		set := map[string]bool{}
		for _, c := range cats {
			set[strings.ToUpper(c.Code)] = true
		}
		for i, c := range in.CategoryCodes {
			c = strings.ToUpper(strings.TrimSpace(c))
			if !set[c] {
				return fmt.Errorf("%w: unknown category %s", services.ErrInvalidFilter, c)
			}
			in.CategoryCodes[i] = c
		}
		in.EventCodes = nil
	case models.PromoScopeEvent:
		if len(in.EventCodes) == 0 {
			return fmt.Errorf("%w: event scope requires at least one event", services.ErrInvalidFilter)
		}
		for i, e := range in.EventCodes {
			e = strings.TrimSpace(e)
			if e == "" {
				return fmt.Errorf("%w: event code is required", services.ErrInvalidFilter)
			}
			in.EventCodes[i] = e
		}
		in.CategoryCodes = nil
	default:
		return fmt.Errorf("%w: scope must be ALL, CATEGORY or EVENT", services.ErrInvalidFilter)
	}
	return nil
}

func (s *Service) ListPromos(ctx context.Context, organizer, status, q string, limit, offset int) ([]models.PromoRow, error) {
	_ = s.repo.PublishScheduled(ctx, nowStr())
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	status = strings.ToLower(strings.TrimSpace(status))
	if status != "" && status != "aktif" && status != "akan_datang" && status != "kedaluwarsa" && status != "kuota_habis" && status != "nonaktif" {
		return nil, fmt.Errorf("%w: unknown promo status", services.ErrInvalidFilter)
	}
	return s.repo.ListPromos(ctx, organizer, status, q, limit, offset)
}

func (s *Service) PromoDetail(ctx context.Context, organizer, code string) (*models.PromoRow, []models.PromoUsageRow, error) {
	row, err := s.repo.PromoByCode(ctx, organizer, code)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil, services.ErrNewsNotFound
		}
		return nil, nil, err
	}
	usage, err := s.repo.PromoUsage(ctx, row.Code, 50, 0)
	if err != nil {
		return nil, nil, err
	}
	return row, usage, nil
}

func (s *Service) CreatePromo(ctx context.Context, organizer string, in models.PromoInput) (string, error) {
	if err := s.validatePromo(ctx, &in, ""); err != nil {
		return "", err
	}
	code, err := utils.GenerateCode("PRM-")
	if err != nil {
		return "", err
	}
	if err := s.repo.CreatePromo(ctx, organizer, code, in); err != nil {
		return "", err
	}
	return code, nil
}

func (s *Service) UpdatePromo(ctx context.Context, organizer, code string, in models.PromoInput) error {
	cur, err := s.repo.PromoByCode(ctx, organizer, code)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return services.ErrNewsNotFound
		}
		return err
	}
	used, err := s.repo.PromoUsed(ctx, cur.Code)
	if err != nil {
		return err
	}
	if used {
		if in.DiscountPercent != cur.DiscountPercent || in.DiscountAmount != cur.DiscountAmount || in.MaxDiscount != cur.MaxDiscount {
			return fmt.Errorf("%w: discount cannot be changed after usage", services.ErrInvalidFilter)
		}
	}
	in.PromoCode = strings.ToUpper(strings.TrimSpace(in.PromoCode))
	if in.PromoCode == "" {
		in.PromoCode = cur.PromoCode
	}
	if err := s.validatePromo(ctx, &in, cur.Code); err != nil {
		return err
	}
	if used {
		in.DiscountPercent = cur.DiscountPercent
		in.DiscountAmount = cur.DiscountAmount
		in.MaxDiscount = cur.MaxDiscount
	}
	return s.repo.UpdatePromo(ctx, organizer, cur.Code, in)
}

func (s *Service) DeletePromo(ctx context.Context, organizer, code string) error {
	row, err := s.repo.PromoByCode(ctx, organizer, code)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return services.ErrNewsNotFound
		}
		return err
	}
	used, err := s.repo.PromoUsed(ctx, row.Code)
	if err != nil {
		return err
	}
	if used {
		return fmt.Errorf("%w: promo already used, deactivate instead", services.ErrInvalidFilter)
	}
	return s.repo.DeletePromo(ctx, organizer, row.Code)
}

func (s *Service) SetPromoActive(ctx context.Context, organizer, code string, active bool) error {
	row, err := s.repo.PromoByCode(ctx, organizer, code)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return services.ErrNewsNotFound
		}
		return err
	}
	status := models.PromoStatusActive
	if !active {
		status = models.PromoStatusDisabled
	}
	return s.repo.SetPromoStatus(ctx, organizer, row.Code, status)
}

func (s *Service) Positions(ctx context.Context) ([]models.AdPosition, error) {
	return s.repo.ListPositions(ctx)
}

func computeCost(pricePerDay float64, start, end string) (float64, int) {
	s, err1 := parseDateTime(start)
	e, err2 := parseDateTime(end)
	if err1 != nil || err2 != nil {
		return 0, 0
	}
	days := int(e.Sub(s).Hours()/24) + 1
	if days < 1 {
		days = 1
	}
	return pricePerDay * float64(days), days
}

func (s *Service) validateAd(in *models.AdCampaignInput) error {
	in.Title = strings.TrimSpace(in.Title)
	if len(in.Title) < 3 || len(in.Title) > 200 {
		return fmt.Errorf("%w: title must be 3-200 chars", services.ErrInvalidFilter)
	}
	if len(in.TargetURL) > 500 {
		return fmt.Errorf("%w: target URL too long", services.ErrInvalidFilter)
	}
	target := strings.TrimSpace(in.TargetURL)
	if in.TargetEventCode = strings.TrimSpace(in.TargetEventCode); in.TargetEventCode != "" {
		target = ""
	}
	if target != "" {
		if !strings.HasPrefix(strings.ToLower(target), "http://") && !strings.HasPrefix(strings.ToLower(target), "https://") {
			return fmt.Errorf("%w: target must be http or https", services.ErrInvalidFilter)
		}
	}
	if target == "" && in.TargetEventCode == "" {
		return fmt.Errorf("%w: target event or URL is required", services.ErrInvalidFilter)
	}
	in.Position = strings.ToUpper(strings.TrimSpace(in.Position))
	if in.Position == "" {
		return fmt.Errorf("%w: position is required", services.ErrInvalidFilter)
	}
	start, err := parseDateTime(in.StartAt)
	if err != nil {
		return fmt.Errorf("%w: invalid start: %v", services.ErrInvalidFilter, err)
	}
	end, err := parseDateTime(in.EndAt)
	if err != nil {
		return fmt.Errorf("%w: invalid end: %v", services.ErrInvalidFilter, err)
	}
	if !end.After(start) {
		return fmt.Errorf("%w: end must be after start", services.ErrInvalidFilter)
	}
	in.StartAt = start.Format("2006-01-02 15:04:05")
	in.EndAt = end.Format("2006-01-02 15:04:05")
	if strings.TrimSpace(in.ImageURL) == "" {
		return fmt.Errorf("%w: banner image is required", services.ErrInvalidFilter)
	}
	return nil
}

func (s *Service) ListAds(ctx context.Context, organizer, status, q string, limit, offset int) ([]models.AdCampaignRow, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	status = strings.ToLower(strings.TrimSpace(status))
	if status != "" && status != "draft" && status != "menunggu_pembayaran" && status != "menunggu_persetujuan" && status != "tayang" && status != "selesai" && status != "ditolak" && status != "dijeda" {
		return nil, fmt.Errorf("%w: unknown ad status", services.ErrInvalidFilter)
	}
	return s.repo.ListAds(ctx, organizer, status, q, limit, offset)
}

func (s *Service) AdDetail(ctx context.Context, organizer, code string) (*models.AdCampaignRow, []models.AdDailyPoint, error) {
	row, err := s.repo.AdByCode(ctx, organizer, code)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil, services.ErrNewsNotFound
		}
		return nil, nil, err
	}
	stats, err := s.repo.AdStats(ctx, row.Code)
	if err != nil {
		return nil, nil, err
	}
	return row, stats, nil
}

func (s *Service) CreateAd(ctx context.Context, organizer string, in models.AdCampaignInput) (string, error) {
	if err := s.validateAd(&in); err != nil {
		return "", err
	}
	if in.TargetEventCode != "" {
		ok, err := s.repo.EventOwned(ctx, organizer, in.TargetEventCode)
		if err != nil {
			return "", err
		}
		if !ok {
			return "", fmt.Errorf("%w: unknown event", services.ErrInvalidFilter)
		}
		in.TargetURL = "/events/" + in.TargetEventCode
	}
	pos, err := s.repo.PositionByCode(ctx, in.Position)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("%w: unknown position", services.ErrInvalidFilter)
		}
		return "", err
	}
	cost, _ := computeCost(pos.PricePerDay, in.StartAt, in.EndAt)
	code, err := utils.GenerateCode("ADC-")
	if err != nil {
		return "", err
	}
	if err := s.repo.CreateAd(ctx, organizer, code, in, cost); err != nil {
		return "", err
	}
	return code, nil
}

func (s *Service) UpdateAd(ctx context.Context, organizer, code string, in models.AdCampaignInput) error {
	cur, err := s.repo.AdByCode(ctx, organizer, code)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return services.ErrNewsNotFound
		}
		return err
	}
	if cur.Status == models.AdStatusLive || cur.Status == "ACTIVE" || cur.Status == models.AdStatusEnded {
		return fmt.Errorf("%w: live or ended campaigns cannot change image or period, only pause", services.ErrInvalidFilter)
	}
	if err := s.validateAd(&in); err != nil {
		return err
	}
	if in.TargetEventCode != "" {
		ok, err := s.repo.EventOwned(ctx, organizer, in.TargetEventCode)
		if err != nil {
			return err
		}
		if !ok {
			return fmt.Errorf("%w: unknown event", services.ErrInvalidFilter)
		}
		in.TargetURL = "/events/" + in.TargetEventCode
	}
	pos, err := s.repo.PositionByCode(ctx, in.Position)
	if err != nil {
		return err
	}
	cost, _ := computeCost(pos.PricePerDay, in.StartAt, in.EndAt)
	return s.repo.UpdateAd(ctx, organizer, cur.Code, in, cost)
}

func (s *Service) SetAdStatus(ctx context.Context, organizer, code, action, note string) error {
	cur, err := s.repo.AdByCode(ctx, organizer, code)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return services.ErrNewsNotFound
		}
		return err
	}
	action = strings.ToLower(strings.TrimSpace(action))
	switch action {
	case "submit":
		if cur.Status != models.AdStatusDraft {
			return fmt.Errorf("%w: only draft can be submitted", services.ErrInvalidFilter)
		}
		return s.repo.SetAdStatus(ctx, organizer, cur.Code, models.AdStatusPendingPayment, models.AdPayUnpaid, models.AdApprovalPending, "")
	case "pay":
		if strings.ToLower(s.appEnv) == "production" {
			return fmt.Errorf("%w: simulation disabled in production", services.ErrInvalidFilter)
		}
		if cur.Status != models.AdStatusPendingPayment {
			return fmt.Errorf("%w: not awaiting payment", services.ErrInvalidFilter)
		}
		if strings.ToLower(s.appEnv) == "development" {
			return s.repo.SetAdStatus(ctx, organizer, cur.Code, models.AdStatusLive, models.AdPayPaid, models.AdApprovalApproved, "")
		}
		return s.repo.SetAdStatus(ctx, organizer, cur.Code, models.AdStatusPendingApproval, models.AdPayPaid, models.AdApprovalPending, "")
	case "approve":
		if cur.PaymentStatus != models.AdPayPaid {
			return fmt.Errorf("%w: not paid", services.ErrInvalidFilter)
		}
		return s.repo.SetAdStatus(ctx, organizer, cur.Code, models.AdStatusLive, cur.PaymentStatus, models.AdApprovalApproved, "")
	case "reject":
		note = strings.TrimSpace(note)
		if note == "" {
			return fmt.Errorf("%w: reject note is required", services.ErrInvalidFilter)
		}
		return s.repo.SetAdStatus(ctx, organizer, cur.Code, models.AdStatusRejected, cur.PaymentStatus, models.AdApprovalRejected, note)
	case "pause":
		if cur.Status != models.AdStatusLive && cur.Status != "ACTIVE" {
			return fmt.Errorf("%w: only live can be paused", services.ErrInvalidFilter)
		}
		return s.repo.SetAdStatus(ctx, organizer, cur.Code, models.AdStatusPaused, cur.PaymentStatus, cur.ApprovalStatus, "")
	case "resume":
		if cur.Status != models.AdStatusPaused {
			return fmt.Errorf("%w: only paused can be resumed", services.ErrInvalidFilter)
		}
		return s.repo.SetAdStatus(ctx, organizer, cur.Code, models.AdStatusLive, cur.PaymentStatus, cur.ApprovalStatus, "")
	case "end":
		return s.repo.SetAdStatus(ctx, organizer, cur.Code, models.AdStatusEnded, cur.PaymentStatus, cur.ApprovalStatus, "")
	default:
		return fmt.Errorf("%w: unknown action", services.ErrInvalidFilter)
	}
}

func (s *Service) SimulatePay(ctx context.Context, organizer, code string) error {
	return s.SetAdStatus(ctx, organizer, code, "pay", "")
}

func (s *Service) ListNews(ctx context.Context, status, category, q string, limit, offset int) ([]models.NewsRow, error) {
	_ = s.repo.PublishScheduled(ctx, nowStr())
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	status = strings.ToLower(strings.TrimSpace(status))
	if status != "" && status != "draft" && status != "terbit" && status != "published" && status != "diarsipkan" && status != "archived" {
		return nil, fmt.Errorf("%w: unknown news status", services.ErrInvalidFilter)
	}
	if status == "terbit" {
		status = "published"
	}
	if status == "diarsipkan" {
		status = "archived"
	}
	category = strings.TrimSpace(category)
	return s.repo.ListNews(ctx, status, category, q, limit, offset)
}

func (s *Service) NewsDetail(ctx context.Context, code string) (*models.NewsRow, error) {
	row, err := s.repo.NewsByCode(ctx, code)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, services.ErrNewsNotFound
		}
		return nil, err
	}
	row.Body = utils.SanitizeHTML(row.Body)
	row.Summary = utils.SanitizeHTML(row.Summary)
	return row, nil
}

func (s *Service) validateNews(ctx context.Context, in *models.NewsInput) error {
	in.Title = strings.TrimSpace(in.Title)
	if len(in.Title) < 3 || len(in.Title) > 200 {
		return fmt.Errorf("%w: title must be 3-200 chars", services.ErrInvalidFilter)
	}
	if len(in.Summary) > 500 {
		return fmt.Errorf("%w: summary max 500 chars", services.ErrInvalidFilter)
	}
	if strings.TrimSpace(in.Body) == "" {
		return fmt.Errorf("%w: body is required", services.ErrInvalidFilter)
	}
	in.Body = utils.SanitizeHTML(in.Body)
	if len(in.ImageURL) > 500 {
		return fmt.Errorf("%w: image URL too long", services.ErrInvalidFilter)
	}
	in.NewsCategoryCode = strings.ToUpper(strings.TrimSpace(in.NewsCategoryCode))
	if in.NewsCategoryCode == "" {
		return fmt.Errorf("%w: news category is required", services.ErrInvalidFilter)
	}
	if ok, err := s.repo.CategoryExists(ctx, in.NewsCategoryCode); err != nil {
		return err
	} else if !ok {
		return fmt.Errorf("%w: unknown news category", services.ErrInvalidFilter)
	}
	if in.EventCategoryCode = strings.ToUpper(strings.TrimSpace(in.EventCategoryCode)); in.EventCategoryCode != "" {
		cats, err := s.repo.EventCategories(ctx)
		if err != nil {
			return err
		}
		found := false
		for _, c := range cats {
			if strings.EqualFold(c.Code, in.EventCategoryCode) {
				found = true
				in.EventCategoryCode = c.Code
				break
			}
		}
		if !found {
			return fmt.Errorf("%w: unknown event category", services.ErrInvalidFilter)
		}
	}
	in.EventCode = strings.TrimSpace(in.EventCode)
	if in.ScheduledAt = strings.TrimSpace(in.ScheduledAt); in.ScheduledAt != "" {
		t, err := parseDateTime(in.ScheduledAt)
		if err != nil {
			return fmt.Errorf("%w: invalid scheduled_at: %v", services.ErrInvalidFilter, err)
		}
		in.ScheduledAt = t.Format("2006-01-02 15:04:05")
	}
	return nil
}

func (s *Service) CreateNews(ctx context.Context, author string, in models.NewsInput) (string, error) {
	if err := s.validateNews(ctx, &in); err != nil {
		return "", err
	}
	slug, err := s.uniqueSlug(ctx, in.Title, "")
	if err != nil {
		return "", err
	}
	code, err := utils.GenerateCode("NWS-")
	if err != nil {
		return "", err
	}
	if err := s.repo.CreateNews(ctx, author, code, slug, in); err != nil {
		return "", err
	}
	return code, nil
}

func (s *Service) UpdateNews(ctx context.Context, code string, in models.NewsInput) error {
	cur, err := s.repo.NewsByCode(ctx, code)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return services.ErrNewsNotFound
		}
		return err
	}
	if err := s.validateNews(ctx, &in); err != nil {
		return err
	}
	slug := cur.Slug
	if !strings.EqualFold(strings.TrimSpace(in.Title), cur.Title) {
		ns, err := s.uniqueSlug(ctx, in.Title, cur.Code)
		if err != nil {
			return err
		}
		slug = ns
	}
	return s.repo.UpdateNews(ctx, cur.Code, slug, in)
}

func (s *Service) DeleteNews(ctx context.Context, code string) error {
	if _, err := s.repo.NewsByCode(ctx, code); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return services.ErrNewsNotFound
		}
		return err
	}
	return s.repo.SoftDeleteNews(ctx, code)
}

func (s *Service) SetNewsStatus(ctx context.Context, code, status string) error {
	if _, err := s.repo.NewsByCode(ctx, code); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return services.ErrNewsNotFound
		}
		return err
	}
	status = strings.ToLower(strings.TrimSpace(status))
	switch status {
	case "publish", "terbit", "published":
		return s.repo.SetNewsStatus(ctx, code, models.NewsStatusPublished)
	case "unpublish", "tarik":
		return s.repo.SetNewsStatus(ctx, code, models.NewsStatusDraft)
	case "archive", "arsipkan", "archived":
		return s.repo.SetNewsStatus(ctx, code, models.NewsStatusArchived)
	default:
		return fmt.Errorf("%w: status must be publish, unpublish or archive", services.ErrInvalidFilter)
	}
}

func (s *Service) NewsCategories(ctx context.Context) ([]models.NewsCategory, error) {
	return s.repo.NewsCategories(ctx)
}

func (s *Service) EventCategories(ctx context.Context) ([]models.EventCategory, error) {
	return s.repo.EventCategories(ctx)
}

func (s *Service) OwnedEvents(ctx context.Context, organizer string) ([]OwnedEvent, error) {
	return s.repo.OwnedEvents(ctx, organizer)
}

func nowStr() string { return time.Now().Format("2006-01-02 15:04:05") }
