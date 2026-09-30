package ticketing

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"golang-backend/internal/export"
	"golang-backend/models"
	"golang-backend/services"
	"golang-backend/utils"
)

// Ticketing service: checkout pricing + holds, simulated settlement (FINANCE
// seam), buyer ticket reads, refund filing, atomic scanning. Buyer identity
// is always the caller's code; ownership violations map to 404/403.
type ServiceInterface interface {
	Checkout(ctx context.Context, buyer string, req models.CheckoutRequest) (*models.CheckoutResponse, error)
	PayOrder(ctx context.Context, buyer, orderCode string, req models.PayRequest, holderName, email, phone string) (*models.PayResponse, error)
	CancelOrder(ctx context.Context, buyer, orderCode string) error
	MyTickets(ctx context.Context, buyer, display, q string, limit, offset int) ([]models.MyTicketItem, error)
	TicketDetail(ctx context.Context, buyer, ticketCode string) (*models.TicketDetail, error)
	RequestRefund(ctx context.Context, buyer, orderCode, reason string) (string, error)
	BuyerRefunds(ctx context.Context, buyer string) ([]models.RefundItem, error)
	ScanTicket(ctx context.Context, officer string, req models.ScanRequest) (*models.ScanResult, error)
	ScanSummary(ctx context.Context, officer, event string) (*models.ScanSummary, error)
	ScanHistory(ctx context.Context, officer, event, result, officerFilter, from, to string, limit, offset int) ([]models.ScanLogRow, models.ScanHistorySummary, error)
	ScanExport(ctx context.Context, officer, event, result, officerFilter, from, to string) ([]export.Table, error)
	ScanEvents(ctx context.Context, officer string) ([]AssignedEvent, error)
}

type Service struct {
	repo        RepositoryInterface
	qrSecret    string
	appEnv      string
	holdMinutes int
	cutoffHours int
}

func NewService(repo RepositoryInterface, qrSecret, appEnv string, holdMinutes, cutoffHours int) *Service {
	if holdMinutes <= 0 {
		holdMinutes = models.DefaultHoldMinutes
	}
	return &Service{repo: repo, qrSecret: qrSecret, appEnv: appEnv, holdMinutes: holdMinutes, cutoffHours: cutoffHours}
}

func nowStr() string { return time.Now().Format("2006-01-02 15:04:05") }

func isDuplicateErr(err error) bool {
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "2627") || strings.Contains(msg, "2601") ||
		strings.Contains(msg, "unique") || strings.Contains(msg, "duplicate")
}

