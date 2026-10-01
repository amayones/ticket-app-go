package marketing

import (
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"golang-backend/features/audit"
	"golang-backend/features/syslog"
	"golang-backend/internal/web"
	"golang-backend/middleware"
	"golang-backend/models"
	"golang-backend/utils"
)

type Handler struct {
	Service ServiceInterface
	Audit   audit.ServiceInterface
	Sys     syslog.ServiceInterface
}

func NewHandler(s ServiceInterface, a audit.ServiceInterface, sys syslog.ServiceInterface) *Handler {
	return &Handler{Service: s, Audit: a, Sys: sys}
}

func (h *Handler) auditLog(r *http.Request, actor, action, entity, code, detail string) {
	if h.Audit == nil {
		return
	}
	_ = h.Audit.Log(r.Context(), actor, action, entity, code, detail, utils.ClientIP(r.RemoteAddr))
}

func (h *Handler) handleErr(w http.ResponseWriter, err error) {
	if web.ServiceError(w, err) == http.StatusInternalServerError && h.Sys != nil {
		_ = h.Sys.Error(context.Background(), "marketing_handler", err.Error())
	}
}

func callerCode(w http.ResponseWriter, r *http.Request) (string, bool) {
	c, ok := middleware.GetUserCode(r)
	if !ok {
		web.WriteError(w, http.StatusUnauthorized, "Missing user code")
		return "", false
	}
	return c, true
}

func (h *Handler) ListPromos(w http.ResponseWriter, r *http.Request) {
	caller, ok := callerCode(w, r)
	if !ok {
		return
	}
	q := r.URL.Query()
	limit, offset := web.Paginate(r)
	rows, err := h.Service.ListPromos(r.Context(), caller, q.Get("status"), q.Get("q"), limit, offset)
	if err != nil {
		h.handleErr(w, err)
		return
	}
	if rows == nil {
		rows = []models.PromoRow{}
	}
	web.WriteJSON(w, http.StatusOK, rows)
}

func (h *Handler) PromoDetail(w http.ResponseWriter, r *http.Request) {
	caller, ok := callerCode(w, r)
	if !ok {
		return
	}
	code, ok := web.PathCode(r, "code")
	if !ok {
		web.WriteError(w, http.StatusBadRequest, "Invalid promo code")
		return
	}
	row, usage, err := h.Service.PromoDetail(r.Context(), caller, code)
	if err != nil {
		h.handleErr(w, err)
		return
	}
	if usage == nil {
		usage = []models.PromoUsageRow{}
	}
	web.WriteJSON(w, http.StatusOK, map[string]interface{}{"promo": row, "usage": usage})
}

func (h *Handler) CreatePromo(w http.ResponseWriter, r *http.Request) {
	caller, ok := callerCode(w, r)
	if !ok {
		return
	}
	var in models.PromoInput
	if !web.DecodeJSON(w, r, &in) {
		return
	}
	code, err := h.Service.CreatePromo(r.Context(), caller, in)
	if err != nil {
		h.handleErr(w, err)
		return
	}
	h.auditLog(r, caller, models.AuditPromoCreate, "PROMO", code, "Promo "+in.PromoCode+" dibuat")
	web.WriteJSON(w, http.StatusCreated, map[string]interface{}{"message": "Promo created", "code": code})
}

func (h *Handler) UpdatePromo(w http.ResponseWriter, r *http.Request) {
	caller, ok := callerCode(w, r)
	if !ok {
		return
	}
	code, ok := web.PathCode(r, "code")
	if !ok {
		web.WriteError(w, http.StatusBadRequest, "Invalid promo code")
		return
	}
	var in models.PromoInput
	if !web.DecodeJSON(w, r, &in) {
		return
	}
	if err := h.Service.UpdatePromo(r.Context(), caller, code, in); err != nil {
		h.handleErr(w, err)
		return
	}
	h.auditLog(r, caller, models.AuditPromoUpdate, "PROMO", code, "Promo diperbarui")
	web.WriteJSON(w, http.StatusOK, map[string]string{"message": "Promo updated"})
}

func (h *Handler) DeletePromo(w http.ResponseWriter, r *http.Request) {
	caller, ok := callerCode(w, r)
	if !ok {
		return
	}
	code, ok := web.PathCode(r, "code")
	if !ok {
		web.WriteError(w, http.StatusBadRequest, "Invalid promo code")
		return
	}
	if err := h.Service.DeletePromo(r.Context(), caller, code); err != nil {
		h.handleErr(w, err)
		return
	}
	h.auditLog(r, caller, models.AuditPromoDelete, "PROMO", code, "Promo dihapus")
	web.WriteJSON(w, http.StatusOK, map[string]string{"message": "Promo deleted"})
}

