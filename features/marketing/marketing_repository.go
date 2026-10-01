package marketing

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"golang-backend/models"
	"golang-backend/repositories"
	"golang-backend/utils"
)

type RepositoryInterface interface {
	ListPromos(ctx context.Context, organizer, statusFilter, q string, limit, offset int) ([]models.PromoRow, error)
	PromoByCode(ctx context.Context, organizer, code string) (*models.PromoRow, error)
	PromoExists(ctx context.Context, promoCode, exclude string) (bool, error)
	CreatePromo(ctx context.Context, organizer, code string, in models.PromoInput) error
	UpdatePromo(ctx context.Context, organizer, code string, in models.PromoInput) error
	SetPromoStatus(ctx context.Context, organizer, code, status string) error
	PromoUsed(ctx context.Context, code string) (bool, error)
	DeletePromo(ctx context.Context, organizer, code string) error
	PromoUsage(ctx context.Context, code string, limit, offset int) ([]models.PromoUsageRow, error)
	UpsertPromoScopes(ctx context.Context, promoCode string, scopeType string, cats, events []string) error
	ListPromoScopes(ctx context.Context, promoCode string) (cats []string, events []string, err error)
	ListPositions(ctx context.Context) ([]models.AdPosition, error)
	PositionByCode(ctx context.Context, code string) (*models.AdPosition, error)
	ListAds(ctx context.Context, organizer, status, q string, limit, offset int) ([]models.AdCampaignRow, error)
	AdByCode(ctx context.Context, organizer, code string) (*models.AdCampaignRow, error)
	CreateAd(ctx context.Context, organizer, code string, in models.AdCampaignInput, cost float64) error
	UpdateAd(ctx context.Context, organizer, code string, in models.AdCampaignInput, cost float64) error
	SetAdStatus(ctx context.Context, organizer, code, status, pay, approval, note string) error
	AdStats(ctx context.Context, code string) ([]models.AdDailyPoint, error)
	ListNews(ctx context.Context, status, category, q string, limit, offset int) ([]models.NewsRow, error)
	NewsByCode(ctx context.Context, code string) (*models.NewsRow, error)
	NewsSlugExists(ctx context.Context, slug, exclude string) (bool, error)
	CreateNews(ctx context.Context, author, code, slug string, in models.NewsInput) error
	UpdateNews(ctx context.Context, code, slug string, in models.NewsInput) error
	SetNewsStatus(ctx context.Context, code, status string) error
	SoftDeleteNews(ctx context.Context, code string) error
	PublishScheduled(ctx context.Context, now string) error
	NewsCategories(ctx context.Context) ([]models.NewsCategory, error)
	EventCategories(ctx context.Context) ([]models.EventCategory, error)
	OwnedEvents(ctx context.Context, organizer string) ([]OwnedEvent, error)
	CategoryExists(ctx context.Context, code string) (bool, error)
	EventOwned(ctx context.Context, organizer, event string) (bool, error)
}

type OwnedEvent struct {
	Code  string `json:"code"`
	Title string `json:"title"`
}

type Repository struct {
	db      *sql.DB
	dialect repositories.Dialect
}

func NewRepository(db *sql.DB, dialect repositories.Dialect) RepositoryInterface {
	return &Repository{db: db, dialect: dialect}
}