func (s *Service) Checkout(ctx context.Context, buyer string, req models.CheckoutRequest) (*models.CheckoutResponse, error) {
	_ = s.repo.ExpirePending(ctx, nowStr()) // opportunistic; checkout must not fail on it
	eventCode := strings.TrimSpace(req.EventCode)
	if eventCode == "" {
		return nil, fmt.Errorf("%w: event is required", services.ErrInvalidCheckout)
	}
	if len(req.Items) == 0 {
		return nil, fmt.Errorf("%w: at least one ticket is required", services.ErrInvalidCheckout)
	}
	ev, err := s.repo.EventInfo(ctx, eventCode)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: event not found", services.ErrInvalidCheckout)
		}
		return nil, fmt.Errorf("event lookup: %w", err)
	}
	if ev.Status != models.EventPublished {
		return nil, fmt.Errorf("%w: event is not on sale", services.ErrInvalidCheckout)
	}
	if ev.EndAt != "" && ev.EndAt <= nowStr() {
		return nil, fmt.Errorf("%w: event has ended", services.ErrInvalidCheckout)
	}
	buyerName := strings.TrimSpace(req.BuyerName)
	if buyerName == "" {
		return nil, fmt.Errorf("%w: buyer name is required", services.ErrInvalidCheckout)
	}
	if req.BuyerEmail != "" && !utils.IsValidEmail(strings.TrimSpace(req.BuyerEmail)) {
		return nil, fmt.Errorf("%w: invalid buyer email", services.ErrInvalidCheckout)
	}

	type line struct {
		info  *TypeInfo
		qty   int
		price float64
	}
	lines := make([]line, 0, len(req.Items))
	seen := make(map[string]bool, len(req.Items))
	var gross float64
	totalQty := 0
	for _, it := range req.Items {
		tcode := strings.TrimSpace(it.TypeCode)
		if tcode == "" || seen[tcode] {
			return nil, fmt.Errorf("%w: duplicate or empty ticket type", services.ErrInvalidCheckout)
		}
		seen[tcode] = true
		if it.Qty < 1 {
			return nil, fmt.Errorf("%w: quantity must be at least 1", services.ErrInvalidCheckout)
		}
		info, err := s.repo.TypeInfo(ctx, tcode)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, fmt.Errorf("%w: unknown ticket type", services.ErrInvalidCheckout)
			}
			return nil, fmt.Errorf("type lookup: %w", err)
		}
		if info.EventCode != eventCode || !info.IsActive {
			return nil, fmt.Errorf("%w: ticket type not on sale", services.ErrInvalidCheckout)
		}
		if info.SaleFrom != "" && info.SaleFrom > nowStr() {
			return nil, fmt.Errorf("%w: sale has not started", services.ErrInvalidCheckout)
		}
		if info.SaleTo != "" && info.SaleTo < nowStr() {
			return nil, fmt.Errorf("%w: sale has ended", services.ErrInvalidCheckout)
		}
		sold, err := s.repo.SoldQty(ctx, tcode)
		if err != nil {
			return nil, fmt.Errorf("sold lookup: %w", err)
		}
		held, err := s.repo.ReservedQty(ctx, tcode, nowStr())
		if err != nil {
			return nil, fmt.Errorf("hold lookup: %w", err)
		}
		if sold+held+it.Qty > info.Quota {
			return nil, fmt.Errorf("%w: not enough quota (remaining %d)", services.ErrInvalidCheckout, info.Quota-sold-held)
		}
		if info.HasMax {
			owned, err := s.repo.BuyerTypeQty(ctx, buyer, tcode, nowStr())
			if err != nil {
				return nil, fmt.Errorf("buyer qty lookup: %w", err)
			}
			if owned+it.Qty > info.MaxPerPerson {
				return nil, fmt.Errorf("%w: max %d per person", services.ErrInvalidCheckout, info.MaxPerPerson)
			}
		}
		lines = append(lines, line{info: info, qty: it.Qty, price: info.Price})
		gross += info.Price * float64(it.Qty)
		totalQty += it.Qty
	}

	// Promo validation (ACTIVE, in period, quota left, scope match).
	var discount float64
	promoCode := strings.TrimSpace(strings.ToUpper(req.PromoCode))
	if promoCode != "" {
		p, err := s.repo.PromoInfo(ctx, promoCode)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, fmt.Errorf("%w: unknown promo code", services.ErrInvalidCheckout)
			}
			return nil, fmt.Errorf("promo lookup: %w", err)
		}
		now := nowStr()
		if p.Status != "ACTIVE" || p.ValidFrom > now || p.ValidTo < now {
			return nil, fmt.Errorf("%w: promo is not active", services.ErrInvalidCheckout)
		}
		if p.HasMax && p.UsedCount >= p.MaxUses {
			return nil, fmt.Errorf("%w: promo quota exhausted", services.ErrInvalidCheckout)
		}
		if p.EventCode != "" && p.EventCode != eventCode {
			return nil, fmt.Errorf("%w: promo does not apply to this event", services.ErrInvalidCheckout)
		}
		if p.EventCode == "" && p.CoverageCat != "" && p.CoverageCat != ev.Category {
			return nil, fmt.Errorf("%w: promo does not apply to this category", services.ErrInvalidCheckout)
		}
		if gross < p.MinPurchase {
			return nil, fmt.Errorf("%w: minimum purchase not met", services.ErrInvalidCheckout)
		}
		if p.DiscountPercent > 0 {
			discount = gross * p.DiscountPercent / 100
			if p.MaxDiscount > 0 && discount > p.MaxDiscount {
				discount = p.MaxDiscount
			}
		} else {
			discount = p.DiscountAmount
		}
		if discount > gross {
			discount = gross
		}
	}
	fee := (gross - discount) * ev.FeePercent / 100
	net := gross - discount + fee

	orderCode, err := utils.GenerateCode(utils.OrderCodePrefix)
	if err != nil {
		return nil, err
	}
	expires := time.Now().Add(time.Duration(s.holdMinutes) * time.Minute)
	po := &PendingOrder{
		OrderCode: orderCode,
		EventCode: eventCode,
		BuyerCode: buyer,
		Qty:       totalQty,
		Gross:     gross,
		Discount:  discount,
		Fee:       fee,
		Net:       net,
		PromoCode: promoCode,
		ExpiresAt: expires.Format("2006-01-02 15:04:05"),
	}
	for _, l := range lines {
		itemCode, err := utils.GenerateCode(utils.OrderItemPrefix)
		if err != nil {
			return nil, err
		}
		resvCode, err := utils.GenerateCode(utils.ResvCodePrefix)
		if err != nil {
			return nil, err
		}
		if po.TypeCode == "" {
			po.TypeCode = l.info.Code // legacy single-type column keeps first line
		}
		po.Items = append(po.Items, PendingItem{
			Code: itemCode, ResvCode: resvCode, TypeCode: l.info.Code,
			TypeName: l.info.Name, Price: l.price, Qty: l.qty,
		})
	}
	if err := s.repo.CheckoutTx(ctx, po); err != nil {
		return nil, fmt.Errorf("create order: %w", err)
	}
	return &models.CheckoutResponse{
		OrderCode:        orderCode,
		Gross:            gross,
		Discount:         discount,
		Fee:              fee,
		Net:              net,
		ExpiresAt:        po.ExpiresAt,
		CountdownSeconds: s.holdMinutes * 60,
	}, nil
}

