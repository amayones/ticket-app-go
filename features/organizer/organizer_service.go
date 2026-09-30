package organizer

import (
	"context"
	"fmt"
	"strings"
	"time"

	"golang-backend/internal/export"
	"golang-backend/models"
	"golang-backend/services"
	"golang-backend/utils"
)

// Dashboard service: validates filters, scopes everything to the caller's
// owned events, and assembles summary payloads with previous-period deltas.
type ServiceInterface interface {
	Summary(ctx context.Context, organizer, preset, from, to, event, city string) (*models.OrganizerSummary, error)
	Orders(ctx context.Context, organizer, preset, from, to, event, city string, limit int) ([]models.OrganizerOrder, error)
	Refunds(ctx context.Context, organizer string) ([]models.OrganizerRefund, error)
	DecideRefund(ctx context.Context, organizer, code string, approve bool) (string, float64, error)
	CheckinLive(ctx context.Context, organizer, eventCode string) (*models.CheckinLive, error)
	ExportTables(ctx context.Context, organizer, preset, from, to, event, city string) ([]export.Table, error)
}

type Service struct {
	repo RepositoryInterface
}

func NewService(repo RepositoryInterface) *Service {
	return &Service{repo: repo}
}

// resolveFilter validates preset/dates + event ownership + city existence.
func (s *Service) resolveFilter(ctx context.Context, organizer, preset, from, to, event, city string) (models.DashboardFilter, utils.DateRange, error) {
	rng, err := utils.ParseDateRange(strings.TrimSpace(preset), strings.TrimSpace(from), strings.TrimSpace(to))
	if err != nil {
		return models.DashboardFilter{}, rng, fmt.Errorf("%w: %v", services.ErrInvalidFilter, err)
	}
	event = strings.TrimSpace(event)
	if event != "" {
		ok, err := s.repo.EventOwned(ctx, organizer, event)
		if err != nil {
			return models.DashboardFilter{}, rng, fmt.Errorf("check event: %w", err)
		}
		if !ok {
			return models.DashboardFilter{}, rng, fmt.Errorf("%w: unknown event", services.ErrInvalidFilter)
		}
	}
	city = strings.TrimSpace(strings.ToUpper(city))
	if city != "" {
		ok, err := s.repo.CityExists(ctx, city)
		if err != nil {
			return models.DashboardFilter{}, rng, fmt.Errorf("check city: %w", err)
		}
		if !ok {
			return models.DashboardFilter{}, rng, fmt.Errorf("%w: unknown city", services.ErrInvalidFilter)
		}
	}
	return models.DashboardFilter{From: rng.From, To: rng.To, Event: event, City: city}, rng, nil
}

// plusOne returns the day after d (YYYY-MM-DD) for half-open ranges.
func plusOne(d string) string {
	t, err := time.Parse("2006-01-02", d)
	if err != nil {
		return d
	}
	return t.AddDate(0, 0, 1).Format("2006-01-02")
}

func kpi(cur, prev float64) models.KPIValue {
	return models.KPIValue{Value: cur, Previous: prev, DeltaPct: utils.DeltaPct(cur, prev)}
}

func rate(part, total float64) float64 {
	if total <= 0 {
		return 0
	}
	return part / total * 100
}

