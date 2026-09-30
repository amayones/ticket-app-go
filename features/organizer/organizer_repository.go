package organizer

import (
	"context"
	"database/sql"

	"golang-backend/models"
	"golang-backend/repositories"
)

// Dashboard reads: read-only aggregation over T_* tables (T_ORDER, T_TICKET,
// T_EVENT, T_PROMO, T_REFUND, T_AD_DAILY, T_PAYOUT, T_ORGANIZER_BALANCE).
//
// Konvensi query: SQL selalu di variabel `query`, placeholder "?" lalu
// di-rebind (Bind), timeout via repositories.WithTimeout. Rentang tanggal
// setengah-terbuka [from, toPlusOne) agar index tanggal tetap kepakai
// di semua engine.
type RepositoryInterface interface {
	// EventCodes returns codes of events owned by the organizer.
	EventCodes(ctx context.Context, organizer string) ([]string, error)
	// EventCities returns distinct cities of the organizer's events.
	EventCities(ctx context.Context, organizer string) ([]models.CityOption, error)
	// OrderSums aggregates PAID orders: tickets, gross, discount, fee, net,
	// promo orders, order count.
	OrderSums(ctx context.Context, organizer, event, city, from, toPlusOne string) (tickets int, gross, discount, fee, net float64, promoOrders, orders int, err error)
	// ActiveEvents counts PUBLISHED events starting in range.
	ActiveEvents(ctx context.Context, organizer, from, toPlusOne string) (int, error)
	// CheckinStats counts sold vs used tickets of PAID orders in range.
	CheckinStats(ctx context.Context, organizer, event, city, from, toPlusOne string) (sold, used int, err error)
	// RefundStats counts refunds + refunded amount in range.
	RefundStats(ctx context.Context, organizer, event, city, from, toPlusOne string) (count int, amount float64, err error)
	// AdSums aggregates ad impressions/clicks/cost in range.
	AdSums(ctx context.Context, organizer, event, city, from, toPlusOne string) (imp, clicks int, cost float64, err error)
	// SalesTrend returns per-day buckets (tickets + net revenue).
	SalesTrend(ctx context.Context, organizer, event, city, from, toPlusOne string) ([]models.TrendPoint, error)
	// SalesByTicketType groups net revenue by ticket type name.
	SalesByTicketType(ctx context.Context, organizer, event, city, from, toPlusOne string) ([]models.SlicePoint, error)
	// SalesByCity groups net revenue by city name.
	SalesByCity(ctx context.Context, organizer, event, city, from, toPlusOne string) ([]models.SlicePoint, error)
	// SalesByPromo groups orders + discount by promo code.
	SalesByPromo(ctx context.Context, organizer, event, city, from, toPlusOne string) ([]models.SlicePoint, error)
	// UpcomingEvents lists PUBLISHED future events with sold-vs-quota.
	UpcomingEvents(ctx context.Context, organizer string) ([]models.UpcomingEvent, error)
	// RecentOrders returns the latest PAID orders (limit).
	RecentOrders(ctx context.Context, organizer, event, city string, limit int) ([]models.OrganizerOrder, error)
	// PendingRefunds lists PENDING refunds with event titles.
	PendingRefunds(ctx context.Context, organizer string) ([]models.OrganizerRefund, error)
	// RefundByCode loads one refund with its order + event owner.
	RefundByCode(ctx context.Context, code string) (orderCode, eventCode, owner, status string, amount float64, err error)
	// DecideRefund marks the refund APPROVED/REJECTED (decider + timestamp).
	DecideRefund(ctx context.Context, code, status, decider string) error
	// MarkOrderRefunded flips an approved order to REFUNDED (+ its tickets).
	MarkOrderRefunded(ctx context.Context, orderCode string) error
	// CheckinLive counts USED vs total ACTIVE+USED tickets of one event.
	CheckinLive(ctx context.Context, organizer, eventCode string) (title string, checkedIn, total int, err error)
	// QuotaUsage returns per-event sold + quota for the low-quota alert.
	QuotaUsage(ctx context.Context, organizer string) ([]models.UpcomingEvent, error)
	// DraftEvents counts unpublished (DRAFT) events.
	DraftEvents(ctx context.Context, organizer string) (int, error)
	// ExpiringPromos counts ACTIVE promos ending within 3 days.
	ExpiringPromos(ctx context.Context, organizer, now, in3 string) (int, error)
	// EventOwned reports whether the event belongs to the organizer.
	EventOwned(ctx context.Context, organizer, event string) (bool, error)
	// CityExists reports whether the city code exists.
	CityExists(ctx context.Context, city string) (bool, error)
	// Balance reads the organizer escrow/available row (zero when absent).
	Balance(ctx context.Context, organizer string) (models.Balance, error)
	// PayoutReady sums READY payouts of the organizer.
	PayoutReady(ctx context.Context, organizer string) (float64, error)
}

