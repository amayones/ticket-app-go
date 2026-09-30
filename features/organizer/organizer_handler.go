package organizer

import (
	"context"
	"fmt"
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
		_ = h.Sys.Error(context.Background(), "organizer_handler", err.Error())
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

func dashboardQuery(r *http.Request) (preset, from, to, event, city string) {
	q := r.URL.Query()
	return q.Get("preset"), q.Get("from"), q.Get("to"), q.Get("event"), q.Get("city")
}

func (h *Handler) Summary(w http.ResponseWriter, r *http.Request) {
	caller, ok := callerCode(w, r)
	if !ok {
		return
	}
	preset, from, to, event, city := dashboardQuery(r)
	sum, err := h.Service.Summary(r.Context(), caller, preset, from, to, event, city)
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
	preset, from, to, event, city := dashboardQuery(r)
	limit, _ := web.Paginate(r)
	orders, err := h.Service.Orders(r.Context(), caller, preset, from, to, event, city, limit)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	if orders == nil {
		orders = []models.OrganizerOrder{}
	}
	web.WriteJSON(w, http.StatusOK, orders)
}

func (h *Handler) Refunds(w http.ResponseWriter, r *http.Request) {
	caller, ok := callerCode(w, r)
	if !ok {
		return
	}
	refunds, err := h.Service.Refunds(r.Context(), caller)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	if refunds == nil {
		refunds = []models.OrganizerRefund{}
	}
	web.WriteJSON(w, http.StatusOK, refunds)
}

// DecideRefund approves (approve=true) or rejects a pending refund.
// Only the owning organizer passes RequireRole + ownership check.
func (h *Handler) DecideRefund(w http.ResponseWriter, r *http.Request, approve bool) {
	caller, ok := callerCode(w, r)
	if !ok {
		return
	}
	code, ok := web.PathCode(r, "code")
	if !ok {
		web.WriteError(w, http.StatusBadRequest, "Invalid refund code")
		return
	}
	status, amount, err := h.Service.DecideRefund(r.Context(), caller, code, approve)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	action := models.AuditRefundReject
	if approve {
		action = models.AuditRefundApprove
	}
	h.audit(r, caller, action, models.EntityRefund, code,
		"Refund "+code+" "+strings.ToLower(status)+" ("+fmt.Sprintf("%.2f", amount)+")")
	web.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Refund " + strings.ToLower(status),
		"status":  status,
		"amount":  amount,
	})
}

func (h *Handler) CheckinLive(w http.ResponseWriter, r *http.Request) {
	caller, ok := callerCode(w, r)
	if !ok {
		return
	}
	live, err := h.Service.CheckinLive(r.Context(), caller, r.URL.Query().Get("event"))
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	web.WriteJSON(w, http.StatusOK, live)
}

// Export renders the summary as CSV or PDF and logs the export to audit.
func (h *Handler) Export(w http.ResponseWriter, r *http.Request) {
	caller, ok := callerCode(w, r)
	if !ok {
		return
	}
	format := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("format")))
	if format == "" {
		format = "csv"
	}
	if format != "csv" && format != "pdf" {
		web.WriteError(w, http.StatusBadRequest, "format must be csv or pdf")
		return
	}
	preset, from, to, event, city := dashboardQuery(r)
	tables, err := h.Service.ExportTables(r.Context(), caller, preset, from, to, event, city)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	var (
		data        []byte
		contentType string
		ext         string
	)
	if format == "pdf" {
		data, err = export.PDF("Organizer Dashboard Summary", tables)
		contentType = "application/pdf"
		ext = "pdf"
	} else {
		data, err = export.CSV(tables)
		contentType = "text/csv"
		ext = "csv"
	}
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	h.audit(r, caller, models.AuditExport, models.EntityDashboard, "organizer",
		"Export organizer summary ("+format+")")
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", `attachment; filename="organizer-summary.`+ext+`"`)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}
