package discovery

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"golang-backend/models"
	"golang-backend/services"
	"golang-backend/utils"
)

// Public discovery service: validates every filter/sort against whitelists
// (no raw request text reaches SQL), sanitizes content for XSS, and assembles
// home/list/detail payloads from publishable rows only.
type ServiceInterface interface {
	Home(ctx context.Context, city string) (*models.HomeResponse, error)
	Categories(ctx context.Context) ([]models.EventCategory, error)
	Cities(ctx context.Context) ([]models.CityOption, error)
	NewsCategories(ctx context.Context) ([]models.NewsCategory, error)
	NewsList(ctx context.Context, category, q, sort string, limit, offset int) ([]models.NewsItem, error)
	NewsDetail(ctx context.Context, slug, viewerHash string) (*models.NewsDetail, error)
	EventsList(ctx context.Context, f models.DiscoveryFilter) ([]models.EventCard, error)
	EventDetail(ctx context.Context, code string) (*models.EventDetail, error)
	Promotions(ctx context.Context, mode, category, q string, limit, offset int) ([]models.PromoItem, error)
	RecordAdEvent(ctx context.Context, campaign, kind, iphash string) (bool, error)
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

func (s *Service) checkCity(ctx context.Context, city string) error {
	if city == "" {
		return nil
	}
	cities, err := s.repo.Cities(ctx)
	if err != nil {
		return fmt.Errorf("list cities: %w", err)
	}
	for _, c := range cities {
		if strings.EqualFold(c.Code, city) {
			return nil
		}
	}
	return fmt.Errorf("%w: unknown city", services.ErrInvalidFilter)
}

func (s *Service) checkEventCategories(ctx context.Context, cats []string) error {
	if len(cats) == 0 {
		return nil
	}
	known, err := s.repo.EventCategories(ctx)
	if err != nil {
		return fmt.Errorf("list categories: %w", err)
	}
	set := make(map[string]bool, len(known))
	for _, c := range known {
		set[strings.ToUpper(c.Code)] = true
	}
	for _, c := range cats {
		if !set[strings.ToUpper(c)] {
			return fmt.Errorf("%w: unknown category", services.ErrInvalidFilter)
		}
	}
	return nil
}

func (s *Service) Home(ctx context.Context, city string) (*models.HomeResponse, error) {
	city = strings.TrimSpace(strings.ToUpper(city))
	if err := s.checkCity(ctx, city); err != nil {
		return nil, err
	}
	now := nowStr()
	banners, err := s.repo.LiveBanners(ctx, "HOME_BANNER")
	if err != nil {
		return nil, fmt.Errorf("live banners: %w", err)
	}
	promos, err := s.repo.FeaturedPromos(ctx, now, 4)
	if err != nil {
		return nil, fmt.Errorf("featured promos: %w", err)
	}
	news, err := s.repo.LatestNews(ctx, 4)
	if err != nil {
		return nil, fmt.Errorf("latest news: %w", err)
	}
	popular, err := s.repo.PopularEvents(ctx, city, 6)
	if err != nil {
		return nil, fmt.Errorf("popular events: %w", err)
	}
	upcoming, err := s.repo.UpcomingEvents(ctx, city, 6)
	if err != nil {
		return nil, fmt.Errorf("upcoming events: %w", err)
	}
	categories, err := s.repo.EventCategories(ctx)
	if err != nil {
		return nil, fmt.Errorf("event categories: %w", err)
	}
	for i := range news {
		news[i].Summary = utils.SanitizeHTML(news[i].Summary)
	}
	return &models.HomeResponse{
		Banners:    banners,
		Promos:     promos,
		News:       news,
		Popular:    popular,
		Upcoming:   upcoming,
		Categories: categories,
	}, nil
}

func (s *Service) Categories(ctx context.Context) ([]models.EventCategory, error) {
	return s.repo.EventCategories(ctx)
}

func (s *Service) Cities(ctx context.Context) ([]models.CityOption, error) {
	return s.repo.Cities(ctx)
}

func (s *Service) NewsCategories(ctx context.Context) ([]models.NewsCategory, error) {
	return s.repo.NewsCategories(ctx)
}

func (s *Service) NewsList(ctx context.Context, category, q, sort string, limit, offset int) ([]models.NewsItem, error) {
	category = strings.TrimSpace(strings.ToUpper(category))
	if category != "" {
		cats, err := s.repo.NewsCategories(ctx)
		if err != nil {
			return nil, fmt.Errorf("list news categories: %w", err)
		}
		known := false
		for _, c := range cats {
			if strings.EqualFold(c.Code, category) {
				category = c.Code
				known = true
				break
			}
		}
		if !known {
			return nil, fmt.Errorf("%w: unknown news category", services.ErrInvalidFilter)
		}
	}
	sort = strings.ToLower(strings.TrimSpace(sort))
	if sort != "" && sort != "newest" && sort != "popular" {
		return nil, fmt.Errorf("%w: sort must be newest or popular", services.ErrInvalidFilter)
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	items, err := s.repo.ListNews(ctx, category, strings.TrimSpace(q), sort, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list news: %w", err)
	}
	for i := range items {
		items[i].Summary = utils.SanitizeHTML(items[i].Summary)
	}
	return items, nil
}

func (s *Service) NewsDetail(ctx context.Context, slug, viewerHash string) (*models.NewsDetail, error) {
	slug = strings.TrimSpace(slug)
	if slug == "" {
		return nil, fmt.Errorf("%w: slug is required", services.ErrInvalidFilter)
	}
	detail, err := s.repo.NewsBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, services.ErrNewsNotFound
		}
		return nil, fmt.Errorf("news by slug: %w", err)
	}
	detail.Body = utils.SanitizeHTML(detail.Body)
	detail.Summary = utils.SanitizeHTML(detail.Summary)
	related, err := s.repo.RelatedNews(ctx, detailCategoryCode(ctx, s, detail), detail.Code, 4)
	if err != nil {
		return nil, fmt.Errorf("related news: %w", err)
	}
	for i := range related {
		related[i].Summary = utils.SanitizeHTML(related[i].Summary)
	}
	detail.RelatedNews = related
	if detail.EventCode != "" {
		if ev, err := s.repo.EventRow(ctx, detail.EventCode); err == nil {
			detail.EventTitle = ev.Title
			detail.RelatedEvent = &models.EventCard{
				Code: ev.Code, Title: ev.Title, Category: ev.Category, City: ev.City,
				Venue: ev.Venue, StartAt: ev.StartAt, EndAt: ev.EndAt,
				PosterURL: ev.PosterURL,
			}
		}
		// Event gone/unpublished: related event simply stays empty (no leak).
	}
	// View counting with 30-minute same-viewer cooldown (best-effort).
	if viewerHash != "" {
		seen, err := s.repo.NewsViewedSince(ctx, detail.Code, viewerHash, 30)
		if err == nil && !seen {
			_ = s.repo.InsertNewsView(ctx, detail.Code, viewerHash)
			_ = s.repo.IncrementNewsView(ctx, detail.Code)
			detail.ViewCount++
		}
	}
	return detail, nil
}