type Repository struct {
	db      *sql.DB
	dialect repositories.Dialect
}

func NewRepository(db *sql.DB, dialect repositories.Dialect) RepositoryInterface {
	return &Repository{db: db, dialect: dialect}
}

func (r *Repository) orderTable() string  { return r.dialect.Table("T_ORDER") }
func (r *Repository) ticketTable() string { return r.dialect.Table("T_TICKET") }
func (r *Repository) eventTable() string  { return r.dialect.Table("T_EVENT") }
func (r *Repository) typeTable() string   { return r.dialect.Table("T_TICKET_TYPE") }
func (r *Repository) promoTable() string  { return r.dialect.Table("T_PROMO") }
func (r *Repository) refundTable() string { return r.dialect.Table("T_REFUND") }
func (r *Repository) adTable() string     { return r.dialect.Table("T_AD_DAILY") }
func (r *Repository) payoutTable() string { return r.dialect.Table("T_PAYOUT") }
func (r *Repository) balTable() string    { return r.dialect.Table("T_ORGANIZER_BALANCE") }
func (r *Repository) cityTable() string   { return r.dialect.Table("T_CITY") }

// citySQL optionally restricts rows to one city (empty = all).
const citySQL = " AND (? = '' OR e.CITY_CODE = ?)"

// dayExpr renders an ISO date (YYYY-MM-DD string) per engine for trend buckets.
func (r *Repository) dayExpr(col string) string {
	switch r.dialect {
	case repositories.DialectPostgres:
		return "TO_CHAR(" + col + ", 'YYYY-MM-DD')"
	case repositories.DialectSQLite:
		return "date(" + col + ")"
	default:
		return "CONVERT(varchar(10), " + col + ", 120)"
	}
}

// nowExpr renders the current-timestamp expression per engine.
func (r *Repository) nowExpr() string { return r.dialect.Now() }

func (r *Repository) EventCodes(ctx context.Context, organizer string) ([]string, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	query := `SELECT CODE FROM ` + r.eventTable() + ` WHERE ORGANIZER_CODE = ? ORDER BY CODE ASC`
	rows, err := r.db.QueryContext(ctx, r.dialect.Bind(query), organizer)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	codes := make([]string, 0)
	for rows.Next() {
		var c string
		if err := rows.Scan(&c); err != nil {
			return nil, err
		}
		codes = append(codes, c)
	}
	return codes, rows.Err()
}

func (r *Repository) OrderSums(ctx context.Context, organizer, event, city, from, toPlusOne string) (int, float64, float64, float64, float64, int, int, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	var tickets, promoOrders, orders int
	var gross, discount, fee, net float64
	args := []any{organizer, from, toPlusOne}
	ef := ""
	if event != "" {
		ef = " AND o.EVENT_CODE = ?"
		args = append(args, event)
	}
	args = append(args, city, city)
	query := `SELECT COUNT(*), COALESCE(SUM(o.QTY),0), COALESCE(SUM(o.GROSS_AMOUNT),0),
	  COALESCE(SUM(o.DISCOUNT_AMOUNT),0), COALESCE(SUM(o.PLATFORM_FEE),0), COALESCE(SUM(o.NET_AMOUNT),0),
	  COALESCE(SUM(CASE WHEN o.PROMO_CODE IS NOT NULL THEN 1 ELSE 0 END),0)
	  FROM ` + r.orderTable() + ` o JOIN ` + r.eventTable() + ` e ON e.CODE = o.EVENT_CODE
	  WHERE e.ORGANIZER_CODE = ? AND o.STATUS = 'PAID'
	  AND o.ORDER_DATE >= ? AND o.ORDER_DATE < ?` + ef + citySQL
	err := r.db.QueryRowContext(ctx, r.dialect.Bind(query), args...).
		Scan(&orders, &tickets, &gross, &discount, &fee, &net, &promoOrders)
	return tickets, gross, discount, fee, net, promoOrders, orders, err
}

