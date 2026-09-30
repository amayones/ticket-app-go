package ticketing

import (
	"context"
	"net/http"
	"strings"

	"github.com/skip2/go-qrcode"

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
	Service   ServiceInterface
	Audit     audit.ServiceInterface
	Sys       syslog.ServiceInterface
	qrSecret  string
	jwtSecret string
}

func NewHandler(service ServiceInterface, audit audit.ServiceInterface, sys syslog.ServiceInterface, qrSecret, jwtSecret string) *Handler {
	return &Handler{Service: service, Audit: audit, Sys: sys, qrSecret: qrSecret, jwtSecret: jwtSecret}
}

func (h *Handler) audit(r *http.Request, actorCode, action, entity, entityCode, detail string) {
	if h.Audit == nil {
		return
	}
	_ = h.Audit.Log(r.Context(), actorCode, action, entity, entityCode, detail, utils.ClientIP(r.RemoteAddr))
}

func (h *Handler) handleServiceError(w http.ResponseWriter, err error) {
	if web.ServiceError(w, err) == http.StatusInternalServerError && h.Sys != nil {
		_ = h.Sys.Error(context.Background(), "ticketing_handler", err.Error())
	}
}

func callerCode(w http.ResponseWriter, r *http.Request, h *Handler) (string, bool) {
	code, ok := middleware.GetUserCode(r)
	if ok {
		return code, true
	}
	// <img> QR tags cannot set headers: accept a query access token
	// (validated with the session secret, like the auth middleware).
	if tok := strings.TrimSpace(r.URL.Query().Get("access_token")); tok != "" {
		if claims, err := utils.ValidateToken(tok, h.jwtSecret); err == nil {
			if userCode, ok := claims["user_code"].(string); ok && strings.TrimSpace(userCode) != "" {
				return strings.TrimSpace(userCode), true
			}
		}
	}
	web.WriteError(w, http.StatusUnauthorized, "Missing user code")
	return "", false
}

func (h *Handler) Checkout(w http.ResponseWriter, r *http.Request) {
	caller, ok := callerCode(w, r, h)
	if !ok {
		return
	}
	var req models.CheckoutRequest
	if !web.DecodeJSON(w, r, &req) {
		return
	}
	res, err := h.Service.Checkout(r.Context(), caller, req)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	h.audit(r, caller, models.AuditOrderCreate, models.EntityOrder, res.OrderCode,
		"Checkout order dibuat (menunggu bayar)")
	web.WriteJSON(w, http.StatusCreated, res)
}

func (h *Handler) PayOrder(w http.ResponseWriter, r *http.Request) {
	caller, ok := callerCode(w, r, h)
	if !ok {
		return
	}
	code, ok := web.PathCode(r, "code")
	if !ok {
		web.WriteError(w, http.StatusBadRequest, "Invalid order code")
		return
	}
	var req models.PayRequest
	if !web.DecodeJSON(w, r, &req) {
		return
	}
	holder := strings.TrimSpace(r.URL.Query().Get("holder"))
	if holder == "" {
		holder = caller
	}
	res, err := h.Service.PayOrder(r.Context(), caller, code, req, holder,
		strings.TrimSpace(r.URL.Query().Get("email")), strings.TrimSpace(r.URL.Query().Get("phone")))
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	h.audit(r, caller, models.AuditOrderPay, models.EntityOrder, code, "Order dibayar")
	web.WriteJSON(w, http.StatusOK, res)
}

func (h *Handler) CancelOrder(w http.ResponseWriter, r *http.Request) {
	caller, ok := callerCode(w, r, h)
	if !ok {
		return
	}
	code, ok := web.PathCode(r, "code")
	if !ok {
		web.WriteError(w, http.StatusBadRequest, "Invalid order code")
		return
	}
	if err := h.Service.CancelOrder(r.Context(), caller, code); err != nil {
		h.handleServiceError(w, err)
		return
	}
	web.WriteJSON(w, http.StatusOK, map[string]string{"message": "Order cancelled"})
}

func (h *Handler) MyTickets(w http.ResponseWriter, r *http.Request) {
	caller, ok := callerCode(w, r, h)
	if !ok {
		return
	}
	q := r.URL.Query()
	limit, offset := web.Paginate(r)
	items, err := h.Service.MyTickets(r.Context(), caller, q.Get("status"), q.Get("q"), limit, offset)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	if items == nil {
		items = []models.MyTicketItem{}
	}
	web.WriteJSON(w, http.StatusOK, items)
}

