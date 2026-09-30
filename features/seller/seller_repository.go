package seller

import (
	"context"
	"database/sql"
	"strconv"

	"golang-backend/models"
	"golang-backend/repositories"
)

// Dashboard reads: read-only aggregation over T_* tables (T_MERCH_ORDER,
// T_PRODUCT, T_COMPLAINT, T_REVIEW, T_PAYOUT, T_SELLER_BALANCE).
//
// Konvensi query: sama seperti features/organizer (variabel `query`,
// placeholder "?" + Bind, WithTimeout, rentang [from, toPlusOne)).
type RepositoryInterface interface {
	// CategoryList returns distinct product categories of the seller.
	CategoryList(ctx context.Context, seller string) ([]string, error)
	// OrderSums aggregates non-cancelled orders: sales, order count.
	OrderSums(ctx context.Context, seller, category, from, toPlusOne string) (sales float64, orders int, err error)
	// OrdersByStatus groups orders by status (for summary + chart).
	OrdersByStatus(ctx context.Context, seller, category, from, toPlusOne string) ([]models.SlicePoint, error)
	// SalesTrend returns per-day buckets (qty + revenue).
	SalesTrend(ctx context.Context, seller, category, from, toPlusOne string) ([]models.TrendPoint, error)
	// TopProducts returns the 5 best sellers by revenue.
	TopProducts(ctx context.Context, seller, category, from, toPlusOne string) ([]models.SlicePoint, error)
	// ActionableOrders lists NEW/PROCESSING orders with ship deadlines.
	ActionableOrders(ctx context.Context, seller string) ([]models.SellerOrder, error)
	// LowStockProducts lists ACTIVE products at/under threshold.
	LowStockProducts(ctx context.Context, seller string) ([]models.SellerProduct, error)
	// RecentOrders returns the latest orders (limit).
	RecentOrders(ctx context.Context, seller, category string, limit int) ([]models.SellerOrder, error)
	// OpenComplaints lists unresolved complaints with order info.
	OpenComplaints(ctx context.Context, seller string) ([]models.SellerComplaint, error)
	// ActiveProducts counts ACTIVE products.
	ActiveProducts(ctx context.Context, seller string) (int, error)
	// LowStockCount counts ACTIVE products at/under threshold.
	LowStockCount(ctx context.Context, seller string) (int, error)
	// OpenComplaintCount counts OPEN complaints.
	OpenComplaintCount(ctx context.Context, seller string) (int, error)
	// Rating averages ACTIVE product ratings (0 when unrated).
	Rating(ctx context.Context, seller string) (float64, error)
	// RejectedProducts counts products rejected by moderation.
	RejectedProducts(ctx context.Context, seller string) (int, error)
	// Balance reads the seller escrow/available row (zero when absent).
	Balance(ctx context.Context, seller string) (models.Balance, error)
	// PayoutReady sums READY payouts of the seller.
	PayoutReady(ctx context.Context, seller string) (float64, error)
}

type Repository struct {
	db      *sql.DB
	dialect repositories.Dialect
}

func NewRepository(db *sql.DB, dialect repositories.Dialect) RepositoryInterface {
	return &Repository{db: db, dialect: dialect}
}

func (r *Repository) orderTable() string     { return r.dialect.Table("T_MERCH_ORDER") }
func (r *Repository) productTable() string  { return r.dialect.Table("T_PRODUCT") }
func (r *Repository) complaintTable() string { return r.dialect.Table("T_COMPLAINT") }
func (r *Repository) reviewTable() string   { return r.dialect.Table("T_REVIEW") }
func (r *Repository) payoutTable() string   { return r.dialect.Table("T_PAYOUT") }
func (r *Repository) balTable() string      { return r.dialect.Table("T_SELLER_BALANCE") }

// catSQL optionally restricts rows to one category (empty = all).
const catSQL = " AND (? = '' OR p.CATEGORY = ?)"

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