func (r *Repository) ActiveEvents(ctx context.Context, organizer, from, toPlusOne string) (int, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	var n int
	query := `SELECT COUNT(*) FROM ` + r.eventTable() + `
	  WHERE ORGANIZER_CODE = ? AND STATUS = 'PUBLISHED'
	  AND START_AT >= ? AND START_AT < ?`
	err := r.db.QueryRowContext(ctx, r.dialect.Bind(query), organizer, from, toPlusOne).Scan(&n)
	return n, err
}

func (r *Repository) CheckinStats(ctx context.Context, organizer, event, city, from, toPlusOne string) (int, int, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	var sold, used int
	args := []any{organizer, from, toPlusOne}
	ef := ""
	if event != "" {
		ef = " AND o.EVENT_CODE = ?"
		args = append(args, event)
	}
	args = append(args, city, city)
	query := `SELECT COUNT(*), COALESCE(SUM(CASE WHEN t.STATUS = 'USED' THEN 1 ELSE 0 END),0)
	  FROM ` + r.ticketTable() + ` t JOIN ` + r.orderTable() + ` o ON o.CODE = t.ORDER_CODE
	  JOIN ` + r.eventTable() + ` e ON e.CODE = o.EVENT_CODE
	  WHERE e.ORGANIZER_CODE = ? AND o.STATUS = 'PAID' AND t.STATUS IN ('ACTIVE','USED')
	  AND o.ORDER_DATE >= ? AND o.ORDER_DATE < ?` + ef + citySQL
	err := r.db.QueryRowContext(ctx, r.dialect.Bind(query), args...).Scan(&sold, &used)
	return sold, used, err
}

func (r *Repository) RefundStats(ctx context.Context, organizer, event, city, from, toPlusOne string) (int, float64, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	var n int
	var amount float64
	args := []any{organizer, from, toPlusOne}
	ef := ""
	if event != "" {
		ef = " AND o.EVENT_CODE = ?"
		args = append(args, event)
	}
	args = append(args, city, city)
	query := `SELECT COUNT(*), COALESCE(SUM(f.AMOUNT),0)
	  FROM ` + r.refundTable() + ` f JOIN ` + r.orderTable() + ` o ON o.CODE = f.ORDER_CODE
	  JOIN ` + r.eventTable() + ` e ON e.CODE = o.EVENT_CODE
	  WHERE e.ORGANIZER_CODE = ? AND f.CREATED_AT >= ? AND f.CREATED_AT < ?` + ef + citySQL
	err := r.db.QueryRowContext(ctx, r.dialect.Bind(query), args...).Scan(&n, &amount)
	return n, amount, err
}

func (r *Repository) AdSums(ctx context.Context, organizer, event, city, from, toPlusOne string) (int, int, float64, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	var imp, clicks int
	var cost float64
	args := []any{organizer, from, toPlusOne}
	ef := ""
	if event != "" {
		ef = " AND a.EVENT_CODE = ?"
		args = append(args, event)
	}
	args = append(args, city, city)
	query := `SELECT COALESCE(SUM(a.IMPRESSIONS),0), COALESCE(SUM(a.CLICKS),0), COALESCE(SUM(a.COST),0)
	  FROM ` + r.adTable() + ` a JOIN ` + r.eventTable() + ` e ON e.CODE = a.EVENT_CODE
	  WHERE e.ORGANIZER_CODE = ? AND a.STAT_DATE >= ? AND a.STAT_DATE < ?` + ef + citySQL
	err := r.db.QueryRowContext(ctx, r.dialect.Bind(query), args...).Scan(&imp, &clicks, &cost)
	return imp, clicks, cost, err
}