// detailCategoryCode resolves the news category code for related lookup.
func detailCategoryCode(ctx context.Context, s *Service, detail *models.NewsDetail) string {
	cats, err := s.repo.NewsCategories(ctx)
	if err != nil {
		return ""
	}
	for _, c := range cats {
		if c.Name == detail.Category {
			return c.Code
		}
	}
	return ""
}

var eventSorts = map[string]bool{
	"nearest":    true,
	"popular":    true,
	"price_asc":  true,
	"price_desc": true,
}

func (s *Service) EventsList(ctx context.Context, f models.DiscoveryFilter) ([]models.EventCard, error) {
	for i := range f.Categories {
		f.Categories[i] = strings.TrimSpace(strings.ToUpper(f.Categories[i]))
	}
	kept := f.Categories[:0]
	for _, c := range f.Categories {
		if c != "" {
			kept = append(kept, c)
		}
	}
	f.Categories = kept
	if err := s.checkEventCategories(ctx, f.Categories); err != nil {
		return nil, err
	}
	f.City = strings.TrimSpace(strings.ToUpper(f.City))
	if err := s.checkCity(ctx, f.City); err != nil {
		return nil, err
	}
	if f.From != "" && !utils.IsValidDate(f.From) {
		return nil, fmt.Errorf("%w: from must be YYYY-MM-DD", services.ErrInvalidFilter)
	}
	if f.To != "" && !utils.IsValidDate(f.To) {
		return nil, fmt.Errorf("%w: to must be YYYY-MM-DD", services.ErrInvalidFilter)
	}
	if f.From != "" && f.To != "" && f.From > f.To {
		return nil, fmt.Errorf("%w: from must not be after to", services.ErrInvalidFilter)
	}
	if f.MinPrice < 0 || f.MaxPrice < 0 {
		return nil, fmt.Errorf("%w: prices must not be negative", services.ErrInvalidFilter)
	}
	if f.MaxPrice > 0 && f.MinPrice > f.MaxPrice {
		return nil, fmt.Errorf("%w: min price must not exceed max price", services.ErrInvalidFilter)
	}
	f.Query = strings.TrimSpace(f.Query)
	sort := strings.ToLower(strings.TrimSpace(f.Sort))
	if sort == "" {
		sort = "nearest"
	}
	if _, ok := eventSorts[sort]; !ok {
		return nil, fmt.Errorf("%w: sort must be nearest, popular, price_asc or price_desc", services.ErrInvalidFilter)
	}
	if f.Limit <= 0 || f.Limit > 100 {
		f.Limit = 20
	}
	if f.Offset < 0 {
		f.Offset = 0
	}
	toPlusOne := ""
	if f.To != "" {
		toPlusOne = plusOne(f.To)
		f.To = toPlusOne // half-open: START_AT < day-after-To includes To fully
	}
	events, err := s.repo.ListEvents(ctx, f, sort)
	if err != nil {
		return nil, fmt.Errorf("list events: %w", err)
	}
	return events, nil
}

