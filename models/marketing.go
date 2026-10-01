package models

const (
	PromoScopeAll      = "ALL"
	PromoScopeCategory = "CATEGORY"
	PromoScopeEvent    = "EVENT"
)

const (
	PromoStatusActive   = "ACTIVE"
	PromoStatusDisabled = "DISABLED"
	PromoStatusExpired  = "EXPIRED"
)

const (
	AdStatusDraft           = "DRAFT"
	AdStatusPendingPayment  = "PENDING_PAYMENT"
	AdStatusPendingApproval = "PENDING_APPROVAL"
	AdStatusLive            = "LIVE"
	AdStatusPaused          = "PAUSED"
	AdStatusEnded           = "ENDED"
	AdStatusRejected        = "REJECTED"
)

const (
	AdPayUnpaid  = "UNPAID"
	AdPayPending = "PENDING"
	AdPayPaid    = "PAID"
)

const (
	AdApprovalPending  = "PENDING"
	AdApprovalApproved = "APPROVED"
	AdApprovalRejected = "REJECTED"
)

const (
	NewsStatusDraft     = "DRAFT"
	NewsStatusPublished = "PUBLISHED"
	NewsStatusArchived  = "ARCHIVED"
)

type PromoInput struct {
	PromoCode       string   `json:"promo_code"`
	Description     string   `json:"description"`
	DiscountPercent float64  `json:"discount_percent"`
	DiscountAmount  float64  `json:"discount_amount"`
	MaxDiscount     float64  `json:"max_discount"`
	MinPurchase     float64  `json:"min_purchase"`
	MaxUses         *int     `json:"max_uses"`
	MaxPerUser      *int     `json:"max_per_user"`
	ValidFrom       string   `json:"valid_from"`
	ValidTo         string   `json:"valid_to"`
	ScopeType       string   `json:"scope_type"`
	CategoryCodes   []string `json:"category_codes"`
	EventCodes      []string `json:"event_codes"`
}

type PromoRow struct {
	Code            string   `json:"code"`
	PromoCode       string   `json:"promo_code"`
	Description     string   `json:"description"`
	DiscountPercent float64  `json:"discount_percent"`
	DiscountAmount  float64  `json:"discount_amount"`
	MaxDiscount     float64  `json:"max_discount"`
	MinPurchase     float64  `json:"min_purchase"`
	MaxUses         *int     `json:"max_uses"`
	MaxPerUser      *int     `json:"max_per_user"`
	UsedCount       int      `json:"used_count"`
	Remaining       int      `json:"remaining"`
	ValidFrom       string   `json:"valid_from"`
	ValidTo         string   `json:"valid_to"`
	Status          string   `json:"status"`
	DisplayStatus   string   `json:"display_status"`
	ScopeType       string   `json:"scope_type"`
	CategoryCodes   []string `json:"category_codes"`
	EventCodes      []string `json:"event_codes"`
	OrganizerCode   string   `json:"organizer_code"`
	CreatedAt       string   `json:"created_at"`
	UpdatedAt       string   `json:"updated_at"`
}

type PromoUsageRow struct {
	Code       string  `json:"code"`
	OrderCode  string  `json:"order_code"`
	UserCode   string  `json:"user_code"`
	Username   string  `json:"username"`
	Discount   float64 `json:"discount"`
	UsedAt     string  `json:"used_at"`
}

type AdPosition struct {
	Code        string  `json:"code"`
	Name        string  `json:"name"`
	PricePerDay float64 `json:"price_per_day"`
	SortOrder   int     `json:"sort_order"`
}

type AdCampaignInput struct {
	Title     string `json:"title"`
	TargetURL string `json:"target_url"`
	TargetEventCode string `json:"target_event_code"`
	Position  string `json:"position"`
	StartAt   string `json:"start_at"`
	EndAt     string `json:"end_at"`
	ImageURL  string `json:"image_url"`
}

type AdCampaignRow struct {
	Code           string  `json:"code"`
	Title          string  `json:"title"`
	ImageURL       string  `json:"image_url"`
	TargetURL      string  `json:"target_url"`
	Position       string  `json:"position"`
	PositionName   string  `json:"position_name"`
	StartAt        string  `json:"start_at"`
	EndAt          string  `json:"end_at"`
	Status         string  `json:"status"`
	DisplayStatus  string  `json:"display_status"`
	PaymentStatus  string  `json:"payment_status"`
	ApprovalStatus string  `json:"approval_status"`
	Cost           float64 `json:"cost"`
	PricePerDay    float64 `json:"price_per_day"`
	Days           int     `json:"days"`
	Impressions    int     `json:"impressions"`
	Clicks         int     `json:"clicks"`
	CTR            float64 `json:"ctr"`
	RejectNote     string  `json:"reject_note"`
	CreatedAt      string  `json:"created_at"`
	UpdatedAt      string  `json:"updated_at"`
	OrganizerCode  string  `json:"organizer_code"`
}

type AdDailyPoint struct {
	Date        string `json:"date"`
	Impressions int    `json:"impressions"`
	Clicks      int    `json:"clicks"`
}

type NewsInput struct {
	Title              string `json:"title"`
	Summary            string `json:"summary"`
	Body               string `json:"body"`
	ImageURL           string `json:"image_url"`
	NewsCategoryCode   string `json:"news_category_code"`
	EventCategoryCode  string `json:"event_category_code"`
	EventCode          string `json:"event_code"`
	ScheduledAt        string `json:"scheduled_at"`
}

type NewsRow struct {
	Code              string `json:"code"`
	Title             string `json:"title"`
	Slug              string `json:"slug"`
	Summary           string `json:"summary"`
	Body              string `json:"body"`
	ImageURL          string `json:"image_url"`
	NewsCategoryCode  string `json:"news_category_code"`
	NewsCategoryName  string `json:"news_category_name"`
	EventCategoryCode string `json:"event_category_code"`
	EventCategoryName string `json:"event_category_name"`
	EventCode         string `json:"event_code"`
	EventTitle        string `json:"event_title"`
	Status            string `json:"status"`
	AuthorCode        string `json:"author_code"`
	AuthorName        string `json:"author_name"`
	ScheduledAt       string `json:"scheduled_at"`
	PublishedAt       string `json:"published_at"`
	ViewCount         int    `json:"view_count"`
	IsDeleted         bool   `json:"is_deleted"`
	CreatedAt         string `json:"created_at"`
	UpdatedAt         string `json:"updated_at"`
}