func (r *Repository) SalesTrend(ctx context.Context, organizer, event, city, from, toPlusOne string) ([]models.TrendPoint, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	args := []any{organizer, from, toPlusOne}
	ef := ""
	if event != "" {
		ef = " AND o.EVENT_CODE = ?"
		args = append(args, event)
	}
	args = append(args, city, city)
	query := `SELECT ` + r.dayExpr("o.ORDER_DATE") + `, COALESCE(SUM(o.QTY),0), COALESCE(SUM(o.NET_AMOUNT),0)
	  FROM ` + r.orderTable() + ` o JOIN ` + r.eventTable() + ` e ON e.CODE = o.EVENT_CODE
	  WHERE e.ORGANIZER_CODE = ? AND o.STATUS = 'PAID'
	  AND o.ORDER_DATE >= ? AND o.ORDER_DATE < ?` + ef + citySQL + `
	  GROUP BY ` + r.dayExpr("o.ORDER_DATE") + ` ORDER BY 1 ASC`
	rows, err := r.db.QueryContext(ctx, r.dialect.Bind(query), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.TrendPoint, 0)
	for rows.Next() {
		var p models.TrendPoint
		if err := rows.Scan(&p.Date, &p.Tickets, &p.Revenue); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *Repository) SalesByTicketType(ctx context.Context, organizer, event, city, from, toPlusOne string) ([]models.SlicePoint, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	args := []any{organizer, from, toPlusOne}
	ef := ""
	if event != "" {
		ef = " AND o.EVENT_CODE = ?"
		args = append(args, event)
	}
	args = append(args, city, city)
	query := `SELECT tt.NAME, COALESCE(SUM(o.QTY),0), COALESCE(SUM(o.NET_AMOUNT),0), COUNT(*)
	  FROM ` + r.orderTable() + ` o JOIN ` + r.eventTable() + ` e ON e.CODE = o.EVENT_CODE
	  JOIN ` + r.typeTable() + ` tt ON tt.CODE = o.TICKET_TYPE_CODE
	  WHERE e.ORGANIZER_CODE = ? AND o.STATUS = 'PAID'
	  AND o.ORDER_DATE >= ? AND o.ORDER_DATE < ?` + ef + citySQL + `
	  GROUP BY tt.NAME ORDER BY 3 DESC`
	return r.scanSlices(ctx, query, args)
}

func (r *Repository) SalesByCity(ctx context.Context, organizer, event, city, from, toPlusOne string) ([]models.SlicePoint, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	args := []any{organizer, from, toPlusOne}
	ef := ""
	if event != "" {
		ef = " AND o.EVENT_CODE = ?"
		args = append(args, event)
	}
	args = append(args, city, city)
	query := `SELECT COALESCE(c.NAME, '-'), COALESCE(SUM(o.QTY),0), COALESCE(SUM(o.NET_AMOUNT),0), COUNT(*)
	  FROM ` + r.orderTable() + ` o JOIN ` + r.eventTable() + ` e ON e.CODE = o.EVENT_CODE
	  LEFT JOIN ` + r.cityTable() + ` c ON c.CODE = e.CITY_CODE
	  WHERE e.ORGANIZER_CODE = ? AND o.STATUS = 'PAID'
	  AND o.ORDER_DATE >= ? AND o.ORDER_DATE < ?` + ef + citySQL + `
	  GROUP BY COALESCE(c.NAME, '-') ORDER BY 3 DESC`
	return r.scanSlices(ctx, query, args)
}

func (r *Repository) SalesByPromo(ctx context.Context, organizer, event, city, from, toPlusOne string) ([]models.SlicePoint, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	args := []any{organizer, from, toPlusOne}
	ef := ""
	if event != "" {
		ef = " AND o.EVENT_CODE = ?"
		args = append(args, event)
	}
	args = append(args, city, city)
	query := `SELECT COALESCE(o.PROMO_CODE, '-'), 0, COALESCE(SUM(o.DISCOUNT_AMOUNT),0), COUNT(*)
	  FROM ` + r.orderTable() + ` o JOIN ` + r.eventTable() + ` e ON e.CODE = o.EVENT_CODE
	  WHERE e.ORGANIZER_CODE = ? AND o.STATUS = 'PAID' AND o.PROMO_CODE IS NOT NULL
	  AND o.ORDER_DATE >= ? AND o.ORDER_DATE < ?` + ef + citySQL + `
	  GROUP BY COALESCE(o.PROMO_CODE, '-') ORDER BY 4 DESC`
	return r.scanSlices(ctx, query, args)
}

func (r *Repository) scanSlices(ctx context.Context, query string, args []any) ([]models.SlicePoint, error) {
	rows, err := r.db.QueryContext(ctx, r.dialect.Bind(query), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.SlicePoint, 0)
	for rows.Next() {
		var p models.SlicePoint
		if err := rows.Scan(&p.Name, &p.Tickets, &p.Revenue, &p.Orders); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *Repository) scanProgress(ctx context.Context, organizer string, futureOnly bool) ([]models.UpcomingEvent, error) {
	future := ""
	if futureOnly {
		future = " AND e.START_AT >= " + r.nowExpr()
	}
	var dateSel string
	switch r.dialect {
	case repositories.DialectPostgres:
		dateSel = "TO_CHAR(e.START_AT, 'YYYY-MM-DD HH24:MI:SS')"
	case repositories.DialectSQLite:
		dateSel = "strftime('%Y-%m-%d %H:%M:%S', e.START_AT)"
	default:
		dateSel = "CONVERT(varchar(19), e.START_AT, 120)"
	}
	query := `SELECT e.CODE, e.TITLE, ` + dateSel + `, COALESCE(e.VENUE, ''),
	  COALESCE((SELECT SUM(o.QTY) FROM ` + r.orderTable() + ` o
	    WHERE o.EVENT_CODE = e.CODE AND o.STATUS = 'PAID'), 0),
	  COALESCE((SELECT SUM(tt.QUOTA) FROM ` + r.typeTable() + ` tt
	    WHERE tt.EVENT_CODE = e.CODE), 0)
	  FROM ` + r.eventTable() + ` e
	  WHERE e.ORGANIZER_CODE = ? AND e.STATUS = 'PUBLISHED'` + future + `
	  ORDER BY e.START_AT ASC`
	rows, err := r.db.QueryContext(ctx, r.dialect.Bind(query), organizer)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.UpcomingEvent, 0)
	for rows.Next() {
		var ev models.UpcomingEvent
		if err := rows.Scan(&ev.Code, &ev.Title, &ev.Start, &ev.Venue, &ev.Sold, &ev.Quota); err != nil {
			return nil, err
		}
		out = append(out, ev)
	}
	return out, rows.Err()
}

func (r *Repository) UpcomingEvents(ctx context.Context, organizer string) ([]models.UpcomingEvent, error) {
	return r.scanProgress(ctx, organizer, true)
}

func (r *Repository) QuotaUsage(ctx context.Context, organizer string) ([]models.UpcomingEvent, error) {
	return r.scanProgress(ctx, organizer, false)
}

func (r *Repository) RecentOrders(ctx context.Context, organizer, event, city string, limit int) ([]models.OrganizerOrder, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	if limit <= 0 || limit > 100 {
		limit = 10
	}
	args := []any{organizer}
	ef := ""
	if event != "" {
		ef = " AND o.EVENT_CODE = ?"
		args = append(args, event)
	}
	args = append(args, city, city)
	if r.dialect != repositories.DialectMSSQL {
		args = append(args, limit)
	}
	var page string
	switch r.dialect {
	case repositories.DialectPostgres, repositories.DialectSQLite:
		page = " LIMIT ?"
	default:
		page = ""
	}
	query := `SELECT o.CODE, e.TITLE, o.BUYER_CODE, tt.NAME, o.QTY, o.NET_AMOUNT,
	  o.STATUS, ` + r.startStr("o.ORDER_DATE") + `
	  FROM ` + r.orderTable() + ` o JOIN ` + r.eventTable() + ` e ON e.CODE = o.EVENT_CODE
	  JOIN ` + r.typeTable() + ` tt ON tt.CODE = o.TICKET_TYPE_CODE
	  WHERE e.ORGANIZER_CODE = ? AND o.STATUS = 'PAID'` + ef + citySQL + `
	  ORDER BY o.ORDER_DATE DESC` + page
	if r.dialect == repositories.DialectMSSQL {
		query = `SELECT TOP (?) o.CODE, e.TITLE, o.BUYER_CODE, tt.NAME, o.QTY, o.NET_AMOUNT,
	  o.STATUS, ` + r.startStr("o.ORDER_DATE") + `
	  FROM ` + r.orderTable() + ` o JOIN ` + r.eventTable() + ` e ON e.CODE = o.EVENT_CODE
	  JOIN ` + r.typeTable() + ` tt ON tt.CODE = o.TICKET_TYPE_CODE
	  WHERE e.ORGANIZER_CODE = ? AND o.STATUS = 'PAID'` + ef + citySQL + `
	  ORDER BY o.ORDER_DATE DESC`
		args = append([]any{limit}, args...)
	}
	rows, err := r.db.QueryContext(ctx, r.dialect.Bind(query), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.OrganizerOrder, 0)
	for rows.Next() {
		var o models.OrganizerOrder
		if err := rows.Scan(&o.Code, &o.Event, &o.Buyer, &o.TicketType, &o.Qty, &o.Net, &o.Status, &o.OrderDate); err != nil {
			return nil, err
		}
		out = append(out, o)
	}
	return out, rows.Err()
}

func (r *Repository) EventCities(ctx context.Context, organizer string) ([]models.CityOption, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	query := `SELECT DISTINCT c.CODE, c.NAME FROM ` + r.cityTable() + ` c
	  JOIN ` + r.eventTable() + ` e ON e.CITY_CODE = c.CODE
	  WHERE e.ORGANIZER_CODE = ? ORDER BY c.NAME ASC`
	rows, err := r.db.QueryContext(ctx, r.dialect.Bind(query), organizer)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.CityOption, 0)
	for rows.Next() {
		var c models.CityOption
		if err := rows.Scan(&c.Code, &c.Name); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *Repository) EventOwned(ctx context.Context, organizer, event string) (bool, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	var n int
	query := `SELECT COUNT(*) FROM ` + r.eventTable() + ` WHERE CODE = ? AND ORGANIZER_CODE = ?`
	err := r.db.QueryRowContext(ctx, r.dialect.Bind(query), event, organizer).Scan(&n)
	return n > 0, err
}

func (r *Repository) CityExists(ctx context.Context, city string) (bool, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	var n int
	query := `SELECT COUNT(*) FROM ` + r.cityTable() + ` WHERE CODE = ?`
	err := r.db.QueryRowContext(ctx, r.dialect.Bind(query), city).Scan(&n)
	return n > 0, err
}

// startStr formats a datetime column as ISO string per engine.
func (r *Repository) startStr(col string) string {
	switch r.dialect {
	case repositories.DialectPostgres:
		return "TO_CHAR(" + col + ", 'YYYY-MM-DD HH24:MI:SS')"
	case repositories.DialectSQLite:
		return "strftime('%Y-%m-%d %H:%M:%S', " + col + ")"
	default:
		return "CONVERT(varchar(19), " + col + ", 120)"
	}
}

func (r *Repository) PendingRefunds(ctx context.Context, organizer string) ([]models.OrganizerRefund, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	query := `SELECT f.CODE, f.ORDER_CODE, e.TITLE, f.AMOUNT, COALESCE(f.REASON, ''),
	  f.STATUS, ` + r.startStr("f.CREATED_AT") + `
	  FROM ` + r.refundTable() + ` f JOIN ` + r.orderTable() + ` o ON o.CODE = f.ORDER_CODE
	  JOIN ` + r.eventTable() + ` e ON e.CODE = o.EVENT_CODE
	  WHERE e.ORGANIZER_CODE = ? AND f.STATUS = 'PENDING'
	  ORDER BY f.CREATED_AT ASC`
	rows, err := r.db.QueryContext(ctx, r.dialect.Bind(query), organizer)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.OrganizerRefund, 0)
	for rows.Next() {
		var f models.OrganizerRefund
		if err := rows.Scan(&f.Code, &f.OrderCode, &f.Event, &f.Amount, &f.Reason, &f.Status, &f.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, f)
	}
	return out, rows.Err()
}