func (s *Service) Summary(ctx context.Context, organizer, preset, from, to, event, city string) (*models.OrganizerSummary, error) {
	filter, rng, err := s.resolveFilter(ctx, organizer, preset, from, to, event, city)
	if err != nil {
		return nil, err
	}
	end := plusOne(rng.To)
	prev := utils.PreviousRange(rng)
	prevEnd := plusOne(prev.To)

	tickets, gross, discount, _, net, promoOrders, orders, err := s.repo.OrderSums(ctx, organizer, event, city, rng.From, end)
	if err != nil {
		return nil, fmt.Errorf("order sums: %w", err)
	}
	ptickets, pgross, pdiscount, _, pnet, ppromoOrders, porders, err := s.repo.OrderSums(ctx, organizer, event, city, prev.From, prevEnd)
	if err != nil {
		return nil, fmt.Errorf("previous order sums: %w", err)
	}
	active, err := s.repo.ActiveEvents(ctx, organizer, rng.From, end)
	if err != nil {
		return nil, fmt.Errorf("active events: %w", err)
	}
	pactive, err := s.repo.ActiveEvents(ctx, organizer, prev.From, prevEnd)
	if err != nil {
		return nil, fmt.Errorf("previous active events: %w", err)
	}
	sold, used, err := s.repo.CheckinStats(ctx, organizer, event, city, rng.From, end)
	if err != nil {
		return nil, fmt.Errorf("checkin stats: %w", err)
	}
	psold, pused, err := s.repo.CheckinStats(ctx, organizer, event, city, prev.From, prevEnd)
	if err != nil {
		return nil, fmt.Errorf("previous checkin stats: %w", err)
	}
	refunds, _, err := s.repo.RefundStats(ctx, organizer, event, city, rng.From, end)
	if err != nil {
		return nil, fmt.Errorf("refund stats: %w", err)
	}
	prefunds, _, err := s.repo.RefundStats(ctx, organizer, event, city, prev.From, prevEnd)
	if err != nil {
		return nil, fmt.Errorf("previous refund stats: %w", err)
	}
	imp, clicks, cost, err := s.repo.AdSums(ctx, organizer, event, city, rng.From, end)
	if err != nil {
		return nil, fmt.Errorf("ad sums: %w", err)
	}
	pimp, pclicks, pcost, err := s.repo.AdSums(ctx, organizer, event, city, prev.From, prevEnd)
	if err != nil {
		return nil, fmt.Errorf("previous ad sums: %w", err)
	}
	trend, err := s.repo.SalesTrend(ctx, organizer, event, city, rng.From, end)
	if err != nil {
		return nil, fmt.Errorf("sales trend: %w", err)
	}
	byType, err := s.repo.SalesByTicketType(ctx, organizer, event, city, rng.From, end)
	if err != nil {
		return nil, fmt.Errorf("sales by type: %w", err)
	}
	byCity, err := s.repo.SalesByCity(ctx, organizer, event, city, rng.From, end)
	if err != nil {
		return nil, fmt.Errorf("sales by city: %w", err)
	}
	byPromo, err := s.repo.SalesByPromo(ctx, organizer, event, city, rng.From, end)
	if err != nil {
		return nil, fmt.Errorf("sales by promo: %w", err)
	}
	upcoming, err := s.repo.UpcomingEvents(ctx, organizer)
	if err != nil {
		return nil, fmt.Errorf("upcoming events: %w", err)
	}
	recent, err := s.repo.RecentOrders(ctx, organizer, event, city, 10)
	if err != nil {
		return nil, fmt.Errorf("recent orders: %w", err)
	}
	pending, err := s.repo.PendingRefunds(ctx, organizer)
	if err != nil {
		return nil, fmt.Errorf("pending refunds: %w", err)
	}
	balance, err := s.repo.Balance(ctx, organizer)
	if err != nil {
		return nil, fmt.Errorf("balance: %w", err)
	}
	payout, err := s.repo.PayoutReady(ctx, organizer)
	if err != nil {
		return nil, fmt.Errorf("payout ready: %w", err)
	}
	cities, err := s.repo.EventCities(ctx, organizer)
	if err != nil {
		return nil, fmt.Errorf("event cities: %w", err)
	}

	alerts := make([]models.DashboardAlert, 0)
	usage, err := s.repo.QuotaUsage(ctx, organizer)
	if err != nil {
		return nil, fmt.Errorf("quota usage: %w", err)
	}
	for _, u := range usage {
		if u.Quota > 0 && float64(u.Sold)/float64(u.Quota) >= 0.9 {
			alerts = append(alerts, models.DashboardAlert{
				Kind:   "QUOTA",
				Title:  "Quota almost sold out",
				Detail: fmt.Sprintf("%s: %d/%d sold", u.Title, u.Sold, u.Quota),
			})
		}
	}
	if drafts, err := s.repo.DraftEvents(ctx, organizer); err != nil {
		return nil, fmt.Errorf("draft events: %w", err)
	} else if drafts > 0 {
		alerts = append(alerts, models.DashboardAlert{
			Kind:   "DRAFT",
			Title:  "Unpublished events",
			Detail: fmt.Sprintf("%d event(s) still in DRAFT", drafts),
		})
	}
	now := time.Now()
	in3 := now.AddDate(0, 0, 3).Format("2006-01-02")
	if exp, err := s.repo.ExpiringPromos(ctx, organizer, now.Format("2006-01-02"), in3); err != nil {
		return nil, fmt.Errorf("expiring promos: %w", err)
	} else if exp > 0 {
		alerts = append(alerts, models.DashboardAlert{
			Kind:   "PROMO",
			Title:  "Promos expiring soon",
			Detail: fmt.Sprintf("%d promo(s) expire within 3 days", exp),
		})
	}
	if payout > 0 {
		alerts = append(alerts, models.DashboardAlert{
			Kind:   "PAYOUT",
			Title:  "Payout ready",
			Detail: fmt.Sprintf("%.2f ready to withdraw", payout),
		})
	}

	return &models.OrganizerSummary{
		Filter:        filter,
		TicketsSold:   kpi(float64(tickets), float64(ptickets)),
		GrossRevenue:  kpi(gross, pgross),
		NetRevenue:    kpi(net, pnet),
		ActiveEvents:  kpi(float64(active), float64(pactive)),
		CheckinRate:   kpi(rate(float64(used), float64(sold)), rate(float64(pused), float64(psold))),
		RefundCount:   kpi(float64(refunds), float64(prefunds)),
		RefundPct:     kpi(rate(float64(refunds), float64(orders)), rate(float64(prefunds), float64(porders))),
		PromoOrders:   kpi(float64(promoOrders), float64(ppromoOrders)),
		PromoDiscount: kpi(discount, pdiscount),
		AdImpressions: kpi(float64(imp), float64(pimp)),
		AdClicks:      kpi(float64(clicks), float64(pclicks)),
		AdCost:        kpi(cost, pcost),
		SalesTrend:    trend,
		ByTicketType:  byType,
		ByCity:        byCity,
		ByPromo:       byPromo,
		Upcoming:      upcoming,
		RecentOrders:  recent,
		PendingRefund: pending,
		Alerts:        alerts,
		Balance:       balance,
		PayoutReady:   payout,
		Cities:        cities,
	}, nil
}