// PayOrder settles an order. FINANCE seam: real gateways call this with
// method=<rail> + ref=<gateway reference>; the SIMULATION rail below only
// runs outside production and issues tickets identically.
func (s *Service) PayOrder(ctx context.Context, buyer, orderCode string, req models.PayRequest, holderName, email, phone string) (*models.PayResponse, error) {
	_ = s.repo.ExpirePending(ctx, nowStr())
	orderCode = strings.TrimSpace(orderCode)
	o, err := s.repo.GetOrder(ctx, orderCode)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, services.ErrOrderNotFound
		}
		return nil, fmt.Errorf("get order: %w", err)
	}
	if o.BuyerCode != buyer {
		return nil, services.ErrOrderNotFound
	}
	if o.Status != models.OrderPending {
		return nil, fmt.Errorf("%w: order is %s", services.ErrInvalidCheckout, strings.ToLower(o.Status))
	}
	if o.ReservedUntil != "" && o.ReservedUntil < nowStr() {
		_ = s.repo.ExpirePending(ctx, nowStr())
		return nil, fmt.Errorf("%w: payment window expired", services.ErrInvalidCheckout)
	}
	method := strings.ToUpper(strings.TrimSpace(req.Method))
	if req.Simulate || method == "" || method == "SIMULATION" {
		if strings.ToLower(strings.TrimSpace(s.appEnv)) == "production" {
			return nil, fmt.Errorf("%w: simulation disabled in production", services.ErrInvalidCheckout)
		}
		method = "SIMULATION"
	}
	items, err := s.repo.OrderItems(ctx, orderCode)
	if err != nil {
		return nil, fmt.Errorf("order items: %w", err)
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("%w: order has no items", services.ErrInvalidCheckout)
	}
	// Revalidate quota under the hold (oversell guard at settle time).
	for _, it := range items {
		info, err := s.repo.TypeInfo(ctx, it.TypeCode)
		if err != nil {
			return nil, fmt.Errorf("type lookup: %w", err)
		}
		sold, err := s.repo.SoldQty(ctx, it.TypeCode)
		if err != nil {
			return nil, fmt.Errorf("sold lookup: %w", err)
		}
		held, err := s.repo.ReservedQty(ctx, it.TypeCode, nowStr())
		if err != nil {
			return nil, fmt.Errorf("hold lookup: %w", err)
		}
		// Own hold counts inside held; settling it frees exactly its qty.
		if sold+held > info.Quota {
			return nil, fmt.Errorf("%w: quota exhausted during payment", services.ErrInvalidCheckout)
		}
	}
	holderName = strings.TrimSpace(holderName)
	if holderName == "" {
		return nil, fmt.Errorf("%w: holder name is required", services.ErrInvalidCheckout)
	}
	tickets := make([]NewTicket, 0, o.Qty)
	codes := make([]string, 0, o.Qty)
	for _, it := range items {
		for i := 0; i < it.Qty; i++ {
			tcode, err := utils.GenerateCode(utils.TicketCodePrefix)
			if err != nil {
				return nil, err
			}
			qrcode, err := utils.GenerateCode(utils.TicketQRPrefix)
			if err != nil {
				return nil, err
			}
			token, err := utils.GenerateTicketToken(s.qrSecret, tcode, o.EventCode)
			if err != nil {
				return nil, err
			}
			tickets = append(tickets, NewTicket{
				Code: tcode, QRCode: qrcode, QRToken: token,
				Holder: holderName, Email: strings.TrimSpace(email), Phone: strings.TrimSpace(phone),
			})
			codes = append(codes, tcode)
		}
	}
	_ = method
	if err := s.repo.PayTx(ctx, orderCode, tickets); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: order no longer payable", services.ErrInvalidCheckout)
		}
		return nil, fmt.Errorf("settle order: %w", err)
	}
	return &models.PayResponse{OrderCode: orderCode, Status: models.OrderPaid, Tickets: codes}, nil
}

