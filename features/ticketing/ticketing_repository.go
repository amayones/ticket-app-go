package ticketing

import (
	"context"
	"database/sql"

	"golang-backend/models"
	"golang-backend/repositories"
)

// Ticketing reads/writes T_ORDER (+ITEM), T_STOCK_RESERVATION, T_TICKET,
// T_REFUND, T_SCAN_LOG and T_EVENT_OFFICER (reusing T_EVENT, T_TICKET_TYPE,
// T_PROMO, T_CITY, CPUSER as-is).
//
// Konvensi query: SQL di variabel `query`, placeholder "?" + Bind,
// WithTimeout. Writes spanning tables run in a transaction (pola
// roles.SetRolePermissions): checkout creates, pay settles.
type RepositoryInterface interface {
	// ExpirePending flips lapsed PENDING orders to EXPIRED + drops holds.
	ExpirePending(ctx context.Context, now string) error
	// EventInfo loads event fields needed by checkout validation.
	EventInfo(ctx context.Context, event string) (*EventInfo, error)
	// TypeInfo loads ticket type fields needed by checkout validation.
	TypeInfo(ctx context.Context, typeCode string) (*TypeInfo, error)
	// PromoInfo loads a promo row by code (any status; service filters).
	PromoInfo(ctx context.Context, promoCode string) (*PromoInfo, error)
	RecordPromoUsage(ctx context.Context, promoCode, orderCode, userCode string, discount float64) error
	// ReservedQty sums live holds of a ticket type.
	ReservedQty(ctx context.Context, typeCode, now string) (int, error)
	// SoldQty sums PAID tickets of a type.
	SoldQty(ctx context.Context, typeCode string) (int, error)
	// BuyerTypeQty sums a buyer's PAID + held qty of a type.
	BuyerTypeQty(ctx context.Context, buyer, typeCode, now string) (int, error)
	// CheckoutTx persists a PENDING order + items + hold (one tx).
	CheckoutTx(ctx context.Context, o *PendingOrder) error
	// GetOrder loads one order row (any buyer; service scopes).
	GetOrder(ctx context.Context, code string) (*OrderRow, error)
	// OrderItems loads price-snapshot lines of an order.
	OrderItems(ctx context.Context, orderCode string) ([]OrderItemRow, error)
	// PayTx settles an order: PAID + tickets + hold drop (one tx).
	PayTx(ctx context.Context, orderCode string, tickets []NewTicket) error
	// CancelPending flips an owned PENDING order to CANCELLED + drops hold.
	CancelPending(ctx context.Context, buyer, code string) error
	// MyTickets lists buyer tickets (displayStatus: active|used|refunded|expired).
	MyTickets(ctx context.Context, buyer, display, q, now string, limit, offset int) ([]models.MyTicketItem, error)
	// TicketDetail loads one owned ticket with event + order info.
	TicketDetail(ctx context.Context, buyer, ticketCode string) (*models.TicketDetail, error)
	// CreateRefund files a PENDING refund (UQ-backed, one per order).
	CreateRefund(ctx context.Context, orderCode, refundCode string, amount float64, reason string) error
	// BuyerRefunds lists a buyer's refund requests.
	BuyerRefunds(ctx context.Context, buyer string) ([]models.RefundItem, error)
	// OrderTickets returns ticket codes+statuses of an order.
	OrderTickets(ctx context.Context, orderCode string) ([]TicketStatus, error)
	// OfficerAssigned reports a scan-scope assignment.
	OfficerAssigned(ctx context.Context, officer, event string) (bool, error)
	// AssignedEvents lists events the caller may scan (assigned + owned).
	AssignedEvents(ctx context.Context, officer string) ([]AssignedEvent, error)
	// EventOwner returns the organizer code of an event.
	EventOwner(ctx context.Context, event string) (string, error)
	// TicketByCode loads a ticket + order + event row for scanning.
	TicketByCode(ctx context.Context, ticketCode string) (*ScanTicket, error)
	// CheckinAtomic flips ACTIVE -> USED once (returns true to winner).
	CheckinAtomic(ctx context.Context, ticketCode, officer, now string) (bool, error)
	// ScanSummary counts USED vs sellable tickets of an event.
	ScanSummary(ctx context.Context, event string) (checkedIn, total int, title string, err error)
	// InsertScanLog records every scan attempt.
	InsertScanLog(ctx context.Context, code, event, ticketCode, officer, result, detail string) error
	// ScanHistory pages the scan log with filters.
	ScanHistory(ctx context.Context, event, result, officer, from, to string, limit, offset int) ([]models.ScanLogRow, error)
	// ScanHistorySummary counts total/valid/denied under the same filters.
	ScanHistorySummary(ctx context.Context, event, result, officer, from, to string) (models.ScanHistorySummary, error)
}