func (h *Handler) TogglePromo(w http.ResponseWriter, r *http.Request) {
	caller, ok := callerCode(w, r)
	if !ok {
		return
	}
	code, ok := web.PathCode(r, "code")
	if !ok {
		web.WriteError(w, http.StatusBadRequest, "Invalid promo code")
		return
	}
	var req struct {
		Active *bool `json:"active"`
	}
	if !web.DecodeJSON(w, r, &req) {
		return
	}
	if req.Active == nil {
		web.WriteError(w, http.StatusBadRequest, "active is required")
		return
	}
	if err := h.Service.SetPromoActive(r.Context(), caller, code, *req.Active); err != nil {
		h.handleErr(w, err)
		return
	}
	web.WriteJSON(w, http.StatusOK, map[string]string{"message": "Promo status updated"})
}

func (h *Handler) GeneratePromoCode(w http.ResponseWriter, r *http.Request) {
	_, ok := callerCode(w, r)
	if !ok {
		return
	}
	web.WriteJSON(w, http.StatusOK, map[string]string{"code": h.Service.GenerateCode()})
}

func (h *Handler) Positions(w http.ResponseWriter, r *http.Request) {
	_, ok := callerCode(w, r)
	if !ok {
		return
	}
	rows, err := h.Service.Positions(r.Context())
	if err != nil {
		h.handleErr(w, err)
		return
	}
	if rows == nil {
		rows = []models.AdPosition{}
	}
	web.WriteJSON(w, http.StatusOK, rows)
}

func (h *Handler) ListAds(w http.ResponseWriter, r *http.Request) {
	caller, ok := callerCode(w, r)
	if !ok {
		return
	}
	q := r.URL.Query()
	limit, offset := web.Paginate(r)
	rows, err := h.Service.ListAds(r.Context(), caller, q.Get("status"), q.Get("q"), limit, offset)
	if err != nil {
		h.handleErr(w, err)
		return
	}
	if rows == nil {
		rows = []models.AdCampaignRow{}
	}
	web.WriteJSON(w, http.StatusOK, rows)
}

func (h *Handler) AdDetail(w http.ResponseWriter, r *http.Request) {
	caller, ok := callerCode(w, r)
	if !ok {
		return
	}
	code, ok := web.PathCode(r, "code")
	if !ok {
		web.WriteError(w, http.StatusBadRequest, "Invalid ad code")
		return
	}
	row, stats, err := h.Service.AdDetail(r.Context(), caller, code)
	if err != nil {
		h.handleErr(w, err)
		return
	}
	if stats == nil {
		stats = []models.AdDailyPoint{}
	}
	web.WriteJSON(w, http.StatusOK, map[string]interface{}{"campaign": row, "stats": stats})
}

func (h *Handler) CreateAd(w http.ResponseWriter, r *http.Request) {
	caller, ok := callerCode(w, r)
	if !ok {
		return
	}
	var in models.AdCampaignInput
	if !web.DecodeJSON(w, r, &in) {
		return
	}
	code, err := h.Service.CreateAd(r.Context(), caller, in)
	if err != nil {
		h.handleErr(w, err)
		return
	}
	h.auditLog(r, caller, models.AuditAdCreate, "AD_CAMPAIGN", code, "Kampanye "+in.Title+" dibuat")
	web.WriteJSON(w, http.StatusCreated, map[string]interface{}{"message": "Campaign created", "code": code})
}

func (h *Handler) UpdateAd(w http.ResponseWriter, r *http.Request) {
	caller, ok := callerCode(w, r)
	if !ok {
		return
	}
	code, ok := web.PathCode(r, "code")
	if !ok {
		web.WriteError(w, http.StatusBadRequest, "Invalid ad code")
		return
	}
	var in models.AdCampaignInput
	if !web.DecodeJSON(w, r, &in) {
		return
	}
	if err := h.Service.UpdateAd(r.Context(), caller, code, in); err != nil {
		h.handleErr(w, err)
		return
	}
	h.auditLog(r, caller, models.AuditAdUpdate, "AD_CAMPAIGN", code, "Kampanye diperbarui")
	web.WriteJSON(w, http.StatusOK, map[string]string{"message": "Campaign updated"})
}