func (r *Repository) RefundByCode(ctx context.Context, code string) (string, string, string, string, float64, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	var orderCode, eventCode, owner, status string
	var amount float64
	query := `SELECT f.ORDER_CODE, o.EVENT_CODE, e.ORGANIZER_CODE, f.STATUS, f.AMOUNT
	  FROM ` + r.refundTable() + ` f JOIN ` + r.orderTable() + ` o ON o.CODE = f.ORDER_CODE
	  JOIN ` + r.eventTable() + ` e ON e.CODE = o.EVENT_CODE
	  WHERE f.CODE = ?`
	err := r.db.QueryRowContext(ctx, r.dialect.Bind(query), code).
		Scan(&orderCode, &eventCode, &owner, &status, &amount)
	return orderCode, eventCode, owner, status, amount, err
}

func (r *Repository) DecideRefund(ctx context.Context, code, status, decider string) error {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	query := `UPDATE ` + r.refundTable() + ` SET STATUS = ?, DECIDED_BY = ?,
	  DECIDED_AT = ` + r.dialect.Now() + `, UPDATED_AT = ` + r.dialect.Now() + ` WHERE CODE = ?`
	_, err := r.db.ExecContext(ctx, r.dialect.Bind(query), status, decider, code)
	return err
}

func (r *Repository) MarkOrderRefunded(ctx context.Context, orderCode string) error {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	query := `UPDATE ` + r.orderTable() + ` SET STATUS = 'REFUNDED',
	  UPDATED_AT = ` + r.dialect.Now() + ` WHERE CODE = ?`
	if _, err := r.db.ExecContext(ctx, r.dialect.Bind(query), orderCode); err != nil {
		return err
	}
	query = `UPDATE ` + r.ticketTable() + ` SET STATUS = 'REFUNDED',
	  UPDATED_AT = ` + r.dialect.Now() + ` WHERE ORDER_CODE = ? AND STATUS = 'ACTIVE'`
	_, err := r.db.ExecContext(ctx, r.dialect.Bind(query), orderCode)
	return err
}