// PendingOrder is the checkout payload persisted by CheckoutTx.
type PendingOrder struct {
	OrderCode string
	EventCode string
	BuyerCode string
	TypeCode  string
	Qty       int
	Gross     float64
	Discount  float64
	Fee       float64
	Net       float64
	PromoCode string
	ResvCode  string
	ExpiresAt string
	Items     []PendingItem
}

// PendingItem is one priced line (snapshot at purchase).
type PendingItem struct {
	Code     string
	ResvCode string
	TypeCode string
	TypeName string
	Price    float64
	Qty      int
}

// NewTicket is one issued ticket row.
type NewTicket struct {
	Code     string
	QRCode   string
	QRToken  string
	Holder   string
	Email    string
	Phone    string
}

// OrderRow is the T_ORDER row plus event context for settlement.
type OrderRow struct {
	Code         string
	EventCode    string
	BuyerCode    string
	TicketType   string
	Qty          int
	Gross        float64
	Discount     float64
	Fee          float64
	Net          float64
	PromoCode    string
	Status       string
	ReservedUntil string
	HolderName   string
	Email        string
	Phone        string
}

// OrderItemRow is one T_ORDER_ITEM row.
type OrderItemRow struct {
	TypeCode string
	TypeName string
	Price    float64
	Qty      int
}

// TicketStatus is a ticket code + status pair.
type TicketStatus struct {
	Code   string
	Status string
}

// ScanTicket is the scan-validation row.
type ScanTicket struct {
	TicketCode  string
	EventCode   string
	OrderCode   string
	OrderStatus string
	Status      string
	HolderName  string
	TypeName    string
	StartAt     string
	EndAt       string
}

type Repository struct {
	db      *sql.DB
	dialect repositories.Dialect
}

func NewRepository(db *sql.DB, dialect repositories.Dialect) RepositoryInterface {
	return &Repository{db: db, dialect: dialect}
}

