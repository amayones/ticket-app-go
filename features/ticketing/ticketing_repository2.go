package ticketing

// Second half: order settlement, buyer ticket reads, refund filing and
// scan validation queries. Conventions identical to ticketing_repository.go.
import (
	"context"
	"database/sql"
	"strings"

	"golang-backend/models"
	"golang-backend/repositories"
	"golang-backend/utils"
)

func nullStr(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func escapeQ(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `%`, `\%`)
	s = strings.ReplaceAll(s, `_`, `\_`)
	return s
}

func (r *Repository) cityTable() string { return r.dialect.Table("T_CITY") }

func (r *Repository) GetOrder(ctx context.Context, code string) (*OrderRow, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	var o OrderRow
	var promo, reserved sql.NullString
	query := `SELECT CODE, EVENT_CODE, BUYER_CODE, TICKET_TYPE_CODE, QTY,
	  GROSS_AMOUNT, DISCOUNT_AMOUNT, PLATFORM_FEE, NET_AMOUNT, PROMO_CODE,
	  STATUS, COALESCE(` + r.startStr("RESERVED_UNTIL") + `, '')
	  FROM ` + r.orderTable() + ` WHERE CODE = ?`
	err := r.db.QueryRowContext(ctx, r.dialect.Bind(query), code).Scan(
		&o.Code, &o.EventCode, &o.BuyerCode, &o.TicketType, &o.Qty,
		&o.Gross, &o.Discount, &o.Fee, &o.Net, &promo, &o.Status, &reserved)
	if err != nil {
		return nil, err
	}
	if promo.Valid {
		o.PromoCode = promo.String
	}
	if reserved.Valid {
		o.ReservedUntil = reserved.String
	}
	return &o, nil
}

func (r *Repository) OrderItems(ctx context.Context, orderCode string) ([]OrderItemRow, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	query := `SELECT TICKET_TYPE_CODE, TYPE_NAME, UNIT_PRICE, QTY
	  FROM ` + r.itemTable() + ` WHERE ORDER_CODE = ? ORDER BY ID ASC`
	rows, err := r.db.QueryContext(ctx, r.dialect.Bind(query), orderCode)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]OrderItemRow, 0)
	for rows.Next() {
		var it OrderItemRow
		if err := rows.Scan(&it.TypeCode, &it.TypeName, &it.Price, &it.Qty); err != nil {
			return nil, err
		}
		out = append(out, it)
	}
	return out, rows.Err()
}

// PayTx settles an order in one transaction: PAID + tickets + hold drop.
// Tickets carry buyer contact; QR codes/tokens come generated.
func (r *Repository) PayTx(ctx context.Context, orderCode string, tickets []NewTicket) error {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	query := `UPDATE ` + r.orderTable() + ` SET STATUS = 'PAID', RESERVED_UNTIL = NULL,
	  UPDATED_AT = ` + r.dialect.Now() + ` WHERE CODE = ? AND STATUS = 'PENDING'`
	res, err := tx.ExecContext(ctx, r.dialect.Bind(query), orderCode)
	if err != nil {
		return err
	}
	if n, err := res.RowsAffected(); err != nil || n == 0 {
		if err == nil {
			return sql.ErrNoRows
		}
		return err
	}
	for _, t := range tickets {
		query = `INSERT INTO ` + r.ticketTable() + ` (CODE, ORDER_CODE, QR_CODE,
		  QR_TOKEN, HOLDER_NAME, EMAIL, PHONE, STATUS)
		  VALUES (?, ?, ?, ?, ?, ?, ?, 'ACTIVE')`
		if _, err := tx.ExecContext(ctx, r.dialect.Bind(query), t.Code, orderCode,
			t.QRCode, t.QRToken, t.Holder, nullStr(t.Email), nullStr(t.Phone)); err != nil {
			return err
		}
	}
	query = `DELETE FROM ` + r.resvTable() + ` WHERE ORDER_CODE = ?`
	if _, err := tx.ExecContext(ctx, r.dialect.Bind(query), orderCode); err != nil {
		return err
	}
	var promo sql.NullString
	var disc float64
	_ = tx.QueryRowContext(ctx, r.dialect.Bind(`SELECT PROMO_CODE, DISCOUNT_AMOUNT FROM `+r.orderTable()+` WHERE CODE = ?`), orderCode).Scan(&promo, &disc)
	if promo.Valid && promo.String != "" && disc > 0 {
		var pcode string
		var buyerCode string
		_ = tx.QueryRowContext(ctx, r.dialect.Bind(`SELECT BUYER_CODE FROM `+r.orderTable()+` WHERE CODE = ?`), orderCode).Scan(&buyerCode)
		_ = tx.QueryRowContext(ctx, r.dialect.Bind(`SELECT CODE FROM `+r.promoTable()+` WHERE PROMO_CODE = ?`), promo.String).Scan(&pcode)
		if pcode != "" && buyerCode != "" {
			ucode, _ := r.genCode("PU-")
			_, _ = tx.ExecContext(ctx, r.dialect.Bind(`INSERT INTO `+r.dialect.Table("T_PROMO_USAGE")+` (CODE, PROMO_CODE, ORDER_CODE, USER_CODE, DISCOUNT_AMOUNT) VALUES (?, ?, ?, ?, ?)`), ucode, pcode, orderCode, buyerCode, disc)
			_, _ = tx.ExecContext(ctx, r.dialect.Bind(`UPDATE `+r.promoTable()+` SET USED_COUNT = USED_COUNT + 1, UPDATED_AT = `+r.dialect.Now()+` WHERE CODE = ?`), pcode)
		}
	}
	return tx.Commit()
}

func (r *Repository) genCode(prefix string) (string, error) {
	return utils.GenerateCode(prefix)
}