func (s *Service) CancelOrder(ctx context.Context, buyer, orderCode string) error {
	if err := s.repo.CancelPending(ctx, buyer, strings.TrimSpace(orderCode)); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return services.ErrOrderNotFound
		}
		return fmt.Errorf("cancel order: %w", err)
	}
	return nil
}

func (s *Service) MyTickets(ctx context.Context, buyer, display, q string, limit, offset int) ([]models.MyTicketItem, error) {
	_ = s.repo.ExpirePending(ctx, nowStr())
	display = strings.ToLower(strings.TrimSpace(display))
	if display != "" && display != "active" && display != "used" && display != "refunded" && display != "expired" {
		return nil, fmt.Errorf("%w: filter must be active, used, refunded or expired", services.ErrInvalidFilter)
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	items, err := s.repo.MyTickets(ctx, buyer, display, strings.TrimSpace(q), nowStr(), limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list tickets: %w", err)
	}
	return items, nil
}

func (s *Service) TicketDetail(ctx context.Context, buyer, ticketCode string) (*models.TicketDetail, error) {
	ticketCode = strings.TrimSpace(ticketCode)
	if ticketCode == "" {
		return nil, fmt.Errorf("%w: ticket code is required", services.ErrInvalidFilter)
	}
	d, err := s.repo.TicketDetail(ctx, buyer, ticketCode)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, services.ErrTicketNotFound
		}
		return nil, fmt.Errorf("ticket detail: %w", err)
	}
	// QR is live only for usable tickets (unused + event not over).
	d.Active = d.Status == "ACTIVE" && (d.EndAt == "" || d.EndAt >= nowStr())
	if !d.Active {
		d.QRToken = ""
	}
	return d, nil
}