func (r *Repository) orderTable() string  { return r.dialect.Table("T_ORDER") }
func (r *Repository) itemTable() string   { return r.dialect.Table("T_ORDER_ITEM") }
func (r *Repository) resvTable() string   { return r.dialect.Table("T_STOCK_RESERVATION") }
func (r *Repository) ticketTable() string { return r.dialect.Table("T_TICKET") }
func (r *Repository) typeTable() string   { return r.dialect.Table("T_TICKET_TYPE") }
func (r *Repository) eventTable() string  { return r.dialect.Table("T_EVENT") }
func (r *Repository) promoTable() string  { return r.dialect.Table("T_PROMO") }
func (r *Repository) refundTable() string { return r.dialect.Table("T_REFUND") }
func (r *Repository) scanTable() string   { return r.dialect.Table("T_SCAN_LOG") }
func (r *Repository) offTable() string    { return r.dialect.Table("T_EVENT_OFFICER") }
func (r *Repository) userTable() string {
	if r.dialect == repositories.DialectMSSQL {
		return "dbo.CPUSER"
	}
	return "CPUSER"
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

// pageClause renders pagination with integer literals (MSSQL FETCH forbids
// parameters; ints are clamped Go values, never request text).
func (r *Repository) pageClause(limit, offset int) string {
	if limit <= 0 || limit > 500 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	if r.dialect == repositories.DialectMSSQL {
		return " OFFSET " + itoa(offset) + " ROWS FETCH NEXT " + itoa(limit) + " ROWS ONLY"
	}
	return " LIMIT " + itoa(limit) + " OFFSET " + itoa(offset)
}

func itoa(n int) string {
	if n < 0 {
		n = 0
	}
	if n == 0 {
		return "0"
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}

func (r *Repository) ExpirePending(ctx context.Context, now string) error {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	query := `UPDATE ` + r.orderTable() + ` SET STATUS = 'EXPIRED',
	  UPDATED_AT = ` + r.dialect.Now() + `
	  WHERE STATUS = 'PENDING' AND RESERVED_UNTIL IS NOT NULL AND RESERVED_UNTIL < ?`
	if _, err := tx.ExecContext(ctx, r.dialect.Bind(query), now); err != nil {
		return err
	}
	query = `DELETE FROM ` + r.resvTable() + ` WHERE EXPIRES_AT < ?`
	if _, err := tx.ExecContext(ctx, r.dialect.Bind(query), now); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *Repository) ReservedQty(ctx context.Context, typeCode, now string) (int, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	var n int
	query := `SELECT COALESCE(SUM(QTY),0) FROM ` + r.resvTable() + `
	  WHERE TICKET_TYPE_CODE = ? AND EXPIRES_AT >= ?`
	err := r.db.QueryRowContext(ctx, r.dialect.Bind(query), typeCode, now).Scan(&n)
	return n, err
}

func (r *Repository) SoldQty(ctx context.Context, typeCode string) (int, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	var n int
	query := `SELECT COALESCE(SUM(o.QTY),0) FROM ` + r.orderTable() + ` o
	  WHERE o.TICKET_TYPE_CODE = ? AND o.STATUS = 'PAID'`
	err := r.db.QueryRowContext(ctx, r.dialect.Bind(query), typeCode).Scan(&n)
	return n, err
}

func (r *Repository) BuyerTypeQty(ctx context.Context, buyer, typeCode, now string) (int, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	var n int
	// PAID qty + live holds (same buyer, unexpired).
	query := `SELECT COALESCE(SUM(o.QTY),0) FROM ` + r.orderTable() + ` o
	  WHERE o.BUYER_CODE = ? AND o.TICKET_TYPE_CODE = ? AND o.STATUS = 'PAID'`
	if err := r.db.QueryRowContext(ctx, r.dialect.Bind(query), buyer, typeCode).Scan(&n); err != nil {
		return 0, err
	}
	var h int
	query = `SELECT COALESCE(SUM(s.QTY),0) FROM ` + r.resvTable() + ` s
	  JOIN ` + r.orderTable() + ` o ON o.CODE = s.ORDER_CODE
	  WHERE o.BUYER_CODE = ? AND s.TICKET_TYPE_CODE = ? AND s.EXPIRES_AT >= ?`
	if err := r.db.QueryRowContext(ctx, r.dialect.Bind(query), buyer, typeCode, now).Scan(&h); err != nil {
		return 0, err
	}
	return n + h, nil
}

func (r *Repository) CheckoutTx(ctx context.Context, o *PendingOrder) error {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	query := `INSERT INTO ` + r.orderTable() + ` (CODE, EVENT_CODE, BUYER_CODE,
	  TICKET_TYPE_CODE, QTY, GROSS_AMOUNT, DISCOUNT_AMOUNT, PLATFORM_FEE, NET_AMOUNT,
	  PROMO_CODE, STATUS, ORDER_DATE, RESERVED_UNTIL)
	  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'PENDING', ` + r.dialect.Now() + `, ?)`
	if _, err := tx.ExecContext(ctx, r.dialect.Bind(query), o.OrderCode, o.EventCode,
		o.BuyerCode, o.TypeCode, o.Qty, o.Gross, o.Discount, o.Fee, o.Net,
		repositories.NullStr(o.PromoCode), o.ExpiresAt); err != nil {
		return err
	}
	for _, it := range o.Items {
		query = `INSERT INTO ` + r.itemTable() + ` (CODE, ORDER_CODE, TICKET_TYPE_CODE,
		  TYPE_NAME, UNIT_PRICE, QTY) VALUES (?, ?, ?, ?, ?, ?)`
		if _, err := tx.ExecContext(ctx, r.dialect.Bind(query), it.Code, o.OrderCode,
			it.TypeCode, it.TypeName, it.Price, it.Qty); err != nil {
			return err
		}
		query = `INSERT INTO ` + r.resvTable() + ` (CODE, ORDER_CODE, TICKET_TYPE_CODE,
		  QTY, EXPIRES_AT) VALUES (?, ?, ?, ?, ?)`
		if _, err := tx.ExecContext(ctx, r.dialect.Bind(query), it.ResvCode, o.OrderCode,
			it.TypeCode, it.Qty, o.ExpiresAt); err != nil {
			return err
		}
	}
	return tx.Commit()
}