// CheckinLive returns title + USED vs total (ACTIVE+USED) of PAID orders.
func (r *Repository) CheckinLive(ctx context.Context, organizer, eventCode string) (string, int, int, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	var title string
	var checkedIn, total int
	query := `SELECT e.TITLE,
	  COALESCE(SUM(CASE WHEN t.STATUS = 'USED' THEN 1 ELSE 0 END),0),
	  COUNT(t.CODE)
	  FROM ` + r.eventTable() + ` e
	  LEFT JOIN ` + r.orderTable() + ` o ON o.EVENT_CODE = e.CODE AND o.STATUS = 'PAID'
	  LEFT JOIN ` + r.ticketTable() + ` t ON t.ORDER_CODE = o.CODE AND t.STATUS IN ('ACTIVE','USED')
	  WHERE e.CODE = ? AND e.ORGANIZER_CODE = ?
	  GROUP BY e.TITLE`
	err := r.db.QueryRowContext(ctx, r.dialect.Bind(query), eventCode, organizer).
		Scan(&title, &checkedIn, &total)
	return title, checkedIn, total, err
}

func (r *Repository) DraftEvents(ctx context.Context, organizer string) (int, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	var n int
	query := `SELECT COUNT(*) FROM ` + r.eventTable() + ` WHERE ORGANIZER_CODE = ? AND STATUS = 'DRAFT'`
	err := r.db.QueryRowContext(ctx, r.dialect.Bind(query), organizer).Scan(&n)
	return n, err
}