func (s *Service) RequestRefund(ctx context.Context, buyer, orderCode, reason string) (string, error) {
	orderCode = strings.TrimSpace(orderCode)
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return "", fmt.Errorf("%w: reason is required", services.ErrInvalidFilter)
	}
	o, err := s.repo.GetOrder(ctx, orderCode)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", services.ErrOrderNotFound
		}
		return "", fmt.Errorf("get order: %w", err)
	}
	if o.BuyerCode != buyer {
		return "", services.ErrOrderNotFound
	}
	if o.Status != models.OrderPaid {
		return "", fmt.Errorf("%w: only paid orders can be refunded", services.ErrNotEligible)
	}
	tickets, err := s.repo.OrderTickets(ctx, orderCode)
	if err != nil {
		return "", fmt.Errorf("order tickets: %w", err)
	}
	if len(tickets) == 0 {
		return "", fmt.Errorf("%w: order has no tickets", services.ErrNotEligible)
	}
	for _, t := range tickets {
		if t.Status != "ACTIVE" {
			return "", fmt.Errorf("%w: only unused tickets can be refunded", services.ErrNotEligible)
		}
	}
	ev, err := s.repo.EventInfo(ctx, o.EventCode)
	if err != nil {
		return "", fmt.Errorf("event lookup: %w", err)
	}
	if ev.EndAt != "" && ev.EndAt <= nowStr() {
		return "", fmt.Errorf("%w: event has ended", services.ErrNotEligible)
	}
	if ev.Status != models.EventPublished {
		return "", fmt.Errorf("%w: event is not active", services.ErrNotEligible)
	}
	if s.cutoffHours > 0 {
		deadline := time.Now().Add(time.Duration(s.cutoffHours) * time.Hour).Format("2006-01-02 15:04:05")
		if ev.StartAt != "" && ev.StartAt <= deadline {
			return "", fmt.Errorf("%w: refund window closed", services.ErrNotEligible)
		}
	}
	refCode, err := utils.GenerateCode(utils.RefundCodePrefix)
	if err != nil {
		return "", err
	}
	if err := s.repo.CreateRefund(ctx, orderCode, refCode, o.Net, reason); err != nil {
		if isDuplicateErr(err) {
			return "", fmt.Errorf("%w: refund already requested", services.ErrNotEligible)
		}
		return "", fmt.Errorf("create refund: %w", err)
	}
	return refCode, nil
}

func (s *Service) BuyerRefunds(ctx context.Context, buyer string) ([]models.RefundItem, error) {
	items, err := s.repo.BuyerRefunds(ctx, buyer)
	if err != nil {
		return nil, fmt.Errorf("list refunds: %w", err)
	}
	return items, nil
}

// scanScope verifies the caller may scan an event: assigned officer or the
// event's own organizer (both read from data, never request claims).
func (s *Service) scanScope(ctx context.Context, officer, event string) error {
	if event == "" {
		return fmt.Errorf("%w: event is required", services.ErrInvalidFilter)
	}
	if ok, err := s.repo.OfficerAssigned(ctx, officer, event); err != nil {
		return fmt.Errorf("check assignment: %w", err)
	} else if ok {
		return nil
	}
	owner, err := s.repo.EventOwner(ctx, event)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("%w: unknown event", services.ErrInvalidFilter)
		}
		return fmt.Errorf("event owner: %w", err)
	}
	if owner == officer {
		return nil
	}
	return services.ErrScanDenied
}

