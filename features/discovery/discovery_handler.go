package discovery

import (
	"net/http"
	"strconv"
	"strings"

	"golang-backend/internal/web"
	"golang-backend/models"
	"golang-backend/utils"
)

// Public handlers (no auth): Home, News, Events, Event Detail, Promotions.
// Errors use the shared envelope ({"error": ...}) via web.ServiceError.
type Handler struct {
	Service ServiceInterface
}

func NewHandler(service ServiceInterface) *Handler {
	return &Handler{Service: service}
}

func (h *Handler) handleServiceError(w http.ResponseWriter, err error) {
	web.ServiceError(w, err)
}

func parseCategories(r *http.Request) []string {
	raw := strings.TrimSpace(r.URL.Query().Get("categories"))
	if raw == "" {
		if single := strings.TrimSpace(r.URL.Query().Get("category")); single != "" {
			return []string{single}
		}
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

func parseFloatQuery(r *http.Request, name string) float64 {
	v := strings.TrimSpace(r.URL.Query().Get(name))
	if v == "" {
		return 0
	}
	n, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return -1 // sentinel: service rejects negatives
	}
	return n
}

func (h *Handler) Home(w http.ResponseWriter, r *http.Request) {
	home, err := h.Service.Home(r.Context(), r.URL.Query().Get("city"))
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	web.WriteJSON(w, http.StatusOK, home)
}

func (h *Handler) Categories(w http.ResponseWriter, r *http.Request) {
	cats, err := h.Service.Categories(r.Context())
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	if cats == nil {
		cats = []models.EventCategory{}
	}
	web.WriteJSON(w, http.StatusOK, cats)
}

func (h *Handler) Cities(w http.ResponseWriter, r *http.Request) {
	cities, err := h.Service.Cities(r.Context())
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	if cities == nil {
		cities = []models.CityOption{}
	}
	web.WriteJSON(w, http.StatusOK, cities)
}

func (h *Handler) NewsCategories(w http.ResponseWriter, r *http.Request) {
	cats, err := h.Service.NewsCategories(r.Context())
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	if cats == nil {
		cats = []models.NewsCategory{}
	}
	web.WriteJSON(w, http.StatusOK, cats)
}

func (h *Handler) News(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit, offset := web.Paginate(r)
	items, err := h.Service.NewsList(r.Context(), q.Get("category"), q.Get("q"), q.Get("sort"), limit, offset)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	if items == nil {
		items = []models.NewsItem{}
	}
	web.WriteJSON(w, http.StatusOK, items)
}

func (h *Handler) NewsDetail(w http.ResponseWriter, r *http.Request) {
	slug, ok := web.PathCode(r, "slug")
	if !ok {
		web.WriteError(w, http.StatusBadRequest, "Invalid news slug")
		return
	}
	viewer := viewerHash(r)
	detail, err := h.Service.NewsDetail(r.Context(), slug, viewer)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	web.WriteJSON(w, http.StatusOK, detail)
}

// viewerHash fingerprints IP + UA per article (30-minute view cooldown).
func viewerHash(r *http.Request) string {
	ip := utils.ClientIP(r.RemoteAddr)
	ua := r.Header.Get("User-Agent")
	if ip == "" && ua == "" {
		return ""
	}
	return utils.HashString(ip + "|" + ua)
}

func (h *Handler) Events(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit, offset := web.Paginate(r)
	filter := models.DiscoveryFilter{
		Categories: parseCategories(r),
		City:       q.Get("city"),
		From:       strings.TrimSpace(q.Get("from")),
		To:         strings.TrimSpace(q.Get("to")),
		MinPrice:   parseFloatQuery(r, "min_price"),
		MaxPrice:   parseFloatQuery(r, "max_price"),
		Query:      q.Get("q"),
		Sort:       q.Get("sort"),
		Limit:      limit,
		Offset:     offset,
	}
	events, err := h.Service.EventsList(r.Context(), filter)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	if events == nil {
		events = []models.EventCard{}
	}
	web.WriteJSON(w, http.StatusOK, events)
}

func (h *Handler) EventDetail(w http.ResponseWriter, r *http.Request) {
	code, ok := web.PathCode(r, "code")
	if !ok {
		web.WriteError(w, http.StatusBadRequest, "Invalid event code")
		return
	}
	detail, err := h.Service.EventDetail(r.Context(), code)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	web.WriteJSON(w, http.StatusOK, detail)
}

func (h *Handler) Promotions(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit, offset := web.Paginate(r)
	items, err := h.Service.Promotions(r.Context(), q.Get("mode"), q.Get("category"), q.Get("q"), limit, offset)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	if items == nil {
		items = []models.PromoItem{}
	}
	web.WriteJSON(w, http.StatusOK, items)
}

func (h *Handler) AdImpression(w http.ResponseWriter, r *http.Request) {
	h.recordAd(w, r, "IMPRESSION")
}

func (h *Handler) AdClick(w http.ResponseWriter, r *http.Request) {
	h.recordAd(w, r, "CLICK")
}

func (h *Handler) recordAd(w http.ResponseWriter, r *http.Request, kind string) {
	code, ok := web.PathCode(r, "code")
	if !ok {
		web.WriteError(w, http.StatusBadRequest, "Invalid campaign code")
		return
	}
	ip := utils.ClientIP(r.RemoteAddr)
	ua := r.Header.Get("User-Agent")
	recorded, err := h.Service.RecordAdEvent(r.Context(), code, kind, utils.HashString(ip+"|"+ua))
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	web.WriteJSON(w, http.StatusOK, map[string]interface{}{"message": "ok", "recorded": recorded})
}