func (s *Service) Orders(ctx context.Context, organizer, preset, from, to, event, city string, limit int) ([]models.OrganizerOrder, error) {
	_, rng, err := s.resolveFilter(ctx, organizer, preset, from, to, event, city)
	if err != nil {
		return nil, err
	}
	_ = rng
	if limit <= 0 || limit > 100 {
		limit = 10
	}
	orders, err := s.repo.RecentOrders(ctx, organizer, event, city, limit)
	if err != nil {
		return nil, fmt.Errorf("recent orders: %w", err)
	}
	return orders, nil
}

func (s *Service) Refunds(ctx context.Context, organizer string) ([]models.OrganizerRefund, error) {
	refunds, err := s.repo.PendingRefunds(ctx, organizer)
	if err != nil {
		return nil, fmt.Errorf("pending refunds: %w", err)
	}
	return refunds, nil
}

// DecideRefund approves/rejects a PENDING refund owned by the caller.
// Approving flips the order to REFUNDED (tickets follow). Returns the
// new status + amount for the audit line.
func (s *Service) DecideRefund(ctx context.Context, organizer, code string, approve bool) (string, float64, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return "", 0, fmt.Errorf("%w: refund code is required", services.ErrInvalidFilter)
	}
	orderCode, _, owner, status, amount, err := s.repo.RefundByCode(ctx, code)
	if err != nil {
		return "", 0, fmt.Errorf("%w: refund not found", services.ErrInvalidFilter)
	}
	if owner != organizer {
		return "", 0, services.ErrForbidden
	}
	if status != "PENDING" {
		return "", 0, fmt.Errorf("%w: refund already decided", services.ErrInvalidFilter)
	}
	next := "REJECTED"
	if approve {
		next = "APPROVED"
	}
	if err := s.repo.DecideRefund(ctx, code, next, organizer); err != nil {
		return "", 0, fmt.Errorf("decide refund: %w", err)
	}
	if approve {
		if err := s.repo.MarkOrderRefunded(ctx, orderCode); err != nil {
			return "", 0, fmt.Errorf("mark order refunded: %w", err)
		}
	}
	return next, amount, nil
}

