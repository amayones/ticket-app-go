package event

import (
	"context"
	"database/sql"
	"strings"

	"golang-backend/models"
	"golang-backend/repositories"
)

// Event reads/writes over T_EVENT + T_TICKET_TYPE (+ T_ORDER/T_TICKET for
// counts and attendees). Every method scopes by owner (ORGANIZER_CODE);
// cross-owner rows behave as not found.
//
// Konvensi query: SQL di variabel `query`, placeholder "?" + Bind,
// WithTimeout. Multi-table writes (duplicate) run in a transaction
// (pola roles.SetRolePermissions).
type RepositoryInterface interface {
	// MineList lists non-deleted owned events with sold/quota/revenue.
	// sort is service-validated (nearest|newest|popular); ORDER BY is
	// assembled from constants, never request text.
	MineList(ctx context.Context, owner, status, category, city, from, toPlusOne, q, sort string, limit, offset int) ([]models.MyEventItem, error)
	// EventByCode loads one non-deleted owned event (any status).
	EventByCode(ctx context.Context, owner, code string) (*models.EventDetailOwner, error)
	// EventStats returns sold tickets + net revenue of an event.
	EventStats(ctx context.Context, event string) (sold int, revenue float64, err error)
	// SlugExists reports slug usage excluding one event code ("" = any).
	SlugExists(ctx context.Context, slug, exclude string) (bool, error)
	// TitleExists? No - titles may repeat; slugs carry uniqueness.
	CreateEvent(ctx context.Context, owner, code, slug string, in models.EventInput) error
	UpdateEvent(ctx context.Context, owner, code string, in models.EventInput) error
	SetStatus(ctx context.Context, owner, code, status, cancelReason string) error
	SoftDelete(ctx context.Context, owner, code string) error
	// MarkEnded flips the owner's past-due PUBLISHED events to ENDED.
	MarkEnded(ctx context.Context, owner, now string) error
	// Duplicate copies an event + its ticket types as DRAFT (tx).
	Duplicate(ctx context.Context, owner, srcCode, newCode, newSlug string) error
	// AffectedBuyers counts distinct PAID buyers (cancel warning).
	AffectedBuyers(ctx context.Context, event string) (int, error)
	// CategoryExists / CityExists validate filter/form codes.
	CategoryExists(ctx context.Context, code string) (bool, error)
	CityExists(ctx context.Context, code string) (bool, error)
	// WasPublished reports whether the event was ever published.
	WasPublished(ctx context.Context, owner, code string) (bool, error)
	// TypeList lists ticket types with sold counts (sort order).
	TypeList(ctx context.Context, owner, event string) ([]models.TicketTypeRow, error)
	// TypeNameTaken reports duplicate type names (excluding one code).
	TypeNameTaken(ctx context.Context, event, name, exclude string) (bool, error)
	// TypeSold returns sold tickets of one type.
	TypeSold(ctx context.Context, typeCode string) (int, error)
	// NextTypeSort returns max sort order + 1 for appended types.
	NextTypeSort(ctx context.Context, event string) (int, error)
	CreateType(ctx context.Context, owner, event, code string, in models.TicketTypeInput, sortOrder int) error
	UpdateType(ctx context.Context, owner, event, typeCode string, in models.TicketTypeInput) error
	SetTypeActive(ctx context.Context, owner, event, typeCode string, active bool) error
	DeleteType(ctx context.Context, owner, event, typeCode string) error
	ReorderTypes(ctx context.Context, owner, event string, codes []string) error
	// Attendees lists tickets with buyer + payment info.
	Attendees(ctx context.Context, owner, event, typeCode, attendance, payment, q string, limit, offset int) ([]models.AttendeeRow, error)
	AttendeeSummary(ctx context.Context, owner, event string) (models.AttendeeSummary, error)
}

type Repository struct {
	db      *sql.DB
	dialect repositories.Dialect
}

func NewRepository(db *sql.DB, dialect repositories.Dialect) RepositoryInterface {
	return &Repository{db: db, dialect: dialect}
}

func (r *Repository) eventTable() string { return r.dialect.Table("T_EVENT") }
func (r *Repository) catTable() string   { return r.dialect.Table("T_EVENT_CATEGORY") }
func (r *Repository) cityTable() string  { return r.dialect.Table("T_CITY") }
func (r *Repository) typeTable() string  { return r.dialect.Table("T_TICKET_TYPE") }
func (r *Repository) orderTable() string { return r.dialect.Table("T_ORDER") }
func (r *Repository) ticketTable() string { return r.dialect.Table("T_TICKET") }
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

