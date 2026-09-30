package seller

import (
	"context"
	"net/http"
	"strings"

	"golang-backend/features/audit"
	"golang-backend/features/syslog"
	"golang-backend/internal/export"
	"golang-backend/internal/web"
	"golang-backend/middleware"
	"golang-backend/models"
	"golang-backend/utils"
)

// Handler depends on abstractions (mockable), not concretes.
// Audit + Syslog are best-effort (nil-safe) so failures never break requests.
type Handler struct {
	Service ServiceInterface
	Audit   audit.ServiceInterface
	Sys     syslog.ServiceInterface
}

func NewHandler(service ServiceInterface, audit audit.ServiceInterface, sys syslog.ServiceInterface) *Handler {
	return &Handler{Service: service, Audit: audit, Sys: sys}
}

func (h *Handler) audit(r *http.Request, actorCode, action, entity, entityCode, detail string) {
	if h.Audit == nil {
		return
	}
	_ = h.Audit.Log(r.Context(), actorCode, action, entity, entityCode, detail, utils.ClientIP(r.RemoteAddr))
}

func (h *Handler) handleServiceError(w http.ResponseWriter, err error) {
	if web.ServiceError(w, err) == http.StatusInternalServerError && h.Sys != nil {
		_ = h.Sys.Error(context.Background(), "seller_handler", err.Error())
	}
}

func callerCode(w http.ResponseWriter, r *http.Request) (string, bool) {
	code, ok := middleware.GetUserCode(r)
	if !ok {
		web.WriteError(w, http.StatusUnauthorized, "Missing user code")
		return "", false
	}
	return code, true
}

func dashboardQuery(r *http.Request) (preset, from, to, category string) {
	q := r.URL.Query()
	return q.Get("preset"), q.Get("from"), q.Get("to"), q.Get("category")
}

func (h *Handler) Summary(w http.ResponseWriter, r *http.Request) {
	caller, ok := callerCode(w, r)
	if !ok {
		return
	}
	preset, from, to, category := dashboardQuery(r)
	sum, err := h.Service.Summary(r.Context(), caller, preset, from, to, category)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	web.WriteJSON(w, http.StatusOK, sum)
}

func (h *Handler) Orders(w http.ResponseWriter, r *http.Request) {
	caller, ok := callerCode(w, r)
	if !ok {
		return
	}
	preset, from, to, category := dashboardQuery(r)
	limit, _ := web.Paginate(r)
	orders, err := h.Service.Orders(r.Context(), caller, preset, from, to, category, limit)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	if orders == nil {
		orders = []models.SellerOrder{}
	}
	web.WriteJSON(w, http.StatusOK, orders)
}

func (h *Handler) Actionable(w http.ResponseWriter, r *http.Request) {
	caller, ok := callerCode(w, r)
	if !ok {
		return
	}
	orders, err := h.Service.Actionable(r.Context(), caller)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	if orders == nil {
		orders = []models.SellerOrder{}
	}
	web.WriteJSON(w, http.StatusOK, orders)
}

func (h *Handler) LowStock(w http.ResponseWriter, r *http.Request) {
	caller, ok := callerCode(w, r)
	if !ok {
		return
	}
	products, err := h.Service.LowStock(r.Context(), caller)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	if products == nil {
		products = []models.SellerProduct{}
	}
	web.WriteJSON(w, http.StatusOK, products)
}

func (h *Handler) Complaints(w http.ResponseWriter, r *http.Request) {
	caller, ok := callerCode(w, r)
	if !ok {
		return
	}
	complaints, err := h.Service.Complaints(r.Context(), caller)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	if complaints == nil {
		complaints = []models.SellerComplaint{}
	}
	web.WriteJSON(w, http.StatusOK, complaints)
}

// Export renders the sales summary as CSV and logs the export to audit.
func (h *Handler) Export(w http.ResponseWriter, r *http.Request) {
	caller, ok := callerCode(w, r)
	if !ok {
		return
	}
	if format := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("format"))); format != "" && format != "csv" {
		web.WriteError(w, http.StatusBadRequest, "format must be csv")
		return
	}
	preset, from, to, category := dashboardQuery(r)
	tables, err := h.Service.ExportTables(r.Context(), caller, preset, from, to, category)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	data, err := export.CSV(tables)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	h.audit(r, caller, models.AuditExport, models.EntityDashboard, "seller",
		"Export seller summary (csv)")
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", `attachment; filename="seller-summary.csv"`)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}
