package models

// Ticketing DTOs (features/ticketing): Checkout, My Tickets, Ticket Detail,
// Refunds, Scan Ticket, Scan History. Buyer identity always comes from the
// caller's token (never request input). Money is float64 (DECIMAL(18,2)).

// Order statuses: PENDING (awaiting payment, quota held), PAID, CANCELLED,
// REFUNDED, EXPIRED (hold lapsed, quota released).
const (
	OrderPending   = "PENDING"
	OrderPaid      = "PAID"
	OrderCancelled = "CANCELLED"
	OrderRefunded  = "REFUNDED"
	OrderExpired   = "EXPIRED"
)

// Default quota hold for unpaid orders (minutes, overridable per request? no:
// fixed constant so expiry is predictable for buyers and sellers).
const DefaultHoldMinutes = 15

// CheckoutItem is one line of the checkout request.
type CheckoutItem struct {
	TypeCode string `json:"type_code"`
	Qty      int    `json:"qty"`
}

// CheckoutRequest is POST /api/ticketing/checkout payload.
type CheckoutRequest struct {
	EventCode  string         `json:"event_code"`
	Items      []CheckoutItem `json:"items"`
	BuyerName  string         `json:"buyer_name"`
	BuyerEmail string         `json:"buyer_email"`
	BuyerPhone string         `json:"buyer_phone"`
	PromoCode  string         `json:"promo_code"`
}

// CheckoutResponse returns the pending order + payment countdown.
type CheckoutResponse struct {
	OrderCode        string  `json:"order_code"`
	Gross            float64 `json:"gross"`
	Discount         float64 `json:"discount"`
	Fee              float64 `json:"fee"`
	Net              float64 `json:"net"`
	ExpiresAt        string  `json:"expires_at"`
	CountdownSeconds int     `json:"countdown_seconds"`
}

// PayRequest is POST /api/ticketing/orders/:code/pay payload.
// FINANCE seam: method names the rail (SIMULATION today, gateway names
// later); ref carries the gateway reference. Simulation only runs outside
// production (enforced in service).
type PayRequest struct {
	Simulate bool   `json:"simulate"`
	Method   string `json:"method"`
	Ref      string `json:"ref"`
}

// PayResponse is the paid order + issued tickets.
type PayResponse struct {
	OrderCode string   `json:"order_code"`
	Status    string   `json:"status"`
	Tickets   []string `json:"tickets"`
}

// MyTicketItem is one row of GET /api/ticketing/tickets.
type MyTicketItem struct {
	TicketCode string  `json:"ticket_code"`
	EventCode  string  `json:"event_code"`
	EventTitle string  `json:"event_title"`
	StartAt    string  `json:"start_at"`
	Venue      string  `json:"venue"`
	TypeName   string  `json:"type_name"`
	HolderName string  `json:"holder_name"`
	Status     string  `json:"status"`
	OrderCode  string  `json:"order_code"`
	Net        float64 `json:"net"`
}

// TicketDetail is GET /api/ticketing/tickets/:code payload.
type TicketDetail struct {
	TicketCode string  `json:"ticket_code"`
	QRToken    string  `json:"qr_token"`
	HolderName string  `json:"holder_name"`
	TypeName   string  `json:"type_name"`
	EventCode  string  `json:"event_code"`
	EventTitle string  `json:"event_title"`
	StartAt    string  `json:"start_at"`
	EndAt      string  `json:"end_at"`
	Venue      string  `json:"venue"`
	Address    string  `json:"address"`
	City       string  `json:"city"`
	PosterURL  string  `json:"poster_url"`
	OrderCode  string  `json:"order_code"`
	Net        float64 `json:"net"`
	Status     string  `json:"status"`
	Active     bool    `json:"active"`
}

// RefundRequest is POST /api/ticketing/refunds payload (per order).
type RefundRequest struct {
	OrderCode string `json:"order_code"`
	Reason    string `json:"reason"`
}

// RefundItem is one row of GET /api/ticketing/refunds.
type RefundItem struct {
	Code        string  `json:"code"`
	OrderCode   string  `json:"order_code"`
	EventTitle  string  `json:"event_title"`
	Amount      float64 `json:"amount"`
	Reason      string  `json:"reason"`
	RejectNote  string  `json:"reject_note"`
	Status      string  `json:"status"`
	CreatedAt   string  `json:"created_at"`
	DecidedAt   string  `json:"decided_at"`
	Payout      string  `json:"payout_status"`
}

// ScanRequest is POST /api/ticketing/scan payload (token or code + event).
type ScanRequest struct {
	Token string `json:"token"`
	Code  string `json:"code"`
	Event string `json:"event"`
}

// ScanResult is the POST /api/ticketing/scan payload response.
type ScanResult struct {
	Result     string `json:"result"`
	TicketCode string `json:"ticket_code,omitempty"`
	HolderName string `json:"holder_name,omitempty"`
	TypeName   string `json:"type_name,omitempty"`
	Detail     string `json:"detail,omitempty"`
	CheckedIn  int    `json:"checked_in"`
	Total      int    `json:"total"`
}

// ScanSummary is GET /api/ticketing/scan/summary payload.
type ScanSummary struct {
	EventCode string `json:"event_code"`
	Title     string `json:"title"`
	CheckedIn int    `json:"checked_in"`
	Total     int    `json:"total"`
}

// ScanLogRow is one row of GET /api/ticketing/scan/history.
type ScanLogRow struct {
	Code       string `json:"code"`
	CreatedAt  string `json:"created_at"`
	TicketCode string `json:"ticket_code"`
	HolderName string `json:"holder_name"`
	TypeName   string `json:"type_name"`
	Result     string `json:"result"`
	Officer    string `json:"officer"`
}

// ScanHistorySummary heads the history page + export.
type ScanHistorySummary struct {
	Total   int `json:"total"`
	Valid   int `json:"valid"`
	Denied  int `json:"denied"`
}