func (s *Service) CheckinLive(ctx context.Context, organizer, eventCode string) (*models.CheckinLive, error) {
	eventCode = strings.TrimSpace(eventCode)
	if eventCode == "" {
		return nil, fmt.Errorf("%w: event is required", services.ErrInvalidFilter)
	}
	ok, err := s.repo.EventOwned(ctx, organizer, eventCode)
	if err != nil {
		return nil, fmt.Errorf("check event: %w", err)
	}
	if !ok {
		return nil, fmt.Errorf("%w: unknown event", services.ErrInvalidFilter)
	}
	title, checkedIn, total, err := s.repo.CheckinLive(ctx, organizer, eventCode)
	if err != nil {
		return nil, fmt.Errorf("checkin live: %w", err)
	}
	return &models.CheckinLive{
		EventCode: eventCode,
		Title:     title,
		CheckedIn: checkedIn,
		Total:     total,
		Rate:      rate(float64(checkedIn), float64(total)),
	}, nil
}

// ExportTables flattens the summary into titled string tables for CSV/PDF.
func (s *Service) ExportTables(ctx context.Context, organizer, preset, from, to, event, city string) ([]export.Table, error) {
	sum, err := s.Summary(ctx, organizer, preset, from, to, event, city)
	if err != nil {
		return nil, err
	}
	f := func(v float64) string { return fmt.Sprintf("%.2f", v) }
	kpis := [][]string{
		{"Tickets sold", fmt.Sprintf("%d", int(sum.TicketsSold.Value)), fmt.Sprintf("%.1f%%", sum.TicketsSold.DeltaPct)},
		{"Gross revenue", f(sum.GrossRevenue.Value), fmt.Sprintf("%.1f%%", sum.GrossRevenue.DeltaPct)},
		{"Net revenue", f(sum.NetRevenue.Value), fmt.Sprintf("%.1f%%", sum.NetRevenue.DeltaPct)},
		{"Active events", fmt.Sprintf("%d", int(sum.ActiveEvents.Value)), fmt.Sprintf("%.1f%%", sum.ActiveEvents.DeltaPct)},
		{"Check-in rate %", f(sum.CheckinRate.Value), fmt.Sprintf("%.1f%%", sum.CheckinRate.DeltaPct)},
		{"Refunds", fmt.Sprintf("%d", int(sum.RefundCount.Value)), fmt.Sprintf("%.1f%%", sum.RefundCount.DeltaPct)},
		{"Promo orders", fmt.Sprintf("%d", int(sum.PromoOrders.Value)), fmt.Sprintf("%.1f%%", sum.PromoOrders.DeltaPct)},
		{"Promo discount", f(sum.PromoDiscount.Value), fmt.Sprintf("%.1f%%", sum.PromoDiscount.DeltaPct)},
		{"Ad impressions", fmt.Sprintf("%d", int(sum.AdImpressions.Value)), fmt.Sprintf("%.1f%%", sum.AdImpressions.DeltaPct)},
		{"Ad clicks", fmt.Sprintf("%d", int(sum.AdClicks.Value)), fmt.Sprintf("%.1f%%", sum.AdClicks.DeltaPct)},
		{"Ad cost", f(sum.AdCost.Value), fmt.Sprintf("%.1f%%", sum.AdCost.DeltaPct)},
	}
	trend := make([][]string, 0, len(sum.SalesTrend))
	for _, p := range sum.SalesTrend {
		trend = append(trend, []string{p.Date, fmt.Sprintf("%d", p.Tickets), f(p.Revenue)})
	}
	upcoming := make([][]string, 0, len(sum.Upcoming))
	for _, u := range sum.Upcoming {
		upcoming = append(upcoming, []string{u.Code, u.Title, u.Start, fmt.Sprintf("%d/%d", u.Sold, u.Quota)})
	}
	orders := make([][]string, 0, len(sum.RecentOrders))
	for _, o := range sum.RecentOrders {
		orders = append(orders, []string{o.Code, o.Event, o.Buyer, o.TicketType, fmt.Sprintf("%d", o.Qty), f(o.Net), o.Status})
	}
	return []export.Table{
		{Title: fmt.Sprintf("Organizer KPI (%s to %s)", sum.Filter.From, sum.Filter.To),
			Headers: []string{"KPI", "Value", "Delta vs previous"}, Rows: kpis},
		{Title: "Sales trend", Headers: []string{"Date", "Tickets", "Revenue"}, Rows: trend},
		{Title: "Upcoming events", Headers: []string{"Code", "Title", "Start", "Sold/Quota"}, Rows: upcoming},
		{Title: "Recent orders", Headers: []string{"Code", "Event", "Buyer", "Type", "Qty", "Net", "Status"}, Rows: orders},
	}, nil
}