func (r *Repository) promoTable() string  { return r.dialect.Table("T_PROMO") }
func (r *Repository) scopeTable() string  { return r.dialect.Table("T_PROMO_SCOPE") }
func (r *Repository) usageTable() string  { return r.dialect.Table("T_PROMO_USAGE") }
func (r *Repository) adTable() string     { return r.dialect.Table("T_AD_CAMPAIGN") }
func (r *Repository) adPosTable() string  { return r.dialect.Table("T_AD_POSITION") }
func (r *Repository) adStatTable() string { return r.dialect.Table("T_AD_STAT_DAILY") }
func (r *Repository) adEvTable() string   { return r.dialect.Table("T_AD_EVENT") }
func (r *Repository) newsTable() string   { return r.dialect.Table("T_NEWS") }
func (r *Repository) ncatTable() string   { return r.dialect.Table("T_NEWS_CATEGORY") }
func (r *Repository) ecatTable() string   { return r.dialect.Table("T_EVENT_CATEGORY") }
func (r *Repository) eventTable() string  { return r.dialect.Table("T_EVENT") }
func (r *Repository) userTable() string {
	if r.dialect == repositories.DialectMSSQL {
		return "dbo.CPUSER"
	}
	return "CPUSER"
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

func (r *Repository) pageClause(limit, offset int) string {
	if limit <= 0 || limit > 200 {
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

func promoDisplay(m models.PromoRow) string {
	if strings.EqualFold(m.Status, models.PromoStatusDisabled) {
		return "nonaktif"
	}
	now := time.Now().Format("2006-01-02 15:04:05")
	if m.ValidTo != "" && m.ValidTo < now {
		return "kedaluwarsa"
	}
	if m.MaxUses != nil && m.UsedCount >= *m.MaxUses {
		return "kuota_habis"
	}
	if m.ValidFrom != "" && m.ValidFrom > now {
		return "akan_datang"
	}
	return "aktif"
}

func adDisplay(m models.AdCampaignRow) string {
	s := strings.ToUpper(m.Status)
	switch s {
	case models.AdStatusDraft:
		return "draft"
	case models.AdStatusPendingPayment:
		return "menunggu_pembayaran"
	case models.AdStatusPendingApproval:
		return "menunggu_persetujuan"
	case models.AdStatusLive, "ACTIVE":
		return "tayang"
	case models.AdStatusPaused:
		return "dijeda"
	case models.AdStatusEnded:
		return "selesai"
	case models.AdStatusRejected:
		return "ditolak"
	default:
		return strings.ToLower(s)
	}
}

func daysBetween(a, b string) int {
	if a == "" || b == "" {
		return 0
	}
	ta, err1 := time.Parse("2006-01-02 15:04:05", a)
	tb, err2 := time.Parse("2006-01-02 15:04:05", b)
	if err1 != nil || err2 != nil {
		ta2, e1 := time.Parse("2006-01-02 15:04", a)
		tb2, e2 := time.Parse("2006-01-02 15:04", b)
		if e1 != nil || e2 != nil {
			return 0
		}
		ta, tb = ta2, tb2
	}
	d := int(tb.Sub(ta).Hours() / 24)
	if d < 1 {
		return 1
	}
	return d + 1
}

func (r *Repository) ListPromos(ctx context.Context, organizer, statusFilter, q string, limit, offset int) ([]models.PromoRow, error) {
	ctx2, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	fetchLimit := limit
	if statusFilter != "" {
		fetchLimit = 200
	}
	args := []any{organizer}
	cond := "p.ORGANIZER_CODE = ?"
	if strings.TrimSpace(q) != "" {
		cond += " AND p.PROMO_CODE LIKE ? ESCAPE '\\'"
		args = append(args, "%"+escapeQ(strings.ToUpper(strings.TrimSpace(q)))+"%")
	}
	query := `SELECT p.CODE, p.PROMO_CODE, COALESCE(p.DESCRIPTION, ''), p.DISCOUNT_PERCENT, p.DISCOUNT_AMOUNT,
	  p.MAX_DISCOUNT_AMOUNT, p.MIN_PURCHASE, p.MAX_USES, p.MAX_PER_USER, p.USED_COUNT,
	  ` + r.startStr("p.VALID_FROM") + `, ` + r.startStr("p.VALID_TO") + `, p.STATUS, p.SCOPE_TYPE, p.ORGANIZER_CODE,
	  ` + r.startStr("p.CREATED_AT") + `, ` + r.startStr("p.UPDATED_AT") + `
	  FROM ` + r.promoTable() + ` p WHERE ` + cond + ` ORDER BY p.CREATED_AT DESC` + r.pageClause(fetchLimit, 0)
	rows, err := r.db.QueryContext(ctx2, r.dialect.Bind(query), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.PromoRow, 0)
	for rows.Next() {
		var m models.PromoRow
		var maxUses, maxPerUser sql.NullInt64
		if err := rows.Scan(&m.Code, &m.PromoCode, &m.Description, &m.DiscountPercent, &m.DiscountAmount,
			&m.MaxDiscount, &m.MinPurchase, &maxUses, &maxPerUser, &m.UsedCount,
			&m.ValidFrom, &m.ValidTo, &m.Status, &m.ScopeType, &m.OrganizerCode,
			&m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, err
		}
		if maxUses.Valid {
			v := int(maxUses.Int64)
			m.MaxUses = &v
		}
		if maxPerUser.Valid {
			v := int(maxPerUser.Int64)
			m.MaxPerUser = &v
		}
		m.Remaining = -1
		if m.MaxUses != nil {
			m.Remaining = *m.MaxUses - m.UsedCount
			if m.Remaining < 0 {
				m.Remaining = 0
			}
		}
		m.DisplayStatus = promoDisplay(m)
		if statusFilter != "" && !strings.EqualFold(m.DisplayStatus, statusFilter) {
			continue
		}
		out = append(out, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for i := range out {
		cats, evs, _ := r.ListPromoScopes(ctx, out[i].Code)
		out[i].CategoryCodes = cats
		out[i].EventCodes = evs
	}
	if statusFilter != "" {
		start := offset
		if start > len(out) {
			start = len(out)
		}
		end := start + limit
		if end > len(out) {
			end = len(out)
		}
		return out[start:end], nil
	}
	if offset > 0 {
		if offset >= len(out) {
			return []models.PromoRow{}, nil
		}
		out = out[offset:]
		if len(out) > limit {
			out = out[:limit]
		}
	} else if len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

func (r *Repository) PromoByCode(ctx context.Context, organizer, code string) (*models.PromoRow, error) {
	ctx2, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	var m models.PromoRow
	var maxUses, maxPerUser sql.NullInt64
	query := `SELECT p.CODE, p.PROMO_CODE, COALESCE(p.DESCRIPTION, ''), p.DISCOUNT_PERCENT, p.DISCOUNT_AMOUNT,
	  p.MAX_DISCOUNT_AMOUNT, p.MIN_PURCHASE, p.MAX_USES, p.MAX_PER_USER, p.USED_COUNT,
	  ` + r.startStr("p.VALID_FROM") + `, ` + r.startStr("p.VALID_TO") + `, p.STATUS, p.SCOPE_TYPE, p.ORGANIZER_CODE,
	  ` + r.startStr("p.CREATED_AT") + `, ` + r.startStr("p.UPDATED_AT") + `
	  FROM ` + r.promoTable() + ` p WHERE p.ORGANIZER_CODE = ? AND p.CODE = ?`
	err := r.db.QueryRowContext(ctx2, r.dialect.Bind(query), organizer, code).Scan(
		&m.Code, &m.PromoCode, &m.Description, &m.DiscountPercent, &m.DiscountAmount,
		&m.MaxDiscount, &m.MinPurchase, &maxUses, &maxPerUser, &m.UsedCount,
		&m.ValidFrom, &m.ValidTo, &m.Status, &m.ScopeType, &m.OrganizerCode,
		&m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		return nil, err
	}
	if maxUses.Valid {
		v := int(maxUses.Int64)
		m.MaxUses = &v
	}
	if maxPerUser.Valid {
		v := int(maxPerUser.Int64)
		m.MaxPerUser = &v
	}
	m.Remaining = -1
	if m.MaxUses != nil {
		m.Remaining = *m.MaxUses - m.UsedCount
	}
	m.DisplayStatus = promoDisplay(m)
	cats, evs, _ := r.ListPromoScopes(ctx, m.Code)
	m.CategoryCodes = cats
	m.EventCodes = evs
	return &m, nil
}

func (r *Repository) PromoExists(ctx context.Context, promoCode, exclude string) (bool, error) {
	ctx2, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	var n int
	query := `SELECT COUNT(*) FROM ` + r.promoTable() + ` WHERE PROMO_CODE = ? AND CODE <> ?`
	err := r.db.QueryRowContext(ctx2, r.dialect.Bind(query), strings.ToUpper(strings.TrimSpace(promoCode)), exclude).Scan(&n)
	return n > 0, err
}

func (r *Repository) CreatePromo(ctx context.Context, organizer, code string, in models.PromoInput) error {
	ctx2, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	query := `INSERT INTO ` + r.promoTable() + ` (CODE, PROMO_CODE, DESCRIPTION, DISCOUNT_PERCENT, DISCOUNT_AMOUNT,
	  MAX_DISCOUNT_AMOUNT, MIN_PURCHASE, MAX_USES, MAX_PER_USER, VALID_FROM, VALID_TO, STATUS, SCOPE_TYPE, ORGANIZER_CODE, USED_COUNT)
	  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'ACTIVE', ?, ?, 0)`
	_, err := r.db.ExecContext(ctx2, r.dialect.Bind(query), code, strings.ToUpper(strings.TrimSpace(in.PromoCode)),
		repositories.NullStr(strings.TrimSpace(in.Description)), in.DiscountPercent, in.DiscountAmount,
		in.MaxDiscount, in.MinPurchase, nullableInt(in.MaxUses), nullableInt(in.MaxPerUser),
		in.ValidFrom, in.ValidTo, strings.ToUpper(strings.TrimSpace(in.ScopeType)), organizer)
	if err != nil {
		return err
	}
	return r.UpsertPromoScopes(ctx, code, strings.ToUpper(strings.TrimSpace(in.ScopeType)), in.CategoryCodes, in.EventCodes)
}

func nullableInt(v *int) any {
	if v == nil {
		return nil
	}
	return *v
}

func (r *Repository) UpdatePromo(ctx context.Context, organizer, code string, in models.PromoInput) error {
	ctx2, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	query := `UPDATE ` + r.promoTable() + ` SET PROMO_CODE = ?, DESCRIPTION = ?, DISCOUNT_PERCENT = ?, DISCOUNT_AMOUNT = ?,
	  MAX_DISCOUNT_AMOUNT = ?, MIN_PURCHASE = ?, MAX_USES = ?, MAX_PER_USER = ?, VALID_FROM = ?, VALID_TO = ?, SCOPE_TYPE = ?,
	  UPDATED_AT = ` + r.dialect.Now() + ` WHERE ORGANIZER_CODE = ? AND CODE = ?`
	_, err := r.db.ExecContext(ctx2, r.dialect.Bind(query), strings.ToUpper(strings.TrimSpace(in.PromoCode)),
		repositories.NullStr(strings.TrimSpace(in.Description)), in.DiscountPercent, in.DiscountAmount,
		in.MaxDiscount, in.MinPurchase, nullableInt(in.MaxUses), nullableInt(in.MaxPerUser),
		in.ValidFrom, in.ValidTo, strings.ToUpper(strings.TrimSpace(in.ScopeType)), organizer, code)
	if err != nil {
		return err
	}
	return r.UpsertPromoScopes(ctx, code, strings.ToUpper(strings.TrimSpace(in.ScopeType)), in.CategoryCodes, in.EventCodes)
}

func (r *Repository) SetPromoStatus(ctx context.Context, organizer, code, status string) error {
	ctx2, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	query := `UPDATE ` + r.promoTable() + ` SET STATUS = ?, UPDATED_AT = ` + r.dialect.Now() + ` WHERE ORGANIZER_CODE = ? AND CODE = ?`
	res, err := r.db.ExecContext(ctx2, r.dialect.Bind(query), status, organizer, code)
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

func (r *Repository) PromoUsed(ctx context.Context, code string) (bool, error) {
	ctx2, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	var n int
	query := `SELECT COUNT(*) FROM ` + r.usageTable() + ` WHERE PROMO_CODE = ?`
	err := r.db.QueryRowContext(ctx2, r.dialect.Bind(query), code).Scan(&n)
	return n > 0, err
}

func (r *Repository) DeletePromo(ctx context.Context, organizer, code string) error {
	ctx2, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	query := `DELETE FROM ` + r.promoTable() + ` WHERE ORGANIZER_CODE = ? AND CODE = ?`
	res, err := r.db.ExecContext(ctx2, r.dialect.Bind(query), organizer, code)
	if err != nil {
		return err
	}
	return checkRows(res)
}

func (r *Repository) PromoUsage(ctx context.Context, code string, limit, offset int) ([]models.PromoUsageRow, error) {
	ctx2, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	query := `SELECT pu.CODE, pu.ORDER_CODE, pu.USER_CODE, COALESCE(u.USERNAME, pu.USER_CODE), pu.DISCOUNT_AMOUNT,
	  ` + r.startStr("pu.USED_AT") + ` FROM ` + r.usageTable() + ` pu
	  LEFT JOIN ` + r.userTable() + ` u ON u.CODE = pu.USER_CODE
	  WHERE pu.PROMO_CODE = ? ORDER BY pu.USED_AT DESC` + r.pageClause(limit, offset)
	rows, err := r.db.QueryContext(ctx2, r.dialect.Bind(query), code)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.PromoUsageRow, 0)
	for rows.Next() {
		var m models.PromoUsageRow
		if err := rows.Scan(&m.Code, &m.OrderCode, &m.UserCode, &m.Username, &m.Discount, &m.UsedAt); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (r *Repository) UpsertPromoScopes(ctx context.Context, promoCode, scopeType string, cats, events []string) error {
	ctx2, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	if _, err := r.db.ExecContext(ctx2, r.dialect.Bind(`DELETE FROM `+r.scopeTable()+` WHERE PROMO_CODE = ?`), promoCode); err != nil {
		return err
	}
	scopeType = strings.ToUpper(strings.TrimSpace(scopeType))
	if scopeType == models.PromoScopeAll || scopeType == "" {
		return nil
	}
	for _, c := range cats {
		c = strings.TrimSpace(strings.ToUpper(c))
		if c == "" {
			continue
		}
		pc, _ := utils.GenerateCode("PSC-")
		if _, err := r.db.ExecContext(ctx2, r.dialect.Bind(`INSERT INTO `+r.scopeTable()+` (CODE, PROMO_CODE, SCOPE_TYPE, CATEGORY_CODE) VALUES (?, ?, 'CATEGORY', ?)`), pc, promoCode, c); err != nil {
			return err
		}
	}
	for _, e := range events {
		e = strings.TrimSpace(e)
		if e == "" {
			continue
		}
		pc, _ := utils.GenerateCode("PSC-")
		if _, err := r.db.ExecContext(ctx2, r.dialect.Bind(`INSERT INTO `+r.scopeTable()+` (CODE, PROMO_CODE, SCOPE_TYPE, EVENT_CODE) VALUES (?, ?, 'EVENT', ?)`), pc, promoCode, e); err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) ListPromoScopes(ctx context.Context, promoCode string) (cats []string, events []string, err error) {
	ctx2, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	rows, e := r.db.QueryContext(ctx2, r.dialect.Bind(`SELECT SCOPE_TYPE, COALESCE(CATEGORY_CODE,''), COALESCE(EVENT_CODE,'') FROM `+r.scopeTable()+` WHERE PROMO_CODE = ?`), promoCode)
	if e != nil {
		return nil, nil, e
	}
	defer rows.Close()
	cats = []string{}
	events = []string{}
	for rows.Next() {
		var t, c, ev string
		if err := rows.Scan(&t, &c, &ev); err != nil {
			return nil, nil, err
		}
		if t == "CATEGORY" && c != "" {
			cats = append(cats, c)
		}
		if t == "EVENT" && ev != "" {
			events = append(events, ev)
		}
	}
	return cats, events, rows.Err()
}

func (r *Repository) ListPositions(ctx context.Context) ([]models.AdPosition, error) {
	ctx2, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	rows, err := r.db.QueryContext(ctx2, `SELECT CODE, NAME, PRICE_PER_DAY, SORT_ORDER FROM `+r.adPosTable()+` ORDER BY SORT_ORDER ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.AdPosition, 0)
	for rows.Next() {
		var p models.AdPosition
		if err := rows.Scan(&p.Code, &p.Name, &p.PricePerDay, &p.SortOrder); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *Repository) PositionByCode(ctx context.Context, code string) (*models.AdPosition, error) {
	ctx2, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	var p models.AdPosition
	query := `SELECT CODE, NAME, PRICE_PER_DAY, SORT_ORDER FROM ` + r.adPosTable() + ` WHERE CODE = ?`
	err := r.db.QueryRowContext(ctx2, r.dialect.Bind(query), strings.ToUpper(strings.TrimSpace(code))).Scan(&p.Code, &p.Name, &p.PricePerDay, &p.SortOrder)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *Repository) ListAds(ctx context.Context, organizer, status, q string, limit, offset int) ([]models.AdCampaignRow, error) {
	ctx2, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	args := []any{organizer}
	cond := "a.ORGANIZER_CODE = ?"
	if q = strings.TrimSpace(q); q != "" {
		cond += " AND a.TITLE LIKE ? ESCAPE '\\'"
		args = append(args, "%"+escapeQ(q)+"%")
	}
	fetchLimit := limit
	if status != "" {
		fetchLimit = 200
	}
	query := `SELECT a.CODE, a.TITLE, COALESCE(a.IMAGE_URL,''), COALESCE(a.TARGET_URL,''), a.POSITION,
	  COALESCE(ap.NAME, a.POSITION), ` + r.startStr("a.START_AT") + `, ` + r.startStr("a.END_AT") + `,
	  a.STATUS, a.PAYMENT_STATUS, a.APPROVAL_STATUS, a.COST, COALESCE(ap.PRICE_PER_DAY,0),
	  COALESCE(a.REJECT_NOTE,''), ` + r.startStr("a.CREATED_AT") + `, ` + r.startStr("a.UPDATED_AT") + `, a.ORGANIZER_CODE
	  FROM ` + r.adTable() + ` a LEFT JOIN ` + r.adPosTable() + ` ap ON ap.CODE = a.POSITION
	  WHERE ` + cond + ` ORDER BY a.CREATED_AT DESC` + r.pageClause(fetchLimit, 0)
	rows, err := r.db.QueryContext(ctx2, r.dialect.Bind(query), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.AdCampaignRow, 0)
	for rows.Next() {
		var m models.AdCampaignRow
		var price float64
		if err := rows.Scan(&m.Code, &m.Title, &m.ImageURL, &m.TargetURL, &m.Position,
			&m.PositionName, &m.StartAt, &m.EndAt, &m.Status, &m.PaymentStatus, &m.ApprovalStatus, &m.Cost, &price,
			&m.RejectNote, &m.CreatedAt, &m.UpdatedAt, &m.OrganizerCode); err != nil {
			return nil, err
		}
		m.PricePerDay = price
		m.DisplayStatus = adDisplay(m)
		if status != "" && !strings.EqualFold(m.DisplayStatus, status) {
			continue
		}
		imp, clk := r.adTotals(ctx, m.Code)
		m.Impressions = imp
		m.Clicks = clk
		if imp > 0 {
			m.CTR = float64(clk) / float64(imp) * 100
		}
		m.Days = daysBetween(m.StartAt, m.EndAt)
		out = append(out, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if status != "" {
		start := offset
		if start > len(out) {
			start = len(out)
		}
		end := start + limit
		if end > len(out) {
			end = len(out)
		}
		return out[start:end], nil
	}
	if offset > 0 {
		if offset >= len(out) {
			return []models.AdCampaignRow{}, nil
		}
		out = out[offset:]
		if len(out) > limit {
			out = out[:limit]
		}
	} else if len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

func (r *Repository) adTotals(ctx context.Context, code string) (int, int) {
	ctx2, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	var imp, clk sql.NullInt64
	_ = r.db.QueryRowContext(ctx2, r.dialect.Bind(`SELECT COALESCE(SUM(IMPRESSIONS),0), COALESCE(SUM(CLICKS),0) FROM `+r.adStatTable()+` WHERE CAMPAIGN_CODE = ?`), code).Scan(&imp, &clk)
	var evImp, evClk sql.NullInt64
	_ = r.db.QueryRowContext(ctx2, r.dialect.Bind(`SELECT COUNT(CASE WHEN KIND='IMPRESSION' THEN 1 END), COUNT(CASE WHEN KIND='CLICK' THEN 1 END) FROM `+r.adEvTable()+` WHERE CAMPAIGN_CODE = ?`), code).Scan(&evImp, &evClk)
	totalImp, totalClk := 0, 0
	if imp.Valid {
		totalImp += int(imp.Int64)
	}
	if evImp.Valid {
		totalImp += int(evImp.Int64)
	}
	if clk.Valid {
		totalClk += int(clk.Int64)
	}
	if evClk.Valid {
		totalClk += int(evClk.Int64)
	}
	return totalImp, totalClk
}

func (r *Repository) AdByCode(ctx context.Context, organizer, code string) (*models.AdCampaignRow, error) {
	ctx2, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	var m models.AdCampaignRow
	var price float64
	query := `SELECT a.CODE, a.TITLE, COALESCE(a.IMAGE_URL,''), COALESCE(a.TARGET_URL,''), a.POSITION,
	  COALESCE(ap.NAME, a.POSITION), ` + r.startStr("a.START_AT") + `, ` + r.startStr("a.END_AT") + `,
	  a.STATUS, a.PAYMENT_STATUS, a.APPROVAL_STATUS, a.COST, COALESCE(ap.PRICE_PER_DAY,0),
	  COALESCE(a.REJECT_NOTE,''), ` + r.startStr("a.CREATED_AT") + `, ` + r.startStr("a.UPDATED_AT") + `, a.ORGANIZER_CODE
	  FROM ` + r.adTable() + ` a LEFT JOIN ` + r.adPosTable() + ` ap ON ap.CODE = a.POSITION
	  WHERE a.ORGANIZER_CODE = ? AND a.CODE = ?`
	err := r.db.QueryRowContext(ctx2, r.dialect.Bind(query), organizer, code).Scan(
		&m.Code, &m.Title, &m.ImageURL, &m.TargetURL, &m.Position, &m.PositionName,
		&m.StartAt, &m.EndAt, &m.Status, &m.PaymentStatus, &m.ApprovalStatus, &m.Cost, &price,
		&m.RejectNote, &m.CreatedAt, &m.UpdatedAt, &m.OrganizerCode)
	if err != nil {
		return nil, err
	}
	m.PricePerDay = price
	m.DisplayStatus = adDisplay(m)
	imp, clk := r.adTotals(ctx, m.Code)
	m.Impressions = imp
	m.Clicks = clk
	if imp > 0 {
		m.CTR = float64(clk) / float64(imp) * 100
	}
	m.Days = daysBetween(m.StartAt, m.EndAt)
	return &m, nil
}

func (r *Repository) CreateAd(ctx context.Context, organizer, code string, in models.AdCampaignInput, cost float64) error {
	ctx2, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	query := `INSERT INTO ` + r.adTable() + ` (CODE, TITLE, IMAGE_URL, TARGET_URL, POSITION, START_AT, END_AT, STATUS, PAYMENT_STATUS, APPROVAL_STATUS, COST, ORGANIZER_CODE)
	  VALUES (?, ?, ?, ?, ?, ?, ?, 'DRAFT', 'UNPAID', 'PENDING', ?, ?)`
	_, err := r.db.ExecContext(ctx2, r.dialect.Bind(query), code, strings.TrimSpace(in.Title),
		repositories.NullStr(strings.TrimSpace(in.ImageURL)), repositories.NullStr(strings.TrimSpace(in.TargetURL)),
		strings.ToUpper(strings.TrimSpace(in.Position)), in.StartAt, in.EndAt, cost, organizer)
	return err
}

func (r *Repository) UpdateAd(ctx context.Context, organizer, code string, in models.AdCampaignInput, cost float64) error {
	ctx2, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	query := `UPDATE ` + r.adTable() + ` SET TITLE = ?, IMAGE_URL = ?, TARGET_URL = ?, POSITION = ?, START_AT = ?, END_AT = ?, COST = ?, UPDATED_AT = ` + r.dialect.Now() + ` WHERE ORGANIZER_CODE = ? AND CODE = ?`
	res, err := r.db.ExecContext(ctx2, r.dialect.Bind(query), strings.TrimSpace(in.Title),
		repositories.NullStr(strings.TrimSpace(in.ImageURL)), repositories.NullStr(strings.TrimSpace(in.TargetURL)),
		strings.ToUpper(strings.TrimSpace(in.Position)), in.StartAt, in.EndAt, cost, organizer, code)
	if err != nil {
		return err
	}
	return checkRows(res)
}

func (r *Repository) SetAdStatus(ctx context.Context, organizer, code, status, pay, approval, note string) error {
	ctx2, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	query := `UPDATE ` + r.adTable() + ` SET STATUS = ?, PAYMENT_STATUS = ?, APPROVAL_STATUS = ?, REJECT_NOTE = ?, UPDATED_AT = ` + r.dialect.Now() + ` WHERE ORGANIZER_CODE = ? AND CODE = ?`
	res, err := r.db.ExecContext(ctx2, r.dialect.Bind(query), status, pay, approval, repositories.NullStr(note), organizer, code)
	if err != nil {
		return err
	}
	return checkRows(res)
}

func (r *Repository) AdStats(ctx context.Context, code string) ([]models.AdDailyPoint, error) {
	ctx2, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	query := `SELECT STAT_DATE, IMPRESSIONS, CLICKS FROM ` + r.adStatTable() + ` WHERE CAMPAIGN_CODE = ? ORDER BY STAT_DATE ASC`
	rows, err := r.db.QueryContext(ctx2, r.dialect.Bind(query), code)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.AdDailyPoint, 0)
	for rows.Next() {
		var p models.AdDailyPoint
		var d sql.NullString
		if r.dialect == repositories.DialectMSSQL {
			var dt sql.NullString
			if err := rows.Scan(&dt, &p.Impressions, &p.Clicks); err != nil {
				return nil, err
			}
			d = dt
			if d.Valid {
				p.Date = strings.Split(d.String, " ")[0]
				if len(p.Date) > 10 {
					p.Date = p.Date[:10]
				}
			}
		} else {
			if err := rows.Scan(&d, &p.Impressions, &p.Clicks); err != nil {
				return nil, err
			}
			p.Date = d.String
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *Repository) ListNews(ctx context.Context, status, category, q string, limit, offset int) ([]models.NewsRow, error) {
	ctx2, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	args := []any{}
	cond := "n.IS_DELETED = 0"
	if status != "" {
		cond += " AND n.STATUS = ?"
		args = append(args, strings.ToUpper(status))
	}
	if category != "" {
		cond += " AND n.NEWS_CATEGORY_CODE = ?"
		args = append(args, strings.ToUpper(category))
	}
	if q = strings.TrimSpace(q); q != "" {
		cond += " AND n.TITLE LIKE ? ESCAPE '\\'"
		args = append(args, "%"+escapeQ(q)+"%")
	}
	query := `SELECT n.CODE, n.TITLE, n.SLUG, COALESCE(n.SUMMARY,''), COALESCE(n.BODY,''), COALESCE(n.IMAGE_URL,''),
	  n.NEWS_CATEGORY_CODE, COALESCE(nc.NAME,''), COALESCE(n.EVENT_CATEGORY_CODE,''), COALESCE(ec.NAME,''),
	  COALESCE(n.EVENT_CODE,''), COALESCE(e.TITLE,''), n.STATUS, COALESCE(n.AUTHOR_CODE,''), COALESCE(u.USERNAME,''),
	  COALESCE(` + r.startStr("n.SCHEDULED_AT") + `,''), COALESCE(` + r.startStr("n.PUBLISHED_AT") + `,''), n.VIEW_COUNT, n.IS_DELETED,
	  ` + r.startStr("n.CREATED_AT") + `, ` + r.startStr("n.UPDATED_AT") + `
	  FROM ` + r.newsTable() + ` n
	  LEFT JOIN ` + r.ncatTable() + ` nc ON nc.CODE = n.NEWS_CATEGORY_CODE
	  LEFT JOIN ` + r.ecatTable() + ` ec ON ec.CODE = n.EVENT_CATEGORY_CODE
	  LEFT JOIN ` + r.eventTable() + ` e ON e.CODE = n.EVENT_CODE
	  LEFT JOIN ` + r.userTable() + ` u ON u.CODE = n.AUTHOR_CODE
	  WHERE ` + cond + ` ORDER BY n.CREATED_AT DESC` + r.pageClause(limit, offset)
	rows, err := r.db.QueryContext(ctx2, r.dialect.Bind(query), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.NewsRow, 0)
	for rows.Next() {
		var m models.NewsRow
		var isDel bool
		if err := rows.Scan(&m.Code, &m.Title, &m.Slug, &m.Summary, &m.Body, &m.ImageURL,
			&m.NewsCategoryCode, &m.NewsCategoryName, &m.EventCategoryCode, &m.EventCategoryName,
			&m.EventCode, &m.EventTitle, &m.Status, &m.AuthorCode, &m.AuthorName,
			&m.ScheduledAt, &m.PublishedAt, &m.ViewCount, &isDel, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, err
		}
		m.IsDeleted = isDel
		out = append(out, m)
	}
	return out, rows.Err()
}

func (r *Repository) NewsByCode(ctx context.Context, code string) (*models.NewsRow, error) {
	ctx2, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	var m models.NewsRow
	var isDel bool
	query := `SELECT n.CODE, n.TITLE, n.SLUG, COALESCE(n.SUMMARY,''), COALESCE(n.BODY,''), COALESCE(n.IMAGE_URL,''),
	  n.NEWS_CATEGORY_CODE, COALESCE(nc.NAME,''), COALESCE(n.EVENT_CATEGORY_CODE,''), COALESCE(ec.NAME,''),
	  COALESCE(n.EVENT_CODE,''), COALESCE(e.TITLE,''), n.STATUS, COALESCE(n.AUTHOR_CODE,''), COALESCE(u.USERNAME,''),
	  COALESCE(` + r.startStr("n.SCHEDULED_AT") + `,''), COALESCE(` + r.startStr("n.PUBLISHED_AT") + `,''), n.VIEW_COUNT, n.IS_DELETED,
	  ` + r.startStr("n.CREATED_AT") + `, ` + r.startStr("n.UPDATED_AT") + `
	  FROM ` + r.newsTable() + ` n
	  LEFT JOIN ` + r.ncatTable() + ` nc ON nc.CODE = n.NEWS_CATEGORY_CODE
	  LEFT JOIN ` + r.ecatTable() + ` ec ON ec.CODE = n.EVENT_CATEGORY_CODE
	  LEFT JOIN ` + r.eventTable() + ` e ON e.CODE = n.EVENT_CODE
	  LEFT JOIN ` + r.userTable() + ` u ON u.CODE = n.AUTHOR_CODE
	  WHERE n.CODE = ? AND n.IS_DELETED = 0`
	err := r.db.QueryRowContext(ctx2, r.dialect.Bind(query), code).Scan(
		&m.Code, &m.Title, &m.Slug, &m.Summary, &m.Body, &m.ImageURL,
		&m.NewsCategoryCode, &m.NewsCategoryName, &m.EventCategoryCode, &m.EventCategoryName,
		&m.EventCode, &m.EventTitle, &m.Status, &m.AuthorCode, &m.AuthorName,
		&m.ScheduledAt, &m.PublishedAt, &m.ViewCount, &isDel, &m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		return nil, err
	}
	m.IsDeleted = isDel
	return &m, nil
}

func (r *Repository) NewsSlugExists(ctx context.Context, slug, exclude string) (bool, error) {
	ctx2, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	var n int
	query := `SELECT COUNT(*) FROM ` + r.newsTable() + ` WHERE SLUG = ? AND CODE <> ?`
	err := r.db.QueryRowContext(ctx2, r.dialect.Bind(query), strings.ToLower(strings.TrimSpace(slug)), exclude).Scan(&n)
	return n > 0, err
}

func (r *Repository) CreateNews(ctx context.Context, author, code, slug string, in models.NewsInput) error {
	ctx2, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	query := `INSERT INTO ` + r.newsTable() + ` (CODE, TITLE, SLUG, SUMMARY, BODY, IMAGE_URL, NEWS_CATEGORY_CODE, EVENT_CATEGORY_CODE, EVENT_CODE, STATUS, AUTHOR_CODE, SCHEDULED_AT, PUBLISHED_AT, VIEW_COUNT, IS_DELETED)
	  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 'DRAFT', ?, ?, NULL, 0, 0)`
	_, err := r.db.ExecContext(ctx2, r.dialect.Bind(query), code, strings.TrimSpace(in.Title), strings.ToLower(slug),
		repositories.NullStr(strings.TrimSpace(in.Summary)), strings.TrimSpace(in.Body),
		repositories.NullStr(strings.TrimSpace(in.ImageURL)), strings.ToUpper(strings.TrimSpace(in.NewsCategoryCode)),
		nullableStr(strings.ToUpper(strings.TrimSpace(in.EventCategoryCode))),
		nullableStr(strings.TrimSpace(in.EventCode)), author, nullableStr(strings.TrimSpace(in.ScheduledAt)))
	return err
}

func nullableStr(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func (r *Repository) UpdateNews(ctx context.Context, code, slug string, in models.NewsInput) error {
	ctx2, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	sched := strings.TrimSpace(in.ScheduledAt)
	query := `UPDATE ` + r.newsTable() + ` SET TITLE = ?, SLUG = ?, SUMMARY = ?, BODY = ?, IMAGE_URL = ?, NEWS_CATEGORY_CODE = ?, EVENT_CATEGORY_CODE = ?, EVENT_CODE = ?, SCHEDULED_AT = ?, UPDATED_AT = ` + r.dialect.Now() + ` WHERE CODE = ? AND IS_DELETED = 0`
	res, err := r.db.ExecContext(ctx2, r.dialect.Bind(query), strings.TrimSpace(in.Title), strings.ToLower(slug),
		repositories.NullStr(strings.TrimSpace(in.Summary)), strings.TrimSpace(in.Body),
		repositories.NullStr(strings.TrimSpace(in.ImageURL)), strings.ToUpper(strings.TrimSpace(in.NewsCategoryCode)),
		nullableStr(strings.ToUpper(strings.TrimSpace(in.EventCategoryCode))), nullableStr(strings.TrimSpace(in.EventCode)),
		nullableStr(sched), code)
	if err != nil {
		return err
	}
	return checkRows(res)
}

func (r *Repository) SetNewsStatus(ctx context.Context, code, status string) error {
	ctx2, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	if strings.ToUpper(status) == models.NewsStatusPublished {
		query := `UPDATE ` + r.newsTable() + ` SET STATUS = ?, PUBLISHED_AT = ` + r.dialect.Now() + `, UPDATED_AT = ` + r.dialect.Now() + `, SCHEDULED_AT = NULL WHERE CODE = ? AND IS_DELETED = 0`
		res, err := r.db.ExecContext(ctx2, r.dialect.Bind(query), strings.ToUpper(status), code)
		if err != nil {
			return err
		}
		return checkRows(res)
	}
	query := `UPDATE ` + r.newsTable() + ` SET STATUS = ?, UPDATED_AT = ` + r.dialect.Now() + ` WHERE CODE = ? AND IS_DELETED = 0`
	res, err := r.db.ExecContext(ctx2, r.dialect.Bind(query), strings.ToUpper(status), code)
	if err != nil {
		return err
	}
	return checkRows(res)
}

func (r *Repository) SoftDeleteNews(ctx context.Context, code string) error {
	ctx2, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	var status string
	_ = r.db.QueryRowContext(ctx2, r.dialect.Bind(`SELECT STATUS FROM `+r.newsTable()+` WHERE CODE = ?`), code).Scan(&status)
	if strings.ToUpper(status) == models.NewsStatusPublished {
		query := `UPDATE ` + r.newsTable() + ` SET STATUS = 'ARCHIVED', IS_DELETED = 0, UPDATED_AT = ` + r.dialect.Now() + ` WHERE CODE = ?`
		res, err := r.db.ExecContext(ctx2, r.dialect.Bind(query), code)
		if err != nil {
			return err
		}
		return checkRows(res)
	}
	query := `UPDATE ` + r.newsTable() + ` SET IS_DELETED = 1, UPDATED_AT = ` + r.dialect.Now() + ` WHERE CODE = ?`
	res, err := r.db.ExecContext(ctx2, r.dialect.Bind(query), code)
	if err != nil {
		return err
	}
	return checkRows(res)
}

func (r *Repository) PublishScheduled(ctx context.Context, now string) error {
	ctx2, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	query := `UPDATE ` + r.newsTable() + ` SET STATUS = 'PUBLISHED', PUBLISHED_AT = ` + r.dialect.Now() + `, SCHEDULED_AT = NULL, UPDATED_AT = ` + r.dialect.Now() + ` WHERE STATUS = 'DRAFT' AND IS_DELETED = 0 AND SCHEDULED_AT IS NOT NULL AND SCHEDULED_AT <= ?`
	_, err := r.db.ExecContext(ctx2, r.dialect.Bind(query), now)
	return err
}

func (r *Repository) NewsCategories(ctx context.Context) ([]models.NewsCategory, error) {
	ctx2, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	rows, err := r.db.QueryContext(ctx2, `SELECT CODE, NAME FROM `+r.ncatTable()+` ORDER BY NAME ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.NewsCategory, 0)
	for rows.Next() {
		var c models.NewsCategory
		if err := rows.Scan(&c.Code, &c.Name); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *Repository) EventCategories(ctx context.Context) ([]models.EventCategory, error) {
	ctx2, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	rows, err := r.db.QueryContext(ctx2, `SELECT CODE, NAME FROM `+r.ecatTable()+` ORDER BY NAME ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.EventCategory, 0)
	for rows.Next() {
		var c models.EventCategory
		if err := rows.Scan(&c.Code, &c.Name); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *Repository) OwnedEvents(ctx context.Context, organizer string) ([]OwnedEvent, error) {
	ctx2, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	rows, err := r.db.QueryContext(ctx2, r.dialect.Bind(`SELECT CODE, TITLE FROM `+r.eventTable()+` WHERE ORGANIZER_CODE = ? AND IS_DELETED = 0 ORDER BY TITLE ASC`), organizer)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]OwnedEvent, 0)
	for rows.Next() {
		var o OwnedEvent
		if err := rows.Scan(&o.Code, &o.Title); err != nil {
			return nil, err
		}
		out = append(out, o)
	}
	return out, rows.Err()
}

func (r *Repository) CategoryExists(ctx context.Context, code string) (bool, error) {
	ctx2, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	var n int
	err := r.db.QueryRowContext(ctx2, r.dialect.Bind(`SELECT COUNT(*) FROM `+r.ncatTable()+` WHERE CODE = ?`), strings.ToUpper(code)).Scan(&n)
	return n > 0, err
}

func (r *Repository) EventOwned(ctx context.Context, organizer, event string) (bool, error) {
	ctx2, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	var n int
	err := r.db.QueryRowContext(ctx2, r.dialect.Bind(`SELECT COUNT(*) FROM `+r.eventTable()+` WHERE CODE = ? AND ORGANIZER_CODE = ? AND IS_DELETED = 0`), event, organizer).Scan(&n)
	return n > 0, err
}