func (h *Handler) AdStatus(w http.ResponseWriter, r *http.Request) {
	caller, ok := callerCode(w, r)
	if !ok {
		return
	}
	code, ok := web.PathCode(r, "code")
	if !ok {
		web.WriteError(w, http.StatusBadRequest, "Invalid ad code")
		return
	}
	var req struct {
		Action string `json:"action"`
		Note   string `json:"note"`
	}
	if !web.DecodeJSON(w, r, &req) {
		return
	}
	if err := h.Service.SetAdStatus(r.Context(), caller, code, req.Action, req.Note); err != nil {
		h.handleErr(w, err)
		return
	}
	h.auditLog(r, caller, models.AuditAdStatus, "AD_CAMPAIGN", code, "Status kampanye: "+req.Action)
	web.WriteJSON(w, http.StatusOK, map[string]string{"message": "Status updated"})
}

func (h *Handler) SimulatePay(w http.ResponseWriter, r *http.Request) {
	caller, ok := callerCode(w, r)
	if !ok {
		return
	}
	code, ok := web.PathCode(r, "code")
	if !ok {
		web.WriteError(w, http.StatusBadRequest, "Invalid ad code")
		return
	}
	if err := h.Service.SimulatePay(r.Context(), caller, code); err != nil {
		h.handleErr(w, err)
		return
	}
	web.WriteJSON(w, http.StatusOK, map[string]string{"message": "Payment simulated"})
}

func (h *Handler) AdStats(w http.ResponseWriter, r *http.Request) {
	caller, ok := callerCode(w, r)
	if !ok {
		return
	}
	code, ok := web.PathCode(r, "code")
	if !ok {
		web.WriteError(w, http.StatusBadRequest, "Invalid ad code")
		return
	}
	_, stats, err := h.Service.AdDetail(r.Context(), caller, code)
	if err != nil {
		h.handleErr(w, err)
		return
	}
	if stats == nil {
		stats = []models.AdDailyPoint{}
	}
	web.WriteJSON(w, http.StatusOK, stats)
}

func (h *Handler) ListNews(w http.ResponseWriter, r *http.Request) {
	_, ok := callerCode(w, r)
	if !ok {
		return
	}
	q := r.URL.Query()
	limit, offset := web.Paginate(r)
	rows, err := h.Service.ListNews(r.Context(), q.Get("status"), q.Get("category"), q.Get("q"), limit, offset)
	if err != nil {
		h.handleErr(w, err)
		return
	}
	if rows == nil {
		rows = []models.NewsRow{}
	}
	web.WriteJSON(w, http.StatusOK, rows)
}

func (h *Handler) NewsDetail(w http.ResponseWriter, r *http.Request) {
	_, ok := callerCode(w, r)
	if !ok {
		return
	}
	code, ok := web.PathCode(r, "code")
	if !ok {
		web.WriteError(w, http.StatusBadRequest, "Invalid news code")
		return
	}
	row, err := h.Service.NewsDetail(r.Context(), code)
	if err != nil {
		h.handleErr(w, err)
		return
	}
	web.WriteJSON(w, http.StatusOK, row)
}

func (h *Handler) CreateNews(w http.ResponseWriter, r *http.Request) {
	caller, ok := callerCode(w, r)
	if !ok {
		return
	}
	var in models.NewsInput
	if !web.DecodeJSON(w, r, &in) {
		return
	}
	code, err := h.Service.CreateNews(r.Context(), caller, in)
	if err != nil {
		h.handleErr(w, err)
		return
	}
	h.auditLog(r, caller, models.AuditNewsCreate, "NEWS", code, "Berita "+in.Title+" dibuat")
	web.WriteJSON(w, http.StatusCreated, map[string]interface{}{"message": "News created", "code": code})
}

func (h *Handler) UpdateNews(w http.ResponseWriter, r *http.Request) {
	_, ok := callerCode(w, r)
	if !ok {
		return
	}
	code, ok := web.PathCode(r, "code")
	if !ok {
		web.WriteError(w, http.StatusBadRequest, "Invalid news code")
		return
	}
	var in models.NewsInput
	if !web.DecodeJSON(w, r, &in) {
		return
	}
	if err := h.Service.UpdateNews(r.Context(), code, in); err != nil {
		h.handleErr(w, err)
		return
	}
	h.auditLog(r, code, models.AuditNewsUpdate, "NEWS", code, "Berita diperbarui")
	web.WriteJSON(w, http.StatusOK, map[string]string{"message": "News updated"})
}

func (h *Handler) DeleteNews(w http.ResponseWriter, r *http.Request) {
	_, ok := callerCode(w, r)
	if !ok {
		return
	}
	code, ok := web.PathCode(r, "code")
	if !ok {
		web.WriteError(w, http.StatusBadRequest, "Invalid news code")
		return
	}
	if err := h.Service.DeleteNews(r.Context(), code); err != nil {
		h.handleErr(w, err)
		return
	}
	h.auditLog(r, code, models.AuditNewsDelete, "NEWS", code, "Berita dihapus/diarsip")
	web.WriteJSON(w, http.StatusOK, map[string]string{"message": "News deleted"})
}