func (r *Repository) ExpiringPromos(ctx context.Context, organizer, now, in3 string) (int, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	var n int
	query := `SELECT COUNT(*) FROM ` + r.promoTable() + ` p
	  LEFT JOIN ` + r.eventTable() + ` e ON e.CODE = p.EVENT_CODE
	  WHERE p.STATUS = 'ACTIVE' AND p.VALID_TO >= ? AND p.VALID_TO < ?
	  AND (p.EVENT_CODE IS NULL OR e.ORGANIZER_CODE = ?)`
	err := r.db.QueryRowContext(ctx, r.dialect.Bind(query), now, in3, organizer).Scan(&n)
	return n, err
}

func (r *Repository) Balance(ctx context.Context, organizer string) (models.Balance, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	var b models.Balance
	query := `SELECT COALESCE(ESCROW_AMOUNT,0), COALESCE(AVAILABLE_AMOUNT,0)
	  FROM ` + r.balTable() + ` WHERE ORGANIZER_CODE = ?`
	err := r.db.QueryRowContext(ctx, r.dialect.Bind(query), organizer).Scan(&b.Escrow, &b.Available)
	if err == sql.ErrNoRows {
		return models.Balance{}, nil
	}
	return b, err
}

func (r *Repository) PayoutReady(ctx context.Context, organizer string) (float64, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	var amount float64
	query := `SELECT COALESCE(SUM(AMOUNT),0) FROM ` + r.payoutTable() + `
	  WHERE OWNER_CODE = ? AND STATUS = 'READY'`
	err := r.db.QueryRowContext(ctx, r.dialect.Bind(query), organizer).Scan(&amount)
	return amount, err
}