func (h *Handler) TicketDetail(w http.ResponseWriter, r *http.Request) {
	caller, ok := callerCode(w, r, h)
	if !ok {
		return
	}
	code, ok := web.PathCode(r, "code")
	if !ok {
		web.WriteError(w, http.StatusBadRequest, "Invalid ticket code")
		return
	}
	detail, err := h.Service.TicketDetail(r.Context(), caller, code)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	web.WriteJSON(w, http.StatusOK, detail)
}

// TicketQR renders the signed QR token as PNG (for <img> display).
func (h *Handler) TicketQR(w http.ResponseWriter, r *http.Request) {
	caller, ok := callerCode(w, r, h)
	if !ok {
		return
	}
	code, ok := web.PathCode(r, "code")
	if !ok {
		web.WriteError(w, http.StatusBadRequest, "Invalid ticket code")
		return
	}
	detail, err := h.Service.TicketDetail(r.Context(), caller, code)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	if detail.QRToken == "" {
		web.WriteError(w, http.StatusBadRequest, "QR inactive for this ticket")
		return
	}
	png, err := qrcode.Encode(detail.QRToken, qrcode.Medium, 256)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(png)
}

func (h *Handler) RequestRefund(w http.ResponseWriter, r *http.Request) {
	caller, ok := callerCode(w, r, h)
	if !ok {
		return
	}
	var req models.RefundRequest
	if !web.DecodeJSON(w, r, &req) {
		return
	}
	code, err := h.Service.RequestRefund(r.Context(), caller, req.OrderCode, req.Reason)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	h.audit(r, caller, models.AuditRefundRequest, models.EntityRefund, code, "Refund diajukan")
	web.WriteJSON(w, http.StatusCreated, map[string]interface{}{
		"message": "Refund requested",
		"code":    code,
	})
}

func (h *Handler) BuyerRefunds(w http.ResponseWriter, r *http.Request) {
	caller, ok := callerCode(w, r, h)
	if !ok {
		return
	}
	items, err := h.Service.BuyerRefunds(r.Context(), caller)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	if items == nil {
		items = []models.RefundItem{}
	}
	web.WriteJSON(w, http.StatusOK, items)
}

func (h *Handler) ScanTicket(w http.ResponseWriter, r *http.Request) {
	caller, ok := callerCode(w, r, h)
	if !ok {
		return
	}
	var req models.ScanRequest
	if !web.DecodeJSON(w, r, &req) {
		return
	}
	res, err := h.Service.ScanTicket(r.Context(), caller, req)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	web.WriteJSON(w, http.StatusOK, res)
}

func (h *Handler) ScanEvents(w http.ResponseWriter, r *http.Request) {
	caller, ok := callerCode(w, r, h)
	if !ok {
		return
	}
	events, err := h.Service.ScanEvents(r.Context(), caller)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	if events == nil {
		events = []AssignedEvent{}
	}
	web.WriteJSON(w, http.StatusOK, events)
}

func (h *Handler) ScanSummary(w http.ResponseWriter, r *http.Request) {
	caller, ok := callerCode(w, r, h)
	if !ok {
		return
	}
	sum, err := h.Service.ScanSummary(r.Context(), caller, r.URL.Query().Get("event"))
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	web.WriteJSON(w, http.StatusOK, sum)
}

func (h *Handler) ScanHistory(w http.ResponseWriter, r *http.Request) {
	caller, ok := callerCode(w, r, h)
	if !ok {
		return
	}
	q := r.URL.Query()
	limit, offset := web.Paginate(r)
	rows, sum, err := h.Service.ScanHistory(r.Context(), caller,
		q.Get("event"), q.Get("result"), q.Get("officer"),
		strings.TrimSpace(q.Get("from")), strings.TrimSpace(q.Get("to")), limit, offset)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	if rows == nil {
		rows = []models.ScanLogRow{}
	}
	web.WriteJSON(w, http.StatusOK, map[string]interface{}{"logs": rows, "summary": sum})
}

func (h *Handler) ScanExport(w http.ResponseWriter, r *http.Request) {
	caller, ok := callerCode(w, r, h)
	if !ok {
		return
	}
	q := r.URL.Query()
	if format := strings.ToLower(strings.TrimSpace(q.Get("format"))); format != "" && format != "csv" {
		web.WriteError(w, http.StatusBadRequest, "format must be csv")
		return
	}
	tables, err := h.Service.ScanExport(r.Context(), caller,
		q.Get("event"), q.Get("result"), q.Get("officer"),
		strings.TrimSpace(q.Get("from")), strings.TrimSpace(q.Get("to")))
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	data, err := export.CSV(tables)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	h.audit(r, caller, models.AuditExport, models.EntityScanLog, q.Get("event"), "Export riwayat scan (csv)")
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", `attachment; filename="scan-history.csv"`)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}