// ScanTicket validates one QR token/code against event, payment, ticket and
// schedule, then checks in atomically. Every attempt is logged.
func (s *Service) ScanTicket(ctx context.Context, officer string, req models.ScanRequest) (*models.ScanResult, error) {
	event := strings.TrimSpace(req.Event)
	if err := s.scanScope(ctx, officer, event); err != nil {
		return nil, err
	}
	token := strings.TrimSpace(req.Token)
	code := strings.TrimSpace(req.Code)
	if token == "" && code == "" {
		return nil, fmt.Errorf("%w: token or code is required", services.ErrInvalidFilter)
	}
	ticketCode := code
	if token != "" {
		tcode, tevent, err := utils.ValidateTicketToken(token, s.qrSecret)
		if err != nil {
			_ = s.logScan(ctx, event, "", officer, "INVALID", "bad signature")
			return s.scanResult(ctx, event, "INVALID", "", "", "", "Kode tidak valid"), nil
		}
		ticketCode = tcode
		if tevent != event {
			_ = s.logScan(ctx, event, ticketCode, officer, "WRONG_EVENT", "tiket event lain")
			return s.scanResult(ctx, event, "WRONG_EVENT", ticketCode, "", "", "Tiket untuk event lain"), nil
		}
	}
	t, err := s.repo.TicketByCode(ctx, ticketCode)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			_ = s.logScan(ctx, event, ticketCode, officer, "INVALID", "tidak ditemukan")
			return s.scanResult(ctx, event, "INVALID", ticketCode, "", "", "Kode tidak valid"), nil
		}
		return nil, fmt.Errorf("ticket lookup: %w", err)
	}
	if t.EventCode != event {
		_ = s.logScan(ctx, event, t.TicketCode, officer, "WRONG_EVENT", "tiket event lain")
		return s.scanResult(ctx, event, "WRONG_EVENT", t.TicketCode, t.HolderName, t.TypeName, "Tiket untuk event lain"), nil
	}
	if t.OrderStatus != models.OrderPaid {
		_ = s.logScan(ctx, event, t.TicketCode, officer, "UNPAID", "order "+strings.ToLower(t.OrderStatus))
		return s.scanResult(ctx, event, "UNPAID", t.TicketCode, t.HolderName, t.TypeName, "Belum dibayar"), nil
	}
	switch t.Status {
	case "USED":
		_ = s.logScan(ctx, event, t.TicketCode, officer, "USED", "sudah masuk")
		return s.scanResult(ctx, event, "USED", t.TicketCode, t.HolderName, t.TypeName, "Sudah dipakai"), nil
	case "REFUNDED":
		_ = s.logScan(ctx, event, t.TicketCode, officer, "REFUNDED", "sudah refund")
		return s.scanResult(ctx, event, "REFUNDED", t.TicketCode, t.HolderName, t.TypeName, "Sudah refund"), nil
	case "ACTIVE":
		// lanjut ke jendela jadwal di bawah
	default:
		_ = s.logScan(ctx, event, t.TicketCode, officer, "INVALID", "status tiket")
		return s.scanResult(ctx, event, "INVALID", t.TicketCode, t.HolderName, t.TypeName, "Kode tidak valid"), nil
	}
	now := nowStr()
	if t.EndAt != "" && t.EndAt < now {
		_ = s.logScan(ctx, event, t.TicketCode, officer, "EXPIRED", "event berakhir")
		return s.scanResult(ctx, event, "EXPIRED", t.TicketCode, t.HolderName, t.TypeName, "Event berakhir"), nil
	}
	if t.StartAt != "" {
		// Scan opens 6 hours before start.
		if now < addHours(t.StartAt, -6) {
			_ = s.logScan(ctx, event, t.TicketCode, officer, "OUTSIDE_SCHEDULE", "di luar jadwal")
			return s.scanResult(ctx, event, "OUTSIDE_SCHEDULE", t.TicketCode, t.HolderName, t.TypeName, "Di luar jadwal"), nil
		}
	}
	won, err := s.repo.CheckinAtomic(ctx, t.TicketCode, officer, now)
	if err != nil {
		return nil, fmt.Errorf("check in: %w", err)
	}
	if !won {
		_ = s.logScan(ctx, event, t.TicketCode, officer, "USED", "rebutan kalah")
		return s.scanResult(ctx, event, "USED", t.TicketCode, t.HolderName, t.TypeName, "Sudah dipakai"), nil
	}
	_ = s.logScan(ctx, event, t.TicketCode, officer, "VALID", t.HolderName)
	return s.scanResult(ctx, event, "VALID", t.TicketCode, t.HolderName, t.TypeName, "Valid, selamat datang"), nil
}

// addHours shifts an ISO datetime string by h hours (schedule window math).
func addHours(iso string, h int) string {
	t, err := time.Parse("2006-01-02 15:04:05", iso)
	if err != nil {
		if t2, err2 := time.Parse("2006-01-02 15:04", iso); err2 == nil {
			t = t2
		} else {
			return iso
		}
	}
	return t.Add(time.Duration(h) * time.Hour).Format("2006-01-02 15:04:05")
}

func (s *Service) logScan(ctx context.Context, event, ticket, officer, result, detail string) error {
	code, err := utils.GenerateCode(utils.ScanCodePrefix)
	if err != nil {
		return err
	}
	return s.repo.InsertScanLog(ctx, code, event, ticket, officer, result, detail)
}

func (s *Service) scanResult(ctx context.Context, event, result, ticket, holder, typeName, detail string) *models.ScanResult {
	checkedIn, total := 0, 0
	if ci, tot, _, err := s.repo.ScanSummary(ctx, event); err == nil {
		checkedIn, total = ci, tot
	}
	return &models.ScanResult{
		Result: result, TicketCode: ticket, HolderName: holder,
		TypeName: typeName, Detail: detail, CheckedIn: checkedIn, Total: total,
	}
}

