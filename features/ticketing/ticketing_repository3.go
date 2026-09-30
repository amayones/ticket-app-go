package ticketing

// Third part: refund filing, officer scope and atomic scan queries.
import (
	"context"
	"database/sql"

	"golang-backend/models"
	"golang-backend/repositories"
)

// EventInfo is the checkout-validation view of an event.
type EventInfo struct {
	Code       string
	Status     string
	StartAt    string
	EndAt      string
	FeePercent float64
	Category   string
}

// TypeInfo is the checkout-validation view of a ticket type.
type TypeInfo struct {
	Code         string
	EventCode    string
	Name         string
	Price        float64
	Quota        int
	IsActive     bool
	MaxPerPerson int
	HasMax       bool
	SaleFrom     string
	SaleTo       string
}

// PromoInfo is the checkout-validation view of a promo.
type PromoInfo struct {
	Code            string
	EventCode       string
	PromoCode       string
	DiscountPercent float64
	DiscountAmount  float64
	MaxDiscount     float64
	MinPurchase     float64
	MaxUses         int
	HasMax          bool
	UsedCount       int
	ValidFrom       string
	ValidTo         string
	Status          string
	CoverageCat     string
}

func (r *Repository) EventInfo(ctx context.Context, event string) (*EventInfo, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	var e EventInfo
	var end, cat sql.NullString
	query := `SELECT CODE, STATUS, ` + r.startStr("START_AT") + `,
	  COALESCE(` + r.startStr("END_AT") + `, ''),
	  PLATFORM_FEE_PERCENT, COALESCE(EVENT_CATEGORY_CODE, '')
	  FROM ` + r.eventTable() + ` WHERE CODE = ?`
	err := r.db.QueryRowContext(ctx, r.dialect.Bind(query), event).Scan(
		&e.Code, &e.Status, &e.StartAt, &end, &e.FeePercent, &cat)
	if err != nil {
		return nil, err
	}
	e.EndAt, e.Category = end.String, cat.String
	return &e, nil
}

func (r *Repository) TypeInfo(ctx context.Context, typeCode string) (*TypeInfo, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	var t TypeInfo
	var maxp sql.NullInt64
	var from, to sql.NullString
	query := `SELECT CODE, EVENT_CODE, NAME, PRICE, QUOTA, IS_ACTIVE,
	  MAX_PER_PERSON, COALESCE(` + r.startStr("SALE_START_AT") + `, ''),
	  COALESCE(` + r.startStr("SALE_END_AT") + `, '')
	  FROM ` + r.typeTable() + ` WHERE CODE = ?`
	err := r.db.QueryRowContext(ctx, r.dialect.Bind(query), typeCode).Scan(
		&t.Code, &t.EventCode, &t.Name, &t.Price, &t.Quota, &t.IsActive,
		&maxp, &from, &to)
	if err != nil {
		return nil, err
	}
	if maxp.Valid {
		t.MaxPerPerson, t.HasMax = int(maxp.Int64), true
	}
	t.SaleFrom, t.SaleTo = from.String, to.String
	return &t, nil
}

func (r *Repository) PromoInfo(ctx context.Context, promoCode string) (*PromoInfo, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	var p PromoInfo
	var maxu sql.NullInt64
	var ev, cov sql.NullString
	query := `SELECT CODE, COALESCE(EVENT_CODE, ''), PROMO_CODE, DISCOUNT_PERCENT,
	  DISCOUNT_AMOUNT, MAX_DISCOUNT_AMOUNT, MIN_PURCHASE, MAX_USES, USED_COUNT,
	  ` + r.startStr("VALID_FROM") + `, ` + r.startStr("VALID_TO") + `,
	  STATUS, COALESCE(COVERAGE_CATEGORY_CODE, '')
	  FROM ` + r.promoTable() + ` WHERE PROMO_CODE = ?`
	err := r.db.QueryRowContext(ctx, r.dialect.Bind(query), promoCode).Scan(
		&p.Code, &ev, &p.PromoCode, &p.DiscountPercent, &p.DiscountAmount,
		&p.MaxDiscount, &p.MinPurchase, &maxu, &p.UsedCount,
		&p.ValidFrom, &p.ValidTo, &p.Status, &cov)
	if err != nil {
		return nil, err
	}
	p.EventCode, p.CoverageCat = ev.String, cov.String
	if maxu.Valid {
		p.MaxUses, p.HasMax = int(maxu.Int64), true
	}
	return &p, nil
}

func (r *Repository) CreateRefund(ctx context.Context, orderCode, refundCode string, amount float64, reason string) error {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	query := `INSERT INTO ` + r.refundTable() + ` (CODE, ORDER_CODE, AMOUNT, REASON,
	  STATUS, REQUESTED_BY, PAYOUT_STATUS)
	  SELECT ?, ?, ?, ?, 'PENDING', BUYER_CODE, 'PENDING_FINANCE'
	  FROM ` + r.orderTable() + ` WHERE CODE = ?`
	_, err := r.db.ExecContext(ctx, r.dialect.Bind(query),
		refundCode, orderCode, amount, reason, orderCode)
	return err
}

