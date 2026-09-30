package seller

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
// own products and orders, and assembles the summary payload.
type ServiceInterface interface {
	Summary(ctx context.Context, seller, preset, from, to, category string) (*models.SellerSummary, error)
	Orders(ctx context.Context, seller, preset, from, to, category string, limit int) ([]models.SellerOrder, error)
	Actionable(ctx context.Context, seller string) ([]models.SellerOrder, error)
	LowStock(ctx context.Context, seller string) ([]models.SellerProduct, error)
	Complaints(ctx context.Context, seller string) ([]models.SellerComplaint, error)
	ExportTables(ctx context.Context, seller, preset, from, to, category string) ([]export.Table, error)
}

type Service struct {
	repo RepositoryInterface
}

func NewService(repo RepositoryInterface) *Service {
	return &Service{repo: repo}
}

// resolveFilter validates preset/dates + category ownership.
func (s *Service) resolveFilter(ctx context.Context, seller, preset, from, to, category string) (models.DashboardFilter, utils.DateRange, error) {
	rng, err := utils.ParseDateRange(strings.TrimSpace(preset), strings.TrimSpace(from), strings.TrimSpace(to))
	if err != nil {
		return models.DashboardFilter{}, rng, fmt.Errorf("%w: %v", services.ErrInvalidFilter, err)
	}
	category = strings.TrimSpace(category)
	if category != "" {
		cats, err := s.repo.CategoryList(ctx, seller)
		if err != nil {
			return models.DashboardFilter{}, rng, fmt.Errorf("list categories: %w", err)
		}
		known := false
		for _, c := range cats {
			if strings.EqualFold(c, category) {
				category = c
				known = true
				break
			}
		}
		if !known {
			return models.DashboardFilter{}, rng, fmt.Errorf("%w: unknown category", services.ErrInvalidFilter)
		}
	}
	return models.DashboardFilter{From: rng.From, To: rng.To, Category: category}, rng, nil
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

func (s *Service) Summary(ctx context.Context, seller, preset, from, to, category string) (*models.SellerSummary, error) {
	filter, rng, err := s.resolveFilter(ctx, seller, preset, from, to, category)
	if err != nil {
		return nil, err
	}
	end := plusOne(rng.To)
	prev := utils.PreviousRange(rng)
	prevEnd := plusOne(prev.To)

	sales, orders, err := s.repo.OrderSums(ctx, seller, category, rng.From, end)
	if err != nil {
		return nil, fmt.Errorf("order sums: %w", err)
	}
	psales, porders, err := s.repo.OrderSums(ctx, seller, category, prev.From, prevEnd)
	if err != nil {
		return nil, fmt.Errorf("previous order sums: %w", err)
	}
	byStatus, err := s.repo.OrdersByStatus(ctx, seller, category, rng.From, end)
	if err != nil {
		return nil, fmt.Errorf("orders by status: %w", err)
	}
	trend, err := s.repo.SalesTrend(ctx, seller, category, rng.From, end)
	if err != nil {
		return nil, fmt.Errorf("sales trend: %w", err)
	}
	top, err := s.repo.TopProducts(ctx, seller, category, rng.From, end)
	if err != nil {
		return nil, fmt.Errorf("top products: %w", err)
	}
	actionable, err := s.repo.ActionableOrders(ctx, seller)
	if err != nil {
		return nil, fmt.Errorf("actionable orders: %w", err)
	}
	lowList, err := s.repo.LowStockProducts(ctx, seller)
	if err != nil {
		return nil, fmt.Errorf("low stock products: %w", err)
	}
	recent, err := s.repo.RecentOrders(ctx, seller, category, 10)
	if err != nil {
		return nil, fmt.Errorf("recent orders: %w", err)
	}
	complaints, err := s.repo.OpenComplaints(ctx, seller)
	if err != nil {
		return nil, fmt.Errorf("open complaints: %w", err)
	}
	active, err := s.repo.ActiveProducts(ctx, seller)
	if err != nil {
		return nil, fmt.Errorf("active products: %w", err)
	}
	low, err := s.repo.LowStockCount(ctx, seller)
	if err != nil {
		return nil, fmt.Errorf("low stock count: %w", err)
	}
	open, err := s.repo.OpenComplaintCount(ctx, seller)
	if err != nil {
		return nil, fmt.Errorf("open complaint count: %w", err)
	}
	rating, err := s.repo.Rating(ctx, seller)
	if err != nil {
		return nil, fmt.Errorf("rating: %w", err)
	}
	rejected, err := s.repo.RejectedProducts(ctx, seller)
	if err != nil {
		return nil, fmt.Errorf("rejected products: %w", err)
	}
	balance, err := s.repo.Balance(ctx, seller)
	if err != nil {
		return nil, fmt.Errorf("balance: %w", err)
	}
	payout, err := s.repo.PayoutReady(ctx, seller)
	if err != nil {
		return nil, fmt.Errorf("payout ready: %w", err)
	}
	categories, err := s.repo.CategoryList(ctx, seller)
	if err != nil {
		return nil, fmt.Errorf("categories: %w", err)
	}

	alerts := make([]models.DashboardAlert, 0)
	soon := time.Now().AddDate(0, 0, 2).Format("2006-01-02 15:04:05")
	for _, o := range actionable {
		if o.ShipDeadline != "" && o.ShipDeadline <= soon {
			alerts = append(alerts, models.DashboardAlert{
				Kind:   "SHIP",
				Title:  "Ship deadline near",
				Detail: fmt.Sprintf("Order %s (%s) ships by %s", o.Code, o.Product, o.ShipDeadline),
			})
		}
	}
	if rejected > 0 {
		alerts = append(alerts, models.DashboardAlert{
			Kind:   "MODERATION",
			Title:  "Products rejected",
			Detail: fmt.Sprintf("%d product(s) rejected by moderation", rejected),
		})
	}
	if payout > 0 {
		alerts = append(alerts, models.DashboardAlert{
			Kind:   "PAYOUT",
			Title:  "Balance ready",
			Detail: fmt.Sprintf("%.2f ready to withdraw", payout),
		})
	}

	return &models.SellerSummary{
		Filter:         filter,
		TotalSales:     kpi(sales, psales),
		TotalOrders:    kpi(float64(orders), float64(porders)),
		Escrow:         kpi(balance.Escrow, balance.Escrow),
		Available:      kpi(balance.Available, balance.Available),
		ActiveProducts: kpi(float64(active), float64(active)),
		LowStock:       kpi(float64(low), float64(low)),
		OpenComplaints: kpi(float64(open), float64(open)),
		Rating:         kpi(rating, rating),
		OrdersByStatus: byStatus,
		SalesTrend:     trend,
		TopProducts:    top,
		Actionable:     actionable,
		LowStockList:   lowList,
		RecentOrders:   recent,
		Complaints:     complaints,
		Alerts:         alerts,
		Balance:        balance,
		Categories:     categories,
	}, nil
}

func (s *Service) Orders(ctx context.Context, seller, preset, from, to, category string, limit int) ([]models.SellerOrder, error) {
	filter, _, err := s.resolveFilter(ctx, seller, preset, from, to, category)
	if err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 100 {
		limit = 10
	}
	orders, err := s.repo.RecentOrders(ctx, seller, filter.Category, limit)
	if err != nil {
		return nil, fmt.Errorf("recent orders: %w", err)
	}
	return orders, nil
}

func (s *Service) Actionable(ctx context.Context, seller string) ([]models.SellerOrder, error) {
	orders, err := s.repo.ActionableOrders(ctx, seller)
	if err != nil {
		return nil, fmt.Errorf("actionable orders: %w", err)
	}
	return orders, nil
}

func (s *Service) LowStock(ctx context.Context, seller string) ([]models.SellerProduct, error) {
	products, err := s.repo.LowStockProducts(ctx, seller)
	if err != nil {
		return nil, fmt.Errorf("low stock products: %w", err)
	}
	return products, nil
}

func (s *Service) Complaints(ctx context.Context, seller string) ([]models.SellerComplaint, error) {
	complaints, err := s.repo.OpenComplaints(ctx, seller)
	if err != nil {
		return nil, fmt.Errorf("open complaints: %w", err)
	}
	return complaints, nil
}

// ExportTables flattens the summary into titled string tables (CSV only).
func (s *Service) ExportTables(ctx context.Context, seller, preset, from, to, category string) ([]export.Table, error) {
	sum, err := s.Summary(ctx, seller, preset, from, to, category)
	if err != nil {
		return nil, err
	}
	f := func(v float64) string { return fmt.Sprintf("%.2f", v) }
	kpis := [][]string{
		{"Total sales", f(sum.TotalSales.Value), fmt.Sprintf("%.1f%%", sum.TotalSales.DeltaPct)},
		{"Total orders", fmt.Sprintf("%d", int(sum.TotalOrders.Value)), fmt.Sprintf("%.1f%%", sum.TotalOrders.DeltaPct)},
		{"Escrow", f(sum.Escrow.Value), ""},
		{"Available", f(sum.Available.Value), ""},
		{"Active products", fmt.Sprintf("%d", int(sum.ActiveProducts.Value)), ""},
		{"Low stock", fmt.Sprintf("%d", int(sum.LowStock.Value)), ""},
		{"Open complaints", fmt.Sprintf("%d", int(sum.OpenComplaints.Value)), ""},
		{"Rating", fmt.Sprintf("%.2f", sum.Rating.Value), ""},
	}
	byStatus := make([][]string, 0, len(sum.OrdersByStatus))
	for _, p := range sum.OrdersByStatus {
		byStatus = append(byStatus, []string{p.Name, fmt.Sprintf("%d", p.Orders), f(p.Revenue)})
	}
	top := make([][]string, 0, len(sum.TopProducts))
	for _, p := range sum.TopProducts {
		top = append(top, []string{p.Name, fmt.Sprintf("%d", p.Tickets), f(p.Revenue)})
	}
	orders := make([][]string, 0, len(sum.RecentOrders))
	for _, o := range sum.RecentOrders {
		orders = append(orders, []string{o.Code, o.Product, o.Buyer, fmt.Sprintf("%d", o.Qty), f(o.Total), o.Status})
	}
	return []export.Table{
		{Title: fmt.Sprintf("Seller KPI (%s to %s)", sum.Filter.From, sum.Filter.To),
			Headers: []string{"KPI", "Value", "Delta vs previous"}, Rows: kpis},
		{Title: "Orders by status", Headers: []string{"Status", "Orders", "Revenue"}, Rows: byStatus},
		{Title: "Top products", Headers: []string{"Product", "Qty", "Revenue"}, Rows: top},
		{Title: "Recent orders", Headers: []string{"Code", "Product", "Buyer", "Qty", "Total", "Status"}, Rows: orders},
	}, nil
}