func (r *Repository) soldSub() string {
	return `(SELECT COALESCE(SUM(o.QTY),0) FROM ` + r.orderTable() + ` o WHERE o.EVENT_CODE = e.CODE AND o.STATUS = 'PAID')`
}

func (r *Repository) quotaSub() string {
	return `(SELECT COALESCE(SUM(tt.QUOTA),0) FROM ` + r.typeTable() + ` tt WHERE tt.EVENT_CODE = e.CODE)`
}

func (r *Repository) revenueSub() string {
	return `(SELECT COALESCE(SUM(o.NET_AMOUNT),0) FROM ` + r.orderTable() + ` o WHERE o.EVENT_CODE = e.CODE AND o.STATUS = 'PAID')`
}

func (r *Repository) MineList(ctx context.Context, owner, status, category, city, from, toPlusOne, q, sort string, limit, offset int) ([]models.MyEventItem, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	args := []any{owner}
	cond := "e.ORGANIZER_CODE = ? AND e.IS_DELETED = 0"
	if status != "" {
		cond += " AND e.STATUS = ?"
		args = append(args, status)
	}
	if category != "" {
		cond += " AND e.EVENT_CATEGORY_CODE = ?"
		args = append(args, category)
	}
	if city != "" {
		cond += " AND e.CITY_CODE = ?"
		args = append(args, city)
	}
	if from != "" {
		cond += " AND e.START_AT >= ?"
		args = append(args, from)
	}
	if toPlusOne != "" {
		cond += " AND e.START_AT < ?"
		args = append(args, toPlusOne)
	}
	if q != "" {
		cond += ` AND e.TITLE LIKE ? ESCAPE '\'`
		args = append(args, "%"+escapeQ(q)+"%")
	}
	limitClause := r.pageClause(limit, offset)
	order := "e.START_AT ASC"
	switch sort {
	case "newest":
		order = "e.CREATED_AT DESC"
	case "popular":
		order = r.soldSub() + " DESC, e.START_AT ASC"
	}
	query := `SELECT e.CODE, COALESCE(e.SLUG, ''), e.TITLE, COALESCE(ec.NAME, ''), COALESCE(c.NAME, ''),
	  ` + r.startStr("e.START_AT") + `, e.STATUS, COALESCE(e.POSTER_URL, ''),
	  ` + r.soldSub() + `, ` + r.quotaSub() + `, ` + r.revenueSub() + `
	  FROM ` + r.eventTable() + ` e
	  LEFT JOIN ` + r.catTable() + ` ec ON ec.CODE = e.EVENT_CATEGORY_CODE
	  LEFT JOIN ` + r.cityTable() + ` c ON c.CODE = e.CITY_CODE
	  WHERE ` + cond + ` ORDER BY ` + order + limitClause
	rows, err := r.db.QueryContext(ctx, r.dialect.Bind(query), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.MyEventItem, 0)
	for rows.Next() {
		var m models.MyEventItem
		if err := rows.Scan(&m.Code, &m.Slug, &m.Title, &m.Category, &m.City,
			&m.StartAt, &m.Status, &m.PosterURL, &m.Sold, &m.Quota, &m.Revenue); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
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

func escapeQ(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `%`, `\%`)
	s = strings.ReplaceAll(s, `_`, `\_`)
	return s
}

func (r *Repository) EventByCode(ctx context.Context, owner, code string) (*models.EventDetailOwner, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	var d models.EventDetailOwner
	var desc, poster, terms, cancelReason, slug sql.NullString
	var catCode, cityCode sql.NullString
	var lat, lng sql.NullFloat64
	var publishedAt sql.NullString
	query := `SELECT e.CODE, COALESCE(e.SLUG, ''), e.TITLE,
	  COALESCE(e.EVENT_CATEGORY_CODE, ''), COALESCE(ec.NAME, ''),
	  COALESCE(e.DESCRIPTION, ''), COALESCE(e.CITY_CODE, ''), COALESCE(c.NAME, ''),
	  COALESCE(e.VENUE, ''), COALESCE(e.ADDRESS, ''), e.LATITUDE, e.LONGITUDE,
	  ` + r.startStr("e.START_AT") + `, COALESCE(` + r.startStr("e.END_AT") + `, ''),
	  COALESCE(e.POSTER_URL, ''), COALESCE(e.TERMS, ''), e.STATUS,
	  COALESCE(e.CANCEL_REASON, ''), ` + r.soldSub() + `,
	  COALESCE(` + r.startStr("e.PUBLISHED_AT") + `, '')
	  FROM ` + r.eventTable() + ` e
	  LEFT JOIN ` + r.catTable() + ` ec ON ec.CODE = e.EVENT_CATEGORY_CODE
	  LEFT JOIN ` + r.cityTable() + ` c ON c.CODE = e.CITY_CODE
	  WHERE e.ORGANIZER_CODE = ? AND e.CODE = ? AND e.IS_DELETED = 0`
	err := r.db.QueryRowContext(ctx, r.dialect.Bind(query), owner, code).Scan(
		&d.Code, &slug, &d.Title, &catCode, &d.Category, &desc, &cityCode, &d.City,
		&d.Venue, &d.Address, &lat, &lng, &d.StartAt, &d.EndAt, &poster, &terms,
		&d.Status, &cancelReason, &d.Sold, &publishedAt)
	if err != nil {
		return nil, err
	}
	d.Slug = slug.String
	d.CategoryCode = catCode.String
	d.Description = desc.String
	d.CityCode = cityCode.String
	d.PosterURL = poster.String
	d.Terms = terms.String
	d.CancelReason = cancelReason.String
	if lat.Valid {
		d.Latitude = lat.Float64
	}
	if lng.Valid {
		d.Longitude = lng.Float64
	}
	d.Missing = []string{}
	return &d, nil
}

func (r *Repository) EventStats(ctx context.Context, event string) (int, float64, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	var sold int
	var revenue float64
	query := `SELECT COALESCE(SUM(CASE WHEN o.STATUS = 'PAID' THEN o.QTY ELSE 0 END),0),
	  COALESCE(SUM(CASE WHEN o.STATUS = 'PAID' THEN o.NET_AMOUNT ELSE 0 END),0)
	  FROM ` + r.orderTable() + ` o WHERE o.EVENT_CODE = ?`
	err := r.db.QueryRowContext(ctx, r.dialect.Bind(query), event).Scan(&sold, &revenue)
	return sold, revenue, err
}

func (r *Repository) SlugExists(ctx context.Context, slug, exclude string) (bool, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	var n int
	query := `SELECT COUNT(*) FROM ` + r.eventTable() + ` WHERE SLUG = ? AND CODE <> ?`
	err := r.db.QueryRowContext(ctx, r.dialect.Bind(query), slug, exclude).Scan(&n)
	return n > 0, err
}

func (r *Repository) CreateEvent(ctx context.Context, owner, code, slug string, in models.EventInput) error {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	query := `INSERT INTO ` + r.eventTable() + ` (CODE, SLUG, ORGANIZER_CODE, TITLE,
	  EVENT_CATEGORY_CODE, DESCRIPTION, CITY_CODE, VENUE, ADDRESS, LATITUDE, LONGITUDE,
	  START_AT, END_AT, POSTER_URL, TERMS, STATUS)
	  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'DRAFT')`
	_, err := r.db.ExecContext(ctx, r.dialect.Bind(query), code, slug, owner, in.Title,
		repositories.NullStr(in.CategoryCode), in.Description,
		repositories.NullStr(in.CityCode), in.Venue, repositories.NullStr(in.Address),
		nullFloat(in.HasLatitude, in.Latitude), nullFloat(in.HasLongitude, in.Longitude),
		in.StartAt, nullStr(in.EndAt), repositories.NullStr(in.PosterURL), repositories.NullStr(in.Terms))
	return err
}

func nullStr(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func nullFloat(ok bool, v float64) any {
	if !ok {
		return nil
	}
	return v
}

func (r *Repository) UpdateEvent(ctx context.Context, owner, code string, in models.EventInput) error {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	query := `UPDATE ` + r.eventTable() + ` SET TITLE = ?, EVENT_CATEGORY_CODE = ?,
	  DESCRIPTION = ?, CITY_CODE = ?, VENUE = ?, ADDRESS = ?, LATITUDE = ?, LONGITUDE = ?,
	  START_AT = ?, END_AT = ?, POSTER_URL = ?, TERMS = ?,
	  UPDATED_AT = ` + r.dialect.Now() + `
	  WHERE ORGANIZER_CODE = ? AND CODE = ? AND IS_DELETED = 0`
	_, err := r.db.ExecContext(ctx, r.dialect.Bind(query), in.Title,
		repositories.NullStr(in.CategoryCode), in.Description,
		repositories.NullStr(in.CityCode), in.Venue, repositories.NullStr(in.Address),
		nullFloat(in.HasLatitude, in.Latitude), nullFloat(in.HasLongitude, in.Longitude),
		in.StartAt, nullStr(in.EndAt), repositories.NullStr(in.PosterURL),
		repositories.NullStr(in.Terms), owner, code)
	return err
}

func (r *Repository) SetStatus(ctx context.Context, owner, code, status, cancelReason string) error {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	now := r.dialect.Now()
	query := `UPDATE ` + r.eventTable() + ` SET STATUS = ?, CANCEL_REASON = ?,
	  PUBLISHED_AT = CASE WHEN ? = 'PUBLISHED' AND PUBLISHED_AT IS NULL THEN ` + now + ` ELSE PUBLISHED_AT END,
	  UPDATED_AT = ` + now + `
	  WHERE ORGANIZER_CODE = ? AND CODE = ? AND IS_DELETED = 0`
	_, err := r.db.ExecContext(ctx, r.dialect.Bind(query),
		status, repositories.NullStr(cancelReason), status, owner, code)
	return err
}

func (r *Repository) SoftDelete(ctx context.Context, owner, code string) error {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	query := `UPDATE ` + r.eventTable() + ` SET IS_DELETED = 1,
	  UPDATED_AT = ` + r.dialect.Now() + `
	  WHERE ORGANIZER_CODE = ? AND CODE = ? AND IS_DELETED = 0`
	_, err := r.db.ExecContext(ctx, r.dialect.Bind(query), owner, code)
	return err
}

func (r *Repository) MarkEnded(ctx context.Context, owner, now string) error {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	query := `UPDATE ` + r.eventTable() + ` SET STATUS = 'ENDED',
	  UPDATED_AT = ` + r.dialect.Now() + `
	  WHERE ORGANIZER_CODE = ? AND IS_DELETED = 0 AND STATUS = 'PUBLISHED'
	  AND END_AT IS NOT NULL AND END_AT < ?`
	_, err := r.db.ExecContext(ctx, r.dialect.Bind(query), owner, now)
	return err
}

// Duplicate copies an event + its ticket types as a new DRAFT (one tx).
func (r *Repository) Duplicate(ctx context.Context, owner, srcCode, newCode, newSlug string) error {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	query := `INSERT INTO ` + r.eventTable() + ` (CODE, SLUG, ORGANIZER_CODE, TITLE,
	  EVENT_CATEGORY_CODE, DESCRIPTION, CITY_CODE, VENUE, ADDRESS, LATITUDE, LONGITUDE,
	  START_AT, END_AT, POSTER_URL, TERMS, STATUS)
	  SELECT ?, ?, ORGANIZER_CODE, TITLE + ' (Copy)', EVENT_CATEGORY_CODE, DESCRIPTION,
	  CITY_CODE, VENUE, ADDRESS, LATITUDE, LONGITUDE, START_AT, END_AT, POSTER_URL, TERMS, 'DRAFT'
	  FROM ` + r.eventTable() + ` WHERE ORGANIZER_CODE = ? AND CODE = ? AND IS_DELETED = 0`
	res, err := tx.ExecContext(ctx, r.dialect.Bind(query), newCode, newSlug, owner, srcCode)
	if err != nil {
		return err
	}
	if n, err := res.RowsAffected(); err != nil || n == 0 {
		if err == nil {
			return sql.ErrNoRows
		}
		return err
	}
	return r.duplicateTypes(ctx, tx, owner, srcCode, newCode)
}

// duplicateTypes copies ticket types with fresh codes (concat per engine).
func (r *Repository) duplicateTypes(ctx context.Context, tx *sql.Tx, owner, srcCode, newCode string) error {
	types, err := r.sourceTypes(ctx, tx, owner, srcCode)
	if err != nil {
		return err
	}
	n := 0
	for _, t := range types {
		n++
		tcode := "TT-" + newCode + "-" + itoa(n)
		query := `INSERT INTO ` + r.typeTable() + ` (CODE, EVENT_CODE, NAME, DESCRIPTION,
		  PRICE, QUOTA, MAX_PER_PERSON, SALE_START_AT, SALE_END_AT, IS_ACTIVE, SORT_ORDER)
		  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
		if _, err := tx.ExecContext(ctx, r.dialect.Bind(query), tcode, newCode,
			t.Name, nullStr(t.Description), t.Price, t.Quota,
			nullableInt(t.MaxPerPerson, t.HasMax),
			nullableStr(t.SaleStartAt), nullableStr(t.SaleEndAt),
			t.IsActive, t.SortOrder); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// typeSeed is one ticket type row copied by Duplicate.
type typeSeed struct {
	Name         string
	Description  string
	Price        float64
	Quota        int
	MaxPerPerson int
	HasMax       bool
	SaleStartAt  string
	SaleEndAt    string
	IsActive     bool
	SortOrder    int
}

func nullableInt(v int, ok bool) any {
	if !ok {
		return nil
	}
	return v
}

func nullableStr(v string) any {
	if v == "" {
		return nil
	}
	return v
}

func (r *Repository) sourceTypes(ctx context.Context, tx *sql.Tx, owner, srcCode string) ([]typeSeed, error) {
	query := `SELECT tt.NAME, COALESCE(tt.DESCRIPTION, ''), tt.PRICE, tt.QUOTA,
	  tt.MAX_PER_PERSON, ` + r.startStr("tt.SALE_START_AT") + `, ` + r.startStr("tt.SALE_END_AT") + `,
	  tt.IS_ACTIVE, tt.SORT_ORDER
	  FROM ` + r.typeTable() + ` tt JOIN ` + r.eventTable() + ` e ON e.CODE = tt.EVENT_CODE
	  WHERE e.ORGANIZER_CODE = ? AND tt.EVENT_CODE = ? AND e.IS_DELETED = 0
	  ORDER BY tt.SORT_ORDER ASC, tt.CODE ASC`
	rows, err := tx.QueryContext(ctx, r.dialect.Bind(query), owner, srcCode)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]typeSeed, 0)
	for rows.Next() {
		// startStr(NULL) yields NULL: scan via NullString then unwrap.
		var saleFrom, saleTo sql.NullString
		var name, desc string
		var price float64
		var quota int
		var maxp sql.NullInt64
		var active bool
		var sort int
		if err := rows.Scan(&name, &desc, &price, &quota, &maxp, &saleFrom, &saleTo, &active, &sort); err != nil {
			return nil, err
		}
		t := typeSeed{Name: name, Description: desc, Price: price, Quota: quota,
			IsActive: active, SortOrder: sort}
		if maxp.Valid {
			t.MaxPerPerson, t.HasMax = int(maxp.Int64), true
		}
		if saleFrom.Valid {
			t.SaleStartAt = saleFrom.String
		}
		if saleTo.Valid {
			t.SaleEndAt = saleTo.String
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (r *Repository) AffectedBuyers(ctx context.Context, event string) (int, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	var n int
	query := `SELECT COUNT(DISTINCT BUYER_CODE) FROM ` + r.orderTable() + `
	  WHERE EVENT_CODE = ? AND STATUS = 'PAID'`
	err := r.db.QueryRowContext(ctx, r.dialect.Bind(query), event).Scan(&n)
	return n, err
}

func (r *Repository) CategoryExists(ctx context.Context, code string) (bool, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	var n int
	query := `SELECT COUNT(*) FROM ` + r.catTable() + ` WHERE CODE = ?`
	err := r.db.QueryRowContext(ctx, r.dialect.Bind(query), code).Scan(&n)
	return n > 0, err
}

// WasPublished reports whether the event was ever published.
func (r *Repository) WasPublished(ctx context.Context, owner, code string) (bool, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	var n int
	query := `SELECT COUNT(*) FROM ` + r.eventTable() + `
	  WHERE ORGANIZER_CODE = ? AND CODE = ? AND PUBLISHED_AT IS NOT NULL`
	err := r.db.QueryRowContext(ctx, r.dialect.Bind(query), owner, code).Scan(&n)
	return n > 0, err
}

func (r *Repository) CityExists(ctx context.Context, code string) (bool, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	var n int
	query := `SELECT COUNT(*) FROM ` + r.cityTable() + ` WHERE CODE = ?`
	err := r.db.QueryRowContext(ctx, r.dialect.Bind(query), code).Scan(&n)
	return n > 0, err
}

func (r *Repository) TypeList(ctx context.Context, owner, event string) ([]models.TicketTypeRow, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	query := `SELECT tt.CODE, tt.NAME, COALESCE(tt.DESCRIPTION, ''), tt.PRICE, tt.QUOTA,
	  COALESCE((SELECT SUM(o.QTY) FROM ` + r.orderTable() + ` o
	    WHERE o.TICKET_TYPE_CODE = tt.CODE AND o.STATUS = 'PAID'), 0),
	  COALESCE(tt.MAX_PER_PERSON, 0), COALESCE(` + r.startStr("tt.SALE_START_AT") + `, ''),
	  COALESCE(` + r.startStr("tt.SALE_END_AT") + `, ''), tt.IS_ACTIVE, tt.SORT_ORDER
	  FROM ` + r.typeTable() + ` tt JOIN ` + r.eventTable() + ` e ON e.CODE = tt.EVENT_CODE
	  WHERE e.ORGANIZER_CODE = ? AND tt.EVENT_CODE = ? AND e.IS_DELETED = 0
	  ORDER BY tt.SORT_ORDER ASC, tt.CODE ASC`
	rows, err := r.db.QueryContext(ctx, r.dialect.Bind(query), owner, event)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.TicketTypeRow, 0)
	for rows.Next() {
		var t models.TicketTypeRow
		var maxp int
		if err := rows.Scan(&t.Code, &t.Name, &t.Description, &t.Price, &t.Quota,
			&t.Sold, &maxp, &t.SaleStartAt, &t.SaleEndAt, &t.IsActive, &t.SortOrder); err != nil {
			return nil, err
		}
		t.MaxPerPerson = maxp
		t.Remaining = t.Quota - t.Sold
		if t.Remaining < 0 {
			t.Remaining = 0
		}
		t.Revenue = float64(t.Sold) * t.Price
		out = append(out, t)
	}
	return out, rows.Err()
}

func (r *Repository) TypeNameTaken(ctx context.Context, event, name, exclude string) (bool, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	var n int
	query := `SELECT COUNT(*) FROM ` + r.typeTable() + `
	  WHERE EVENT_CODE = ? AND LOWER(NAME) = LOWER(?) AND CODE <> ?`
	err := r.db.QueryRowContext(ctx, r.dialect.Bind(query), event, name, exclude).Scan(&n)
	return n > 0, err
}

func (r *Repository) TypeSold(ctx context.Context, typeCode string) (int, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	var n int
	query := `SELECT COALESCE(SUM(o.QTY),0) FROM ` + r.orderTable() + ` o
	  WHERE o.TICKET_TYPE_CODE = ? AND o.STATUS = 'PAID'`
	err := r.db.QueryRowContext(ctx, r.dialect.Bind(query), typeCode).Scan(&n)
	return n, err
}

func (r *Repository) NextTypeSort(ctx context.Context, event string) (int, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	var n sql.NullInt64
	query := `SELECT MAX(SORT_ORDER) FROM ` + r.typeTable() + ` WHERE EVENT_CODE = ?`
	if err := r.db.QueryRowContext(ctx, r.dialect.Bind(query), event).Scan(&n); err != nil {
		return 99, err
	}
	if !n.Valid {
		return 1, nil
	}
	return int(n.Int64) + 1, nil
}

func (r *Repository) CreateType(ctx context.Context, owner, event, code string, in models.TicketTypeInput, sortOrder int) error {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	// Ownership enforced via event join guard (row must belong to caller).
	if ok, err := r.ownsType(ctx, owner, event, ""); err != nil || !ok {
		if err == nil {
			return sql.ErrNoRows
		}
		return err
	}
	active := true
	if in.IsActive != nil {
		active = *in.IsActive
	}
	query := `INSERT INTO ` + r.typeTable() + ` (CODE, EVENT_CODE, NAME, DESCRIPTION,
	  PRICE, QUOTA, MAX_PER_PERSON, SALE_START_AT, SALE_END_AT, IS_ACTIVE, SORT_ORDER)
	  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, r.dialect.Bind(query), code, event, in.Name,
		repositories.NullStr(in.Description), in.Price, in.Quota,
		nullableIntPtr(in.MaxPerPerson), nullStr(in.SaleStartAt), nullStr(in.SaleEndAt),
		active, sortOrder)
	return err
}

func nullableIntPtr(v *int) any {
	if v == nil {
		return nil
	}
	return *v
}

// ownsType verifies an event (and optionally a type) belongs to the owner.
func (r *Repository) ownsType(ctx context.Context, owner, event, typeCode string) (bool, error) {
	var n int
	query := `SELECT COUNT(*) FROM ` + r.eventTable() + ` e
	  LEFT JOIN ` + r.typeTable() + ` tt ON tt.EVENT_CODE = e.CODE AND tt.CODE = ?
	  WHERE e.ORGANIZER_CODE = ? AND e.CODE = ? AND e.IS_DELETED = 0`
	err := r.db.QueryRowContext(ctx, r.dialect.Bind(query), typeCode, owner, event).Scan(&n)
	if err != nil {
		return false, err
	}
	if typeCode == "" {
		return n > 0, nil
	}
	// With a type code, the join must have matched (count the matched row).
	var m int
	query = `SELECT COUNT(*) FROM ` + r.typeTable() + ` tt
	  JOIN ` + r.eventTable() + ` e ON e.CODE = tt.EVENT_CODE
	  WHERE e.ORGANIZER_CODE = ? AND tt.EVENT_CODE = ? AND tt.CODE = ? AND e.IS_DELETED = 0`
	if err := r.db.QueryRowContext(ctx, r.dialect.Bind(query), owner, event, typeCode).Scan(&m); err != nil {
		return false, err
	}
	return m > 0, nil
}

func (r *Repository) UpdateType(ctx context.Context, owner, event, typeCode string, in models.TicketTypeInput) error {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	if ok, err := r.ownsType(ctx, owner, event, typeCode); err != nil || !ok {
		if err == nil {
			return sql.ErrNoRows
		}
		return err
	}
	query := `UPDATE ` + r.typeTable() + ` SET NAME = ?, DESCRIPTION = ?, PRICE = ?,
	  QUOTA = ?, MAX_PER_PERSON = ?, SALE_START_AT = ?, SALE_END_AT = ?,
	  UPDATED_AT = ` + r.dialect.Now() + ` WHERE CODE = ? AND EVENT_CODE = ?`
	res, err := r.db.ExecContext(ctx, r.dialect.Bind(query), in.Name,
		repositories.NullStr(in.Description), in.Price, in.Quota,
		nullableIntPtr(in.MaxPerPerson), nullStr(in.SaleStartAt), nullStr(in.SaleEndAt),
		typeCode, event)
	if err != nil {
		return err
	}
	return checkRows(res)
}

func checkRows(res sql.Result) error {
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *Repository) SetTypeActive(ctx context.Context, owner, event, typeCode string, active bool) error {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	if ok, err := r.ownsType(ctx, owner, event, typeCode); err != nil || !ok {
		if err == nil {
			return sql.ErrNoRows
		}
		return err
	}
	query := `UPDATE ` + r.typeTable() + ` SET IS_ACTIVE = ?,
	  UPDATED_AT = ` + r.dialect.Now() + ` WHERE CODE = ? AND EVENT_CODE = ?`
	res, err := r.db.ExecContext(ctx, r.dialect.Bind(query), active, typeCode, event)
	if err != nil {
		return err
	}
	return checkRows(res)
}

func (r *Repository) DeleteType(ctx context.Context, owner, event, typeCode string) error {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	if ok, err := r.ownsType(ctx, owner, event, typeCode); err != nil || !ok {
		if err == nil {
			return sql.ErrNoRows
		}
		return err
	}
	query := `DELETE FROM ` + r.typeTable() + ` WHERE CODE = ? AND EVENT_CODE = ?`
	res, err := r.db.ExecContext(ctx, r.dialect.Bind(query), typeCode, event)
	if err != nil {
		return err
	}
	return checkRows(res)
}

func (r *Repository) ReorderTypes(ctx context.Context, owner, event string, codes []string) error {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for i, c := range codes {
		query := `UPDATE ` + r.typeTable() + ` SET SORT_ORDER = ?,
		  UPDATED_AT = ` + r.dialect.Now() + `
		  WHERE CODE = ? AND EVENT_CODE = ?
		  AND EXISTS (SELECT 1 FROM ` + r.eventTable() + ` e
		    WHERE e.CODE = ? AND e.ORGANIZER_CODE = ? AND e.IS_DELETED = 0)`
		res, err := tx.ExecContext(ctx, r.dialect.Bind(query), i+1, c, event, event, owner)
		if err != nil {
			return err
		}
		if n, err := res.RowsAffected(); err != nil || n == 0 {
			if err == nil {
				return sql.ErrNoRows
			}
			return err
		}
	}
	return tx.Commit()
}

func (r *Repository) Attendees(ctx context.Context, owner, event, typeCode, attendance, payment, q string, limit, offset int) ([]models.AttendeeRow, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	args := []any{owner, event}
	cond := `e.ORGANIZER_CODE = ? AND t.ORDER_CODE IN
	  (SELECT o.CODE FROM ` + r.orderTable() + ` o WHERE o.EVENT_CODE = ?)`
	if typeCode != "" {
		cond += " AND o.TICKET_TYPE_CODE = ?"
		args = append(args, typeCode)
	}
	if payment != "" {
		cond += " AND o.STATUS = ?"
		args = append(args, payment)
	}
	if attendance == "HADIR" {
		cond += " AND t.STATUS = 'USED'"
	} else if attendance == "BELUM" {
		cond += " AND t.STATUS <> 'USED'"
	}
	if q != "" {
		cond += ` AND (t.HOLDER_NAME LIKE ? ESCAPE '\' OR t.EMAIL LIKE ? ESCAPE '\'
		  OR t.QR_CODE LIKE ? ESCAPE '\' OR u.USERNAME LIKE ? ESCAPE '\')`
		like := "%" + escapeQ(q) + "%"
		args = append(args, like, like, like, like)
	}
	limitClause := r.pageClause(limit, offset)
	query := `SELECT t.QR_CODE, COALESCE(t.HOLDER_NAME, u.USERNAME), u.USERNAME,
	  COALESCE(t.EMAIL, u.EMAIL), COALESCE(t.PHONE, ''), tt.NAME, o.CODE, o.STATUS,
	  CASE WHEN t.STATUS = 'USED' THEN 'HADIR' ELSE 'BELUM' END,
	  COALESCE(` + r.startStr("t.CHECKED_IN_AT") + `, ''), ` + r.startStr("o.ORDER_DATE") + `
	  FROM ` + r.ticketTable() + ` t
	  JOIN ` + r.orderTable() + ` o ON o.CODE = t.ORDER_CODE
	  JOIN ` + r.eventTable() + ` e ON e.CODE = o.EVENT_CODE
	  JOIN ` + r.typeTable() + ` tt ON tt.CODE = o.TICKET_TYPE_CODE
	  JOIN ` + r.userTable() + ` u ON u.CODE = o.BUYER_CODE
	  WHERE ` + cond + ` ORDER BY o.ORDER_DATE DESC` + limitClause
	rows, err := r.db.QueryContext(ctx, r.dialect.Bind(query), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.AttendeeRow, 0)
	for rows.Next() {
		var a models.AttendeeRow
		if err := rows.Scan(&a.TicketCode, &a.BuyerName, &a.BuyerUsername, &a.Email,
			&a.Phone, &a.TicketType, &a.OrderCode, &a.PaymentStatus, &a.Attendance,
			&a.CheckedInAt, &a.PurchaseDate); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (r *Repository) AttendeeSummary(ctx context.Context, owner, event string) (models.AttendeeSummary, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	var s models.AttendeeSummary
	query := `SELECT COUNT(*),
	  COALESCE(SUM(CASE WHEN t.STATUS = 'USED' THEN 1 ELSE 0 END),0)
	  FROM ` + r.ticketTable() + ` t
	  JOIN ` + r.orderTable() + ` o ON o.CODE = t.ORDER_CODE
	  JOIN ` + r.eventTable() + ` e ON e.CODE = o.EVENT_CODE
	  WHERE e.ORGANIZER_CODE = ? AND o.EVENT_CODE = ? AND o.STATUS = 'PAID'`
	err := r.db.QueryRowContext(ctx, r.dialect.Bind(query), owner, event).Scan(&s.Total, &s.CheckedIn)
	if err != nil {
		return s, err
	}
	s.Pending = s.Total - s.CheckedIn
	if s.Total > 0 {
		s.Rate = float64(s.CheckedIn) / float64(s.Total) * 100
	}
	return s, nil
}
