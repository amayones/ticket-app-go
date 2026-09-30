package models

// Dashboard DTOs shared by the organizer and seller dashboards
// (features/organizer, features/seller). Money is float64 (columns are
// DECIMAL(18,2)); every KPI carries the previous-period value plus the
// percent delta so the frontend can render up/down comparisons.

const (
	// Date presets accepted by dashboard filters.
	PresetToday  = "today"
	Preset7Days  = "7days"
	Preset30Days = "30days"
	PresetMonth  = "month"
	PresetCustom = "custom"
)

// DashboardFilter is the validated request filter. Dates are inclusive;
// Event/City/Category empty means "all".
type DashboardFilter struct {
	From     string
	To       string
	Event    string
	City     string
	Category string
}

// KPIValue is one KPI with its previous-period comparison.
type KPIValue struct {
	Value    float64 `json:"value"`
	Previous float64 `json:"previous"`
	DeltaPct float64 `json:"delta_pct"`
}

// TrendPoint is one bucket of the sales trend chart.
type TrendPoint struct {
	Date    string  `json:"date"`
	Tickets int     `json:"tickets"`
	Revenue float64 `json:"revenue"`
}

// SlicePoint is one row of a breakdown chart (by ticket type, city, promo,
// product, or status).
type SlicePoint struct {
	Name    string  `json:"name"`
	Tickets int     `json:"tickets"`
	Revenue float64 `json:"revenue"`
	Orders  int     `json:"orders"`
}

// DashboardAlert is one actionable warning on the dashboard.
type DashboardAlert struct {
	Kind   string `json:"kind"`
	Title  string `json:"title"`
	Detail string `json:"detail"`
}

// Balance is escrow + available payout balance.
type Balance struct {
	Escrow    float64 `json:"escrow"`
	Available float64 `json:"available"`
}

// OrganizerOrder is one row of the recent-orders list.
type OrganizerOrder struct {
	Code       string  `json:"code"`
	Event      string  `json:"event"`
	Buyer      string  `json:"buyer"`
	TicketType string  `json:"ticket_type"`
	Qty        int     `json:"qty"`
	Net        float64 `json:"net"`
	Status     string  `json:"status"`
	OrderDate  string  `json:"order_date"`
}

// OrganizerRefund is one refund row (pending list + approve/reject target).
type OrganizerRefund struct {
	Code      string  `json:"code"`
	OrderCode string  `json:"order_code"`
	Event     string  `json:"event"`
	Amount    float64 `json:"amount"`
	Reason    string  `json:"reason"`
	Status    string  `json:"status"`
	CreatedAt string  `json:"created_at"`
}

// UpcomingEvent is one upcoming event with sold-vs-quota progress.
type UpcomingEvent struct {
	Code  string `json:"code"`
	Title string `json:"title"`
	Start string `json:"start_at"`
	Venue string `json:"venue"`
	Sold  int    `json:"sold"`
	Quota int    `json:"quota"`
}

// CityOption is one city filter option (owned events' cities).
type CityOption struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

// OrganizerSummary is the full GET /api/dashboard/organizer/summary payload.
type OrganizerSummary struct {
	Filter        DashboardFilter  `json:"filter"`
	TicketsSold   KPIValue         `json:"tickets_sold"`
	GrossRevenue  KPIValue         `json:"gross_revenue"`
	NetRevenue    KPIValue         `json:"net_revenue"`
	ActiveEvents  KPIValue         `json:"active_events"`
	CheckinRate   KPIValue         `json:"checkin_rate"`
	RefundCount   KPIValue         `json:"refund_count"`
	RefundPct     KPIValue         `json:"refund_pct"`
	PromoOrders   KPIValue         `json:"promo_orders"`
	PromoDiscount KPIValue         `json:"promo_discount"`
	AdImpressions KPIValue         `json:"ad_impressions"`
	AdClicks      KPIValue         `json:"ad_clicks"`
	AdCost        KPIValue         `json:"ad_cost"`
	SalesTrend    []TrendPoint     `json:"sales_trend"`
	ByTicketType  []SlicePoint     `json:"by_ticket_type"`
	ByCity        []SlicePoint     `json:"by_city"`
	ByPromo       []SlicePoint     `json:"by_promo"`
	Upcoming      []UpcomingEvent  `json:"upcoming_events"`
	RecentOrders  []OrganizerOrder `json:"recent_orders"`
	PendingRefund []OrganizerRefund `json:"pending_refunds"`
	Alerts        []DashboardAlert `json:"alerts"`
	Balance       Balance         `json:"balance"`
	PayoutReady   float64          `json:"payout_ready"`
	Cities        []CityOption    `json:"cities"`
}

// CheckinLive is the GET /api/dashboard/organizer/checkin-live payload:
// checked-in vs total tickets on the event day.
type CheckinLive struct {
	EventCode string `json:"event_code"`
	Title     string `json:"title"`
	CheckedIn int    `json:"checked_in"`
	Total     int    `json:"total"`
	Rate      float64 `json:"rate"`
}

// RefundDecision is the POST approve/reject payload.
type RefundDecision struct {
	Note string `json:"note"`
}

// SellerOrder is one row of the seller order lists.
type SellerOrder struct {
	Code        string  `json:"code"`
	Product     string  `json:"product"`
	Buyer       string  `json:"buyer"`
	Qty         int     `json:"qty"`
	Total       float64 `json:"total"`
	Status      string  `json:"status"`
	OrderDate   string  `json:"order_date"`
	ShipDeadline string `json:"ship_deadline,omitempty"`
}

// SellerProduct is one product row (low-stock list).
type SellerProduct struct {
	Code      string  `json:"code"`
	Name      string  `json:"name"`
	Category  string  `json:"category"`
	Price     float64 `json:"price"`
	Stock     int     `json:"stock"`
	Threshold int     `json:"threshold"`
	Status    string  `json:"status"`
}

// SellerComplaint is one open complaint row.
type SellerComplaint struct {
	Code      string `json:"code"`
	OrderCode string `json:"order_code"`
	Buyer     string `json:"buyer"`
	Message   string `json:"message"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

// SellerSummary is the full GET /api/dashboard/seller/summary payload.
type SellerSummary struct {
	Filter         DashboardFilter `json:"filter"`
	TotalSales     KPIValue        `json:"total_sales"`
	TotalOrders    KPIValue        `json:"total_orders"`
	Escrow         KPIValue        `json:"escrow"`
	Available      KPIValue        `json:"available"`
	ActiveProducts KPIValue        `json:"active_products"`
	LowStock       KPIValue        `json:"low_stock"`
	OpenComplaints KPIValue        `json:"open_complaints"`
	Rating         KPIValue        `json:"rating"`
	OrdersByStatus []SlicePoint    `json:"orders_by_status"`
	SalesTrend     []TrendPoint    `json:"sales_trend"`
	TopProducts    []SlicePoint    `json:"top_products"`
	Actionable     []SellerOrder   `json:"actionable_orders"`
	LowStockList   []SellerProduct `json:"low_stock_products"`
	RecentOrders   []SellerOrder   `json:"recent_orders"`
	Complaints     []SellerComplaint `json:"complaints"`
	Alerts         []DashboardAlert `json:"alerts"`
	Balance        Balance        `json:"balance"`
	Categories     []string        `json:"categories"`
}