func (s *Service) EventDetail(ctx context.Context, code string) (*models.EventDetail, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return nil, fmt.Errorf("%w: event code is required", services.ErrInvalidFilter)
	}
	detail, err := s.repo.EventRow(ctx, code)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, services.ErrEventNotFound
		}
		return nil, fmt.Errorf("event row: %w", err)
	}
	detail.Description = utils.SanitizeHTML(detail.Description)
	types, err := s.repo.EventTicketTypes(ctx, code, nowStr())
	if err != nil {
		return nil, fmt.Errorf("event ticket types: %w", err)
	}
	detail.TicketTypes = types
	promos, err := s.repo.EventPromos(ctx, code, detail.CategoryCode)
	if err != nil {
		return nil, fmt.Errorf("event promos: %w", err)
	}
	detail.Promos = promos
	related, err := s.repo.RelatedEvents(ctx, code, detail.CategoryCode, detailCityCode(ctx, s, detail), 4)
	if err != nil {
		return nil, fmt.Errorf("related events: %w", err)
	}
	detail.Related = related
	sold, quota := 0, 0
	for _, t := range types {
		sold += t.Sold
		quota += t.Quota
	}
	detail.SoldOut = quota > 0 && sold >= quota
	return detail, nil
}

// detailCityCode resolves the city code of a detail row for related lookup.
func detailCityCode(ctx context.Context, s *Service, detail *models.EventDetail) string {
	cities, err := s.repo.Cities(ctx)
	if err != nil {
		return ""
	}
	for _, c := range cities {
		if c.Name == detail.City {
			return c.Code
		}
	}
	return ""
}

func (s *Service) Promotions(ctx context.Context, mode, category, q string, limit, offset int) ([]models.PromoItem, error) {
	mode = strings.ToLower(strings.TrimSpace(mode))
	if mode == "" {
		mode = "active"
	}
	if mode != "active" && mode != "upcoming" {
		return nil, fmt.Errorf("%w: mode must be active or upcoming", services.ErrInvalidFilter)
	}
	category = strings.TrimSpace(strings.ToUpper(category))
	if category != "" {
		if err := s.checkEventCategories(ctx, []string{category}); err != nil {
			return nil, err
		}
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	items, err := s.repo.ListPromos(ctx, nowStr(), mode, category, strings.TrimSpace(q), limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list promos: %w", err)
	}
	for i := range items {
		items[i].Description = utils.SanitizeHTML(items[i].Description)
	}
	return items, nil
}

// RecordAdEvent logs one impression/click with anti-gaming cooldowns
// (impression 60 min, click 10 min per IP+campaign). Returns recorded.
func (s *Service) RecordAdEvent(ctx context.Context, campaign, kind, iphash string) (bool, error) {
	campaign = strings.TrimSpace(campaign)
	kind = strings.ToUpper(strings.TrimSpace(kind))
	if campaign == "" {
		return false, fmt.Errorf("%w: campaign is required", services.ErrInvalidFilter)
	}
	if kind != "IMPRESSION" && kind != "CLICK" {
		return false, fmt.Errorf("%w: kind must be impression or click", services.ErrInvalidFilter)
	}
	live, err := s.repo.CampaignLive(ctx, campaign)
	if err != nil {
		return false, fmt.Errorf("check campaign: %w", err)
	}
	if !live {
		return false, fmt.Errorf("%w: campaign not live", services.ErrInvalidFilter)
	}
	window := 60
	if kind == "CLICK" {
		window = 10
	}
	seen, err := s.repo.RecentAdEvent(ctx, campaign, kind, iphash, window)
	if err != nil {
		return false, fmt.Errorf("check ad event: %w", err)
	}
	if seen {
		return false, nil
	}
	if err := s.repo.InsertAdEvent(ctx, campaign, kind, iphash); err != nil {
		return false, fmt.Errorf("insert ad event: %w", err)
	}
	return true, nil
}