func (h *Handler) NewsStatus(w http.ResponseWriter, r *http.Request) {
	_, ok := callerCode(w, r)
	if !ok {
		return
	}
	code, ok := web.PathCode(r, "code")
	if !ok {
		web.WriteError(w, http.StatusBadRequest, "Invalid news code")
		return
	}
	var req struct {
		Action string `json:"action"`
		Status string `json:"status"`
	}
	if !web.DecodeJSON(w, r, &req) {
		return
	}
	action := req.Action
	if action == "" {
		action = req.Status
	}
	if err := h.Service.SetNewsStatus(r.Context(), code, action); err != nil {
		h.handleErr(w, err)
		return
	}
	web.WriteJSON(w, http.StatusOK, map[string]string{"message": "Status updated"})
}

func (h *Handler) NewsCategories(w http.ResponseWriter, r *http.Request) {
	_, ok := callerCode(w, r)
	if !ok {
		return
	}
	cats, err := h.Service.NewsCategories(r.Context())
	if err != nil {
		h.handleErr(w, err)
		return
	}
	if cats == nil {
		cats = []models.NewsCategory{}
	}
	web.WriteJSON(w, http.StatusOK, cats)
}

func (h *Handler) EventCategories(w http.ResponseWriter, r *http.Request) {
	_, ok := callerCode(w, r)
	if !ok {
		return
	}
	cats, err := h.Service.EventCategories(r.Context())
	if err != nil {
		h.handleErr(w, err)
		return
	}
	if cats == nil {
		cats = []models.EventCategory{}
	}
	web.WriteJSON(w, http.StatusOK, cats)
}

func (h *Handler) OwnedEvents(w http.ResponseWriter, r *http.Request) {
	caller, ok := callerCode(w, r)
	if !ok {
		return
	}
	evs, err := h.Service.OwnedEvents(r.Context(), caller)
	if err != nil {
		h.handleErr(w, err)
		return
	}
	if evs == nil {
		evs = []OwnedEvent{}
	}
	web.WriteJSON(w, http.StatusOK, evs)
}

func (h *Handler) UploadBanner(w http.ResponseWriter, r *http.Request) {
	if _, ok := callerCode(w, r); !ok {
		return
	}
	const maxBytes = 2 << 20
	r.Body = http.MaxBytesReader(w, r.Body, maxBytes+1024)
	if err := r.ParseMultipartForm(maxBytes); err != nil {
		web.WriteError(w, http.StatusBadRequest, "Banner too large (max 2 MB)")
		return
	}
	file, _, err := r.FormFile("banner")
	if err != nil {
		web.WriteError(w, http.StatusBadRequest, "Field banner is required")
		return
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, maxBytes+1))
	if err != nil {
		h.handleErr(w, err)
		return
	}
	if len(data) > maxBytes {
		web.WriteError(w, http.StatusBadRequest, "Banner too large (max 2 MB)")
		return
	}
	ext, ok := bannerExt(data)
	if !ok {
		web.WriteError(w, http.StatusBadRequest, "Banner must be JPEG, PNG, WebP, or GIF")
		return
	}
	name, err := utils.GenerateCode("BNR-")
	if err != nil {
		h.handleErr(w, err)
		return
	}
	dir := filepath.Join("uploads", "banners")
	if err := os.MkdirAll(dir, 0755); err != nil {
		h.handleErr(w, err)
		return
	}
	filename := name + ext
	if err := os.WriteFile(filepath.Join(dir, filename), data, 0644); err != nil {
		h.handleErr(w, err)
		return
	}
	web.WriteJSON(w, http.StatusCreated, map[string]string{"message": "Banner uploaded", "url": "/uploads/banners/" + filename})
}

func bannerExt(data []byte) (string, bool) {
	if len(data) < 12 {
		return "", false
	}
	switch {
	case data[0] == 0xFF && data[1] == 0xD8 && data[2] == 0xFF:
		return ".jpg", true
	case data[0] == 0x89 && data[1] == 'P' && data[2] == 'N' && data[3] == 'G':
		return ".png", true
	case data[0] == 'R' && data[1] == 'I' && data[2] == 'F' && data[3] == 'F' && data[8] == 'W' && data[9] == 'E' && data[10] == 'B' && data[11] == 'P':
		return ".webp", true
	case data[0] == 'G' && data[1] == 'I' && data[2] == 'F':
		return ".gif", true
	default:
		return "", false
	}
}

func isValidURL(s string) bool {
	s = strings.ToLower(strings.TrimSpace(s))
	return strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://")
}

var _ = isValidURL
