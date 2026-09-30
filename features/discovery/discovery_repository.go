package discovery

import (
	"context"
	"database/sql"
	"strings"

	"golang-backend/models"
	"golang-backend/repositories"
)

// Public reads over T_* tables (events, news, promos, campaigns).
// Only publishable rows are returned (see liveClauses below).
//
// Konvensi query: SQL di variabel `query`, placeholder "?" + Bind,
// timeout via repositories.WithTimeout. Sort/filter tidak pernah ditempel
// mentah dari request (whitelist di service, fragment konstan di sini).
type RepositoryInterface interface {
	EventCategories(ctx context.Context) ([]models.EventCategory, error)
	NewsCategories(ctx context.Context) ([]models.NewsCategory, error)
	Cities(ctx context.Context) ([]models.CityOption, error)
	// LiveBanners returns campaigns in their flight period (optional position).
	LiveBanners(ctx context.Context, position string) ([]models.Banner, error)
	// CampaignLive reports whether a campaign may record events.
	CampaignLive(ctx context.Context, code string) (bool, error)
	// RecentAdEvent reports a same-source event inside the cooldown window
	// (windowMin minutes, evaluated in database time).
	RecentAdEvent(ctx context.Context, campaign, kind, iphash string, windowMin int) (bool, error)
	InsertAdEvent(ctx context.Context, campaign, kind, iphash string) error
	// FeaturedPromos returns active promos ending soonest (home shelf).
	FeaturedPromos(ctx context.Context, now string, limit int) ([]models.PromoItem, error)
	// ListPromos lists active|upcoming promos with category + text filter.
	ListPromos(ctx context.Context, now, mode, category, q string, limit, offset int) ([]models.PromoItem, error)
	// LatestNews returns the newest published articles.
	LatestNews(ctx context.Context, limit int) ([]models.NewsItem, error)
	// ListNews filters by category + title search, newest|popular sort.
	ListNews(ctx context.Context, category, q, orderBy string, limit, offset int) ([]models.NewsItem, error)
	// NewsBySlug loads one published article + names.
	NewsBySlug(ctx context.Context, slug string) (*models.NewsDetail, error)
	// RelatedNews lists same-category articles (excluding one).
	RelatedNews(ctx context.Context, category, exclude string, limit int) ([]models.NewsItem, error)
	// NewsViewedSince reports a viewer hit inside the cooldown window
	// (windowMin minutes, evaluated in database time).
	NewsViewedSince(ctx context.Context, news, hash string, windowMin int) (bool, error)
	InsertNewsView(ctx context.Context, news, hash string) error
	IncrementNewsView(ctx context.Context, news string) error
	// ListEvents is the filtered/paginated published event list.
	// sort is service-validated (nearest|popular|price_asc|price_desc);
	// ORDER BY is assembled from constants, never request text.
	ListEvents(ctx context.Context, f models.DiscoveryFilter, sort string) ([]models.EventCard, error)
	// EventRow loads one published, not-ended event.
	EventRow(ctx context.Context, code string) (*models.EventDetail, error)
	// EventTicketTypes lists types with sold counts.
	EventTicketTypes(ctx context.Context, event, now string) ([]models.TicketTypeInfo, error)
	// EventPromos lists promos applicable to an event right now.
	EventPromos(ctx context.Context, event, category string) ([]models.PromoItem, error)
	// RelatedEvents lists same-category/city events (excluding one).
	RelatedEvents(ctx context.Context, event, category, city string, limit int) ([]models.EventCard, error)
	// PopularEvents orders by tickets sold (home shelf).
	PopularEvents(ctx context.Context, city string, limit int) ([]models.EventCard, error)
	// UpcomingEvents orders by nearest start (home shelf).
	UpcomingEvents(ctx context.Context, city string, limit int) ([]models.EventCard, error)
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
func (r *Repository) promoTable() string { return r.dialect.Table("T_PROMO") }
func (r *Repository) newsTable() string  { return r.dialect.Table("T_NEWS") }
func (r *Repository) ncatTable() string  { return r.dialect.Table("T_NEWS_CATEGORY") }
func (r *Repository) nviewTable() string { return r.dialect.Table("T_NEWS_VIEW") }
func (r *Repository) adTable() string    { return r.dialect.Table("T_AD_CAMPAIGN") }
func (r *Repository) adEvTable() string  { return r.dialect.Table("T_AD_EVENT") }

// nowFn is the current-timestamp expression in database time. The app and
// database servers may live in different timezones, so minute-scale windows
// (ad/news cooldowns) and live predicates must never use app time.
func (r *Repository) nowFn() string { return r.dialect.Now() }

// liveEvent is the publishable-event predicate evaluated in database time.
func (r *Repository) liveEvent() string {
	return `e.STATUS = 'PUBLISHED' AND (e.END_AT IS NULL OR e.END_AT >= ` + r.nowFn() + `)`
}

// minutesAgo renders "windowMin minutes before now" in database time.
func (r *Repository) minutesAgo(windowMin int) (string, []any) {
	switch r.dialect {
	case repositories.DialectPostgres:
		return `NOW() - (? * INTERVAL '1 minute')`, []any{windowMin}
	case repositories.DialectSQLite:
		return `datetime('now', '-' || ? || ' minutes')`, []any{windowMin}
	default:
		return `DATEADD(minute, -?, GETDATE())`, []any{windowMin}
	}
}

// soldSub counts PAID tickets of an event; quotaSub sums type quotas.
func (r *Repository) soldSub() string {
	return `(SELECT COALESCE(SUM(o.QTY),0) FROM ` + r.orderTable() + ` o WHERE o.EVENT_CODE = e.CODE AND o.STATUS = 'PAID')`
}

func (r *Repository) quotaSub() string {
	return `(SELECT COALESCE(SUM(tt.QUOTA),0) FROM ` + r.typeTable() + ` tt WHERE tt.EVENT_CODE = e.CODE)`
}

func (r *Repository) minPriceSub() string {
	return `(SELECT COALESCE(MIN(tt.PRICE),0) FROM ` + r.typeTable() + ` tt WHERE tt.EVENT_CODE = e.CODE)`
}

// cardCols selects EventCard columns (aliases must match scanEventCard).
func (r *Repository) cardCols() string {
	return `e.CODE, e.TITLE, COALESCE(ec.NAME, ''), COALESCE(c.NAME, ''), COALESCE(e.VENUE, ''),
	  ` + r.startStr("e.START_AT") + `, COALESCE(` + r.startStr("e.END_AT") + `, ''),
	  COALESCE(e.POSTER_URL, ''), ` + r.minPriceSub() + `, ` + r.soldSub() + `, ` + r.quotaSub()
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

// limitClause renders OFFSET/FETCH (mssql) or LIMIT/OFFSET per engine.
// MSSQL FETCH counts must be integer literals (never parameters), so limit
// and offset are interpolated here - both are Go ints clamped by callers,
// never request text. Every caller already has ORDER BY, which FETCH needs.
func (r *Repository) limitClause(limit, offset int) (prefix []any, selectMod, suffix string) {
	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	if r.dialect == repositories.DialectMSSQL {
		return nil, "", " OFFSET " + itoa(offset) + " ROWS FETCH NEXT " + itoa(limit) + " ROWS ONLY"
	}
	return nil, "", " LIMIT " + itoa(limit) + " OFFSET " + itoa(offset)
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

// escapeLike escapes % _ \ for LIKE ... ESCAPE '\'.
func escapeLike(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `%`, `\%`)
	s = strings.ReplaceAll(s, `_`, `\_`)
	return "%" + s + "%"
}

func badgeFor(sold, quota int) string {
	if quota > 0 && sold >= quota {
		return "SOLD_OUT"
	}
	if quota > 0 && float64(sold)/float64(quota) >= 0.9 {
		return "HAMPIR_HABIS"
	}
	return ""
}

func (r *Repository) scanCards(rows *sql.Rows) ([]models.EventCard, error) {
	defer rows.Close()
	out := make([]models.EventCard, 0)
	for rows.Next() {
		var c models.EventCard
		if err := rows.Scan(&c.Code, &c.Title, &c.Category, &c.City, &c.Venue,
			&c.StartAt, &c.EndAt, &c.PosterURL, &c.MinPrice, &c.Sold, &c.Quota); err != nil {
			return nil, err
		}
		c.Badge = badgeFor(c.Sold, c.Quota)
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *Repository) EventCategories(ctx context.Context) ([]models.EventCategory, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	query := `SELECT CODE, NAME FROM ` + r.catTable() + ` ORDER BY SORT_ORDER ASC, NAME ASC`
	rows, err := r.db.QueryContext(ctx, query)
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

func (r *Repository) NewsCategories(ctx context.Context) ([]models.NewsCategory, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	query := `SELECT CODE, NAME FROM ` + r.ncatTable() + ` ORDER BY NAME ASC`
	rows, err := r.db.QueryContext(ctx, query)
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

func (r *Repository) Cities(ctx context.Context) ([]models.CityOption, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	query := `SELECT CODE, NAME FROM ` + r.cityTable() + ` ORDER BY NAME ASC`
	rows, err := r.db.QueryContext(ctx, query)
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

func (r *Repository) LiveBanners(ctx context.Context, position string) ([]models.Banner, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	args := []any{}
	pos := ""
	if position != "" {
		pos = " AND POSITION = ?"
		args = append(args, position)
	}
	query := `SELECT CODE, TITLE, COALESCE(IMAGE_URL, ''), COALESCE(TARGET_URL, ''), POSITION
	  FROM ` + r.adTable() + `
	  WHERE STATUS = 'ACTIVE' AND START_AT <= ` + r.nowFn() + ` AND END_AT >= ` + r.nowFn() + pos + `
	  ORDER BY START_AT DESC`
	rows, err := r.db.QueryContext(ctx, r.dialect.Bind(query), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.Banner, 0)
	for rows.Next() {
		var b models.Banner
		if err := rows.Scan(&b.Code, &b.Title, &b.ImageURL, &b.TargetURL, &b.Position); err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

func (r *Repository) CampaignLive(ctx context.Context, code string) (bool, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	var n int
	query := `SELECT COUNT(*) FROM ` + r.adTable() + `
	  WHERE CODE = ? AND STATUS = 'ACTIVE' AND START_AT <= ` + r.nowFn() + ` AND END_AT >= ` + r.nowFn()
	err := r.db.QueryRowContext(ctx, r.dialect.Bind(query), code).Scan(&n)
	return n > 0, err
}

func (r *Repository) RecentAdEvent(ctx context.Context, campaign, kind, iphash string, windowMin int) (bool, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	var n int
	cutoff, cutArgs := r.minutesAgo(windowMin)
	query := `SELECT COUNT(*) FROM ` + r.adEvTable() + `
	  WHERE CAMPAIGN_CODE = ? AND KIND = ? AND IP_HASH = ? AND CREATED_AT >= ` + cutoff
	args := append([]any{campaign, kind, iphash}, cutArgs...)
	err := r.db.QueryRowContext(ctx, r.dialect.Bind(query), args...).Scan(&n)
	return n > 0, err
}

func (r *Repository) InsertAdEvent(ctx context.Context, campaign, kind, iphash string) error {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	query := `INSERT INTO ` + r.adEvTable() + ` (CAMPAIGN_CODE, KIND, IP_HASH) VALUES (?, ?, ?)`
	_, err := r.db.ExecContext(ctx, r.dialect.Bind(query), campaign, kind, iphash)
	return err
}

// promoCols selects PromoItem columns + live state (aliases must match
// scanPromos). State is computed in database time, like the predicates.
func (r *Repository) promoCols() string {
	return `p.CODE, p.PROMO_CODE, COALESCE(p.DESCRIPTION, ''),
	  p.DISCOUNT_PERCENT, p.DISCOUNT_AMOUNT, p.MAX_DISCOUNT_AMOUNT, p.MIN_PURCHASE,
	  p.MAX_USES, p.USED_COUNT, ` + r.startStr("p.VALID_FROM") + `, ` + r.startStr("p.VALID_TO") + `,
	  COALESCE(p.EVENT_CODE, ''), COALESCE(pe.TITLE, ''),
	  COALESCE(p.COVERAGE_CATEGORY_CODE, ''), COALESCE(ec.NAME, ''),
	  CASE WHEN p.VALID_FROM > ` + r.nowFn() + ` THEN 'UPCOMING' ELSE 'ACTIVE' END`
}

func (r *Repository) promoJoins() string {
	return ` LEFT JOIN ` + r.eventTable() + ` pe ON pe.CODE = p.EVENT_CODE
	  LEFT JOIN ` + r.catTable() + ` ec ON ec.CODE = p.COVERAGE_CATEGORY_CODE`
}

func (r *Repository) scanPromos(rows *sql.Rows) ([]models.PromoItem, error) {
	defer rows.Close()
	out := make([]models.PromoItem, 0)
	for rows.Next() {
		var m models.PromoItem
		var maxUses sql.NullInt64
		if err := rows.Scan(&m.Code, &m.PromoCode, &m.Description, &m.DiscountPercent,
			&m.DiscountAmount, &m.MaxDiscount, &m.MinPurchase, &maxUses, &m.Remaining,
			&m.ValidFrom, &m.ValidTo, &m.EventCode, &m.EventTitle,
			&m.CategoryCode, &m.CategoryName, &m.State); err != nil {
			return nil, err
		}
		// Remaining holds USED_COUNT from the scan; convert to remaining.
		used := m.Remaining
		if maxUses.Valid {
			m.Remaining = int(maxUses.Int64) - used
		} else {
			m.Remaining = -1 // unlimited
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (r *Repository) FeaturedPromos(ctx context.Context, now string, limit int) ([]models.PromoItem, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	pre, sel, suf := r.limitClause(limit, 0)
	query := `SELECT ` + sel + ` ` + r.promoCols() + ` FROM ` + r.promoTable() + ` p` + r.promoJoins() + `
	  WHERE p.STATUS = 'ACTIVE' AND p.VALID_FROM <= ` + r.nowFn() + ` AND p.VALID_TO >= ` + r.nowFn() + `
	  AND (p.MAX_USES IS NULL OR p.USED_COUNT < p.MAX_USES)
	  ORDER BY p.VALID_TO ASC` + suf
	rows, err := r.db.QueryContext(ctx, r.dialect.Bind(query), pre...)
	if err != nil {
		return nil, err
	}
	return r.scanPromos(rows)
}

func (r *Repository) ListPromos(ctx context.Context, now, mode, category, q string, limit, offset int) ([]models.PromoItem, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	args := []any{}
	cond := "p.STATUS = 'ACTIVE'"
	switch mode {
	case "upcoming":
		cond += " AND p.VALID_FROM > " + r.nowFn()
	default: // active: running now with remaining quota
		cond += " AND p.VALID_FROM <= " + r.nowFn() + " AND p.VALID_TO >= " + r.nowFn() + " AND (p.MAX_USES IS NULL OR p.USED_COUNT < p.MAX_USES)"
	}
	if category != "" {
		cond += " AND (p.COVERAGE_CATEGORY_CODE = ? OR p.COVERAGE_CATEGORY_CODE IS NULL)"
		args = append(args, category)
	}
	if q != "" {
		cond += ` AND (p.PROMO_CODE LIKE ? ESCAPE '\' OR p.DESCRIPTION LIKE ? ESCAPE '\')`
		args = append(args, escapeLike(q), escapeLike(q))
	}
	pre, sel, suf := r.limitClause(limit, offset)
	query := `SELECT ` + sel + ` ` + r.promoCols() + ` FROM ` + r.promoTable() + ` p` + r.promoJoins() + `
	  WHERE ` + cond + ` ORDER BY p.VALID_FROM DESC` + suf
	rows, err := r.db.QueryContext(ctx, r.dialect.Bind(query), append(pre, args...)...)
	if err != nil {
		return nil, err
	}
	return r.scanPromos(rows)
}

func (r *Repository) newsCols() string {
	return `n.CODE, n.TITLE, n.SLUG, COALESCE(n.SUMMARY, ''), COALESCE(n.IMAGE_URL, ''),
	  nc.NAME, ` + r.startStr("n.PUBLISHED_AT") + `, n.VIEW_COUNT`
}

func (r *Repository) scanNews(rows *sql.Rows) ([]models.NewsItem, error) {
	defer rows.Close()
	out := make([]models.NewsItem, 0)
	for rows.Next() {
		var n models.NewsItem
		var published sql.NullString
		if err := rows.Scan(&n.Code, &n.Title, &n.Slug, &n.Summary, &n.ImageURL,
			&n.Category, &published, &n.ViewCount); err != nil {
			return nil, err
		}
		if published.Valid {
			n.PublishedAt = published.String
		}
		out = append(out, n)
	}
	return out, rows.Err()
}

func (r *Repository) LatestNews(ctx context.Context, limit int) ([]models.NewsItem, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	pre, sel, suf := r.limitClause(limit, 0)
	query := `SELECT ` + sel + ` ` + r.newsCols() + ` FROM ` + r.newsTable() + ` n
	  JOIN ` + r.ncatTable() + ` nc ON nc.CODE = n.NEWS_CATEGORY_CODE
	  WHERE n.STATUS = 'PUBLISHED' ORDER BY n.PUBLISHED_AT DESC` + suf
	rows, err := r.db.QueryContext(ctx, r.dialect.Bind(query), pre...)
	if err != nil {
		return nil, err
	}
	return r.scanNews(rows)
}

func (r *Repository) ListNews(ctx context.Context, category, q, orderBy string, limit, offset int) ([]models.NewsItem, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	args := []any{}
	cond := "n.STATUS = 'PUBLISHED'"
	if category != "" {
		cond += " AND n.NEWS_CATEGORY_CODE = ?"
		args = append(args, category)
	}
	if q != "" {
		cond += ` AND n.TITLE LIKE ? ESCAPE '\'`
		args = append(args, escapeLike(q))
	}
	order := "n.PUBLISHED_AT DESC"
	if orderBy == "popular" {
		order = "n.VIEW_COUNT DESC, n.PUBLISHED_AT DESC"
	}
	pre, sel, suf := r.limitClause(limit, offset)
	query := `SELECT ` + sel + ` ` + r.newsCols() + ` FROM ` + r.newsTable() + ` n
	  JOIN ` + r.ncatTable() + ` nc ON nc.CODE = n.NEWS_CATEGORY_CODE
	  WHERE ` + cond + ` ORDER BY ` + order + suf
	rows, err := r.db.QueryContext(ctx, r.dialect.Bind(query), append(pre, args...)...)
	if err != nil {
		return nil, err
	}
	return r.scanNews(rows)
}

func (r *Repository) NewsBySlug(ctx context.Context, slug string) (*models.NewsDetail, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	var d models.NewsDetail
	var published, body sql.NullString
	var eventCode, eventCat sql.NullString
	query := `SELECT n.CODE, n.TITLE, n.SLUG, COALESCE(n.SUMMARY, ''), COALESCE(n.IMAGE_URL, ''),
	  nc.NAME, ` + r.startStr("n.PUBLISHED_AT") + `, n.VIEW_COUNT, n.BODY,
	  n.EVENT_CODE, n.EVENT_CATEGORY_CODE
	  FROM ` + r.newsTable() + ` n
	  JOIN ` + r.ncatTable() + ` nc ON nc.CODE = n.NEWS_CATEGORY_CODE
	  WHERE n.SLUG = ? AND n.STATUS = 'PUBLISHED'`
	err := r.db.QueryRowContext(ctx, r.dialect.Bind(query), slug).Scan(
		&d.Code, &d.Title, &d.Slug, &d.Summary, &d.ImageURL, &d.Category,
		&published, &d.ViewCount, &body, &eventCode, &eventCat)
	if err != nil {
		return nil, err
	}
	if published.Valid {
		d.PublishedAt = published.String
	}
	if body.Valid {
		d.Body = body.String
	}
	if eventCode.Valid {
		d.EventCode = eventCode.String
	}
	d.RelatedNews = []models.NewsItem{}
	return &d, nil
}

func (r *Repository) RelatedNews(ctx context.Context, category, exclude string, limit int) ([]models.NewsItem, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	pre, sel, suf := r.limitClause(limit, 0)
	query := `SELECT ` + sel + ` ` + r.newsCols() + ` FROM ` + r.newsTable() + ` n
	  JOIN ` + r.ncatTable() + ` nc ON nc.CODE = n.NEWS_CATEGORY_CODE
	  WHERE n.STATUS = 'PUBLISHED' AND n.NEWS_CATEGORY_CODE = ? AND n.CODE <> ?
	  ORDER BY n.PUBLISHED_AT DESC` + suf
	rows, err := r.db.QueryContext(ctx, r.dialect.Bind(query), append(pre, category, exclude)...)
	if err != nil {
		return nil, err
	}
	return r.scanNews(rows)
}

func (r *Repository) NewsViewedSince(ctx context.Context, news, hash string, windowMin int) (bool, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	var n int
	cutoff, cutArgs := r.minutesAgo(windowMin)
	query := `SELECT COUNT(*) FROM ` + r.nviewTable() + `
	  WHERE NEWS_CODE = ? AND VIEWER_HASH = ? AND VIEWED_AT >= ` + cutoff
	args := append([]any{news, hash}, cutArgs...)
	err := r.db.QueryRowContext(ctx, r.dialect.Bind(query), args...).Scan(&n)
	return n > 0, err
}

func (r *Repository) InsertNewsView(ctx context.Context, news, hash string) error {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	query := `INSERT INTO ` + r.nviewTable() + ` (NEWS_CODE, VIEWER_HASH) VALUES (?, ?)`
	_, err := r.db.ExecContext(ctx, r.dialect.Bind(query), news, hash)
	return err
}

func (r *Repository) IncrementNewsView(ctx context.Context, news string) error {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	query := `UPDATE ` + r.newsTable() + ` SET VIEW_COUNT = VIEW_COUNT + 1,
	  UPDATED_AT = ` + r.dialect.Now() + ` WHERE CODE = ?`
	_, err := r.db.ExecContext(ctx, r.dialect.Bind(query), news)
	return err
}

// eventJoins brings category + city names for card queries.
func (r *Repository) eventJoins() string {
	return ` LEFT JOIN ` + r.catTable() + ` ec ON ec.CODE = e.EVENT_CATEGORY_CODE
	  LEFT JOIN ` + r.cityTable() + ` c ON c.CODE = e.CITY_CODE`
}

func (r *Repository) ListEvents(ctx context.Context, f models.DiscoveryFilter, sort string) ([]models.EventCard, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	args := []any{}
	cond := ""
	if len(f.Categories) > 0 {
		holders := make([]string, 0, len(f.Categories))
		for _, c := range f.Categories {
			holders = append(holders, "?")
			args = append(args, c)
		}
		cond += " AND e.EVENT_CATEGORY_CODE IN (" + strings.Join(holders, ",") + ")"
	}
	if f.City != "" {
		cond += " AND e.CITY_CODE = ?"
		args = append(args, f.City)
	}
	if f.From != "" {
		cond += " AND e.START_AT >= ?"
		args = append(args, f.From)
	}
	if f.To != "" {
		cond += " AND e.START_AT < ?"
		args = append(args, f.To)
	}
	if f.Query != "" {
		cond += ` AND (e.TITLE LIKE ? ESCAPE '\' OR e.VENUE LIKE ? ESCAPE '\')`
		args = append(args, escapeLike(f.Query), escapeLike(f.Query))
	}
	priceWrap := ""
	if f.MinPrice > 0 || f.MaxPrice > 0 {
		priceWrap = " AND " + r.minPriceSub() + " >= ? AND (" + r.minPriceSub() + " <= ? OR ? = 0)"
		args = append(args, f.MinPrice, f.MaxPrice, f.MaxPrice)
	}
	pre, sel, suf := r.limitClause(f.Limit, f.Offset)
	order := "e.START_AT ASC"
	switch sort {
	case "popular":
		order = r.soldSub() + " DESC, e.START_AT ASC"
	case "price_asc":
		order = r.minPriceSub() + " ASC, e.START_AT ASC"
	case "price_desc":
		order = r.minPriceSub() + " DESC, e.START_AT ASC"
	}
	query := `SELECT ` + sel + ` ` + r.cardCols() + ` FROM ` + r.eventTable() + ` e` + r.eventJoins() + `
	  WHERE ` + r.liveEvent() + cond + priceWrap + ` ORDER BY ` + order + suf
	rows, err := r.db.QueryContext(ctx, r.dialect.Bind(query), append(pre, args...)...)
	if err != nil {
		return nil, err
	}
	return r.scanCards(rows)
}

func (r *Repository) EventRow(ctx context.Context, code string) (*models.EventDetail, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	var d models.EventDetail
	var userTable string
	switch r.dialect {
	case repositories.DialectPostgres, repositories.DialectSQLite:
		userTable = "CPUSER"
	default:
		userTable = "dbo.CPUSER"
	}
	query := `SELECT e.CODE, e.TITLE, COALESCE(e.DESCRIPTION, ''), COALESCE(ec.NAME, ''),
	  COALESCE(e.EVENT_CATEGORY_CODE, ''), e.ORGANIZER_CODE, ` + r.startStr("e.START_AT") + `, COALESCE(` + r.startStr("e.END_AT") + `, ''),
	  COALESCE(e.VENUE, ''), COALESCE(e.ADDRESS, ''), COALESCE(c.NAME, ''),
	  e.LATITUDE, e.LONGITUDE, COALESCE(e.POSTER_URL, ''), e.STATUS, COALESCE(u.USERNAME, '')
	  FROM ` + r.eventTable() + ` e` + r.eventJoins() + `
	  LEFT JOIN ` + userTable + ` u ON u.CODE = e.ORGANIZER_CODE
	  WHERE e.CODE = ? AND ` + r.liveEvent()
	var desc, addr, poster, end sql.NullString
	var lat, lng sql.NullFloat64
	var orgName sql.NullString
	err := r.db.QueryRowContext(ctx, r.dialect.Bind(query), code).Scan(
		&d.Code, &d.Title, &desc, &d.Category, &d.CategoryCode, &d.Organizer,
		&d.StartAt, &end, &d.Venue, &addr, &d.City,
		&lat, &lng, &poster, &d.Status, &orgName)
	if err != nil {
		return nil, err
	}
	d.Description = desc.String
	d.EndAt = end.String
	d.Address = addr.String
	d.PosterURL = poster.String
	if lat.Valid {
		d.Latitude = lat.Float64
	}
	if lng.Valid {
		d.Longitude = lng.Float64
	}
	if orgName.Valid {
		d.Organizer = orgName.String
	}
	d.TicketTypes = []models.TicketTypeInfo{}
	d.Promos = []models.PromoItem{}
	d.Related = []models.EventCard{}
	return &d, nil
}

func (r *Repository) EventTicketTypes(ctx context.Context, event, now string) ([]models.TicketTypeInfo, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	query := `SELECT tt.CODE, tt.NAME, tt.PRICE, tt.QUOTA,
	  COALESCE((SELECT SUM(o.QTY) FROM ` + r.orderTable() + ` o
	    WHERE o.TICKET_TYPE_CODE = tt.CODE AND o.STATUS = 'PAID'), 0),
	  COALESCE(` + r.startStr("tt.SALE_START_AT") + `, ''),
	  COALESCE(` + r.startStr("tt.SALE_END_AT") + `, '')
	  FROM ` + r.typeTable() + ` tt WHERE tt.EVENT_CODE = ? ORDER BY tt.PRICE ASC`
	rows, err := r.db.QueryContext(ctx, r.dialect.Bind(query), event)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.TicketTypeInfo, 0)
	for rows.Next() {
		var t models.TicketTypeInfo
		var saleFrom, saleTo string
		if err := rows.Scan(&t.Code, &t.Name, &t.Price, &t.Quota, &t.Sold, &saleFrom, &saleTo); err != nil {
			return nil, err
		}
		t.Remaining = t.Quota - t.Sold
		if t.Remaining < 0 {
			t.Remaining = 0
		}
		t.State = ticketState(t, saleFrom, saleTo, now)
		out = append(out, t)
	}
	return out, rows.Err()
}

func ticketState(t models.TicketTypeInfo, saleFrom, saleTo, now string) string {
	if t.Quota > 0 && t.Sold >= t.Quota {
		return "SOLD_OUT"
	}
	if saleFrom != "" && now < saleFrom {
		return "NOT_OPEN"
	}
	if saleTo != "" && now > saleTo {
		return "CLOSED"
	}
	if t.Quota > 0 && float64(t.Sold)/float64(t.Quota) >= 0.9 {
		return "LOW"
	}
	return "OPEN"
}

func (r *Repository) EventPromos(ctx context.Context, event, category string) ([]models.PromoItem, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	query := `SELECT ` + r.promoCols() + ` FROM ` + r.promoTable() + ` p` + r.promoJoins() + `
	  WHERE p.STATUS = 'ACTIVE' AND p.VALID_FROM <= ` + r.nowFn() + ` AND p.VALID_TO >= ` + r.nowFn() + `
	  AND (p.MAX_USES IS NULL OR p.USED_COUNT < p.MAX_USES)
	  AND (p.EVENT_CODE = ? OR p.EVENT_CODE IS NULL)
	  AND (p.COVERAGE_CATEGORY_CODE IS NULL OR p.COVERAGE_CATEGORY_CODE = ?)
	  ORDER BY p.VALID_TO ASC`
	rows, err := r.db.QueryContext(ctx, r.dialect.Bind(query), event, category)
	if err != nil {
		return nil, err
	}
	return r.scanPromos(rows)
}

func (r *Repository) RelatedEvents(ctx context.Context, event, category, city string, limit int) ([]models.EventCard, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	pre, sel, suf := r.limitClause(limit, 0)
	query := `SELECT ` + sel + ` ` + r.cardCols() + ` FROM ` + r.eventTable() + ` e` + r.eventJoins() + `
	  WHERE ` + r.liveEvent() + ` AND e.CODE <> ?
	  AND (e.EVENT_CATEGORY_CODE = ? OR e.CITY_CODE = ?)
	  ORDER BY e.START_AT ASC` + suf
	rows, err := r.db.QueryContext(ctx, r.dialect.Bind(query), append(pre, event, category, city)...)
	if err != nil {
		return nil, err
	}
	return r.scanCards(rows)
}

func (r *Repository) PopularEvents(ctx context.Context, city string, limit int) ([]models.EventCard, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	args := []any{}
	cityCond := ""
	if city != "" {
		cityCond = " AND e.CITY_CODE = ?"
		args = append(args, city)
	}
	pre, sel, suf := r.limitClause(limit, 0)
	query := `SELECT ` + sel + ` ` + r.cardCols() + ` FROM ` + r.eventTable() + ` e` + r.eventJoins() + `
	  WHERE ` + r.liveEvent() + cityCond + ` ORDER BY ` + r.soldSub() + ` DESC` + suf
	rows, err := r.db.QueryContext(ctx, r.dialect.Bind(query), append(pre, args...)...)
	if err != nil {
		return nil, err
	}
	return r.scanCards(rows)
}

func (r *Repository) UpcomingEvents(ctx context.Context, city string, limit int) ([]models.EventCard, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	args := []any{}
	cityCond := ""
	if city != "" {
		cityCond = " AND e.CITY_CODE = ?"
		args = append(args, city)
	}
	pre, sel, suf := r.limitClause(limit, 0)
	query := `SELECT ` + sel + ` ` + r.cardCols() + ` FROM ` + r.eventTable() + ` e` + r.eventJoins() + `
	  WHERE ` + r.liveEvent() + ` AND e.START_AT >= ` + r.nowFn() + cityCond + ` ORDER BY e.START_AT ASC` + suf
	rows, err := r.db.QueryContext(ctx, r.dialect.Bind(query), append(pre, args...)...)
	if err != nil {
		return nil, err
	}
	return r.scanCards(rows)
}