func (r *Repository) RecordPromoUsage(ctx context.Context, promoCode, orderCode, userCode string, discount float64) error {
	ctx2, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	var pcode string
	_ = r.db.QueryRowContext(ctx2, r.dialect.Bind(`SELECT CODE FROM `+r.promoTable()+` WHERE PROMO_CODE = ?`), promoCode).Scan(&pcode)
	if pcode == "" {
		return nil
	}
	ucode, err := r.genCode("PU-")
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx2, r.dialect.Bind(`INSERT INTO `+r.dialect.Table("T_PROMO_USAGE")+` (CODE, PROMO_CODE, ORDER_CODE, USER_CODE, DISCOUNT_AMOUNT) VALUES (?, ?, ?, ?, ?)`), ucode, pcode, orderCode, userCode, discount)
	return err
}

func (r *Repository) CancelPending(ctx context.Context, buyer, code string) error {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	query := `UPDATE ` + r.orderTable() + ` SET STATUS = 'CANCELLED',
	  RESERVED_UNTIL = NULL, UPDATED_AT = ` + r.dialect.Now() + `
	  WHERE CODE = ? AND BUYER_CODE = ? AND STATUS = 'PENDING'`
	res, err := tx.ExecContext(ctx, r.dialect.Bind(query), code, buyer)
	if err != nil {
		return err
	}
	if n, err := res.RowsAffected(); err != nil || n == 0 {
		if err == nil {
			return sql.ErrNoRows
		}
		return err
	}
	query = `DELETE FROM ` + r.resvTable() + ` WHERE ORDER_CODE = ?`
	if _, err := tx.ExecContext(ctx, r.dialect.Bind(query), code); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *Repository) MyTickets(ctx context.Context, buyer, display, q, now string, limit, offset int) ([]models.MyTicketItem, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	args := []any{buyer}
	cond := "o.BUYER_CODE = ? AND o.STATUS = 'PAID'"
	switch display {
	case "used":
		cond += " AND t.STATUS = 'USED'"
	case "refunded":
		cond += " AND t.STATUS = 'REFUNDED'"
	case "expired":
		cond += " AND t.STATUS = 'ACTIVE' AND e.END_AT IS NOT NULL AND e.END_AT < ?"
		args = append(args, now)
	default: // active: usable now (event not over)
		cond += " AND t.STATUS = 'ACTIVE' AND (e.END_AT IS NULL OR e.END_AT >= ?)"
		args = append(args, now)
	}
	if q != "" {
		cond += ` AND (e.TITLE LIKE ? ESCAPE '\' OR t.QR_CODE LIKE ? ESCAPE '\' OR tt.NAME LIKE ? ESCAPE '\')`
		like := "%" + escapeQ(q) + "%"
		args = append(args, like, like, like)
	}
	query := `SELECT t.CODE, e.CODE, e.TITLE, ` + r.startStr("e.START_AT") + `,
	  COALESCE(e.VENUE, ''), tt.NAME, COALESCE(t.HOLDER_NAME, ''),
	  t.STATUS, o.CODE, o.NET_AMOUNT
	  FROM ` + r.ticketTable() + ` t
	  JOIN ` + r.orderTable() + ` o ON o.CODE = t.ORDER_CODE
	  JOIN ` + r.eventTable() + ` e ON e.CODE = o.EVENT_CODE
	  JOIN ` + r.typeTable() + ` tt ON tt.CODE = o.TICKET_TYPE_CODE
	  WHERE ` + cond + ` ORDER BY e.START_AT ASC` + r.pageClause(limit, offset)
	rows, err := r.db.QueryContext(ctx, r.dialect.Bind(query), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.MyTicketItem, 0)
	for rows.Next() {
		var m models.MyTicketItem
		if err := rows.Scan(&m.TicketCode, &m.EventCode, &m.EventTitle, &m.StartAt,
			&m.Venue, &m.TypeName, &m.HolderName, &m.Status, &m.OrderCode, &m.Net); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (r *Repository) TicketDetail(ctx context.Context, buyer, ticketCode string) (*models.TicketDetail, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	var d models.TicketDetail
	var end, addr, city, poster sql.NullString
	query := `SELECT t.CODE, COALESCE(t.QR_TOKEN, ''), COALESCE(t.HOLDER_NAME, ''),
	  tt.NAME, e.CODE, e.TITLE, ` + r.startStr("e.START_AT") + `,
	  COALESCE(` + r.startStr("e.END_AT") + `, ''), COALESCE(e.VENUE, ''),
	  COALESCE(e.ADDRESS, ''), COALESCE(c.NAME, ''), COALESCE(e.POSTER_URL, ''),
	  t.STATUS, o.CODE, o.NET_AMOUNT
	  FROM ` + r.ticketTable() + ` t
	  JOIN ` + r.orderTable() + ` o ON o.CODE = t.ORDER_CODE
	  JOIN ` + r.eventTable() + ` e ON e.CODE = o.EVENT_CODE
	  JOIN ` + r.typeTable() + ` tt ON tt.CODE = o.TICKET_TYPE_CODE
	  LEFT JOIN ` + r.cityTable() + ` c ON c.CODE = e.CITY_CODE
	  WHERE t.CODE = ? AND o.BUYER_CODE = ?`
	err := r.db.QueryRowContext(ctx, r.dialect.Bind(query), ticketCode, buyer).Scan(
		&d.TicketCode, &d.QRToken, &d.HolderName, &d.TypeName, &d.EventCode,
		&d.EventTitle, &d.StartAt, &end, &d.Venue, &addr, &city, &poster,
		&d.Status, &d.OrderCode, &d.Net)
	if err != nil {
		return nil, err
	}
	d.EndAt, d.Address, d.City, d.PosterURL = end.String, addr.String, city.String, poster.String
	return &d, nil
}
