package syslog

import (
	"net/http"
	"strconv"
	"strings"

	"golang-backend/features/audit"
	"golang-backend/internal/web"
	"golang-backend/middleware"
	"golang-backend/models"
	"golang-backend/utils"
)

// Handler melayani menu Error / System Log.
//
// Konvensi menu: raw SQL di variabel `query`, di-run, dipetakan ke response
// (lihat syslog_repository.go); error service via web.ServiceError.
type Handler struct {
	Service ServiceInterface
	Audit   audit.ServiceInterface
}

func NewHandler(service ServiceInterface, audit audit.ServiceInterface) *Handler {
	return &Handler{Service: service, Audit: audit}
}

func (h *Handler) audit(r *http.Request, action, entity, entityCode, detail string) {
	if h.Audit == nil {
		return
	}
	code, _ := middleware.GetUserCode(r)
	_ = h.Audit.Log(r.Context(), code, action, entity, entityCode, detail, utils.ClientIP(r.RemoteAddr))
}

func (h *Handler) ListSyslog(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit, offset := web.Paginate(r)
	logs, err := h.Service.List(r.Context(), models.SyslogFilter{
		Level: strings.ToUpper(q.Get("level")), Limit: limit, Offset: offset,
	})
	if err != nil {
		web.ServiceError(w, err)
		return
	}
	web.WriteJSON(w, http.StatusOK, logs)
}

func (h *Handler) PruneSyslog(w http.ResponseWriter, r *http.Request) {
	days := 30
	if v := r.URL.Query().Get("days"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			days = n
		}
	}
	n, err := h.Service.Prune(r.Context(), days)
	if err != nil {
		web.ServiceError(w, err)
		return
	}
	h.audit(r, models.AuditSyslogPrune, models.EntitySystem, "",
		"System log lebih tua dari "+strconv.Itoa(days)+" hari dihapus ("+strconv.FormatInt(n, 10)+" baris)")
	web.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Old system logs pruned",
		"deleted": n,
	})
}
