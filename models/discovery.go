package models

// Discovery DTOs (features/discovery): public read models for Home, News,
// Events, Event Detail, Promotions. Only publishable rows are ever returned
// (event PUBLISHED + not ended, news PUBLISHED, promo ACTIVE in period,
// campaign ACTIVE in period). Money is float64 (columns DECIMAL(18,2)).

// Event categories on the home shortcut row (code rallies the Events filter).
type EventCategory struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

// NewsCategory is one news filter option.
type NewsCategory struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

// Banner is one live ad campaign (always labelled "Iklan" in the UI).
type Banner struct {
	Code      string `json:"code"`
	Title     string `json:"title"`
	ImageURL  string `json:"image_url"`
	TargetURL string `json:"target_url"`
	Position  string `json:"position"`
}

// PromoItem is one promotion (active or upcoming - never expired/exhausted).
type PromoItem struct {
	Code            string  `json:"code"`
	PromoCode       string  `json:"promo_code"`
	Description     string  `json:"description"`
	DiscountPercent float64 `json:"discount_percent"`
	DiscountAmount  float64 `json:"discount_amount"`
	MaxDiscount     float64 `json:"max_discount"`
	MinPurchase     float64 `json:"min_purchase"`
	Remaining       int     `json:"remaining"`
	ValidFrom       string  `json:"valid_from"`
	ValidTo         string  `json:"valid_to"`
	State           string  `json:"state"`
	EventCode       string  `json:"event_code,omitempty"`
	EventTitle      string  `json:"event_title,omitempty"`
	CategoryCode    string  `json:"category_code,omitempty"`
	CategoryName    string  `json:"category_name,omitempty"`
}

// NewsItem is one row of the news list.
type NewsItem struct {
	Code        string `json:"code"`
	Title       string `json:"title"`
	Slug        string `json:"slug"`
	Summary     string `json:"summary"`
	ImageURL    string `json:"image_url"`
	Category    string `json:"category"`
	PublishedAt string `json:"published_at"`
	ViewCount   int    `json:"view_count"`
}

// NewsDetail is the article reader payload (body sanitized for XSS).
type NewsDetail struct {
	NewsItem
	Body         string    `json:"body"`
	EventCode    string    `json:"event_code,omitempty"`
	EventTitle   string    `json:"event_title,omitempty"`
	RelatedNews  []NewsItem `json:"related_news"`
	RelatedEvent *EventCard `json:"related_event"`
}

// EventCard is one event row (list, popular, upcoming, related).
type EventCard struct {
	Code      string  `json:"code"`
	Title     string  `json:"title"`
	Category  string  `json:"category"`
	City      string  `json:"city"`
	Venue     string  `json:"venue"`
	StartAt   string  `json:"start_at"`
	EndAt     string  `json:"end_at"`
	PosterURL string  `json:"poster_url"`
	MinPrice  float64 `json:"min_price"`
	Sold      int     `json:"sold"`
	Quota     int     `json:"quota"`
	Badge     string  `json:"badge"`
}

// TicketTypeInfo is one ticket row on the event detail page.
type TicketTypeInfo struct {
	Code      string  `json:"code"`
	Name      string  `json:"name"`
	Price     float64 `json:"price"`
	Quota     int     `json:"quota"`
	Sold      int     `json:"sold"`
	Remaining int     `json:"remaining"`
	State     string  `json:"state"`
}

// EventDetail is the full GET /api/discovery/events/:code payload.
type EventDetail struct {
	Code         string           `json:"code"`
	Title        string           `json:"title"`
	Description  string           `json:"description"`
	Category     string           `json:"category"`
	CategoryCode string           `json:"category_code"`
	Organizer    string           `json:"organizer"`
	StartAt     string           `json:"start_at"`
	EndAt       string           `json:"end_at"`
	Venue       string           `json:"venue"`
	Address     string           `json:"address"`
	City        string           `json:"city"`
	Latitude    float64          `json:"latitude"`
	Longitude   float64          `json:"longitude"`
	PosterURL   string           `json:"poster_url"`
	Status      string           `json:"status"`
	TicketTypes []TicketTypeInfo `json:"ticket_types"`
	Promos      []PromoItem      `json:"promos"`
	Related     []EventCard      `json:"related"`
	SoldOut     bool             `json:"sold_out"`
}

// HomeResponse is the full GET /api/discovery/home payload.
type HomeResponse struct {
	Banners    []Banner         `json:"banners"`
	Promos     []PromoItem      `json:"promos"`
	News       []NewsItem       `json:"news"`
	Popular    []EventCard      `json:"popular"`
	Upcoming   []EventCard      `json:"upcoming"`
	Categories []EventCategory  `json:"categories"`
}

// DiscoveryFilter carries validated list filters (whitelisted sorts only).
type DiscoveryFilter struct {
	Categories []string
	City       string
	Category   string
	From       string
	To         string
	MinPrice   float64
	MaxPrice   float64
	Query      string
	Sort       string
	Limit      int
	Offset     int
}