func (s *Service) ScanEvents(ctx context.Context, officer string) ([]AssignedEvent, error) {
	events, err := s.repo.AssignedEvents(ctx, officer)
	if err != nil {
		return nil, fmt.Errorf("assigned events: %w", err)
	}
	return events, nil
}

func (s *Service) ScanSummary(ctx context.Context, officer, event string) (*models.ScanSummary, error) {
	if err := s.scanScope(ctx, officer, event); err != nil {
		return nil, err
	}
	checkedIn, total, title, err := s.repo.ScanSummary(ctx, event)
	if err != nil {
		return nil, fmt.Errorf("scan summary: %w", err)
	}
	return &models.ScanSummary{EventCode: event, Title: title, CheckedIn: checkedIn, Total: total}, nil
}

func (s *Service) ScanHistory(ctx context.Context, officer, event, result, officerFilter, from, to string, limit, offset int) ([]models.ScanLogRow, models.ScanHistorySummary, error) {
	event = strings.TrimSpace(event)
	if event == "" {
		return nil, models.ScanHistorySummary{}, fmt.Errorf("%w: event is required", services.ErrInvalidFilter)
	}
	if err := s.scanScope(ctx, officer, event); err != nil {
		return nil, models.ScanHistorySummary{}, err
	}
	// Officers read their own logs; event owners read the whole event log.
	scopeOfficer := strings.TrimSpace(officerFilter)
	owner, err := s.repo.EventOwner(ctx, event)
	if err == nil && owner != officer && scopeOfficer == "" {
		scopeOfficer = officer
	}
	result = strings.TrimSpace(strings.ToUpper(result))
	if result != "" {
		switch result {
		case "VALID", "USED", "INVALID", "WRONG_EVENT", "UNPAID", "REFUNDED", "EXPIRED", "OUTSIDE_SCHEDULE":
		default:
			return nil, models.ScanHistorySummary{}, fmt.Errorf("%w: unknown result", services.ErrInvalidFilter)
		}
	}
	if from != "" && !utils.IsValidDate(from) {
		return nil, models.ScanHistorySummary{}, fmt.Errorf("%w: from must be YYYY-MM-DD", services.ErrInvalidFilter)
	}
	toPlusOne := ""
	if to != "" {
		if !utils.IsValidDate(to) {
			return nil, models.ScanHistorySummary{}, fmt.Errorf("%w: to must be YYYY-MM-DD", services.ErrInvalidFilter)
		}
		t, _ := time.Parse("2006-01-02", to)
		toPlusOne = t.AddDate(0, 0, 1).Format("2006-01-02")
	}
	if limit <= 0 || limit > 500 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	rows, err := s.repo.ScanHistory(ctx, event, result, scopeOfficer, from, toPlusOne, limit, offset)
	if err != nil {
		return nil, models.ScanHistorySummary{}, fmt.Errorf("scan history: %w", err)
	}
	sum, err := s.repo.ScanHistorySummary(ctx, event, result, scopeOfficer, from, toPlusOne)
	if err != nil {
		return nil, models.ScanHistorySummary{}, fmt.Errorf("scan summary: %w", err)
	}
	return rows, sum, nil
}

// ScanExport flattens filtered history for CSV.
func (s *Service) ScanExport(ctx context.Context, officer, event, result, officerFilter, from, to string) ([]export.Table, error) {
	rows, sum, err := s.ScanHistory(ctx, officer, event, result, officerFilter, from, to, 5000, 0)
	if err != nil {
		return nil, err
	}
	lines := make([][]string, 0, len(rows))
	for _, r := range rows {
		lines = append(lines, []string{r.CreatedAt, r.TicketCode, r.HolderName, r.TypeName, r.Result, r.Officer})
	}
	_ = sum
	return []export.Table{{
		Title:   fmt.Sprintf("Scan history %s", event),
		Headers: []string{"Time", "Ticket", "Holder", "Type", "Result", "Officer"},
		Rows:    lines,
	}}, nil
}