// liveStatuses are orders counted as sales (everything but cancelled/returned).
const liveStatuses = "('NEW','PROCESSING','SHIPPED','COMPLETED')"

func (r *Repository) CategoryList(ctx context.Context, seller string) ([]string, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	query := `SELECT DISTINCT CATEGORY FROM ` + r.productTable() + ` WHERE SELLER_CODE = ? ORDER BY CATEGORY ASC`
	rows, err := r.db.QueryContext(ctx, r.dialect.Bind(query), seller)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]string, 0)
	for rows.Next() {
		var c string
		if err := rows.Scan(&c); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *Repository) OrderSums(ctx context.Context, seller, category, from, toPlusOne string) (float64, int, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	var sales float64
	var orders int
	query := `SELECT COALESCE(SUM(o.TOTAL_AMOUNT),0), COUNT(*)
	  FROM ` + r.orderTable() + ` o JOIN ` + r.productTable() + ` p ON p.CODE = o.PRODUCT_CODE
	  WHERE o.SELLER_CODE = ? AND o.STATUS IN ` + liveStatuses + `
	  AND o.ORDER_DATE >= ? AND o.ORDER_DATE < ?` + catSQL
	err := r.db.QueryRowContext(ctx, r.dialect.Bind(query), seller, from, toPlusOne, category, category).
		Scan(&sales, &orders)
	return sales, orders, err
}