func (r *Repository) BuyerRefunds(ctx context.Context, buyer string) ([]models.RefundItem, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	query := `SELECT f.CODE, f.ORDER_CODE, COALESCE(e.TITLE, ''), f.AMOUNT,
	  COALESCE(f.REASON, ''), COALESCE(f.REJECT_NOTE, ''), f.STATUS,
	  ` + r.startStr("f.CREATED_AT") + `, COALESCE(` + r.startStr("f.DECIDED_AT") + `, ''),
	  COALESCE(f.PAYOUT_STATUS, 'NONE')
	  FROM ` + r.refundTable() + ` f
	  JOIN ` + r.orderTable() + ` o ON o.CODE = f.ORDER_CODE
	  LEFT JOIN ` + r.eventTable() + ` e ON e.CODE = o.EVENT_CODE
	  WHERE o.BUYER_CODE = ?
	  ORDER BY f.CREATED_AT DESC`
	rows, err := r.db.QueryContext(ctx, r.dialect.Bind(query), buyer)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.RefundItem, 0)
	for rows.Next() {
		var m models.RefundItem
		if err := rows.Scan(&m.Code, &m.OrderCode, &m.EventTitle, &m.Amount,
			&m.Reason, &m.RejectNote, &m.Status, &m.CreatedAt, &m.DecidedAt,
			&m.Payout); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (r *Repository) OrderTickets(ctx context.Context, orderCode string) ([]TicketStatus, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	query := `SELECT CODE, STATUS FROM ` + r.ticketTable() + ` WHERE ORDER_CODE = ?`
	rows, err := r.db.QueryContext(ctx, r.dialect.Bind(query), orderCode)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]TicketStatus, 0)
	for rows.Next() {
		var t TicketStatus
		if err := rows.Scan(&t.Code, &t.Status); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (r *Repository) OfficerAssigned(ctx context.Context, officer, event string) (bool, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	var n int
	query := `SELECT COUNT(*) FROM ` + r.offTable() + ` WHERE OFFICER_CODE = ? AND EVENT_CODE = ?`
	err := r.db.QueryRowContext(ctx, r.dialect.Bind(query), officer, event).Scan(&n)
	return n > 0, err
}

// AssignedEvent is one scannable event option.
type AssignedEvent struct {
	Code  string `json:"code"`
	Title string `json:"title"`
}

func (r *Repository) AssignedEvents(ctx context.Context, officer string) ([]AssignedEvent, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	query := `SELECT e.CODE, e.TITLE FROM ` + r.eventTable() + ` e
	  WHERE e.IS_DELETED = 0 AND (e.ORGANIZER_CODE = ?
	  OR EXISTS (SELECT 1 FROM ` + r.offTable() + ` a
	    WHERE a.EVENT_CODE = e.CODE AND a.OFFICER_CODE = ?))
	  ORDER BY e.START_AT ASC`
	rows, err := r.db.QueryContext(ctx, r.dialect.Bind(query), officer, officer)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]AssignedEvent, 0)
	for rows.Next() {
		var a AssignedEvent
		if err := rows.Scan(&a.Code, &a.Title); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (r *Repository) EventOwner(ctx context.Context, event string) (string, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	var owner string
	query := `SELECT ORGANIZER_CODE FROM ` + r.eventTable() + ` WHERE CODE = ?`
	err := r.db.QueryRowContext(ctx, r.dialect.Bind(query), event).Scan(&owner)
	if err != nil {
		return "", err
	}
	return owner, nil
}

func (r *Repository) TicketByCode(ctx context.Context, ticketCode string) (*ScanTicket, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	var s ScanTicket
	query := `SELECT t.CODE, o.EVENT_CODE, t.ORDER_CODE, o.STATUS, t.STATUS,
	  COALESCE(t.HOLDER_NAME, ''), tt.NAME,
	  ` + r.startStr("e.START_AT") + `, COALESCE(` + r.startStr("e.END_AT") + `, '')
	  FROM ` + r.ticketTable() + ` t
	  JOIN ` + r.orderTable() + ` o ON o.CODE = t.ORDER_CODE
	  JOIN ` + r.eventTable() + ` e ON e.CODE = o.EVENT_CODE
	  JOIN ` + r.typeTable() + ` tt ON tt.CODE = o.TICKET_TYPE_CODE
	  WHERE t.CODE = ? OR t.QR_CODE = ?`
	err := r.db.QueryRowContext(ctx, r.dialect.Bind(query), ticketCode, ticketCode).Scan(
		&s.TicketCode, &s.EventCode, &s.OrderCode, &s.OrderStatus, &s.Status,
		&s.HolderName, &s.TypeName, &s.StartAt, &s.EndAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// CheckinAtomic flips one ACTIVE ticket to USED (the single UPDATE is the
// lock: concurrent scans race here and only one reports VALID).
func (r *Repository) CheckinAtomic(ctx context.Context, ticketCode, officer, now string) (bool, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	query := `UPDATE ` + r.ticketTable() + ` SET STATUS = 'USED',
	  CHECKED_IN_AT = ` + r.dialect.Now() + `, CHECKED_IN_BY = ?,
	  UPDATED_AT = ` + r.dialect.Now() + `
	  WHERE CODE = ? AND STATUS = 'ACTIVE'`
	_ = now
	res, err := r.db.ExecContext(ctx, r.dialect.Bind(query), officer, ticketCode)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return n == 1, nil
}

func (r *Repository) ScanSummary(ctx context.Context, event string) (int, int, string, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	var checkedIn, total int
	var title string
	query := `SELECT e.TITLE,
	  COALESCE(SUM(CASE WHEN t.STATUS = 'USED' THEN 1 ELSE 0 END),0),
	  COUNT(t.CODE)
	  FROM ` + r.eventTable() + ` e
	  LEFT JOIN ` + r.orderTable() + ` o ON o.EVENT_CODE = e.CODE AND o.STATUS = 'PAID'
	  LEFT JOIN ` + r.ticketTable() + ` t ON t.ORDER_CODE = o.CODE AND t.STATUS IN ('ACTIVE','USED')
	  WHERE e.CODE = ?
	  GROUP BY e.TITLE`
	err := r.db.QueryRowContext(ctx, r.dialect.Bind(query), event).Scan(&title, &checkedIn, &total)
	if err != nil {
		return 0, 0, "", err
	}
	return checkedIn, total, title, nil
}

func (r *Repository) InsertScanLog(ctx context.Context, code, event, ticketCode, officer, result, detail string) error {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	query := `INSERT INTO ` + r.scanTable() + ` (CODE, EVENT_CODE, TICKET_CODE,
	  OFFICER_CODE, RESULT, DETAIL) VALUES (?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, r.dialect.Bind(query), code, event,
		nullStr(ticketCode), officer, result, nullStr(detail))
	return err
}

func (r *Repository) ScanHistory(ctx context.Context, event, result, officer, from, to string, limit, offset int) ([]models.ScanLogRow, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	args := []any{}
	cond := "1 = 1"
	if event != "" {
		cond += " AND s.EVENT_CODE = ?"
		args = append(args, event)
	}
	if result != "" {
		cond += " AND s.RESULT = ?"
		args = append(args, result)
	}
	if officer != "" {
		cond += " AND s.OFFICER_CODE = ?"
		args = append(args, officer)
	}
	if from != "" {
		cond += " AND s.CREATED_AT >= ?"
		args = append(args, from)
	}
	if to != "" {
		cond += " AND s.CREATED_AT < ?"
		args = append(args, to)
	}
	query := `SELECT s.CODE, ` + r.startStr("s.CREATED_AT") + `,
	  COALESCE(s.TICKET_CODE, ''), COALESCE(t.HOLDER_NAME, ''), COALESCE(tt.NAME, ''),
	  s.RESULT, u.USERNAME
	  FROM ` + r.scanTable() + ` s
	  LEFT JOIN ` + r.ticketTable() + ` t ON t.CODE = s.TICKET_CODE
	  LEFT JOIN ` + r.orderTable() + ` o ON o.CODE = t.ORDER_CODE
	  LEFT JOIN ` + r.typeTable() + ` tt ON tt.CODE = o.TICKET_TYPE_CODE
	  JOIN ` + r.userTable() + ` u ON u.CODE = s.OFFICER_CODE
	  WHERE ` + cond + ` ORDER BY s.CREATED_AT DESC` + r.pageClause(limit, offset)
	rows, err := r.db.QueryContext(ctx, r.dialect.Bind(query), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.ScanLogRow, 0)
	for rows.Next() {
		var m models.ScanLogRow
		if err := rows.Scan(&m.Code, &m.CreatedAt, &m.TicketCode, &m.HolderName,
			&m.TypeName, &m.Result, &m.Officer); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (r *Repository) ScanHistorySummary(ctx context.Context, event, result, officer, from, to string) (models.ScanHistorySummary, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	var s models.ScanHistorySummary
	args := []any{}
	cond := "1 = 1"
	if event != "" {
		cond += " AND EVENT_CODE = ?"
		args = append(args, event)
	}
	if result != "" {
		cond += " AND RESULT = ?"
		args = append(args, result)
	}
	if officer != "" {
		cond += " AND OFFICER_CODE = ?"
		args = append(args, officer)
	}
	if from != "" {
		cond += " AND CREATED_AT >= ?"
		args = append(args, from)
	}
	if to != "" {
		cond += " AND CREATED_AT < ?"
		args = append(args, to)
	}
	query := `SELECT COUNT(*),
	  COALESCE(SUM(CASE WHEN RESULT = 'VALID' THEN 1 ELSE 0 END),0)
	  FROM ` + r.scanTable() + ` WHERE ` + cond
	err := r.db.QueryRowContext(ctx, r.dialect.Bind(query), args...).Scan(&s.Total, &s.Valid)
	if err != nil {
		return s, err
	}
	s.Denied = s.Total - s.Valid
	return s, nil
}
