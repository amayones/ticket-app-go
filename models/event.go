package models

// Event DTOs (features/event): My Events, Create Event, Ticket Types,
// Attendees. Ownership is always the caller's CPUSER code (never request
// input). Money is float64 (columns DECIMAL(18,2)).

// Event statuses and allowed transitions (enforced in service):
//   DRAFT -> PUBLISHED | CANCELLED | DELETED(soft)
//   PUBLISHED -> DRAFT (unpublish, no sales) | CANCELLED | ENDED (auto)
//   ENDED, CANCELLED: terminal.
const (
	EventDraft     = "DRAFT"
	EventPublished = "PUBLISHED"
	EventCompleted = "COMPLETED"
	EventEnded     = "ENDED"
	EventCancelled = "CANCELLED"
)

// EventInput is POST / PUT /api/events payload (create + edit, one form).
type EventInput struct {
	Title        string  `json:"title"`
	CategoryCode string  `json:"category_code"`
	Description  string  `json:"description"`
	CityCode     string  `json:"city_code"`
	Venue        string  `json:"venue"`
	Address      string  `json:"address"`
	Latitude     float64 `json:"latitude"`
	HasLatitude  bool    `json:"has_latitude"`
	Longitude    float64 `json:"longitude"`
	HasLongitude bool    `json:"has_longitude"`
	StartAt      string  `json:"start_at"`
	EndAt        string  `json:"end_at"`
	PosterURL    string  `json:"poster_url"`
	Terms        string  `json:"terms"`
}

// MyEventItem is one row of GET /api/events/mine.
type MyEventItem struct {
	Code      string  `json:"code"`
	Slug      string  `json:"slug"`
	Title     string  `json:"title"`
	Category  string  `json:"category"`
	City      string  `json:"city"`
	StartAt   string  `json:"start_at"`
	EndAt     string  `json:"end_at"`
	Status    string  `json:"status"`
	PosterURL string  `json:"poster_url"`
	Sold      int     `json:"sold"`
	Quota     int     `json:"quota"`
	Revenue   float64 `json:"revenue"`
}

// EventDetail is GET /api/events/:code (owner edit view) + publish checklist.
type EventDetailOwner struct {
	Code         string   `json:"code"`
	Slug         string   `json:"slug"`
	Title        string   `json:"title"`
	CategoryCode string   `json:"category_code"`
	Category     string   `json:"category"`
	Description  string   `json:"description"`
	CityCode     string   `json:"city_code"`
	City         string   `json:"city"`
	Venue        string   `json:"venue"`
	Address      string   `json:"address"`
	Latitude     float64  `json:"latitude"`
	Longitude    float64  `json:"longitude"`
	StartAt      string   `json:"start_at"`
	EndAt        string   `json:"end_at"`
	PosterURL    string   `json:"poster_url"`
	Terms        string   `json:"terms"`
	Status       string   `json:"status"`
	CancelReason string   `json:"cancel_reason"`
	Sold         int      `json:"sold"`
	// Missing lists unmet publish requirements (empty = ready to publish).
	Missing []string `json:"missing"`
	// NotifyBuyers warns that schedule/venue edits affect sold tickets
	// (notification sending belongs to another module).
	NotifyBuyers bool `json:"notify_buyers"`
}

// TicketTypeInput is POST / PUT ticket type payload.
type TicketTypeInput struct {
	Name         string  `json:"name"`
	Description  string  `json:"description"`
	Price        float64 `json:"price"`
	Quota        int     `json:"quota"`
	MaxPerPerson *int    `json:"max_per_person"`
	SaleStartAt  string  `json:"sale_start_at"`
	SaleEndAt    string  `json:"sale_end_at"`
	IsActive     *bool   `json:"is_active"`
	SortOrder    *int    `json:"sort_order"`
}

// TicketTypeRow is one row of the ticket type list.
type TicketTypeRow struct {
	Code         string  `json:"code"`
	Name         string  `json:"name"`
	Description  string  `json:"description"`
	Price        float64 `json:"price"`
	Quota        int     `json:"quota"`
	Sold         int     `json:"sold"`
	Remaining    int     `json:"remaining"`
	Revenue      float64 `json:"revenue"`
	MaxPerPerson int     `json:"max_per_person"`
	SaleStartAt  string  `json:"sale_start_at"`
	SaleEndAt    string  `json:"sale_end_at"`
	IsActive     bool    `json:"is_active"`
	SortOrder    int     `json:"sort_order"`
}

// TicketTypeSummary heads the Ticket Types page.
type TicketTypeSummary struct {
	TotalQuota   int     `json:"total_quota"`
	TotalSold    int     `json:"total_sold"`
	Potential    float64 `json:"potential"`
	ActiveCount  int     `json:"active_count"`
	EventCode    string  `json:"event_code"`
	EventTitle   string  `json:"event_title"`
	EventStatus  string  `json:"event_status"`
	PublishReady bool    `json:"publish_ready"`
	Missing      []string `json:"missing"`
}

// AttendeeRow is one row of GET /api/events/:code/attendees.
type AttendeeRow struct {
	TicketCode    string `json:"ticket_code"`
	BuyerName     string `json:"buyer_name"`
	BuyerUsername string `json:"buyer_username"`
	Email         string `json:"email"`
	Phone         string `json:"phone"`
	TicketType    string `json:"ticket_type"`
	OrderCode     string `json:"order_code"`
	PaymentStatus string `json:"payment_status"`
	Attendance    string `json:"attendance"`
	CheckedInAt   string `json:"checked_in_at"`
	PurchaseDate  string `json:"purchase_date"`
}

// AttendeeSummary heads the Attendees page.
type AttendeeSummary struct {
	Total     int     `json:"total"`
	CheckedIn int     `json:"checked_in"`
	Pending   int     `json:"pending"`
	Rate      float64 `json:"rate"`
}

// CancelRequest is POST /api/events/:code/cancel payload.
type CancelRequest struct {
	Reason string `json:"reason"`
}

// TicketOrderRequest is PUT ticket-types/order payload (display order).
type TicketOrderRequest struct {
	Order []string `json:"order"`
}