func (r *Repository) OrdersByStatus(ctx context.Context, seller, category, from, toPlusOne string) ([]models.SlicePoint, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	query := `SELECT o.STATUS, COALESCE(SUM(o.QTY),0), COALESCE(SUM(o.TOTAL_AMOUNT),0), COUNT(*)
	  FROM ` + r.orderTable() + ` o JOIN ` + r.productTable() + ` p ON p.CODE = o.PRODUCT_CODE
	  WHERE o.SELLER_CODE = ? AND o.ORDER_DATE >= ? AND o.ORDER_DATE < ?` + catSQL + `
	  GROUP BY o.STATUS ORDER BY 4 DESC`
	rows, err := r.db.QueryContext(ctx, r.dialect.Bind(query), seller, from, toPlusOne, category, category)
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

func (r *Repository) SalesTrend(ctx context.Context, seller, category, from, toPlusOne string) ([]models.TrendPoint, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	query := `SELECT ` + r.dayExpr("o.ORDER_DATE") + `, COALESCE(SUM(o.QTY),0), COALESCE(SUM(o.TOTAL_AMOUNT),0)
	  FROM ` + r.orderTable() + ` o JOIN ` + r.productTable() + ` p ON p.CODE = o.PRODUCT_CODE
	  WHERE o.SELLER_CODE = ? AND o.STATUS IN ` + liveStatuses + `
	  AND o.ORDER_DATE >= ? AND o.ORDER_DATE < ?` + catSQL + `
	  GROUP BY ` + r.dayExpr("o.ORDER_DATE") + ` ORDER BY 1 ASC`
	rows, err := r.db.QueryContext(ctx, r.dialect.Bind(query), seller, from, toPlusOne, category, category)
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

func (r *Repository) TopProducts(ctx context.Context, seller, category, from, toPlusOne string) ([]models.SlicePoint, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	limit, page, top := r.limitClause(5)
	query := `SELECT ` + top + ` p.NAME, COALESCE(SUM(o.QTY),0), COALESCE(SUM(o.TOTAL_AMOUNT),0), COUNT(*)
	  FROM ` + r.orderTable() + ` o JOIN ` + r.productTable() + ` p ON p.CODE = o.PRODUCT_CODE
	  WHERE o.SELLER_CODE = ? AND o.STATUS IN ` + liveStatuses + `
	  AND o.ORDER_DATE >= ? AND o.ORDER_DATE < ?` + catSQL + `
	  GROUP BY p.NAME ORDER BY 3 DESC` + page
	args := []any{seller, from, toPlusOne, category, category}
	if limit > 0 {
		args = append([]any{limit}, args...)
	}
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

// limitClause renders TOP (?) / LIMIT n per engine (limit<=0 = plain query).
func (r *Repository) limitClause(n int) (int, string, string) {
	if r.dialect == repositories.DialectMSSQL {
		return n, "", "TOP (?)"
	}
	return 0, " LIMIT " + strconv.Itoa(n), ""
}

func (r *Repository) scanOrders(ctx context.Context, query string, args []any) ([]models.SellerOrder, error) {
	rows, err := r.db.QueryContext(ctx, r.dialect.Bind(query), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.SellerOrder, 0)
	for rows.Next() {
		var o models.SellerOrder
		var deadline sql.NullString
		if err := rows.Scan(&o.Code, &o.Product, &o.Buyer, &o.Qty, &o.Total, &o.Status, &o.OrderDate, &deadline); err != nil {
			return nil, err
		}
		if deadline.Valid {
			o.ShipDeadline = deadline.String
		}
		out = append(out, o)
	}
	return out, rows.Err()
}

func (r *Repository) ActionableOrders(ctx context.Context, seller string) ([]models.SellerOrder, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	limit, page, top := r.limitClause(20)
	query := `SELECT ` + top + ` o.CODE, p.NAME, o.BUYER_CODE, o.QTY, o.TOTAL_AMOUNT,
	  o.STATUS, ` + r.startStr("o.ORDER_DATE") + `, ` + r.startStr("o.SHIP_DEADLINE") + `
	  FROM ` + r.orderTable() + ` o JOIN ` + r.productTable() + ` p ON p.CODE = o.PRODUCT_CODE
	  WHERE o.SELLER_CODE = ? AND o.STATUS IN ('NEW','PROCESSING')
	  ORDER BY o.SHIP_DEADLINE ASC` + page
	args := []any{seller}
	if limit > 0 {
		args = append([]any{limit}, args...)
	}
	// SHIP_DEADLINE may be NULL: startStr(NULL) yields NULL, scanned via NullString.
	return r.scanOrders(ctx, query, args)
}

func (r *Repository) LowStockProducts(ctx context.Context, seller string) ([]models.SellerProduct, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	query := `SELECT CODE, NAME, CATEGORY, PRICE, STOCK, LOW_STOCK_THRESHOLD, STATUS
	  FROM ` + r.productTable() + `
	  WHERE SELLER_CODE = ? AND STATUS = 'ACTIVE' AND STOCK <= LOW_STOCK_THRESHOLD
	  ORDER BY STOCK ASC`
	rows, err := r.db.QueryContext(ctx, r.dialect.Bind(query), seller)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.SellerProduct, 0)
	for rows.Next() {
		var p models.SellerProduct
		if err := rows.Scan(&p.Code, &p.Name, &p.Category, &p.Price, &p.Stock, &p.Threshold, &p.Status); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *Repository) RecentOrders(ctx context.Context, seller, category string, limit int) ([]models.SellerOrder, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	if limit <= 0 || limit > 100 {
		limit = 10
	}
	n, page, top := r.limitClause(limit)
	_ = n
	query := `SELECT ` + top + ` o.CODE, p.NAME, o.BUYER_CODE, o.QTY, o.TOTAL_AMOUNT,
	  o.STATUS, ` + r.startStr("o.ORDER_DATE") + `, ` + r.startStr("o.SHIP_DEADLINE") + `
	  FROM ` + r.orderTable() + ` o JOIN ` + r.productTable() + ` p ON p.CODE = o.PRODUCT_CODE
	  WHERE o.SELLER_CODE = ?` + catSQL + `
	  ORDER BY o.ORDER_DATE DESC` + page
	args := []any{seller, category, category}
	if r.dialect == repositories.DialectMSSQL {
		args = append([]any{limit}, args...)
	}
	return r.scanOrders(ctx, query, args)
}

func (r *Repository) OpenComplaints(ctx context.Context, seller string) ([]models.SellerComplaint, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	query := `SELECT c.CODE, c.ORDER_CODE, c.BUYER_CODE, c.MESSAGE, c.STATUS,
	  ` + r.startStr("c.CREATED_AT") + `
	  FROM ` + r.complaintTable() + ` c JOIN ` + r.orderTable() + ` o ON o.CODE = c.ORDER_CODE
	  WHERE o.SELLER_CODE = ? AND c.STATUS = 'OPEN'
	  ORDER BY c.CREATED_AT ASC`
	rows, err := r.db.QueryContext(ctx, r.dialect.Bind(query), seller)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.SellerComplaint, 0)
	for rows.Next() {
		var c models.SellerComplaint
		if err := rows.Scan(&c.Code, &c.OrderCode, &c.Buyer, &c.Message, &c.Status, &c.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *Repository) ActiveProducts(ctx context.Context, seller string) (int, error) {
	return r.countProducts(ctx, seller, " AND STATUS = 'ACTIVE'")
}

func (r *Repository) LowStockCount(ctx context.Context, seller string) (int, error) {
	return r.countProducts(ctx, seller, " AND STATUS = 'ACTIVE' AND STOCK <= LOW_STOCK_THRESHOLD")
}

func (r *Repository) RejectedProducts(ctx context.Context, seller string) (int, error) {
	return r.countProducts(ctx, seller, " AND STATUS = 'REJECTED'")
}

func (r *Repository) countProducts(ctx context.Context, seller, extra string) (int, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	var n int
	query := `SELECT COUNT(*) FROM ` + r.productTable() + ` WHERE SELLER_CODE = ?` + extra
	err := r.db.QueryRowContext(ctx, r.dialect.Bind(query), seller).Scan(&n)
	return n, err
}

func (r *Repository) OpenComplaintCount(ctx context.Context, seller string) (int, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	var n int
	query := `SELECT COUNT(*) FROM ` + r.complaintTable() + ` c
	  JOIN ` + r.orderTable() + ` o ON o.CODE = c.ORDER_CODE
	  WHERE o.SELLER_CODE = ? AND c.STATUS = 'OPEN'`
	err := r.db.QueryRowContext(ctx, r.dialect.Bind(query), seller).Scan(&n)
	return n, err
}

func (r *Repository) Rating(ctx context.Context, seller string) (float64, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	var v sql.NullFloat64
	query := `SELECT AVG(CAST(RATING AS FLOAT)) FROM ` + r.productTable() + `
	  WHERE SELLER_CODE = ? AND STATUS = 'ACTIVE' AND REVIEW_COUNT > 0`
	if err := r.db.QueryRowContext(ctx, r.dialect.Bind(query), seller).Scan(&v); err != nil {
		return 0, err
	}
	if !v.Valid {
		return 0, nil
	}
	return v.Float64, nil
}

func (r *Repository) Balance(ctx context.Context, seller string) (models.Balance, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	var b models.Balance
	query := `SELECT COALESCE(ESCROW_AMOUNT,0), COALESCE(AVAILABLE_AMOUNT,0)
	  FROM ` + r.balTable() + ` WHERE SELLER_CODE = ?`
	err := r.db.QueryRowContext(ctx, r.dialect.Bind(query), seller).Scan(&b.Escrow, &b.Available)
	if err == sql.ErrNoRows {
		return models.Balance{}, nil
	}
	return b, err
}

func (r *Repository) PayoutReady(ctx context.Context, seller string) (float64, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	var amount float64
	query := `SELECT COALESCE(SUM(AMOUNT),0) FROM ` + r.payoutTable() + `
	  WHERE OWNER_CODE = ? AND STATUS = 'READY'`
	err := r.db.QueryRowContext(ctx, r.dialect.Bind(query), seller).Scan(&amount)
	return amount, err
}
